package datasources

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
)

// Make sure to implement the BitcoinDatasource interface
var _ Datasource[[]*types.Block] = (*BitcoinNodeDatasource)(nil)

// BitcoinNodeDatasource fetch data from Bitcoin node for Bitcoin Indexer
type BitcoinNodeDatasource struct{}

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
	return nil, nil
}
