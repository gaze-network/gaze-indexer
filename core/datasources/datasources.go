package datasources

import (
	"context"

	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/internal/subscription"
)

// Datasource is an interface for indexer data sources.
type Datasource[T any] interface {
	Name() string
	Fetch(ctx context.Context, from, to int64) ([]T, error)
	FetchAsync(ctx context.Context, from, to int64, ch chan<- []T) (*subscription.ClientSubscription[[]T], error)
	GetBlockHeader(ctx context.Context, height int64) (types.BlockHeader, error)
}
