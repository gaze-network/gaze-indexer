package cmd

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/datasources"
	"github.com/gaze-network/indexer-network/core/indexers"
	"github.com/gaze-network/indexer-network/internal/config"
	"github.com/gaze-network/indexer-network/internal/postgres"
	"github.com/gaze-network/indexer-network/modules/bitcoin"
	btcdatagateway "github.com/gaze-network/indexer-network/modules/bitcoin/datagateway"
	btcpostgres "github.com/gaze-network/indexer-network/modules/bitcoin/repository/postgres"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/spf13/cobra"
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

	// Wait for interrupt signal to gracefully stop the server with
	<-ctx.Done()
	logger.InfoContext(ctx, "Shutting down server")
	return nil
}
