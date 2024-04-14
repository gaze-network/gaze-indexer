package datasources

import (
	"context"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	cstream "github.com/planxnx/concurrent-stream"
	"github.com/samber/lo"
)

// Make sure to implement the BitcoinDatasource interface
var _ Datasource[[]*types.Block] = (*BitcoinNodeDatasource)(nil)

// BitcoinNodeDatasource fetch data from Bitcoin node for Bitcoin Indexer
type BitcoinNodeDatasource struct {
	btcclient *rpcclient.Client
}

// Fetch polling blocks from Bitcoin node
//
//   - from: block height to start fetching, if -1, it will start from genesis block
//   - to: block height to stop fetching, if -1, it will fetch until the latest block
func (d *BitcoinNodeDatasource) Fetch(ctx context.Context, from, to int64) ([]*types.Block, error) {
	ch := make(chan []*types.Block)
	subscription, err := d.FetchAsync(ctx, from, to, ch)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer subscription.Unsubscribe()

	blocks := make([]*types.Block, 0)
	for {
		select {
		case b, ok := <-ch:
			if !ok {
				return blocks, nil
			}
			blocks = append(blocks, b...)
		case <-subscription.Done():
			if err := ctx.Err(); err != nil {
				return nil, errors.Wrap(err, "context done")
			}
			return blocks, nil
		case err := <-subscription.Err():
			if err != nil {
				return nil, errors.Wrap(err, "got error while fetch async")
			}
			return blocks, nil
		case <-ctx.Done():
			return nil, errors.Wrap(ctx.Err(), "context done")
		}
	}
}

// FetchAsync polling blocks from Bitcoin node asynchronously (non-blocking)
//
//   - from: block height to start fetching, if -1, it will start from genesis block
//   - to: block height to stop fetching, if -1, it will fetch until the latest block
func (d *BitcoinNodeDatasource) FetchAsync(ctx context.Context, from, to int64, ch chan<- []*types.Block) (*ClientSubscription[[]*types.Block], error) {
	from, to, skip, err := d.prepareRange(from, to)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare fetch range")
	}

	subscription := newClientSubscription(ch)
	if skip {
		if err := subscription.UnsubscribeWithContext(ctx); err != nil {
			return nil, errors.Wrap(err, "failed to unsubscribe")
		}
		return subscription, nil
	}

	// Create parallel stream
	out := make(chan []*types.Block)
	stream := cstream.NewStream(ctx, 8, out)

	// create slice of block height to fetch
	blockHeights := make([]int64, 0, to-from+1)
	for i := from; i <= to; i++ {
		blockHeights = append(blockHeights, i)
	}

	// Wait for stream to finish and close out channel
	go func() {
		defer close(out)
		_ = stream.Wait()
	}()

	// Fan-out blocks to subscription channel
	go func() {
		defer subscription.Unsubscribe()
		for {
			select {
			case data, ok := <-out:
				// stream closed
				if !ok {
					return
				}

				// empty blocks
				if len(data) == 0 {
					continue
				}

				// send blocks to subscription channel
				if err := subscription.send(ctx, data); err != nil {
					logger.ErrorContext(ctx, "failed while dispatch block", err,
						slogx.Int64("start", data[0].Header.Height),
						slogx.Int64("end", data[len(data)-1].Header.Height),
					)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Parallel fetch blocks from Bitcoin node until complete all block heights
	// or subscription is done.
	go func() {
		defer stream.Close()
		done := subscription.Done()
		chunks := lo.Chunk(blockHeights, 100)
		for _, chunk := range chunks {
			chunk := chunk
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			default:
				stream.Go(func() []*types.Block {
					// TODO: should concurrent fetch block or not ?
					blocks := make([]*types.Block, 0, len(chunk))
					for _, height := range chunk {
						hash, err := d.btcclient.GetBlockHash(height)
						if err != nil {
							logger.ErrorContext(ctx, "failed to get block hash", err, slogx.Int64("height", height))
							if err := subscription.sendError(ctx, errors.Wrapf(err, "failed to get block hash: height: %d", height)); err != nil {
								logger.ErrorContext(ctx, "failed to send error", err)
							}
						}

						block, err := d.btcclient.GetBlock(hash)
						if err != nil {
							logger.ErrorContext(ctx, "failed to get block", err, slogx.Int64("height", height))
							if err := subscription.sendError(ctx, errors.Wrapf(err, "failed to get block: height: %d, hash: %s", height, hash)); err != nil {
								logger.ErrorContext(ctx, "failed to send error", err)
							}
						}

						blocks = append(blocks, types.ParseMsgBlock(block, height))
					}
					return blocks
				})
			}
		}
	}()

	return subscription, nil
}

func (d *BitcoinNodeDatasource) prepareRange(fromHeight, toHeight int64) (start, end int64, skip bool, err error) {
	start = fromHeight
	end = toHeight

	// get current bitcoin block height
	latestBlockHeight, err := d.btcclient.GetBlockCount()
	if err != nil {
		return -1, -1, false, errors.Wrap(err, "failed to get block count")
	}

	// set start to genesis block height
	if start < 0 {
		start = 0
	}

	// set end to current bitcoin block height if
	// - end is -1
	// - end is greater that current bitcoin block height
	if end < 0 || end > latestBlockHeight {
		end = latestBlockHeight
	}

	// if start is greater than end, skip this round
	if start > end {
		return -1, -1, true, nil
	}

	return start, end, false, nil
}
