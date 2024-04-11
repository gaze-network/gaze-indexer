package indexers

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/pkg/logger"
)

type (
	BitcoinProcessor  Processor[*types.Block]
	BitcoinDatasource Datasource[*types.Block]
)

// Make sure to implement the IndexerWorker interface
var _ core.IndexerWorker = (*BitcoinIndexer)(nil)

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

// Prepare range of blocks to sync
func (b *BitcoinIndexer) prepareRange(currentBlockHeight int64) (start, end int64, skip bool, err error) {
	latestBlockHeight, err := b.btcclient.GetBlockCount()
	if err != nil {
		return -1, -1, false, errors.Wrap(err, "failed to get block count")
	}

	endHeight := latestBlockHeight
	startHeight := currentBlockHeight + 1
	if startHeight < 0 {
		startHeight = 0
	}

	if startHeight > latestBlockHeight {
		return -1, -1, true, nil
	}

	return startHeight, endHeight, false, nil
}
