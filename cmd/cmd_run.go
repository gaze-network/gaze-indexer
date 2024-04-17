package cmd

import (
	"context"
	"fmt"
	"log/slog"
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
	runesdatagateway "github.com/gaze-network/indexer-network/modules/runes/datagateway"
	runespostgres "github.com/gaze-network/indexer-network/modules/runes/repository/postgres"
	"github.com/gaze-network/indexer-network/pkg/errorhandler"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

type runCmdOptions struct {
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

	// Initialize context
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Add logger context
	ctx = logger.WithContext(ctx, slogx.Stringer("network", conf.Network))

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

	// use gracefulEG to coordinate graceful shutdown after context is done. (e.g. shut down http server, shutdown logic of each module, etc.)
	gracefulEG, gctx := errgroup.WithContext(context.Background())

	// Initialize Bitcoin Indexer
	if opts.Bitcoin {
		var db btcdatagateway.BitcoinDataGateway
		switch strings.ToLower(conf.Modules.Bitcoin.Database) {
		case "postgresql", "postgres", "pg":
			pg, err := postgres.NewPool(ctx, conf.Modules.Bitcoin.Postgres)
			if err != nil {
				logger.PanicContext(ctx, "Failed to create Postgres connection pool", slogx.Error(err))
			}
			defer pg.Close()
			db = btcpostgres.NewRepository(pg)
		default:
			return errors.Wrapf(errs.Unsupported, "%q database is not supported", conf.Modules.Bitcoin.Database)
		}
		bitcoinProcessor := bitcoin.NewProcessor(db)
		bitcoinNodeDatasource := datasources.NewBitcoinNode(client)
		bitcoinIndexer := indexers.NewBitcoinIndexer(bitcoinProcessor, bitcoinNodeDatasource)

		// Run Indexer
		go func() {
			logger.InfoContext(ctx, "Starting Bitcoin Indexer")
			if err := bitcoinIndexer.Run(ctx); err != nil {
				logger.ErrorContext(ctx, "Failed to run Bitcoin Indexer", slogx.Error(err))
			}

			// stop main process if Bitcoin Indexer failed
			logger.InfoContext(ctx, "Bitcoin Indexer stopped. Stopping main process...")
			stop()
		}()
	}

	// Initialize Runes Indexer
	if opts.Runes {
		var db runesdatagateway.RunesDataGateway
		switch strings.ToLower(conf.Modules.Runes.Database) {
		case "postgres", "pg":
			pg, err := postgres.NewPool(ctx, conf.Modules.Runes.Postgres)
			if err != nil {
				logger.PanicContext(ctx, "Failed to create Postgres connection pool", slogx.Error(err))
			}
			defer pg.Close()
			db = runespostgres.NewRepository(pg)
		default:
			logger.PanicContext(ctx, "Unsupported database", slogx.String("database", conf.Modules.Runes.Database))
		}
		// TODO: add option to change bitcoinNodeDatasource implementation
		var bitcoinDatasource indexers.BitcoinDatasource
		var bitcoinClient btcclient.Contract
		switch strings.ToLower(conf.Modules.Runes.Datasource) {
		case "bitcoin-node":
			bitcoinNodeDatasource := datasources.NewBitcoinNode(client)
			bitcoinDatasource = bitcoinNodeDatasource
			bitcoinClient = bitcoinNodeDatasource
		case "database":
			pg, err := postgres.NewPool(ctx, conf.Modules.Runes.Postgres)
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
		runesProcessor := runes.NewProcessor(db, bitcoinClient, bitcoinDatasource, conf.Network)
		runesIndexer := indexers.NewBitcoinIndexer(runesProcessor, bitcoinDatasource)

		if err := runesProcessor.Init(ctx); err != nil {
			logger.PanicContext(ctx, "Failed to initialize Runes Processor", slogx.Error(err))
		}

		// Run Indexer
		go func() {
			logger.InfoContext(ctx, "Starting Runes Indexer...")
			if err := runesIndexer.Run(ctx); err != nil {
				logger.ErrorContext(ctx, "Failed to run Runes Indexer", slogx.Error(err))
			}

			// stop main process if Runes Indexer failed
			logger.InfoContext(ctx, "Runes Indexer stopped. Stopping main process...")
			stop()
		}()
	}

	// Wait for interrupt signal to gracefully stop the server with
	// Setup HTTP server if there are any HTTP handlers
	if len(httpHandlers) > 0 {
		app := fiber.New(fiber.Config{
			AppName:      "gaze",
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

		// mount http handlers from each http-enabled module
		for module, handler := range httpHandlers {
			if err := handler.Mount(app); err != nil {
				logger.PanicContext(ctx, "Failed to mount HTTP handler", slogx.Error(err), slog.String("module", module))
			}
			logger.InfoContext(ctx, "Mounted HTTP handler", slog.String("module", module))
		}
		go func() {
			logger.InfoContext(ctx, "Started HTTP server", slog.Int("port", conf.HTTPServer.Port))
			if err := app.Listen(fmt.Sprintf(":%d", conf.HTTPServer.Port)); err != nil {
				logger.PanicContext(ctx, "Failed to start HTTP server", slogx.Error(err))
			}
		}()
		// handle graceful shutdown
		gracefulEG.Go(func() error {
			<-ctx.Done()
			logger.InfoContext(gctx, "Stopping HTTP server...")
			if err := app.ShutdownWithTimeout(60 * time.Second); err != nil {
				logger.ErrorContext(gctx, "Error during shutdown HTTP server", slogx.Error(err))
			}
			logger.InfoContext(gctx, "HTTP server stopped gracefully")
			return nil
		})
	}

	<-ctx.Done()
	// wait until all graceful shutdown goroutines are done before returning
	if err := gracefulEG.Wait(); err != nil {
		logger.ErrorContext(ctx, "Failed to shutdown gracefully", slogx.Error(err))
	} else {
		logger.InfoContext(ctx, "Successfully shut down gracefully")
	}
	return nil
}
