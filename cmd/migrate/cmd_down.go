package migrate

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

type migrateDownCmdOptions struct {
	DatabaseURL   string
	BitcoinSource string
	RunesSource   string
	Bitcoin       bool
	Runes         bool
	All           bool
}

type migrateDownCmdArgs struct {
	N int
}

func (a *migrateDownCmdArgs) ParseArgs(args []string) error {
	if len(args) > 0 {
		// assume args already validated by cobra to be len(args) <= 1
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return errors.Wrap(err, "failed to parse N")
		}
		if n < 0 {
			return errors.New("N must be a positive integer")
		}
		a.N = n
	}
	return nil
}

func NewMigrateDownCommand() *cobra.Command {
	opts := &migrateDownCmdOptions{}

	cmd := &cobra.Command{
		Use:     "down [N]",
		Short:   "Apply all or N down migrations",
		Args:    cobra.MaximumNArgs(1),
		Example: `gaze migrate down --database "postgres://postgres:postgres@localhost:5432/gaze-indexer?sslmode=disable"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// args already validated by cobra
			var downArgs migrateDownCmdArgs
			if err := downArgs.ParseArgs(args); err != nil {
				return errors.Wrap(err, "failed to parse args")
			}
			return migrateDownHandler(opts, cmd, downArgs)
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&opts.Bitcoin, "bitcoin", false, "Apply Bitcoin down migrations")
	flags.StringVar(&opts.BitcoinSource, "bitcoin-source", "modules/bitcoin/database/postgresql/migrations", "Path to Bitcoin migrations directory. Default is \"modules/bitcoin/database/postgresql/migrations\".")
	flags.BoolVar(&opts.Runes, "runes", false, "Apply Runes down migrations")
	flags.StringVar(&opts.RunesSource, "runes-source", "modules/runes/database/postgresql/migrations", "Path to Runes migrations directory. Default is \"modules/runes/database/postgresql/migrations\".")
	flags.StringVar(&opts.DatabaseURL, "database", "", "Database url to run migration on")
	flags.BoolVar(&opts.All, "all", false, "Confirm apply ALL down migrations without prompt")

	return cmd
}

func migrateDownHandler(opts *migrateDownCmdOptions, _ *cobra.Command, args migrateDownCmdArgs) error {
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
	// prevent accidental down all migrations
	if args.N == 0 && !opts.All {
		input := ""
		fmt.Print("Are you sure you want to apply all down migrations? (y/N):")
		fmt.Scanln(&input)
		if !lo.Contains([]string{"y", "yes"}, strings.ToLower(input)) {
			return nil
		}
	}

	applyDownMigrations := func(module string, sourcePath string, migrationTable string) error {
		newDatabaseURL := cloneURLWithQuery(databaseURL, url.Values{"x-migrations-table": {migrationTable}})
		sourceURL := "file://" + sourcePath
		m, err := migrate.New(sourceURL, newDatabaseURL.String())
		m.Log = &consoleLogger{
			prefix: fmt.Sprintf("[%s] ", module),
		}
		if err != nil {
			return errors.Wrap(err, "failed to create Migrate instance")
		}
		if args.N == 0 {
			m.Log.Printf("Applying down migrations...\n")
			err = m.Down()
		} else {
			m.Log.Printf("Applying %d down migrations...\n", args.N)
			err = m.Steps(-args.N)
		}
		if err != nil {
			if !errors.Is(err, migrate.ErrNoChange) {
				return errors.Wrapf(err, "failed to apply %s down migrations", module)
			}
			m.Log.Printf("No more down migrations to apply\n")
		}
		return nil
	}

	if opts.Bitcoin {
		if err := applyDownMigrations("Bitcoin", opts.BitcoinSource, "bitcoin_schema_migrations"); err != nil {
			return errors.WithStack(err)
		}
	}
	if opts.Runes {
		if err := applyDownMigrations("Runes", opts.RunesSource, "runes_schema_migrations"); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
