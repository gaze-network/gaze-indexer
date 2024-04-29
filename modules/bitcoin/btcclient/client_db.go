package btcclient

import (
	"context"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/datasources"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/internal/subscription"
	"github.com/gaze-network/indexer-network/modules/bitcoin/datagateway"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	cstream "github.com/planxnx/concurrent-stream"
	"github.com/samber/lo"
)

// TODO: Refactor this, datasources.BitcoinNode and This package is the same.

const (
	blockStreamChunkSize = 100
)

// Make sure to implement the BitcoinDatasource interface
var _ datasources.Datasource[[]*types.Block] = (*ClientDatabase)(nil)

// ClientDatabase is a client to connect to the bitcoin database.
type ClientDatabase struct {
	bitcoinDg datagateway.BitcoinDataGateway
}

func NewClientDatabase(bitcoinDg datagateway.BitcoinDataGateway) *ClientDatabase {
	return &ClientDatabase{
		bitcoinDg: bitcoinDg,
	}
}

func (d ClientDatabase) Name() string {
	return "bitcoin_database"
}

func (d *ClientDatabase) Fetch(ctx context.Context, from, to int64) ([]*types.Block, error) {
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

func (d *ClientDatabase) FetchAsync(ctx context.Context, from, to int64, ch chan<- []*types.Block) (*subscription.ClientSubscription[[]*types.Block], error) {
	ctx = logger.WithContext(ctx,
		slogx.String("package", "datasources"),
		slogx.String("datasource", d.Name()),
	)

	from, to, skip, err := d.prepareRange(ctx, from, to)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare fetch range")
	}

	subscription := subscription.NewSubscription(ch)
	if skip {
		if err := subscription.UnsubscribeWithContext(ctx); err != nil {
			return nil, errors.Wrap(err, "failed to unsubscribe")
		}
		return subscription.Client(), nil
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
		defer func() {
			// add a bit delay to prevent shutdown before client receive all blocks
			time.Sleep(100 * time.Millisecond)

			subscription.Unsubscribe()
		}()
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
				if err := subscription.Send(ctx, data); err != nil {
					if errors.Is(err, errs.Closed) {
						return
					}
					logger.WarnContext(ctx, "Failed to send bitcoin blocks to subscription client",
						slogx.Int64("start", data[0].Header.Height),
						slogx.Int64("end", data[len(data)-1].Header.Height),
						slogx.Error(err),
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
		chunks := lo.Chunk(blockHeights, blockStreamChunkSize)
		for _, chunk := range chunks {
			chunk := chunk
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			default:
				if len(chunk) == 0 {
					continue
				}
				stream.Go(func() []*types.Block {
					startAt := time.Now()
					defer func() {
						logger.DebugContext(ctx, "Fetched chunk of blocks from Bitcoin node",
							slogx.Int("total_blocks", len(chunk)),
							slogx.Int64("from", chunk[0]),
							slogx.Int64("to", chunk[len(chunk)-1]),
							slogx.Duration("duration", time.Since(startAt)),
						)
					}()

					fromHeight, toHeight := chunk[0], chunk[len(chunk)-1]
					blocks, err := d.bitcoinDg.GetBlocksByHeightRange(ctx, fromHeight, toHeight)
					if err != nil {
						logger.ErrorContext(ctx, "Can't get block data from Bitcoin database",
							slogx.Error(err),
							slogx.Int64("from", fromHeight),
							slogx.Int64("to", toHeight),
						)
						if err := subscription.SendError(ctx, errors.Wrapf(err, "failed to get blocks: from_height: %d, to_height: %d", fromHeight, toHeight)); err != nil {
							logger.WarnContext(ctx, "Failed to send datasource error to subscription client", slogx.Error(err))
						}
						return nil
					}
					return blocks
				})
			}
		}
	}()

	return subscription.Client(), nil
}

func (c *ClientDatabase) GetBlockHeader(ctx context.Context, height int64) (types.BlockHeader, error) {
	header, err := c.bitcoinDg.GetBlockHeaderByHeight(ctx, height)
	if err != nil {
		return types.BlockHeader{}, errors.WithStack(err)
	}
	return header, nil
}

func (c *ClientDatabase) prepareRange(ctx context.Context, fromHeight, toHeight int64) (start, end int64, skip bool, err error) {
	start = fromHeight
	end = toHeight

	// get current bitcoin block height
	latestBlock, err := c.bitcoinDg.GetLatestBlockHeader(ctx)
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
	if end < 0 || end > latestBlock.Height {
		end = latestBlock.Height
	}

	// if start is greater than end, skip this round
	if start > end {
		return -1, -1, true, nil
	}

	return start, end, false, nil
}

// GetTransactionByHash returns a transaction with the given hash. Returns errs.NotFound if transaction does not exist.
func (c *ClientDatabase) GetTransactionByHash(ctx context.Context, txHash chainhash.Hash) (*types.Transaction, error) {
	tx, err := c.bitcoinDg.GetTransactionByHash(ctx, txHash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get transaction by hash")
	}
	return tx, nil
}
