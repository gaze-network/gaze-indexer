package indexers

import (
	"context"

	"github.com/gaze-network/indexer-network/core/types"
)

type Processor[T any] interface {
	Name() string

	// Process processes the input data and indexes it.
	Process(ctx context.Context, inputs []T) error

	// CurrentBlock returns the latest indexed block header.
	CurrentBlock(ctx context.Context) (types.BlockHeader, error)

	// RevertData revert synced data to the specified block height for re-indexing.
	RevertData(ctx context.Context, from int64) error
}

type Datasource[T any] interface {
	Fetch(ctx context.Context, from, to int64) ([]T, error)
	FetchAsync(ctx context.Context, from, to int64) (stream <-chan []T, stop func(), err error)
}
