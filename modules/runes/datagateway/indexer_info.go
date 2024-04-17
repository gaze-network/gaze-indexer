package datagateway

import (
	"context"

	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
)

type IndexerInfoDataGateway interface {
	GetLatestIndexerState(ctx context.Context) (entity.IndexerState, error)
	GetLatestIndexerStats(ctx context.Context) (version string, network common.Network, err error)
	SetIndexerState(ctx context.Context, state entity.IndexerState) error
	UpdateIndexerStats(ctx context.Context, clientVersion string, network common.Network) error
}
