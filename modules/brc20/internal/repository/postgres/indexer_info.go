package postgres

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/datagateway"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
	"github.com/jackc/pgx/v5"
)

var _ datagateway.IndexerInfoDataGateway = (*Repository)(nil)

func (r *Repository) GetLatestIndexerState(ctx context.Context) (entity.IndexerState, error) {
	model, err := r.queries.GetLatestIndexerState(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.IndexerState{}, errors.WithStack(errs.NotFound)
		}
		return entity.IndexerState{}, errors.Wrap(err, "error during query")
	}
	state := mapIndexerStatesModelToType(model)
	return state, nil
}

func (r *Repository) CreateIndexerState(ctx context.Context, state entity.IndexerState) error {
	params := mapIndexerStatesTypeToParams(state)
	if err := r.queries.CreateIndexerState(ctx, params); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}
