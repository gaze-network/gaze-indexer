package runes

import (
	"bytes"
	"context"
	"time"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/uint128"
	"github.com/samber/lo"
)

func (p *Processor) Process(ctx context.Context, blocks []*types.Block) error {
	err := p.runesDg.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer func() {
		if err := p.runesDg.Rollback(ctx); err != nil {
			logger.ErrorContext(ctx, "failed to rollback transaction", err)
		}
	}()

	for _, block := range blocks {
		for _, tx := range block.Transactions {
			if err := p.processTx(ctx, tx, block.Header); err != nil {
				return errors.Wrap(err, "failed to process tx")
			}
		}
	}
	if err := p.runesDg.Commit(ctx); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}
	return nil
}

func (p *Processor) processTx(ctx context.Context, tx *types.Transaction, blockHeader types.BlockHeader) error {
	runestone, err := runes.DecipherRunestone(tx)
	if err != nil {
		return errors.Wrap(err, "failed to decipher runestone")
	}

	unallocated, err := p.getUnallocatedRunes(ctx, tx.TxIn)
	allocated := make(map[int]map[runes.RuneId]uint128.Uint128)
	if err != nil {
		return errors.Wrap(err, "failed to get unallocated runes")
	}
	allocate := func(output int, runeId runes.RuneId, amount uint128.Uint128) {
		if _, ok := unallocated[runeId]; !ok {
			return
		}
		// cap amount to unallocated amount
		if amount.Cmp(unallocated[runeId]) > 0 {
			amount = unallocated[runeId]
		}
		if amount.IsZero() {
			return
		}
		if _, ok := allocated[output]; !ok {
			allocated[output] = make(map[runes.RuneId]uint128.Uint128)
		}
		allocated[output][runeId] = allocated[output][runeId].Add(amount)
		unallocated[runeId] = unallocated[runeId].Sub(amount)
	}

	if runestone != nil {
		if runestone.Mint != nil {
			mintRuneId := *runestone.Mint
			amount, err := p.mint(ctx, mintRuneId, uint64(blockHeader.Height))
			if err != nil {
				return errors.Wrap(err, "error during mint")
			}
			unallocated[mintRuneId] = unallocated[mintRuneId].Add(amount)
		}

		etching, etchedRuneId, etchedRune, err := p.getEtchedRune(ctx, tx, runestone)
		if err != nil {
			return errors.Wrap(err, "error during getting etched rune")
		}

		if !runestone.Cenotaph {
			// include premine in unallocated, if exists
			if etching != nil {
				premine := lo.FromPtr(etching.Premine)
				if !premine.IsZero() {
					unallocated[etchedRuneId] = unallocated[etchedRuneId].Add(premine)
				}
			}

			// allocate runes
			for _, edict := range runestone.Edicts {
				// sanity check, should not happen since it is already checked in runes.MessageFromIntegers
				if edict.Output > len(tx.TxOut) {
					return errors.New("edict output index is out of range")
				}

				var emptyRuneId runes.RuneId
				// if rune id is empty, then use etched rune id
				if edict.Id == emptyRuneId {
					// empty rune id is only allowed for runestones with etching
					if etching == nil {
						continue
					}
					edict.Id = etchedRuneId
				}

				if edict.Output == len(tx.TxOut) {
					// if output == len(tx.TxOut), then allocate the amount to all outputs

					// find all non-OP_RETURN outputs
					var destinations []int
					for i, txOut := range tx.TxOut {
						if txOut.IsOpReturn() {
							destinations = append(destinations, i)
						}
					}

					if edict.Amount.IsZero() {
						// if amount is zero, divide ALL unallocated amount to all destinations
						amount, remainder := unallocated[edict.Id].QuoRem64(uint64(len(destinations)))
						for i, dest := range destinations {
							// if i < remainder, then add 1 to amount
							allocate(dest, edict.Id, lo.Ternary(i < int(remainder), amount.Add64(1), amount))
						}
					} else {
						// if amount is not zero, allocate the amount to all destinations, sequentially.
						// If there is no more amount to allocate the rest of outputs, then no more will be allocated.
						for _, dest := range destinations {
							allocate(dest, edict.Id, edict.Amount)
						}
					}
				} else {
					// allocate amount to specific output
					var amount uint128.Uint128
					if edict.Amount.IsZero() {
						// if amount is zero, allocate the whole unallocated amount
						amount = unallocated[edict.Id]
					} else {
						amount = edict.Amount
					}

					allocate(edict.Output, edict.Id, amount)
				}
			}
		}

		if etching != nil {
			if err := p.createRuneEntry(ctx, runestone, etchedRuneId, etchedRune); err != nil {
				return errors.Wrap(err, "failed to create rune entry")
			}
		}
	}

	burned := make(map[runes.RuneId]uint128.Uint128)
	if runestone != nil && runestone.Cenotaph {
		// all input runes and minted runes in a tx with cenotaph are burned
		for runeId, amount := range unallocated {
			burned[runeId] = burned[runeId].Add(amount)
		}
	} else {
		// assign all un-allocated runes to the default output (pointer), or the first non
		// OP_RETURN output if there is no default, or if the default output exceeds the number of outputs
		var pointer *uint64
		if runestone != nil && !runestone.Cenotaph && runestone.Pointer != nil && *runestone.Pointer < uint64(len(tx.TxOut)) {
			pointer = runestone.Pointer
		}

		// if no pointer is provided, use the first non-OP_RETURN output
		if pointer == nil {
			for i, txOut := range tx.TxOut {
				if txOut.IsOpReturn() {
					pointer = lo.ToPtr(uint64(i))
					break
				}
			}
		}

		if pointer != nil {
			// allocate all unallocated runes to the pointer
			output := int(*pointer)
			for runeId, amount := range unallocated {
				allocate(output, runeId, amount)
			}
		} else {
			// if pointer is still nil, then no output is available. Burn all unallocated runes.
			for runeId, amount := range unallocated {
				burned[runeId] = burned[runeId].Add(amount)
			}
		}
	}

	// update outpoint balances
	for output, balances := range allocated {
		if tx.TxOut[output].IsOpReturn() {
			// burn all allocated runes to OP_RETURN outputs
			for runeId, amount := range balances {
				burned[runeId] = burned[runeId].Add(amount)
			}
			continue
		}

		if err := p.runesDg.CreateRuneBalancesAtOutPoint(ctx, wire.OutPoint{
			Hash:  tx.TxHash,
			Index: uint32(output),
		}, balances); err != nil {
			return errors.Wrap(err, "failed to create rune balances at out point")
		}
	}

	// increment burned amounts in rune entries
	if err := p.updatedBurnedAmount(ctx, burned); err != nil {
		return errors.Wrap(err, "failed to update burned amount")
	}
	return nil
}

func (p *Processor) getUnallocatedRunes(ctx context.Context, txInputs []*types.TxIn) (map[runes.RuneId]uint128.Uint128, error) {
	unallocatedRunes := make(map[runes.RuneId]uint128.Uint128)
	for _, txIn := range txInputs {
		balances, err := p.runesDg.GetRunesBalancesAtOutPoint(ctx, wire.OutPoint{
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

func (p *Processor) mint(ctx context.Context, runeId runes.RuneId, height uint64) (uint128.Uint128, error) {
	runeEntry, err := p.runesDg.GetRuneEntryByRuneId(ctx, runeId)
	if err != nil {
		if errors.Is(err, errs.NotFound) {
			return uint128.Zero, nil
		}
		return uint128.Uint128{}, errors.Wrap(err, "failed to get rune entry by rune id")
	}

	amount, err := runeEntry.GetMintableAmount(height)
	if err != nil {
		return uint128.Zero, nil
	}

	runeEntry.Mints = runeEntry.Mints.Add64(1)
	if err := p.runesDg.SetRuneEntry(ctx, runeEntry); err != nil {
		return uint128.Uint128{}, errors.Wrap(err, "failed to set rune entry")
	}
	return amount, nil
}

func (p *Processor) getEtchedRune(ctx context.Context, tx *types.Transaction, runestone *runes.Runestone) (*runes.Etching, runes.RuneId, runes.Rune, error) {
	if runestone.Etching == nil {
		return nil, runes.RuneId{}, runes.Rune{}, nil
	}
	rune := runestone.Etching.Rune
	if rune != nil {
		minimumRune := runes.MinimumRuneAtHeight(p.network, uint64(tx.BlockHeight))
		if rune.Cmp(minimumRune) < 0 || rune.IsReserved() {
			return nil, runes.RuneId{}, runes.Rune{}, nil
		}

		_, err := p.runesDg.GetRuneEntryByRune(ctx, *rune)
		if err != nil && errors.Is(err, errs.NotFound) {
			return nil, runes.RuneId{}, runes.Rune{}, errors.Wrap(err, "error during get rune entry by rune")
		}
		// if found, then this is duplicate
		if err == nil {
			return nil, runes.RuneId{}, runes.Rune{}, nil
		}

		// check if tx commits to the rune
		commit, err := p.txCommitsToRune(ctx, tx, *rune)
		if err != nil {
			return nil, runes.RuneId{}, runes.Rune{}, errors.Wrap(err, "error during check tx commits to rune")
		}
		if !commit {
			return nil, runes.RuneId{}, runes.Rune{}, nil
		}
	} else {
		rune = lo.ToPtr(runes.GetReservedRune(uint64(tx.BlockHeight), tx.Index))
	}

	runeId, err := runes.NewRuneId(uint64(tx.BlockHeight), tx.Index)
	if err != nil {
		return nil, runes.RuneId{}, runes.Rune{}, errors.Wrap(err, "failed to create rune id")
	}
	return runestone.Etching, runeId, *rune, nil
}

func (p *Processor) txCommitsToRune(ctx context.Context, tx *types.Transaction, rune runes.Rune) (bool, error) {
	commitment := rune.Commitment()
	for _, txIn := range tx.TxIn {
		tapscript, ok := extractTapScript(txIn.Witness)
		if !ok {
			continue
		}
		for tapscript.Next() {
			// ignore errors and continue to next input
			if tapscript.Err() != nil {
				break
			}
			data := tapscript.Data()
			// tapscript must contain commitment of the rune
			if !bytes.Equal(data, commitment) {
				continue
			}

			// It is impossible to verify that input utxo is a P2TR output with just the input.
			// Need to verify with utxo's pk script.

			// TODO: what is behavior for not found? nil result or error?
			prevTx, err := p.bitcoinClient.GetTransaction(ctx, txIn.PreviousOutTxHash)
			if err != nil {
				logger.ErrorContext(ctx, "failed to get pk script at out point", err)
				continue
			}
			pkScript := prevTx.TxOut[txIn.PreviousOutIndex].PkScript
			// input utxo must be P2TR
			if !txscript.IsPayToTaproot(pkScript) {
				continue
			}
			// input must be mature enough
			confirmations := tx.BlockHeight - prevTx.BlockHeight
			if confirmations < runes.RUNE_COMMIT_BLOCKS {
				continue
			}

			return true, nil
		}
	}
	return false, nil
}

func extractTapScript(witness [][]byte) (txscript.ScriptTokenizer, bool) {
	var script []byte

	witness = removeAnnexFromWitness(witness)
	if len(witness) < 2 {
		return txscript.ScriptTokenizer{}, false
	}

	return txscript.MakeScriptTokenizer(0, script), true
}

func removeAnnexFromWitness(witness [][]byte) [][]byte {
	if len(witness) >= 2 && len(witness[len(witness)-1]) > 0 && witness[len(witness)-1][0] == txscript.TaprootAnnexTag {
		return witness[:len(witness)-1]
	}
	return witness
}

func (p *Processor) createRuneEntry(ctx context.Context, runestone *runes.Runestone, runeId runes.RuneId, rune runes.Rune) error {
	var runeEntry *runes.RuneEntry
	if runestone.Cenotaph {
		runeEntry = &runes.RuneEntry{
			RuneId:         runeId,
			SpacedRune:     runes.NewSpacedRune(rune, 0),
			Mints:          uint128.Zero,
			BurnedAmount:   uint128.Zero,
			Premine:        uint128.Zero,
			Symbol:         0,
			Divisibility:   0,
			Terms:          nil,
			CompletionTime: time.Time{},
		}
	} else {
		etching := runestone.Etching
		runeEntry = &runes.RuneEntry{
			RuneId:         runeId,
			SpacedRune:     runes.NewSpacedRune(rune, lo.FromPtr(etching.Spacers)),
			Mints:          uint128.Zero,
			BurnedAmount:   uint128.Zero,
			Premine:        lo.FromPtr(etching.Premine),
			Symbol:         lo.FromPtr(etching.Symbol),
			Divisibility:   lo.FromPtr(etching.Divisibility),
			Terms:          etching.Terms,
			CompletionTime: time.Time{},
		}
	}
	if err := p.runesDg.SetRuneEntry(ctx, runeEntry); err != nil {
		return errors.Wrap(err, "failed to set rune entry")
	}
	return nil
}

func (p *Processor) updatedBurnedAmount(ctx context.Context, burned map[runes.RuneId]uint128.Uint128) error {
	runeEntries, err := p.runesDg.GetRuneEntryByRuneIdBatch(ctx, lo.Keys(burned))
	if err != nil {
		return errors.Wrap(err, "failed to get rune entry by rune id batch")
	}
	for runeId, amount := range burned {
		runeEntry, ok := runeEntries[runeId]
		if !ok {
			continue
		}
		runeEntry.BurnedAmount = runeEntry.BurnedAmount.Add(amount)
		if err := p.runesDg.SetRuneEntry(ctx, runeEntry); err != nil {
			return errors.Wrap(err, "failed to set rune entry")
		}
	}
	return nil
}
