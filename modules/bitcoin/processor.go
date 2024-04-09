package bitcoin

import (
	"context"

	"github.com/gaze-network/indexer-network/core/indexers"
	"github.com/gaze-network/indexer-network/core/types"
)

// Make sure to implement the BitcoinProcessor interface
var _ indexers.BitcoinProcessor = (*Processor)(nil)

type Processor struct{}

func (p *Processor) Process(ctx context.Context, inputs []*types.Block) error {
	return nil
}

func (p *Processor) CurrentBlock() (types.BlockHeader, error) {
	// TODO: return the latest indexed block header
	return types.BlockHeader{}, nil
}

func (p *Processor) PrepareData(ctx context.Context, from, to int64) ([]*types.Block, error) {
	// TODO: move out to a separate interface (e.g. DataFetcher)
	return nil, nil
}

func (p *Processor) RevertData(ctx context.Context, from types.BlockHeader) error {
	// TODO: revert synced data to the specified block for re-indexing
	return nil
}
