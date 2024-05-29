package datagateway

import (
	"context"

	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
)

type IndexerInfoDataGateway interface {
	GetLatestIndexerState(ctx context.Context) (entity.IndexerState, error)
	CreateIndexerState(ctx context.Context, state entity.IndexerState) error
}
