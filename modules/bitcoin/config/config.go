package config

import "github.com/gaze-network/indexer-network/internal/postgres"

type Config struct {
	Database string          `mapstructure:"database"` // Database to store bitcoin data.
	Postgres postgres.Config `mapstructure:"postgresql"`
}
