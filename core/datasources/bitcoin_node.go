package datasources

import (
	"context"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
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
	_, _, skip, err := d.prepareRange(from, to)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare fetch range")
	}

	subscription := newClientSubscription(ch)
	if skip {
		if err := subscription.UnsubscribeWithContext(ctx); err != nil {
			return nil, errors.Wrap(err, "failed to unsubscribe")
		}
	}

	// TODO: async fetching

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
