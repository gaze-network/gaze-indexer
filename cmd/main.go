package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btclog"
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
		Output: "text",
		Debug:  true,
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
		panic(err)
	}
	defer client.Shutdown()

	if err := client.Ping(); err != nil {
		panic(err)
	}

	peerInfo, err := client.GetPeerInfo()
	if err != nil {
		panic(err)
	}

	log.Infof("Connected to Bitcoin Core RPC Server, %d peers", len(peerInfo))

	// Wait for interrupt signal to gracefully stop the server with
	<-ctx.Done()
	log.Info("Shutting down server")
}
