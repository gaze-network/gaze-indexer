package indexers

import (
	"context"
	"log/slog"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/datasources"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
)

type (
	BitcoinProcessor  Processor[[]*types.Block]
	BitcoinDatasource datasources.Datasource[[]*types.Block]
)

// Make sure to implement the IndexerWorker interface
var _ IndexerWorker = (*BitcoinIndexer)(nil)

// BitcoinIndexer is the polling indexer for sync Bitcoin data to the database.
type BitcoinIndexer struct {
	Processor    BitcoinProcessor
	Datasource   BitcoinDatasource
	currentBlock types.BlockHeader
}

func (i *BitcoinIndexer) Run(ctx context.Context) (err error) {
	ctx = logger.WithContext(ctx, slog.String("module", i.Processor.Name()))

	// set to -1 to start from genesis block
	i.currentBlock, err = i.Processor.CurrentBlock(ctx)
	if err != nil {
		if errors.Is(err, errs.NotFound) {
			i.currentBlock.Height = -1
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
			ctx = logger.WithContext(ctx, slog.Int64("current_block_height", i.currentBlock.Height))

			if err := i.process(ctx); err != nil {
				logger.ErrorContext(ctx, "failed to process", err)
				return errors.Wrap(err, "failed to process")
			}
		}
	}
}

func (i *BitcoinIndexer) process(ctx context.Context) (err error) {
	ch := make(chan []*types.Block)
	subscription, err := i.Datasource.FetchAsync(ctx, i.currentBlock.Height, -1, ch)
	if err != nil {
		return errors.Wrap(err, "failed to call fetch async")
	}
	defer subscription.Unsubscribe()

	for {
		select {
		case blocks := <-ch:
			// empty blocks
			if len(blocks) == 0 {
				continue
			}

			// validate reorg from prev current block
			if blocks[0].Header.Height == i.currentBlock.Height {
				currentBlockHeader := blocks[0].Header
				if !currentBlockHeader.Hash.IsEqual(&i.currentBlock.Hash) {
					logger.WarnContext(ctx, "reorg detected",
						slogx.Stringer("current_hash", i.currentBlock.Hash),
						slogx.Stringer("expected_hash", currentBlockHeader.Hash),
					)

					// TODO: find start reorg block
					startReorgHeight := currentBlockHeader.Height // TODO: use start reorg block height
					if err := i.Processor.RevertData(ctx, startReorgHeight); err != nil {
						return errors.Wrap(err, "failed to revert data")
					}
					i.currentBlock = currentBlockHeader
					return nil
				}
			}

			// remove old block
			newBlocks := blocks[1:]
			if len(newBlocks) == 0 {
				continue
			}

			// validate is block is continuous and no reorg
			for i := 1; i < len(newBlocks); i++ {
				if newBlocks[i].Header.Height != newBlocks[i-1].Header.Height+1 {
					return errors.Wrapf(errs.InternalError, "block is not continuous, block[%d] height: %d, block[%d] height: %d", i-1, newBlocks[i-1].Header.Height, i, newBlocks[i].Header.Height)
				}

				if !newBlocks[i].Header.PrevBlock.IsEqual(&newBlocks[i-1].Header.Hash) {
					logger.WarnContext(ctx, "reorg occurred while batch fetching blocks, need to try to fetch again")
					// end current round
					return nil
				}
			}

			// Start processing blocks
			if err := i.Processor.Process(ctx, blocks[1:]); err != nil {
				return errors.WithStack(err)
			}

			// Update current state
			i.currentBlock = blocks[len(blocks)-1].Header
		case <-subscription.Done():
			// end current round
			return nil
		case <-ctx.Done():
			return errors.WithStack(ctx.Err())
		case err := <-subscription.Err():
			if err != nil {
				return errors.Wrap(err, "got error while fetch async")
			}
		}
	}
}
