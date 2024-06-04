package indexer

import (
	"context"
	"log/slog"
	"sync"
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

	// pollingInterval is the default polling interval for the indexer polling worker
	pollingInterval = 15 * time.Second
)

// Indexer generic indexer for fetching and processing data
type Indexer[T Input] struct {
	Processor    Processor[T]
	Datasource   datasources.Datasource[T]
	currentBlock types.BlockHeader

	quitOnce sync.Once
	quit     chan struct{}
	done     chan struct{}
}

// New create new generic indexer
func New[T Input](processor Processor[T], datasource datasources.Datasource[T]) *Indexer[T] {
	return &Indexer[T]{
		Processor:  processor,
		Datasource: datasource,

		quit: make(chan struct{}),
		done: make(chan struct{}),
	}
}

func (i *Indexer[T]) Shutdown() error {
	return i.ShutdownWithContext(context.Background())
}

func (i *Indexer[T]) ShutdownWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return i.ShutdownWithContext(ctx)
}

func (i *Indexer[T]) ShutdownWithContext(ctx context.Context) (err error) {
	i.quitOnce.Do(func() {
		close(i.quit)
		select {
		case <-i.done:
		case <-time.After(180 * time.Second):
			err = errors.Wrap(errs.Timeout, "indexer shutdown timeout")
		case <-ctx.Done():
			err = errors.Wrap(ctx.Err(), "indexer shutdown context canceled")
		}
	})
	return
}

func (i *Indexer[T]) Run(ctx context.Context) (err error) {
	defer close(i.done)

	ctx = logger.WithContext(ctx,
		slog.String("package", "indexers"),
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

	ticker := time.NewTicker(pollingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-i.quit:
			logger.InfoContext(ctx, "Got quit signal, stopping indexer")
			if err := i.Processor.Shutdown(ctx); err != nil {
				logger.ErrorContext(ctx, "Failed to shutdown processor", slogx.Error(err))
				return errors.Wrap(err, "processor shutdown failed")
			}
			return nil
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := i.process(ctx); err != nil {
				logger.ErrorContext(ctx, "Indexer failed while processing", slogx.Error(err))
				return errors.Wrap(err, "process failed")
			}
			logger.DebugContext(ctx, "Waiting for next polling interval")
		}
	}
}

func (i *Indexer[T]) process(ctx context.Context) (err error) {
	// height range to fetch data
	from, to := i.currentBlock.Height+1, int64(-1)

	logger.InfoContext(ctx, "Start fetching input data", slog.Int64("from", from))
	ch := make(chan []T)
	subscription, err := i.Datasource.FetchAsync(ctx, from, to, ch)
	if err != nil {
		return errors.Wrap(err, "failed to fetch input data")
	}
	defer subscription.Unsubscribe()

	for {
		select {
		case <-i.quit:
			return nil
		case inputs := <-ch:
			// empty inputs
			if len(inputs) == 0 {
				continue
			}

			firstInput := inputs[0]
			firstInputHeader := firstInput.BlockHeader()

			startAt := time.Now()
			ctx := logger.WithContext(ctx,
				slogx.Int64("from", firstInputHeader.Height),
				slogx.Int64("to", inputs[len(inputs)-1].BlockHeader().Height),
			)

			// validate reorg from first input
			{
				remoteBlockHeader := firstInputHeader
				if !remoteBlockHeader.PrevBlock.IsEqual(&i.currentBlock.Hash) {
					logger.WarnContext(ctx, "Detected chain reorganization. Searching for fork point...",
						slogx.String("event", "reorg_detected"),
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

					logger.InfoContext(ctx, "Found reorg fork point, starting to revert data...",
						slogx.String("event", "reorg_forkpoint"),
						slogx.Int64("since", beforeReorgBlockHeader.Height+1),
						slogx.Int64("total_blocks", i.currentBlock.Height-beforeReorgBlockHeader.Height),
						slogx.Duration("search_duration", time.Since(start)),
					)

					// Revert all data since the reorg block
					start = time.Now()
					if err := i.Processor.RevertData(ctx, beforeReorgBlockHeader.Height+1); err != nil {
						return errors.Wrap(err, "failed to revert data")
					}

					// Set current block to before reorg block and
					// end current round to fetch again
					i.currentBlock = beforeReorgBlockHeader
					logger.Info("Fixing chain reorganization completed",
						slogx.Int64("current_block", i.currentBlock.Height),
						slogx.Duration("duration", time.Since(start)),
					)
					return nil
				}
			}

			// validate is input is continuous and no reorg
			for i := 1; i < len(inputs); i++ {
				header := inputs[i].BlockHeader()
				prevHeader := inputs[i-1].BlockHeader()
				if header.Height != prevHeader.Height+1 {
					return errors.Wrapf(errs.InternalError, "input is not continuous, input[%d] height: %d, input[%d] height: %d", i-1, prevHeader.Height, i, header.Height)
				}

				if !header.PrevBlock.IsEqual(&prevHeader.Hash) {
					logger.WarnContext(ctx, "Chain Reorganization occurred in the middle of batch fetching inputs, need to try to fetch again")

					// end current round
					return nil
				}
			}

			ctx = logger.WithContext(ctx, slog.Int("total_inputs", len(inputs)))

			// Start processing input
			logger.InfoContext(ctx, "Processing inputs")
			if err := i.Processor.Process(ctx, inputs); err != nil {
				return errors.WithStack(err)
			}

			// Update current state
			i.currentBlock = inputs[len(inputs)-1].BlockHeader()

			logger.InfoContext(ctx, "Processed inputs successfully",
				slogx.String("event", "processed_inputs"),
				slogx.Int64("current_block", i.currentBlock.Height),
				slogx.Duration("duration", time.Since(startAt)),
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
