package postgres

import (
	"context"
	"fmt"

	"github.com/Cleverse/go-utilities/utils"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	pgxslog "github.com/mcosta74/pgx-slog"
)

const (
	DefaultMaxConns = 16
	DefaultMinConns = 0
	DefaultLogLevel = tracelog.LogLevelDebug
)

type Config struct {
	Host     string `mapstructure:"host"`     // Default is 127.0.0.1
	Port     string `mapstructure:"port"`     // Default is 5432
	User     string `mapstructure:"user"`     // Default is empty
	Password string `mapstructure:"password"` // Default is empty
	DBName   string `mapstructure:"db_name"`  // Default is postgres
	SSLMode  string `mapstructure:"ssl_mode"` // Default is prefer
	URL      string `mapstructure:"url"`      // If URL is provided, other fields are ignored

	MaxConns int32 `mapstructure:"max_conns"` // Default is 16
	MinConns int32 `mapstructure:"min_conns"` // Default is 0

	Debug bool `mapstructure:"debug"`
}

// New creates a new connection to the database
func New(ctx context.Context, conf Config) (*pgx.Conn, error) {
	// Prepare connection pool configuration
	connConfig, err := pgx.ParseConfig(conf.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse config to create a new connection")
	}
	connConfig.Tracer = conf.QueryTracer()

	// Create a new connection
	conn, err := pgx.ConnectConfig(ctx, connConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a new connection")
	}

	// Test the connection
	if err := conn.Ping(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to connect to the database")
	}

	return conn, nil
}

// NewPool creates a new connection pool to the database
func NewPool(ctx context.Context, conf Config) (*pgxpool.Pool, error) {
	// Prepare connection pool configuration
	connConfig, err := pgxpool.ParseConfig(conf.String())
	if err != nil {
		return nil, errors.Join(errs.InvalidArgument, errors.Wrap(err, "failed while parse config"))
	}
	connConfig.MaxConns = utils.Default(conf.MaxConns, DefaultMaxConns)
	connConfig.MinConns = utils.Default(conf.MinConns, DefaultMinConns)
	connConfig.ConnConfig.Tracer = conf.QueryTracer()

	// Create a new connection pool
	connPool, err := pgxpool.NewWithConfig(ctx, connConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a new connection pool")
	}

	// Test the connection
	if err := connPool.Ping(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to connect to the database")
	}

	return connPool, nil
}

// String returns the connection string (DSN format or URL format)
func (conf Config) String() string {
	if conf.Host == "" {
		conf.Host = "127.0.0.1"
	}
	if conf.Port == "" {
		conf.Port = "5432"
	}
	if conf.SSLMode == "" {
		conf.SSLMode = "prefer"
	}
	if conf.DBName == "" {
		conf.DBName = "postgres"
	}

	// Construct DSN
	connString := fmt.Sprintf("host=%s dbname=%s port=%s sslmode=%s", conf.Host, conf.DBName, conf.Port, conf.SSLMode)
	if conf.User != "" {
		connString = fmt.Sprintf("%s user=%s", connString, conf.User)
	}
	if conf.Password != "" {
		connString = fmt.Sprintf("%s password=%s", connString, conf.Password)
	}

	// Prefer URL over DSN format
	if conf.URL != "" {
		connString = conf.URL
	}

	return connString
}

func (conf Config) QueryTracer() pgx.QueryTracer {
	loglevel := DefaultLogLevel
	if conf.Debug {
		loglevel = tracelog.LogLevelTrace
	}
	return &tracelog.TraceLog{
		Logger:   pgxslog.NewLogger(logger.With("package", "postgres")),
		LogLevel: loglevel,
	}
}
