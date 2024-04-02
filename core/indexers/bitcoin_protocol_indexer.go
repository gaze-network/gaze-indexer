package indexers

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
)

// BitcoinProtocolProcessor is indexer processor for Bitcoin Protocol Indexer. E.g. OrdinalsProcessor, RunesProcessor, AtomicalsProcessor
type BitcoinProtocolProcessor interface {
	// Process processes the input data and indexes it.
	Process(ctx context.Context, input types.Block) error

	// CurrentBlock returns the latest indexed block header.
	CurrentBlock() (types.BlockHeader, error)
}

// BitcoinProtocolIndexer is the indexer for processing & sync Bitcoin Meta-protocols that rely on Bitcoin Data.
type BitcoinProtocolIndexer struct {
	Processor    BitcoinProtocolProcessor
	currentBlock types.BlockHeader
}

func (b *BitcoinProtocolIndexer) Run(ctx context.Context) error {
	cur, err := b.Processor.CurrentBlock()
	if err != nil {
		return errors.Wrap(err, "can't init state, failed to get indexer current block")
	}
	b.currentBlock = cur

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:

			// TODO: Get the latest block from Database (Phase #1)
			var block types.Block

			if block.Height <= b.currentBlock.Height && block.Hash != b.currentBlock.Hash {
				// TODO: Chain reorganization detected, need to re-index
				_ = block
			}

			if err := b.Processor.Process(ctx, block); err != nil {
				return errors.WithStack(err)
			}
		}
	}
}
