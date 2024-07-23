package config

import "github.com/gaze-network/indexer-network/internal/postgres"

type Config struct {
	Postgres         postgres.Config `mapstructure:"postgres"`
	LastBlockDefault int64           `mapstructure:"last_block_default"`
}
