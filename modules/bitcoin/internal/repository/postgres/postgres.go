package postgres

import (
	"github.com/gaze-network/indexer-network/internal/postgres"
	"github.com/gaze-network/indexer-network/modules/bitcoin/internal/repository/postgres/gen"
)

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
