package workers

import (
	"context"
	"log"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/types"
)

// BitcoinProcessor is processor for Bitcoin Indexer.
type BitcoinProcessor interface {
	// Process processes the input data and indexes it.
	Process(ctx context.Context, input *wire.MsgBlock) error

	// CurrentBlock returns the latest indexed block header.
	CurrentBlock() (types.BlockHeader, error)
}

// BitcoinIndexer is the indexer for processing Bitcoin data.
type BitcoinIndexer struct {
	Processor    BitcoinProcessor
	btcclient    *rpcclient.Client
	currentBlock types.BlockHeader
}

func (b *BitcoinIndexer) Run(ctx context.Context) error {
	cur, err := b.Processor.CurrentBlock()
	if err != nil {
		if errors.Is(err, errs.NotFound) {
			b.currentBlock.Height = -1 // set to -1 to start from genesis block
		} else {
			return errors.Wrap(err, "can't init state, failed to get indexer current block")
		}
	} else {
		b.currentBlock = cur
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			// TODO: supports chain reorganization (get latest block from network to compare with the current block, if less than current block, then re-index)
			// b.btcclient.GetBlockCount()
			incomingBlockHeight := b.currentBlock.Height + 1

			incomingHash, err := b.btcclient.GetBlockHash(incomingBlockHeight)
			if err != nil {
				return errors.Wrap(err, "failed to get block hash")
			}

			if incomingBlockHeight <= b.currentBlock.Height && !incomingHash.IsEqual(b.currentBlock.Hash) {
				// TODO: Chain reorganization detected, need to re-index
				log.Println("Chain reorganization detected, need to re-index")
			}

			incomingBlock, err := b.btcclient.GetBlock(incomingHash)
			if err != nil {
				return errors.Wrap(err, "failed to get block")
			}

			if err := b.Processor.Process(ctx, incomingBlock); err != nil {
				return errors.WithStack(err)
			}

			b.currentBlock.Height = incomingBlockHeight
			b.currentBlock.Hash = incomingHash
			b.currentBlock.BlockHeader = incomingBlock.Header
		}
	}
}
