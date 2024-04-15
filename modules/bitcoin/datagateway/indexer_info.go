package datagateway

import "context"

type IndexerInformationDataGateway interface {
	GetCurrentDBVersion(ctx context.Context) (int, error)
	GetCurrentIndexerStats(ctx context.Context) (clientVersion string, network string, err error)
	UpdateIndexerStats(ctx context.Context, clientVersion string, network string) error
}
