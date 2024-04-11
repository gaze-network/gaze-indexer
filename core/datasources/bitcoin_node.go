package datasources

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/indexers"
	"github.com/gaze-network/indexer-network/core/types"
)

// Make sure to implement the BitcoinDatasource interface
var _ indexers.BitcoinDatasource = (*BitcoinNodeDatasource)(nil)

// BitcoinNodeDatasource fetch data from Bitcoin node for Bitcoin Indexer
type BitcoinNodeDatasource struct{}

func (d *BitcoinNodeDatasource) Fetch(ctx context.Context, from, to int64) ([]*types.Block, error) {
	ch, stop, err := d.FetchAsync(ctx, from, to)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	blocks := make([]*types.Block, 0)
	for {
		select {
		case b, ok := <-ch:
			if !ok {
				return blocks, nil
			}
			blocks = append(blocks, b...)
		case <-ctx.Done():
			stop()
			return nil, errors.Wrap(ctx.Err(), "context done")
		}
	}
}

func (d *BitcoinNodeDatasource) FetchAsync(ctx context.Context, from, to int64) (<-chan []*types.Block, func(), error) {
	ctx, cancel := context.WithCancel(ctx)
	_ = ctx
	return nil, cancel, nil
}
