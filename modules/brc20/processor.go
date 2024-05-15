package brc20

import (
	"context"

	"github.com/gaze-network/indexer-network/core/indexer"
	"github.com/gaze-network/indexer-network/core/types"
)

// Make sure to implement the Bitcoin Processor interface
var _ indexer.Processor[*types.Block] = (*Processor)(nil)

type Processor struct{}

func NewProcessor() *Processor {
	return &Processor{}
}

// VerifyStates implements indexer.Processor.
func (p *Processor) VerifyStates(ctx context.Context) error {
	panic("unimplemented")
}

// CurrentBlock implements indexer.Processor.
func (p *Processor) CurrentBlock(ctx context.Context) (types.BlockHeader, error) {
	panic("unimplemented")
}

// GetIndexedBlock implements indexer.Processor.
func (p *Processor) GetIndexedBlock(ctx context.Context, height int64) (types.BlockHeader, error) {
	panic("unimplemented")
}

// Name implements indexer.Processor.
func (p *Processor) Name() string {
	panic("unimplemented")
}

// RevertData implements indexer.Processor.
func (p *Processor) RevertData(ctx context.Context, from int64) error {
	panic("unimplemented")
}
