package datasources

import (
	"context"

	"github.com/gaze-network/indexer-network/core/indexers"
	"github.com/gaze-network/indexer-network/core/types"
)

// Make sure to implement the BitcoinDatasource interface
var _ indexers.BitcoinDatasource = (*BitcoinNodeDatasource)(nil)

type BitcoinNodeDatasource struct{}

func (d *BitcoinNodeDatasource) Fetch(ctx context.Context, from, to int64) ([]*types.Block, error) {
	return nil, nil
}

func (d *BitcoinNodeDatasource) FetchAsync(ctx context.Context, from, to int64) (chan<- []*types.Block, func(), error) {
	ctx, cancel := context.WithCancel(ctx)
	return nil, cancel, nil
}
