package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gaze-network/indexer-network/cmd"
	_ "go.uber.org/automaxprocs"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cmd.Execute(ctx)
}
