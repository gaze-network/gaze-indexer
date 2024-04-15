package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gaze-network/indexer-network/core/datasources"
	"github.com/gaze-network/indexer-network/core/indexers"
	"github.com/gaze-network/indexer-network/internal/config"
	"github.com/gaze-network/indexer-network/internal/postgres"
	"github.com/gaze-network/indexer-network/modules/bitcoin"
	btcpostgres "github.com/gaze-network/indexer-network/modules/bitcoin/repository/postgres"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/spf13/cobra"
)

func (c *CommandHandlers) RunHandler(cmd *cobra.Command, args []string) {
	conf := config.Load()

	// Initialize context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
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
	logger.InfoContext(ctx, "Connected to Bitcoin Core RPC Server", slogx.String("host", conf.BitcoinNode.Host))

	// Validate network
	if !conf.Network.IsSupported() {
		logger.PanicContext(ctx, "Unsupported network", slogx.String("network", conf.Network.String()))
	}

	// Initialize Bitcoin Indexer
	{
		pg, err := postgres.NewPool(ctx, conf.Modules["bitcoin"].Postgres)
		if err != nil {
			logger.PanicContext(ctx, "Failed to create Postgres connection pool", slogx.Error(err))
		}
		defer pg.Close()
		bitcoinRepository := btcpostgres.NewRepository(pg)
		bitcoinProcessor := bitcoin.NewProcessor(bitcoinRepository)
		bitcoinNodeDatasource := datasources.NewBitcoinNode(client)
		bitcoinIndexer := indexers.NewBitcoinIndexer(bitcoinProcessor, bitcoinNodeDatasource)

		// Run Indexer
		go func() {
			logger.InfoContext(ctx, "Starting Bitcoin Indexer")
			if err := bitcoinIndexer.Run(ctx); err != nil {
				logger.ErrorContext(ctx, "Failed to run Bitcoin Indexer", slogx.Error(err))
			}
		}()
	}

	// Wait for interrupt signal to gracefully stop the server with
	<-ctx.Done()
	logger.InfoContext(ctx, "Shutting down server")
}
