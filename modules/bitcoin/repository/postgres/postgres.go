package postgres

import (
	"github.com/gaze-network/indexer-network/internal/postgres"
	"github.com/gaze-network/indexer-network/modules/bitcoin/internal/datagateway"
	"github.com/gaze-network/indexer-network/modules/bitcoin/repository/postgres/gen"
)

// Make sure Repository implements the BitcoinDataGateway interface
var _ datagateway.BitcoinDataGateway = (*Repository)(nil)

type Repository struct {
	db      postgres.DB
	queries *gen.Queries
}

func NewRepository(db postgres.DB) *Repository {
	return &Repository{
		db:      db,
		queries: gen.New(db),
	}
}
