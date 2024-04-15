package cmd

import (
	"log/slog"

	"github.com/gaze-network/indexer-network/internal/config"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/spf13/cobra"
)

type CommandHandlers struct{}

var cmd = &cobra.Command{
	Use:  "gaze",
	Long: `Description of gaze indexer`,
}

func init() {
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
}

func Execute() {
	// Initialize command handlers
	cmds := &CommandHandlers{}

	// Register sub-commands and handlers
	cmd.AddCommand(
		&cobra.Command{
			Use:   "version",
			Short: "Show indexer-network version",
			Run:   cmds.VersionHandler,
		},
		&cobra.Command{
			Use:   "run",
			Short: "Start indexer-network service",
			Run:   cmds.RunHandler,
		},
	)

	// Execute command
	if err := cmd.Execute(); err != nil {
		logger.Panic("Failed to execute root command")
	}
}
