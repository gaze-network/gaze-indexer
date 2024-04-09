package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btclog"
	"github.com/gaze-network/indexer-network/pkg/logger"
)

var (
	logbackend = btclog.NewBackend(os.Stdout)
	log        = logbackend.Logger("local")
)

func init() {
	rpcclient.UseLogger(logbackend.Logger("rpcclient"))
}

func main() {
	// Initialize context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := logger.Init(logger.Config{
		Env:      "Production",
		Platform: "GCP",
	}); err != nil {
		logger.Panic("Failed to initialize logger: %v", logger.AttrError(err))
	}

	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:         os.Getenv("BITCOIN_HOST"),
		User:         "user",
		Pass:         "pass",
		HTTPPostMode: true,
		// DisableTLS:   true,
	}, nil)
	if err != nil {
		logger.Panic("Failed to create Bitcoin Core RPC Client", logger.AttrError(err))
	}
	defer client.Shutdown()

	if err := client.Ping(); err != nil {
		logger.Panic("Failed to ping Bitcoin Core RPC Server", logger.AttrError(err))
	}

	peerInfo, err := client.GetPeerInfo()
	if err != nil {
		logger.Panic("Failed to get peer info", logger.AttrError(err))
	}

	logger.Info("Connected to Bitcoin Core RPC Server", slog.Int("peers", len(peerInfo)))

	// Wait for interrupt signal to gracefully stop the server with
	<-ctx.Done()
	log.Info("Shutting down server")
}
