package indexers

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/datasources"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/pkg/logger"
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
	btcclient    *rpcclient.Client
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
			ch := make(chan []*types.Block)
			subscription, err := i.Datasource.FetchAsync(ctx, i.currentBlock.Height-1, -1, ch)
			if err != nil {
				return errors.Wrap(err, "failed to call fetch async")
			}

			for {
				select {
				case blocks := <-ch:
					// empty blocks
					if len(blocks) == 0 {
						continue
					}

					if err := i.Processor.Process(ctx, blocks); err != nil {
						return errors.WithStack(err)
					}

					// Update current state
					i.currentBlock = blocks[len(blocks)-1].Header
				case <-subscription.Done():
					// end current loop
					break
				case <-ctx.Done():
					subscription.Unsubscribe()
					return errors.WithStack(ctx.Err())
				case err := <-subscription.Err():
					if err != nil {
						return errors.Wrap(err, "got error while fetch async")
					}
				}
			}

			// Prepare range of blocks to sync
			startHeight, endHeight, skip, err := i.prepareRange(i.currentBlock.Height)
			if err != nil {
				return errors.Wrap(err, "failed to prepare range")
			}
			if skip {
				continue
			}

			logger.InfoContext(ctx, "Syncing blocks from %d to %d\n", startHeight, endHeight)

			// TODO: supports chain reorganization

			/*
				Concurrent block processing (fetch + process)
				1. Fetch blocks concurrently
				2. Process blocks sequentially
			*/

			blockHash, err := i.btcclient.GetBlockHash(endHeight)
			if err != nil {
				return errors.Wrap(err, "failed to get block hash")
			}

			if endHeight <= i.currentBlock.Height && !blockHash.IsEqual(&i.currentBlock.Hash) {
				// TODO: Chain reorganization detected, need to re-index
				log.Println("Chain reorganization detected, need to re-index")
			}

			block, err := i.btcclient.GetBlock(blockHash)
			if err != nil {
				return errors.Wrap(err, "failed to get block")
			}

			b := types.ParseMsgBlock(block, endHeight)
			if err := i.Processor.Process(ctx, []*types.Block{b}); err != nil {
				return errors.WithStack(err)
			}
			i.currentBlock = b.Header
		}
	}
}
