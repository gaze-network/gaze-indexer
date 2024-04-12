package indexers

import (
	"context"

	"github.com/gaze-network/indexer-network/core/types"
)

type IndexerWorker interface {
	Run(ctx context.Context) error
}

type Processor[T any] interface {
	Name() string

	// Process processes the input data and indexes it.
	Process(ctx context.Context, inputs T) error

	// CurrentBlock returns the latest indexed block header.
	CurrentBlock(ctx context.Context) (types.BlockHeader, error)

	// GetIndexedBlock returns the indexed block header by the specified block height.
	GetIndexedBlock(ctx context.Context, height int64) (types.BlockHeader, error)

	// RevertData revert synced data to the specified block height for re-indexing.
	RevertData(ctx context.Context, from int64) error
}
