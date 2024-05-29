package migrate

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

type migrateUpCmdOptions struct {
	DatabaseURL string
	Modules     string
}

type migrateUpCmdArgs struct {
	N int
}

func (a *migrateUpCmdArgs) ParseArgs(args []string) error {
	if len(args) > 0 {
		// assume args already validated by cobra to be len(args) <= 1
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return errors.Wrap(err, "failed to parse N")
		}
		a.N = n
	}
	return nil
}

func NewMigrateUpCommand() *cobra.Command {
	opts := &migrateUpCmdOptions{}

	cmd := &cobra.Command{
		Use:     "up [N]",
		Short:   "Apply all or N up migrations",
		Args:    cobra.MaximumNArgs(1),
		Example: `gaze migrate up --database "postgres://postgres:postgres@localhost:5432/gaze-indexer?sslmode=disable"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// args already validated by cobra
			var upArgs migrateUpCmdArgs
			if err := upArgs.ParseArgs(args); err != nil {
				return errors.Wrap(err, "failed to parse args")
			}
			return migrateUpHandler(opts, cmd, upArgs)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.Modules, "modules", "", "Modules to apply up migrations")
	flags.StringVar(&opts.DatabaseURL, "database", "", "Database url to run migration on")

	return cmd
}

func migrateUpHandler(opts *migrateUpCmdOptions, _ *cobra.Command, args migrateUpCmdArgs) error {
	if opts.DatabaseURL == "" {
		return errors.New("--database is required")
	}
	databaseURL, err := url.Parse(opts.DatabaseURL)
	if err != nil {
		return errors.Wrap(err, "failed to parse database URL")
	}
	if _, ok := supportedDrivers[databaseURL.Scheme]; !ok {
		return errors.Errorf("unsupported database driver: %s", databaseURL.Scheme)
	}

	modules := strings.Split(opts.Modules, ",")

	applyUpMigrations := func(module string, sourcePath string, migrationTable string) error {
		newDatabaseURL := cloneURLWithQuery(databaseURL, url.Values{"x-migrations-table": {migrationTable}})
		sourceURL := "file://" + sourcePath
		m, err := migrate.New(sourceURL, newDatabaseURL.String())
		if err != nil {
			if strings.Contains(err.Error(), "no such file or directory") {
				return errors.Wrap(errs.InternalError, "migrations directory not found")
			}
			return errors.Wrap(err, "failed to open database")
		}
		m.Log = &consoleLogger{
			prefix: fmt.Sprintf("[%s] ", module),
		}
		if args.N == 0 {
			m.Log.Printf("Applying up migrations...\n")
			err = m.Up()
		} else {
			m.Log.Printf("Applying %d up migrations...\n", args.N)
			err = m.Steps(args.N)
		}
		if err != nil {
			if !errors.Is(err, migrate.ErrNoChange) {
				return errors.Wrapf(err, "failed to apply %s up migrations", module)
			}
			m.Log.Printf("Migrations already up-to-date\n")
		}
		return nil
	}

	if lo.Contains(modules, "runes") {
		if err := applyUpMigrations("Runes", runesMigrationSource, "runes_schema_migrations"); err != nil {
			return errors.WithStack(err)
		}
	}
	if lo.Contains(modules, "brc20") {
		if err := applyUpMigrations("BRC20", brc20MigrationSource, "brc20_schema_migrations"); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
