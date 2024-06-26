package indexer

import (
	"context"
	"time"

	"github.com/gaze-network/indexer-network/core/types"
)

type Input interface {
	BlockHeader() types.BlockHeader
}

type Processor[T Input] interface {
	Name() string

	// Process processes the input data and indexes it.
	Process(ctx context.Context, inputs []T) error

	// CurrentBlock returns the latest indexed block header.
	CurrentBlock(ctx context.Context) (types.BlockHeader, error)

	// GetIndexedBlock returns the indexed block header by the specified block height.
	GetIndexedBlock(ctx context.Context, height int64) (types.BlockHeader, error)

	// RevertData revert synced data to the specified block height for re-indexing.
	RevertData(ctx context.Context, from int64) error

	// VerifyStates verifies the states of the indexed data and the indexer
	// to ensure the last shutdown was graceful and no missing data.
	VerifyStates(ctx context.Context) error

	// Shutdown gracefully stops the processor. Database connections, network calls, leftover states, etc. should be closed and cleaned up here.
	Shutdown(ctx context.Context) error
}

type IndexerWorker interface {
	Shutdown() error
	ShutdownWithTimeout(timeout time.Duration) error
	ShutdownWithContext(ctx context.Context) (err error)
	Run(ctx context.Context) (err error)
}
