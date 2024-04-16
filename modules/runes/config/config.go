package config

import "github.com/gaze-network/indexer-network/internal/postgres"

type Config struct {
	Datasource string          `mapstructure:"datasource"` // Datasource to fetch bitcoin data for Meta-Protocol e.g. `bitcoin-node` | `database`
	Database   string          `mapstructure:"database"`   // Database to store runes data.
	Postgres   postgres.Config `mapstructure:"postgres"`
}
