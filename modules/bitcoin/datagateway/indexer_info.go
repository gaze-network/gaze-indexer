package datagateway

import (
	"context"

	"github.com/gaze-network/indexer-network/common"
)

type IndexerInformationDataGateway interface {
	GetCurrentDBVersion(ctx context.Context) (int32, error)
	GetLatestIndexerStats(ctx context.Context) (version string, network common.Network, err error)
	UpdateIndexerStats(ctx context.Context, clientVersion string, network common.Network) error
}
