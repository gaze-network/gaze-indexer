package runes

import (
	"context"

	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/indexers"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/runes/internal/datagateway"
	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
	"github.com/gaze-network/uint128"
)

var _ indexers.BitcoinProcessor = (*Processor)(nil)

type Processor struct {
	runesProcessorDG datagateway.RunesProcessorDataGateway
	entriesToUpdate  map[runes.RuneId]*runes.RuneEntry
}

type NewRunesProcessorParams struct {
	RunesProcessorDG datagateway.RunesProcessorDataGateway
}

func NewRunesProcessor(params NewRunesProcessorParams) *Processor {
	return &Processor{
		runesProcessorDG: params.RunesProcessorDG,
	}
}

func (r *Processor) Process(ctx context.Context, blocks []*types.Block) error {
	for _, block := range blocks {
		for _, tx := range block.Transactions {
			if err := r.processTx(ctx, tx, block.Header); err != nil {
				return errors.Wrap(err, "failed to process tx")
			}
		}
	}
	return nil
}

func (r *Processor) processTx(ctx context.Context, tx *types.Transaction, blockHeader types.BlockHeader) error {
	runestone, err := runes.DecipherRunestone(tx)
	if err != nil {
		return errors.Wrap(err, "failed to decipher runestone")
	}

	unallocated, err := r.getUnallocatedRunes(ctx, tx.TxIn)
	if err != nil {
		return errors.Wrap(err, "failed to get unallocated runes")
	}

	if runestone != nil {
		if runestone.Mint != nil {
			mintRuneId := *runestone.Mint
			amount, err := r.mint(ctx, mintRuneId, uint64(blockHeader.Height))
			if err != nil {
				return errors.Wrap(err, "error during mint")
			}
			unallocated[mintRuneId] = unallocated[mintRuneId].Add(amount)
		}
	}
	// TODO: finish implementation
}

func (r *Processor) getUnallocatedRunes(ctx context.Context, txInputs []*types.TxIn) (map[runes.RuneId]uint128.Uint128, error) {
	unallocatedRunes := make(map[runes.RuneId]uint128.Uint128)
	for _, txIn := range txInputs {
		balances, err := r.runesProcessorDG.GetRunesBalancesAtOutPoint(ctx, wire.OutPoint{
			Hash:  txIn.PreviousOutTxHash,
			Index: txIn.PreviousOutIndex,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to get runes balances in ")
		}
		for runeId, balance := range balances {
			unallocatedRunes[runeId] = unallocatedRunes[runeId].Add(balance)
		}
	}
	return unallocatedRunes, nil
}

func (r *Processor) mint(ctx context.Context, runeId runes.RuneId, height uint64) (uint128.Uint128, error) {
	runeEntry, ok := r.entriesToUpdate[runeId]
	if !ok {
		var err error
		runeEntry, err = r.runesProcessorDG.GetRuneEntryByRuneId(ctx, runeId)
		if err != nil {
			return uint128.Uint128{}, errors.Wrap(err, "failed to get rune entry by rune id")
		}
	}

	amount, err := runeEntry.GetMintableAmount(height)
	if err != nil {
		return uint128.Zero, nil
	}

	runeEntry.Mints = runeEntry.Mints.Add64(1)
	r.entriesToUpdate[runeId] = runeEntry
	return amount, nil
}

func (r *Processor) CurrentBlock() (types.BlockHeader, error) {
	panic("implement me")
}

func (r *Processor) PrepareData(ctx context.Context, from, to int64) ([]*types.Block, error) {
	panic("implement me")
}

func (r *Processor) RevertData(ctx context.Context, from types.BlockHeader) error {
	panic("implement me")
}
