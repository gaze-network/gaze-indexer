package runes

import (
	"context"

	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/indexers"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/runes/internal/datagateway"
)

var _ indexers.BitcoinProcessor = (*Processor)(nil)

type Processor struct {
	runesProcessorDG datagateway.RunesProcessorDataGateway
}

type NewRunesProcessorParams struct {
	RunesProcessorDG datagateway.RunesProcessorDataGateway
}

func NewRunesProcessor(params NewRunesProcessorParams) *Processor {
	return &Processor{
		runesProcessorDG: params.RunesProcessorDG,
	}
}

func (r *Processor) Process(ctx context.Context, blocks []types.Block) error {
	for _, block := range blocks {
		for _, tx := range block.Transactions {
			if err := r.processTx(tx, block.BlockHeader); err != nil {
				return errors.Wrap(err, "failed to process tx")
			}
		}
	}
	return nil
}

func (r *Processor) processTx(tx *wire.MsgTx, blockHeader types.BlockHeader) error {
	panic("implement me")
}

func (r *Processor) CurrentBlock() (types.BlockHeader, error) {
	panic("implement me")
}

func (r *Processor) PrepareData(ctx context.Context, from, to int64) ([]types.Block, error) {
	panic("implement me")
}

func (r *Processor) RevertData(ctx context.Context, from types.BlockHeader) error {
	panic("implement me")
}
