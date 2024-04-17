package postgres

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/bitcoin/datagateway"
	"github.com/gaze-network/indexer-network/modules/bitcoin/repository/postgres/gen"
	"github.com/jackc/pgx/v5"
)

// Make sure Repository implements the IndexerInformationDataGateway interface
var _ datagateway.IndexerInformationDataGateway = (*Repository)(nil)

func (r *Repository) GetCurrentDBVersion(ctx context.Context) (int32, error) {
	version, err := r.queries.GetCurrentDBVersion(ctx)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return version, nil
}

func (r *Repository) GetLatestIndexerStats(ctx context.Context) (string, common.Network, error) {
	stats, err := r.queries.GetCurrentIndexerStats(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", errors.Join(errs.NotFound, err)
		}
		return "", "", errors.WithStack(err)
	}
	return stats.ClientVersion, common.Network(stats.Network), nil
}

func (r *Repository) UpdateIndexerStats(ctx context.Context, clientVersion string, network common.Network) error {
	if err := r.queries.UpdateIndexerStats(ctx, gen.UpdateIndexerStatsParams{
		ClientVersion: clientVersion,
		Network:       network.String(),
	}); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
