package datasources

import (
	"bytes"
	"context"
	"encoding/hex"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/internal/subscription"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	cstream "github.com/planxnx/concurrent-stream"
	"github.com/samber/lo"
)

const (
	blockStreamChunkSize = 5
)

// Make sure to implement the BitcoinDatasource interface
var _ Datasource[*types.Block] = (*BitcoinNodeDatasource)(nil)

// BitcoinNodeDatasource fetch data from Bitcoin node for Bitcoin Indexer
type BitcoinNodeDatasource struct {
	btcclient *rpcclient.Client
}

// NewBitcoinNode create new BitcoinNodeDatasource	with Bitcoin Core RPC Client
func NewBitcoinNode(btcclient *rpcclient.Client) *BitcoinNodeDatasource {
	return &BitcoinNodeDatasource{
		btcclient: btcclient,
	}
}

func (p BitcoinNodeDatasource) Name() string {
	return "bitcoin_node"
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
func (d *BitcoinNodeDatasource) FetchAsync(ctx context.Context, from, to int64, ch chan<- []*types.Block) (*subscription.ClientSubscription[[]*types.Block], error) {
	ctx = logger.WithContext(ctx,
		slogx.String("package", "datasources"),
		slogx.String("datasource", d.Name()),
	)

	from, to, skip, err := d.prepareRange(from, to)
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
			// TODO: Implement throttling logic to control the rate of fetching blocks (block/sec)
			chunk := chunk
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			default:
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
					// TODO: should concurrent fetch block or not ?
					blocks := make([]*types.Block, 0, len(chunk))
					for _, height := range chunk {
						hash, err := d.btcclient.GetBlockHash(height)
						if err != nil {
							logger.ErrorContext(ctx, "Can't get block hash from Bitcoin node rpc", slogx.Error(err), slogx.Int64("height", height))
							if err := subscription.SendError(ctx, errors.Wrapf(err, "failed to get block hash: height: %d", height)); err != nil {
								logger.WarnContext(ctx, "Failed to send datasource error to subscription client", slogx.Error(err))
							}
							return nil
						}

						block, err := d.btcclient.GetBlock(hash)
						if err != nil {
							logger.ErrorContext(ctx, "Can't get block data from Bitcoin node rpc", slogx.Error(err), slogx.Int64("height", height))
							if err := subscription.SendError(ctx, errors.Wrapf(err, "failed to get block: height: %d, hash: %s", height, hash)); err != nil {
								logger.WarnContext(ctx, "Failed to send datasource error to subscription client", slogx.Error(err))
							}
							return nil
						}

						blocks = append(blocks, types.ParseMsgBlock(block, height))
					}
					return blocks
				})
			}
		}
	}()

	return subscription.Client(), nil
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

// GetTransaction fetch transaction from Bitcoin node
func (d *BitcoinNodeDatasource) GetTransactionByHash(ctx context.Context, txHash chainhash.Hash) (*types.Transaction, error) {
	rawTxVerbose, err := d.btcclient.GetRawTransactionVerbose(&txHash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get raw transaction")
	}

	blockHash, err := chainhash.NewHashFromStr(rawTxVerbose.BlockHash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse block hash")
	}
	block, err := d.btcclient.GetBlockVerboseTx(blockHash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get block header")
	}

	// parse tx
	txBytes, err := hex.DecodeString(rawTxVerbose.Hex)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode transaction hex")
	}
	var msgTx wire.MsgTx
	if err := msgTx.Deserialize(bytes.NewReader(txBytes)); err != nil {
		return nil, errors.Wrap(err, "failed to deserialize transaction")
	}
	var txIndex uint32
	for i, tx := range block.Tx {
		if tx.Hex == rawTxVerbose.Hex {
			txIndex = uint32(i)
			break
		}
	}

	return types.ParseMsgTx(&msgTx, block.Height, *blockHash, txIndex), nil
}

// GetBlockHeader fetch block header from Bitcoin node
func (d *BitcoinNodeDatasource) GetBlockHeader(ctx context.Context, height int64) (types.BlockHeader, error) {
	hash, err := d.btcclient.GetBlockHash(height)
	if err != nil {
		return types.BlockHeader{}, errors.Wrap(err, "failed to get block hash")
	}

	block, err := d.btcclient.GetBlockHeader(hash)
	if err != nil {
		return types.BlockHeader{}, errors.Wrap(err, "failed to get block header")
	}

	return types.ParseMsgBlockHeader(*block, height), nil
}
