package cmd

import (
	"github.com/gaze-network/indexer-network/cmd/migrate"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
)

func NewMigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate database schema",
	}
	cmd.AddCommand(
		migrate.NewMigrateUpCommand(),
		migrate.NewMigrateDownCommand(),
	)
	return cmd
}
