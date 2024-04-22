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

const (
	maxReorgLookBack = 1000
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

// NewBitcoinIndexer create new BitcoinIndexer
func NewBitcoinIndexer(processor BitcoinProcessor, datasource BitcoinDatasource) *BitcoinIndexer {
	return &BitcoinIndexer{
		Processor:  processor,
		Datasource: datasource,
	}
}

func (i *BitcoinIndexer) Run(ctx context.Context) (err error) {
	ctx = logger.WithContext(ctx,
		slog.String("indexer", "bitcoin"),
		slog.String("processor", i.Processor.Name()),
		slog.String("datasource", i.Datasource.Name()),
	)

	// set to -1 to start from genesis block
	i.currentBlock, err = i.Processor.CurrentBlock(ctx)
	if err != nil {
		if !errors.Is(err, errs.NotFound) {
			return errors.Wrap(err, "can't init state, failed to get indexer current block")
		}
		i.currentBlock.Height = -1
	}

	// TODO:
	// - compare db version in constants and database
	// - compare current network and local indexed network
	// - update indexer stats

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := i.process(ctx); err != nil {
				logger.ErrorContext(ctx, "failed to process", slogx.Error(err))
				return errors.Wrap(err, "failed to process")
			}
		}
	}
}

func (i *BitcoinIndexer) process(ctx context.Context) (err error) {
	ch := make(chan []*types.Block)
	from, to := i.currentBlock.Height+1, int64(-1)
	logger.InfoContext(ctx, "Fetching blocks", slog.Int64("from", from), slog.Int64("to", to))
	subscription, err := i.Datasource.FetchAsync(ctx, from, to, ch)
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

			startAt := time.Now()
			ctx := logger.WithContext(ctx,
				slog.Int("total_blocks", len(blocks)),
				slogx.Int64("from", blocks[0].Header.Height),
				slogx.Int64("to", blocks[len(blocks)-1].Header.Height),
			)

			// validate reorg from first block
			{
				remoteBlockHeader := blocks[0].Header
				if !remoteBlockHeader.PrevBlock.IsEqual(&i.currentBlock.Hash) {
					logger.WarnContext(ctx, "Reorg detected",
						slogx.Stringer("current_hash", i.currentBlock.Hash),
						slogx.Stringer("expected_hash", remoteBlockHeader.PrevBlock),
					)

					var (
						start                  = time.Now()
						targetHeight           = i.currentBlock.Height - 1
						beforeReorgBlockHeader = types.BlockHeader{
							Height: -1,
						}
					)
					for n := 0; n < maxReorgLookBack; n++ {
						// TODO: concurrent fetch
						indexedHeader, err := i.Processor.GetIndexedBlock(ctx, targetHeight)
						if err != nil {
							return errors.Wrapf(err, "failed to get indexed block, height: %d", targetHeight)
						}

						remoteHeader, err := i.Datasource.GetBlockHeader(ctx, targetHeight)
						if err != nil {
							return errors.Wrapf(err, "failed to get remote block header, height: %d", targetHeight)
						}

						// Found no reorg block
						if indexedHeader.Hash.IsEqual(&remoteHeader.Hash) {
							beforeReorgBlockHeader = remoteHeader
							break
						}

						// Walk back to find fork point
						targetHeight -= 1
					}

					// Reorg look back limit reached
					if beforeReorgBlockHeader.Height < 0 {
						return errors.Wrap(errs.SomethingWentWrong, "reorg look back limit reached")
					}

					// Revert all data since the reorg block
					logger.WarnContext(ctx, "reverting reorg data",
						slogx.Int64("reorg_from", beforeReorgBlockHeader.Height+1),
						slogx.Int64("total_reorg_blocks", i.currentBlock.Height-beforeReorgBlockHeader.Height),
						slogx.Stringer("detect_duration", time.Since(start)),
					)
					if err := i.Processor.RevertData(ctx, beforeReorgBlockHeader.Height+1); err != nil {
						return errors.Wrap(err, "failed to revert data")
					}

					// Set current block to before reorg block and
					// end current round to fetch again
					i.currentBlock = beforeReorgBlockHeader
					logger.Info("Reverted data successfully",
						slogx.Any("current_block", i.currentBlock),
						slogx.Stringer("duration", time.Since(start)),
						slogx.Int64("duration_ms", time.Since(start).Milliseconds()),
					)
					return nil
				}
			}

			// validate is block is continuous and no reorg
			for i := 1; i < len(blocks); i++ {
				if blocks[i].Header.Height != blocks[i-1].Header.Height+1 {
					return errors.Wrapf(errs.InternalError, "block is not continuous, block[%d] height: %d, block[%d] height: %d", i-1, blocks[i-1].Header.Height, i, blocks[i].Header.Height)
				}

				if !blocks[i].Header.PrevBlock.IsEqual(&blocks[i-1].Header.Hash) {
					logger.WarnContext(ctx, "reorg occurred while batch fetching blocks, need to try to fetch again")
					// end current round
					return nil
				}
			}

			// Start processing blocks
			logger.InfoContext(ctx, "Processing blocks")
			if err := i.Processor.Process(ctx, blocks); err != nil {
				return errors.WithStack(err)
			}

			// Update current state
			i.currentBlock = blocks[len(blocks)-1].Header

			logger.InfoContext(ctx, "Processed blocks successfully",
				slogx.Stringer("duration", time.Since(startAt)),
				slogx.Int64("duration_ms", time.Since(startAt).Milliseconds()),
			)
		case <-subscription.Done():
			// end current round
			if err := ctx.Err(); err != nil {
				return errors.Wrap(err, "context done")
			}
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
