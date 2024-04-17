package postgres

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/datagateway"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/repository/postgres/gen"
	"github.com/jackc/pgx/v5"
)

var _ datagateway.IndexerInfoDataGateway = (*Repository)(nil)

func (r *Repository) GetLatestIndexerState(ctx context.Context) (entity.IndexerState, error) {
	indexerStateModel, err := r.getQueries().GetLatestIndexerState(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.IndexerState{}, errors.WithStack(errs.NotFound)
		}
		return entity.IndexerState{}, errors.Wrap(err, "error during query")
	}
	indexerState := mapIndexerStateModelToType(indexerStateModel)
	return indexerState, nil
}

func (r *Repository) GetLatestIndexerStats(ctx context.Context) (string, common.Network, error) {
	stats, err := r.getQueries().GetLatestIndexerStats(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", errors.WithStack(errs.NotFound)
		}
		return "", "", errors.Wrap(err, "error during query")
	}
	return stats.ClientVersion, common.Network(stats.Network), nil
}

func (r *Repository) SetIndexerState(ctx context.Context, state entity.IndexerState) error {
	params := mapIndexerStateTypeToParams(state)
	if err := r.getQueries().SetIndexerState(ctx, params); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}

func (r *Repository) UpdateIndexerStats(ctx context.Context, clientVersion string, network common.Network) error {
	if err := r.getQueries().UpdateIndexerStats(ctx, gen.UpdateIndexerStatsParams{
		ClientVersion: clientVersion,
		Network:       string(network),
	}); err != nil {
		return errors.Wrap(err, "error during exec")
	}
	return nil
}
