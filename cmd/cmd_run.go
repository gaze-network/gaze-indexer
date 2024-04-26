package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/datasources"
	"github.com/gaze-network/indexer-network/core/indexers"
	"github.com/gaze-network/indexer-network/internal/config"
	"github.com/gaze-network/indexer-network/internal/postgres"
	"github.com/gaze-network/indexer-network/modules/bitcoin"
	"github.com/gaze-network/indexer-network/modules/bitcoin/btcclient"
	btcdatagateway "github.com/gaze-network/indexer-network/modules/bitcoin/datagateway"
	btcpostgres "github.com/gaze-network/indexer-network/modules/bitcoin/repository/postgres"
	"github.com/gaze-network/indexer-network/modules/runes"
	runesapi "github.com/gaze-network/indexer-network/modules/runes/api"
	runesdatagateway "github.com/gaze-network/indexer-network/modules/runes/datagateway"
	runespostgres "github.com/gaze-network/indexer-network/modules/runes/repository/postgres"
	runesusecase "github.com/gaze-network/indexer-network/modules/runes/usecase"
	"github.com/gaze-network/indexer-network/pkg/errorhandler"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/gaze-network/indexer-network/pkg/reportingclient"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

const (
	shutdownTimeout = 60 * time.Second
)

type runCmdOptions struct {
	APIOnly bool
	Bitcoin bool
	Runes   bool
}

func NewRunCommand() *cobra.Command {
	opts := &runCmdOptions{}

	// Create command
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Start indexer-network service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHandler(opts, cmd, args)
		},
	}

	// TODO: separate flags and bind flags to each module cmd package.

	// Add local flags
	flags := runCmd.Flags()
	flags.BoolVar(&opts.APIOnly, "api-only", false, "Run only API server")
	flags.BoolVar(&opts.Bitcoin, "bitcoin", false, "Enable Bitcoin indexer module")
	flags.String("bitcoin-db", "postgres", `Database to store bitcoin data. current supported databases: "postgres"`)
	flags.BoolVar(&opts.Runes, "runes", false, "Enable Runes indexer module")
	flags.String("runes-db", "postgres", `Database to store runes data. current supported databases: "postgres"`)
	flags.String("runes-datasource", "bitcoin-node", `Datasource to fetch bitcoin data for processing Meta-Protocol data. current supported datasources: "bitcoin-node" | "database"`)

	// Bind flags to configuration
	config.BindPFlag("modules.bitcoin.database", flags.Lookup("bitcoin-db"))
	config.BindPFlag("modules.runes.database", flags.Lookup("runes-db"))
	config.BindPFlag("modules.runes.datasource", flags.Lookup("runes-datasource"))

	return runCmd
}

type HttpHandler interface {
	Mount(router fiber.Router) error
}

func runHandler(opts *runCmdOptions, cmd *cobra.Command, _ []string) error {
	conf := config.Load()

	// Initialize application process context
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize worker context to separate worker's lifecycle from main process
	ctxWorker, stopWorker := context.WithCancel(context.Background())
	defer stopWorker()

	// Add logger context
	ctx = logger.WithContext(ctx, slogx.Stringer("network", conf.Network))
	ctxWorker = logger.WithContext(ctxWorker, slogx.Stringer("network", conf.Network))

	// Initialize Bitcoin Core RPC Client
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:         conf.BitcoinNode.Host,
		User:         conf.BitcoinNode.User,
		Pass:         conf.BitcoinNode.Pass,
		DisableTLS:   conf.BitcoinNode.DisableTLS,
		HTTPPostMode: true,
	}, nil)
	if err != nil {
		logger.PanicContext(ctx, "Failed to create Bitcoin Core RPC Client", slogx.Error(err))
	}
	defer client.Shutdown()

	logger.InfoContext(ctx, "Connecting to Bitcoin Core RPC Server...", slogx.String("host", conf.BitcoinNode.Host))
	if err := client.Ping(); err != nil {
		logger.PanicContext(ctx, "Failed to ping Bitcoin Core RPC Server", slogx.Error(err))
	}
	logger.InfoContext(ctx, "Connected to Bitcoin Core RPC Server")

	// Validate network
	if !conf.Network.IsSupported() {
		return errors.Wrapf(errs.Unsupported, "%q network is not supported", conf.Network.String())
	}

	// TODO: create module command package.
	// each module should have its own command package and main package will routing the command to the module command package.

	// TODO: refactor module name to specific type instead of string?
	httpHandlers := make(map[string]HttpHandler, 0)

	var reportingClient *reportingclient.ReportingClient
	if !conf.Reporting.Disabled {
		reportingClient, err = reportingclient.New(conf.Reporting)
		if err != nil {
			logger.PanicContext(ctx, "Failed to create reporting client", slogx.Error(err))
		}
	}

	// Initialize Bitcoin Indexer
	if opts.Bitcoin {
		var (
			btcDB         btcdatagateway.BitcoinDataGateway
			indexerInfoDB btcdatagateway.IndexerInformationDataGateway
		)
		switch strings.ToLower(conf.Modules.Bitcoin.Database) {
		case "postgresql", "postgres", "pg":
			pg, err := postgres.NewPool(ctx, conf.Modules.Bitcoin.Postgres)
			if err != nil {
				logger.PanicContext(ctx, "Failed to create Postgres connection pool", slogx.Error(err))
			}
			defer pg.Close()
			repo := btcpostgres.NewRepository(pg)
			btcDB = repo
			indexerInfoDB = repo
		default:
			return errors.Wrapf(errs.Unsupported, "%q database is not supported", conf.Modules.Bitcoin.Database)
		}
		if !opts.APIOnly {
			processor := bitcoin.NewProcessor(conf, btcDB, indexerInfoDB)
			datasource := datasources.NewBitcoinNode(client)
			indexer := indexers.NewBitcoinIndexer(processor, datasource)
			defer func() {
				if err := indexer.ShutdownWithTimeout(shutdownTimeout); err != nil {
					logger.ErrorContext(ctx, "Error during shutdown Bitcoin indexer", slogx.Error(err))
				}
				logger.InfoContext(ctx, "Bitcoin indexer stopped gracefully")
			}()

			// Verify states before running Indexer
			if err := processor.VerifyStates(ctx); err != nil {
				return errors.WithStack(err)
			}

			// Run Indexer
			go func() {
				// stop main process if indexer stopped
				defer stop()

				logger.InfoContext(ctx, "Starting Bitcoin Indexer")
				if err := indexer.Run(ctxWorker); err != nil {
					logger.PanicContext(ctx, "Failed to run Bitcoin Indexer", slogx.Error(err))
				}
			}()
		}
	}

	// Initialize Runes Indexer
	if opts.Runes {
		var runesDg runesdatagateway.RunesDataGateway
		var indexerInfoDg runesdatagateway.IndexerInfoDataGateway
		switch strings.ToLower(conf.Modules.Runes.Database) {
		case "postgresql", "postgres", "pg":
			pg, err := postgres.NewPool(ctx, conf.Modules.Runes.Postgres)
			if err != nil {
				logger.PanicContext(ctx, "Failed to create Postgres connection pool", slogx.Error(err))
			}
			defer pg.Close()
			runesRepo := runespostgres.NewRepository(pg)
			runesDg = runesRepo
			indexerInfoDg = runesRepo
		default:
			logger.PanicContext(ctx, "Unsupported database", slogx.String("database", conf.Modules.Runes.Database))
		}
		var bitcoinDatasource indexers.BitcoinDatasource
		var bitcoinClient btcclient.Contract
		switch strings.ToLower(conf.Modules.Runes.Datasource) {
		case "bitcoin-node":
			bitcoinNodeDatasource := datasources.NewBitcoinNode(client)
			bitcoinDatasource = bitcoinNodeDatasource
			bitcoinClient = bitcoinNodeDatasource
		case "database":
			pg, err := postgres.NewPool(ctx, conf.Modules.Bitcoin.Postgres)
			if err != nil {
				logger.PanicContext(ctx, "Failed to create Postgres connection pool", slogx.Error(err))
			}
			defer pg.Close()
			btcRepo := btcpostgres.NewRepository(pg)
			btcClientDB := btcclient.NewClientDatabase(btcRepo)
			bitcoinDatasource = btcClientDB
			bitcoinClient = btcClientDB
		default:
			return errors.Wrapf(errs.Unsupported, "%q datasource is not supported", conf.Modules.Runes.Datasource)
		}

		if !opts.APIOnly {
			processor := runes.NewProcessor(runesDg, indexerInfoDg, bitcoinClient, bitcoinDatasource, conf.Network, reportingClient)
			indexer := indexers.NewBitcoinIndexer(processor, bitcoinDatasource)
			defer func() {
				if err := indexer.ShutdownWithTimeout(shutdownTimeout); err != nil {
					logger.ErrorContext(ctx, "Error during shutdown Runes indexer", slogx.Error(err))
				}
				logger.InfoContext(ctx, "Runes indexer stopped gracefully")
			}()

			if err := processor.VerifyStates(ctx); err != nil {
				return errors.WithStack(err)
			}

			// Run Indexer
			go func() {
				// stop main process if indexer stopped
				defer stop()

				logger.InfoContext(ctx, "Started Runes Indexer")
				if err := indexer.Run(ctxWorker); err != nil {
					logger.PanicContext(ctx, "Failed to run Runes Indexer", slogx.Error(err))
				}
			}()
		}

		// Mount API
		apiHandlers := lo.Uniq(conf.Modules.Runes.APIHandlers)
		for _, handler := range apiHandlers {
			switch handler { // TODO: support more handlers (e.g. gRPC)
			case "http":
				runesUsecase := runesusecase.New(runesDg, bitcoinClient)
				runesHTTPHandler := runesapi.NewHTTPHandler(conf.Network, runesUsecase)
				httpHandlers["runes"] = runesHTTPHandler
			default:
				logger.PanicContext(ctx, "Unsupported API handler", slogx.String("handler", handler))
			}
		}
	}

	// Wait for interrupt signal to gracefully stop the server with
	// Setup HTTP server if there are any HTTP handlers
	if len(httpHandlers) > 0 {
		app := fiber.New(fiber.Config{
			AppName:      "Gaze Indexer",
			ErrorHandler: errorhandler.NewHTTPErrorHandler(),
		})
		app.
			Use(fiberrecover.New(fiberrecover.Config{
				EnableStackTrace: true,
				StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
					buf := make([]byte, 1024) // bufLen = 1024
					buf = buf[:runtime.Stack(buf, false)]
					logger.ErrorContext(c.UserContext(), "panic in http handler", slogx.Any("panic", e), slog.String("stacktrace", string(buf)))
				},
			})).
			Use(compress.New(compress.Config{
				Level: compress.LevelDefault,
			}))

		defer func() {
			if err := app.ShutdownWithTimeout(shutdownTimeout); err != nil {
				logger.ErrorContext(ctx, "Error during shutdown HTTP server", slogx.Error(err))
			}
			logger.InfoContext(ctx, "HTTP server stopped gracefully")
		}()

		// Health check
		app.Get("/", func(c *fiber.Ctx) error {
			return errors.WithStack(c.SendStatus(http.StatusOK))
		})

		// mount http handlers from each http-enabled module
		for module, handler := range httpHandlers {
			if err := handler.Mount(app); err != nil {
				logger.PanicContext(ctx, "Failed to mount HTTP handler", slogx.Error(err), slogx.String("module", module))
			}
			logger.InfoContext(ctx, "Mounted HTTP handler", slogx.String("module", module))
		}

		go func() {
			// stop main process if API stopped
			defer stop()

			logger.InfoContext(ctx, "Started HTTP server", slog.Int("port", conf.HTTPServer.Port))
			if err := app.Listen(fmt.Sprintf(":%d", conf.HTTPServer.Port)); err != nil {
				logger.PanicContext(ctx, "Failed to start HTTP server", slogx.Error(err))
			}
		}()
	}

	// Stop application if worker context is done
	go func() {
		<-ctxWorker.Done()
		defer stop()

		logger.InfoContext(ctx, "Worker is stopped. Stopping main process...")
	}()

	// Wait for interrupt signal to gracefully stop the server
	<-ctx.Done()

	// Force shutdown if timeout exceeded or got signal again
	go func() {
		defer os.Exit(1)

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		select {
		case <-ctx.Done():
			logger.FatalContext(ctx, "Received exit signal again. Force shutdown...")
		case <-time.After(shutdownTimeout + 15*time.Second):
			logger.FatalContext(ctx, "Shutdown timeout exceeded. Force shutdown...")
		}
	}()

	return nil
}
