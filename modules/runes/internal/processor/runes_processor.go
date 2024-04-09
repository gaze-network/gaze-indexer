package processor

import (
	"context"

	"github.com/gaze-network/indexer-network/core/indexers"
	"github.com/gaze-network/indexer-network/core/types"
)

var _ indexers.BitcoinProcessor = (*runesProcessor)(nil)

type runesProcessor struct{}

func NewRunesProcessor() *runesProcessor {
	return &runesProcessor{}
}

func (r *runesProcessor) Process(ctx context.Context, blocks []types.Block) error {
	panic("implement me")
}

func (r *runesProcessor) CurrentBlock() (types.BlockHeader, error) {
	panic("implement me")
}

func (r *runesProcessor) PrepareData(ctx context.Context, from, to int64) ([]types.Block, error) {
	panic("implement me")
}

func (r *runesProcessor) RevertData(ctx context.Context, from types.BlockHeader) error {
	panic("implement me")
}
