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
	"github.com/gaze-network/indexer-network/core/indexer"
	"github.com/gaze-network/indexer-network/internal/config"
	"github.com/gaze-network/indexer-network/modules/nodesale"
	"github.com/gaze-network/indexer-network/modules/runes"
	"github.com/gaze-network/indexer-network/pkg/automaxprocs"
	"github.com/gaze-network/indexer-network/pkg/errorhandler"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/gaze-network/indexer-network/pkg/middleware/requestcontext"
	"github.com/gaze-network/indexer-network/pkg/middleware/requestlogger"
	"github.com/gaze-network/indexer-network/pkg/reportingclient"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/samber/do/v2"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

// Register Modules
var Modules = do.Package(
	do.LazyNamed("runes", runes.New),
	do.LazyNamed("nodesale", nodesale.New),
)

func NewRunCommand() *cobra.Command {
	// Create command
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Start indexer-network service",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := automaxprocs.Init(); err != nil {
				logger.Error("Failed to set GOMAXPROCS", slogx.Error(err))
			}
			return runHandler(cmd, args)
		},
	}

	// Add local flags
	flags := runCmd.Flags()
	flags.Bool("api-only", false, "Run only API server")
	flags.String("modules", "", "Enable specific modules to run. E.g. `runes,brc20`")

	// Bind flags to configuration
	config.BindPFlag("api_only", flags.Lookup("api-only"))
	config.BindPFlag("enable_modules", flags.Lookup("modules"))

	return runCmd
}

const (
	shutdownTimeout = 60 * time.Second
)

func runHandler(cmd *cobra.Command, _ []string) error {
	conf := config.Load()

	// Validate inputs and configurations
	{
		if !conf.Network.IsSupported() {
			return errors.Wrapf(errs.Unsupported, "%q network is not supported", conf.Network.String())
		}
	}

	// Initialize application process context
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	injector := do.New(Modules)
	do.ProvideValue(injector, conf)
	do.ProvideValue(injector, ctx)

	// Initialize Bitcoin RPC client
	do.Provide(injector, func(i do.Injector) (*rpcclient.Client, error) {
		conf := do.MustInvoke[config.Config](i)

		client, err := rpcclient.New(&rpcclient.ConnConfig{
			Host:         conf.BitcoinNode.Host,
			User:         conf.BitcoinNode.User,
			Pass:         conf.BitcoinNode.Pass,
			DisableTLS:   conf.BitcoinNode.DisableTLS,
			HTTPPostMode: true,
		}, nil)
		if err != nil {
			return nil, errors.Wrap(err, "invalid Bitcoin node configuration")
		}

		// Check Bitcoin RPC connection
		{
			start := time.Now()
			logger.InfoContext(ctx, "Connecting to Bitcoin Core RPC Server...", slogx.String("host", conf.BitcoinNode.Host))
			if err := client.Ping(); err != nil {
				return nil, errors.Wrapf(err, "can't connect to Bitcoin Core RPC Server %q", conf.BitcoinNode.Host)
			}
			logger.InfoContext(ctx, "Connected to Bitcoin Core RPC Server", slog.Duration("latency", time.Since(start)))
		}

		return client, nil
	})

	// Initialize reporting client
	do.Provide(injector, func(i do.Injector) (*reportingclient.ReportingClient, error) {
		conf := do.MustInvoke[config.Config](i)
		if conf.Reporting.Disabled {
			return nil, nil
		}

		reportingClient, err := reportingclient.New(conf.Reporting)
		if err != nil {
			if errors.Is(err, errs.InvalidArgument) {
				return nil, errors.Wrap(err, "invalid reporting configuration")
			}
			return nil, errors.Wrap(err, "can't create reporting client")
		}
		return reportingClient, nil
	})

	// Initialize HTTP server
	do.Provide(injector, func(i do.Injector) (*fiber.App, error) {
		app := fiber.New(fiber.Config{
			AppName:      "Gaze Indexer",
			ErrorHandler: errorhandler.NewHTTPErrorHandler(),
		})
		app.
			Use(favicon.New()).
			Use(cors.New()).
			Use(requestid.New()).
			Use(requestcontext.New(
				requestcontext.WithRequestId(),
				requestcontext.WithClientIP(conf.HTTPServer.RequestIP),
			)).
			Use(requestlogger.New(conf.HTTPServer.Logger)).
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

		// Health check
		app.Get("/", func(c *fiber.Ctx) error {
			return errors.WithStack(c.SendStatus(http.StatusOK))
		})

		return app, nil
	})

	// Initialize worker context to separate worker's lifecycle from main process
	ctxWorker, stopWorker := context.WithCancel(context.Background())
	defer stopWorker()

	// Add logger context
	ctxWorker = logger.WithContext(ctxWorker, slogx.Stringer("network", conf.Network))

	// Run modules
	{
		modules := lo.Uniq(conf.EnableModules)
		modules = lo.Map(modules, func(item string, _ int) string { return strings.TrimSpace(item) })
		modules = lo.Filter(modules, func(item string, _ int) bool { return item != "" })
		for _, module := range modules {
			ctx := logger.WithContext(ctxWorker, slogx.String("module", module))

			indexer, err := do.InvokeNamed[indexer.IndexerWorker](injector, module)
			if err != nil {
				if errors.Is(err, do.ErrServiceNotFound) {
					return errors.Errorf("Module %q is not supported", module)
				}
				return errors.Wrapf(err, "can't init module %q", module)
			}

			// Run Indexer
			if !conf.APIOnly {
				go func() {
					// stop main process if indexer stopped
					defer stop()

					logger.InfoContext(ctx, "Starting Gaze Indexer")
					if err := indexer.Run(ctx); err != nil {
						logger.PanicContext(ctx, "Something went wrong, error during running indexer", slogx.Error(err))
					}
				}()
			}
		}
	}

	// Run API server
	httpServer := do.MustInvoke[*fiber.App](injector)
	go func() {
		// stop main process if API stopped
		defer stop()

		logger.InfoContext(ctx, "Started HTTP server", slog.Int("port", conf.HTTPServer.Port))
		if err := httpServer.Listen(fmt.Sprintf(":%d", conf.HTTPServer.Port)); err != nil {
			logger.PanicContext(ctx, "Something went wrong, error during running HTTP server", slogx.Error(err))
		}
	}()

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

	if err := injector.Shutdown(); err != nil {
		logger.PanicContext(ctx, "Failed while gracefully shutting down", slogx.Error(err))
	}

	return nil
}
