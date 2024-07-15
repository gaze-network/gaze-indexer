package postgres

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/nodesale/datagateway"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/jackc/pgx/v5"
)

var ErrTxAlreadyExists = errors.New("Transaction already exists. Call Commit() or Rollback() first.")

func (r *Repository) begin(ctx context.Context) (*Repository, error) {
	if r.tx != nil {
		return nil, errors.WithStack(ErrTxAlreadyExists)
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	return &Repository{
		db:      r.db,
		queries: r.queries.WithTx(tx),
		tx:      tx,
	}, nil
}

func (r *Repository) BeginNodesaleTx(ctx context.Context) (datagateway.NodesaleDataGatewayWithTx, error) {
	repo, err := r.begin(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return repo, nil
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
		logger.DebugContext(ctx, "rolled back transaction")
	}
	r.tx = nil
	return nil
}
