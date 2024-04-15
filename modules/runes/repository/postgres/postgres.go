package postgres

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/internal/postgres"
	"github.com/gaze-network/indexer-network/modules/runes/repository/postgres/gen"
	"github.com/gaze-network/indexer-network/pkg/logger"
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

var ErrTxAlreadyExists = errors.New("Transaction already exists. Call Commit() or Rollback() first.")

func (r *Repository) Begin(ctx context.Context) (err error) {
	if r.tx != nil {
		return errors.WithStack(ErrTxAlreadyExists)
	}
	r.tx, err = r.db.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	return nil
}

func (r *Repository) Commit(ctx context.Context) error {
	if r.tx == nil {
		return nil
	}
	err := r.tx.Commit(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}
	r.tx = nil
	return nil
}

func (r *Repository) Rollback(ctx context.Context) error {
	if r.tx == nil {
		return nil
	}
	err := r.tx.Rollback(ctx)
	if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		return errors.Wrap(err, "failed to rollback transaction")
	}
	if err == nil {
		logger.InfoContext(ctx, "rolled back transaction")
	}
	r.tx = nil
	return nil
}
