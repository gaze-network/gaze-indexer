package cmd

import (
	"context"
	"log/slog"

	"github.com/gaze-network/indexer-network/internal/config"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/spf13/cobra"
)

var (
	// root command
	cmd = &cobra.Command{
		Use:  "gaze",
		Long: `Description of gaze indexer`,
	}

	// sub-commands
	cmds = []*cobra.Command{
		NewVersionCommand(),
		NewRunCommand(),
		NewMigrateCommand(),
	}
)

// Execute runs the root command
func Execute(ctx context.Context) {
	var configFile string

	// Add global flags
	flags := cmd.PersistentFlags()
	flags.StringVar(&configFile, "config", "", "config file, E.g.  `./config.yaml`")
	flags.String("network", "mainnet", "network to connect to, E.g. `mainnet` or `testnet`")

	// Bind flags to configuration
	config.BindPFlag("network", flags.Lookup("network"))

	// Initialize configuration and logger on start command
	cobra.OnInitialize(func() {
		// Initialize configuration
		config := config.Parse(configFile)

		// Initialize logger
		if err := logger.Init(config.Logger); err != nil {
			logger.Panic("Failed to initialize logger: %v", slogx.Error(err), slog.Any("config", config.Logger))
		}
	})

	// Register sub-commands
	cmd.AddCommand(cmds...)

	// Execute command
	if err := cmd.ExecuteContext(ctx); err != nil {
		// use cobra to log error message by default
		logger.Debug("Failed to execute root command", slogx.Error(err))
	}
}
