package indexers

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

// BitcoinProcessor is indexer processor for Bitcoin Indexer.
type BitcoinProcessor interface {
	// Process processes the input data and indexes it.
	Process(ctx context.Context, input *wire.MsgBlock) error

	// CurrentBlock returns the latest indexed block header.
	CurrentBlock() (types.BlockHeader, error)
}

// BitcoinIndexer is the indexer for sync Bitcoin data to the database.
type BitcoinIndexer struct {
	Processor    BitcoinProcessor
	btcclient    *rpcclient.Client
	currentBlock types.BlockHeader
}

func (b *BitcoinIndexer) Run(ctx context.Context) (err error) {
	// set to -1 to start from genesis block
	b.currentBlock, err = b.Processor.CurrentBlock()
	if err != nil {
		if errors.Is(err, errs.NotFound) {
			b.currentBlock.Height = -1
		}
		return errors.Wrap(err, "can't init state, failed to get indexer current block")
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			blockHeight := b.currentBlock.Height + 1
			if blockHeight < 0 {
				blockHeight = 0
			}

			latestBlockHeight, err := b.btcclient.GetBlockCount()
			if err != nil {
				return errors.Wrap(err, "failed to get block count")
			}

			if blockHeight > latestBlockHeight {
				continue
			}

			// TODO: supports chain reorganization (get latest block from network to compare with the current block, if less than current block, then re-index)

			blockHash, err := b.btcclient.GetBlockHash(blockHeight)
			if err != nil {
				return errors.Wrap(err, "failed to get block hash")
			}

			if blockHeight <= b.currentBlock.Height && !blockHash.IsEqual(b.currentBlock.Hash) {
				// TODO: Chain reorganization detected, need to re-index
				log.Println("Chain reorganization detected, need to re-index")
			}

			block, err := b.btcclient.GetBlock(blockHash)
			if err != nil {
				return errors.Wrap(err, "failed to get block")
			}

			if err := b.Processor.Process(ctx, block); err != nil {
				return errors.WithStack(err)
			}

			b.currentBlock.Hash = blockHash
			b.currentBlock.Height = blockHeight
			b.currentBlock.BlockHeader = block.Header
		}
	}
}
