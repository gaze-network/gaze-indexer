package cmd

import (
	"log/slog"

	"github.com/gaze-network/indexer-network/internal/config"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/spf13/cobra"
)

func (c *CommandHandlers) RunHandler(cmd *cobra.Command, args []string) {
	config := config.Load()
	logger.Info("Starting indexer", slog.Any("network", config.Network))
}
