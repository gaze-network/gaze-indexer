package bitcoin

import (
	"cmp"
	"context"
	"slices"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/indexers"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/bitcoin/internal/datagateway"
	"github.com/gaze-network/indexer-network/pkg/logger"
)

// Make sure to implement the BitcoinProcessor interface
var _ indexers.BitcoinProcessor = (*Processor)(nil)

type Processor struct {
	bitcoinDg datagateway.BitcoinDataGateway
}

func (p *Processor) Name() string {
	return "Bitcoin"
}

func (p *Processor) Process(ctx context.Context, inputs []*types.Block) error {
	if len(inputs) == 0 {
		return nil
	}

	// Sort ASC by block height
	slices.SortFunc(inputs, func(t1, t2 *types.Block) int {
		return cmp.Compare(t1.Header.Height, t2.Header.Height)
	})

	latestBlock, err := p.bitcoinDg.GetLatestBlockHeader(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get latest indexed block header")
	}

	// check if the given blocks are continue from the latest indexed block
	// return an error to prevent inserting out-of-order blocks or duplicate blocks
	if inputs[0].Header.Height != latestBlock.Height+1 {
		return errors.New("given blocks are not continue from the latest indexed block")
	}

	// check if the given blocks are in sequence and not missing any block
	for i := 1; i < len(inputs); i++ {
		if inputs[i].Header.Height != inputs[i-1].Header.Height+1 {
			return errors.New("given blocks are not in sequence")
		}
	}

	// Insert blocks
	for _, b := range inputs {
		err := p.bitcoinDg.InsertBlock(ctx, b)
		if err != nil {
			return errors.Wrapf(err, "failed to insert block, height: %d, hash: %s", b.Header.Height, b.Header.Hash)
		}
		logger.InfoContext(ctx, "Block inserted", "height", b.Header.Height, "hash", b.Header.Hash)
	}

	return nil
}

func (p *Processor) CurrentBlock(ctx context.Context) (types.BlockHeader, error) {
	b, err := p.bitcoinDg.GetLatestBlockHeader(ctx)
	if err != nil {
		return types.BlockHeader{}, errors.WithStack(err)
	}
	return b, nil
}

func (p *Processor) PrepareData(ctx context.Context, from, to int64) ([]*types.Block, error) {
	// TODO: move out to a separate interface (e.g. DataFetcher)
	return nil, nil
}

func (p *Processor) RevertData(ctx context.Context, from types.BlockHeader) error {
	// TODO: revert synced data to the specified block for re-indexing
	return nil
}
