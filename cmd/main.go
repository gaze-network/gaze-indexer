package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gaze-network/indexer-network/internal/config"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
)

var conf = config.LoadConfig()

func main() {
	// Initialize context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize logger
	if err := logger.Init(conf.Logger); err != nil {
		logger.PanicContext(ctx, "Failed to initialize logger: %v", slogx.Error(err), slog.Any("config", conf.Logger))
	}

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

	if err := client.Ping(); err != nil {
		logger.PanicContext(ctx, "Failed to ping Bitcoin Core RPC Server", slogx.Error(err))
	}

	peerInfo, err := client.GetPeerInfo()
	if err != nil {
		logger.PanicContext(ctx, "Failed to get peer info", slogx.Error(err))
	}

	logger.InfoContext(ctx, "Connected to Bitcoin Core RPC Server", slog.Int("peers", len(peerInfo)))

	// Wait for interrupt signal to gracefully stop the server with
	<-ctx.Done()
	logger.InfoContext(ctx, "Shutting down server")
}
