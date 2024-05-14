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
	"github.com/gaze-network/indexer-network/core/indexer"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/internal/config"
	"github.com/gaze-network/indexer-network/internal/postgres"
	"github.com/gaze-network/indexer-network/modules/runes"
	runesapi "github.com/gaze-network/indexer-network/modules/runes/api"
	runesdatagateway "github.com/gaze-network/indexer-network/modules/runes/datagateway"
	runespostgres "github.com/gaze-network/indexer-network/modules/runes/repository/postgres"
	runesusecase "github.com/gaze-network/indexer-network/modules/runes/usecase"
	"github.com/gaze-network/indexer-network/pkg/automaxprocs"
	"github.com/gaze-network/indexer-network/pkg/btcclient"
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
	Runes   bool
}

func NewRunCommand() *cobra.Command {
	opts := &runCmdOptions{}

	// Create command
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Start indexer-network service",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := automaxprocs.Init(); err != nil {
				logger.Error("Failed to set GOMAXPROCS", slogx.Error(err))
			}
			return runHandler(opts, cmd, args)
		},
	}

	// TODO: separate flags and bind flags to each module cmd package.

	// Add local flags
	flags := runCmd.Flags()
	flags.BoolVar(&opts.APIOnly, "api-only", false, "Run only API server")
	flags.BoolVar(&opts.Runes, "runes", false, "Enable Runes indexer module")
	flags.String("runes-db", "postgres", `Database to store runes data. current supported databases: "postgres"`)
	flags.String("runes-datasource", "bitcoin-node", `Datasource to fetch bitcoin data for processing Meta-Protocol data. current supported datasources: "bitcoin-node"`)

	// Bind flags to configuration
	config.BindPFlag("modules.runes.database", flags.Lookup("runes-db"))
	config.BindPFlag("modules.runes.datasource", flags.Lookup("runes-datasource"))

	return runCmd
}

type HttpHandler interface {
	Mount(router fiber.Router) error
}

func runHandler(opts *runCmdOptions, cmd *cobra.Command, _ []string) error {
	conf := config.Load()

	// Validate inputs
	{
		if !conf.Network.IsSupported() {
			return errors.Wrapf(errs.Unsupported, "%q network is not supported", conf.Network.String())
		}
	}

	// Initialize application process context
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize worker context to separate worker's lifecycle from main process
	ctxWorker, stopWorker := context.WithCancel(context.Background())
	defer stopWorker()

	// Add logger context
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
		logger.PanicContext(ctx, "Invalid Bitcoin node configuration", slogx.Error(err))
	}
	defer client.Shutdown()

	// Check Bitcoin RPC connection
	{
		start := time.Now()
		logger.InfoContext(ctx, "Connecting to Bitcoin Core RPC Server...", slogx.String("host", conf.BitcoinNode.Host))
		if err := client.Ping(); err != nil {
			logger.PanicContext(ctx, "Can't connect to Bitcoin Core RPC Server", slogx.String("host", conf.BitcoinNode.Host), slogx.Error(err))
		}
		logger.InfoContext(ctx, "Connected to Bitcoin Core RPC Server", slog.Duration("latency", time.Since(start)))
	}

	// TODO: create module command package.
	// each module should have its own command package and main package will routing the command to the module command package.

	// TODO: refactor module name to specific type instead of string?
	httpHandlers := make(map[string]HttpHandler, 0)

	var reportingClient *reportingclient.ReportingClient
	if !conf.Reporting.Disabled {
		reportingClient, err = reportingclient.New(conf.Reporting)
		if err != nil {
			if errors.Is(err, errs.InvalidArgument) {
				logger.PanicContext(ctx, "Invalid reporting configuration", slogx.Error(err))
			}
			logger.PanicContext(ctx, "Something went wrong, can't create reporting client", slogx.Error(err))
		}
	}

	// Initialize Runes Indexer
	if opts.Runes {
		ctx := logger.WithContext(ctx, slogx.String("module", "runes"))
		var (
			runesDg       runesdatagateway.RunesDataGateway
			indexerInfoDg runesdatagateway.IndexerInfoDataGateway
		)
		switch strings.ToLower(conf.Modules.Runes.Database) {
		case "postgresql", "postgres", "pg":
			pg, err := postgres.NewPool(ctx, conf.Modules.Runes.Postgres)
			if err != nil {
				if errors.Is(err, errs.InvalidArgument) {
					logger.PanicContext(ctx, "Invalid Postgres configuration for indexer", slogx.Error(err))
				}
				logger.PanicContext(ctx, "Something went wrong, can't create Postgres connection pool", slogx.Error(err))
			}
			defer pg.Close()
			runesRepo := runespostgres.NewRepository(pg)
			runesDg = runesRepo
			indexerInfoDg = runesRepo
		default:
			return errors.Wrapf(errs.Unsupported, "%q database for indexer is not supported", conf.Modules.Runes.Database)
		}
		var bitcoinDatasource datasources.Datasource[*types.Block]
		var bitcoinClient btcclient.Contract
		switch strings.ToLower(conf.Modules.Runes.Datasource) {
		case "bitcoin-node":
			bitcoinNodeDatasource := datasources.NewBitcoinNode(client)
			bitcoinDatasource = bitcoinNodeDatasource
			bitcoinClient = bitcoinNodeDatasource
		default:
			return errors.Wrapf(errs.Unsupported, "%q datasource is not supported", conf.Modules.Runes.Datasource)
		}

		if !opts.APIOnly {
			processor := runes.NewProcessor(runesDg, indexerInfoDg, bitcoinClient, conf.Network, reportingClient)
			indexer := indexer.New(processor, bitcoinDatasource)
			defer func() {
				if err := indexer.ShutdownWithTimeout(shutdownTimeout); err != nil {
					logger.ErrorContext(ctx, "Error during shutdown indexer", slogx.Error(err))
					return
				}
				logger.InfoContext(ctx, "Indexer stopped gracefully")
			}()

			if err := processor.VerifyStates(ctx); err != nil {
				return errors.WithStack(err)
			}

			// Run Indexer
			go func() {
				// stop main process if indexer stopped
				defer stop()

				logger.InfoContext(ctx, "Starting Gaze Indexer")
				if err := indexer.Run(ctxWorker); err != nil {
					logger.PanicContext(ctx, "Something went wrong, error during running indexer", slogx.Error(err))
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
				logger.PanicContext(ctx, "Something went wrong, unsupported API handler", slogx.String("handler", handler))
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
					logger.ErrorContext(c.UserContext(), "Something went wrong, panic in http handler", slogx.Any("panic", e), slog.String("stacktrace", string(buf)))
				},
			})).
			Use(compress.New(compress.Config{
				Level: compress.LevelDefault,
			}))

		defer func() {
			if err := app.ShutdownWithTimeout(shutdownTimeout); err != nil {
				logger.ErrorContext(ctx, "Error during shutdown HTTP server", slogx.Error(err))
				return
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
				logger.PanicContext(ctx, "Something went wrong, can't mount HTTP handler", slogx.Error(err), slogx.String("module", module))
			}
			logger.InfoContext(ctx, "Mounted HTTP handler", slogx.String("module", module))
		}

		go func() {
			// stop main process if API stopped
			defer stop()

			logger.InfoContext(ctx, "Started HTTP server", slog.Int("port", conf.HTTPServer.Port))
			if err := app.Listen(fmt.Sprintf(":%d", conf.HTTPServer.Port)); err != nil {
				logger.PanicContext(ctx, "Something went wrong, error during running HTTP server", slogx.Error(err))
			}
		}()
	}

	// Stop application if worker context is done
	go func() {
		<-ctxWorker.Done()
		defer stop()

		logger.InfoContext(ctx, "Gaze Indexer Worker is stopped. Stopping application...")
	}()

	logger.InfoContext(ctxWorker, "Gaze Indexer started")

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
