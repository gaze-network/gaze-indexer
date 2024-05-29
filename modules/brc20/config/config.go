package config

import "github.com/gaze-network/indexer-network/internal/postgres"

type Config struct {
	Datasource  string          `mapstructure:"datasource"`   // Datasource to fetch bitcoin data for Meta-Protocol e.g. `bitcoin-node`
	Database    string          `mapstructure:"database"`     // Database to store data.
	APIHandlers []string        `mapstructure:"api_handlers"` // List of API handlers to enable. (e.g. `http`)
	Postgres    postgres.Config `mapstructure:"postgres"`
}
