package postgres

import (
	"github.com/gaze-network/indexer-network/internal/postgres"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/repository/postgres/gen"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	db      postgres.DB
	queries *gen.Queries
	tx      pgx.Tx
}

func NewRepository(db postgres.DB) *Repository {
	return &Repository{
		db:      db,
		queries: gen.New(db),
	}
}
