package postgres

import (
	db "github.com/gaze-network/indexer-network/internal/postgres"
	"github.com/gaze-network/indexer-network/modules/nodesale/repository/postgres/gen"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	Db      db.TxQueryable
	Queries gen.Querier
}

func NewRepository(db db.DB) *Repository {
	return &Repository{
		Db:      db,
		Queries: gen.New(db),
	}
}

func (q *Repository) WithTx(tx pgx.Tx) gen.Querier {
	queries := gen.Queries{}
	return queries.WithTx(tx)
}
