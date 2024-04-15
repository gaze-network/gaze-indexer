package runes

import (
	"bytes"
	"context"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/runes/internal/datagateway"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
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

	logger.DebugContext(ctx, "[RunesProcessor] Received blocks", slogx.Any("blocks", lo.Map(blocks, func(b *types.Block, _ int) int64 {
		return b.Header.Height
	})))
	for _, block := range blocks {
		ctx := logger.WithContext(ctx, slog.Int("block_height", int(block.Header.Height)))
		logger.DebugContext(ctx, "[RunesProcessor] Processing block", slog.Int("txs", len(block.Transactions)))
		for _, tx := range block.Transactions {
			if err := p.processTx(ctx, tx, block.Header); err != nil {
				return errors.Wrap(err, "failed to process tx")
			}
		}
		if err := p.flushBlock(ctx, block.Header); err != nil {
			return errors.Wrap(err, "failed to flush block")
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

	inputBalances, err := p.getInputBalances(ctx, tx.TxIn)
	if err != nil {
		return errors.Wrap(err, "failed to get input balances")
	}
	txInputsPkScripts, err := p.getTxInputsPkScripts(ctx, tx, lo.Keys(inputBalances))
	if err != nil {
		return errors.Wrap(err, "failed to get tx inputs pk scripts")
	}

	if runestone == nil && len(inputBalances) == 0 {
		// no runes involved in this tx
		return nil
	}

	unallocated := make(map[runes.RuneId]uint128.Uint128)
	allocated := make(map[int]map[runes.RuneId]uint128.Uint128)
	for inputIndex, balances := range inputBalances {
		for runeId, amount := range balances {
			unallocated[runeId] = unallocated[runeId].Add(amount)
			p.newSpendOutPoints = append(p.newSpendOutPoints, wire.OutPoint{
				Hash:  tx.TxIn[inputIndex].PreviousOutTxHash,
				Index: tx.TxIn[inputIndex].PreviousOutIndex,
			})
			logger.DebugContext(ctx, "[RunesProcessor] Found runes in tx input",
				slogx.Any("runeId", runeId),
				slogx.Stringer("amount", amount),
				slogx.Stringer("txHash", tx.TxHash),
				slog.Int("blockHeight", int(tx.BlockHeight)),
			)
		}
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
		logger.DebugContext(ctx, "[RunesProcessor] Allocated runes to tx output",
			slogx.Any("runeId", runeId),
			slogx.Stringer("amount", amount),
			slog.Int("output", output),
			slogx.Stringer("txHash", tx.TxHash),
			slog.Int("blockHeight", int(tx.BlockHeight)),
		)
	}

	mints := make(map[runes.RuneId]uint128.Uint128)
	if runestone != nil {
		if runestone.Mint != nil {
			mintRuneId := *runestone.Mint
			amount, err := p.mint(ctx, mintRuneId, blockHeader)
			if err != nil {
				return errors.Wrap(err, "error during mint")
			}
			unallocated[mintRuneId] = unallocated[mintRuneId].Add(amount)
			mints[mintRuneId] = amount
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
					mints[etchedRuneId] = mints[etchedRuneId].Add(premine)
					logger.DebugContext(ctx, "[RunesProcessor] Minted premine", slogx.Any("runeId", etchedRuneId), slogx.Stringer("amount", premine))
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
			if err := p.createRuneEntry(ctx, runestone, etchedRuneId, etchedRune, uint64(tx.BlockHeight)); err != nil {
				return errors.Wrap(err, "failed to create rune entry")
			}
		}
	}

	burns := make(map[runes.RuneId]uint128.Uint128)
	if runestone != nil && runestone.Cenotaph {
		// all input runes and minted runes in a tx with cenotaph are burned
		for runeId, amount := range unallocated {
			burns[runeId] = burns[runeId].Add(amount)
			logger.DebugContext(ctx, "[RunesProcessor] Burned runes in cenotaph", slogx.Any("runeId", runeId), slogx.Stringer("amount", amount))
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
				burns[runeId] = burns[runeId].Add(amount)
				logger.DebugContext(ctx, "[RunesProcessor] Burned runes to no output", slogx.Any("runeId", runeId), slogx.Stringer("amount", amount))
			}
		}
	}

	// update outpoint balances
	for output, balances := range allocated {
		if tx.TxOut[output].IsOpReturn() {
			// burn all allocated runes to OP_RETURN outputs
			for runeId, amount := range balances {
				burns[runeId] = burns[runeId].Add(amount)
				logger.DebugContext(ctx, "[RunesProcessor] Burned runes to OP_RETURN output", slogx.Any("runeId", runeId), slogx.Stringer("amount", amount))
			}
			continue
		}

		p.newOutPointBalances[wire.OutPoint{
			Hash:  tx.TxHash,
			Index: uint32(output),
		}] = balances
	}

	if err := p.updateNewBalances(ctx, tx, txInputsPkScripts, inputBalances, allocated); err != nil {
		return errors.Wrap(err, "failed to update new balances")
	}

	// increment burned amounts in rune entries
	if err := p.incrementBurnedAmount(ctx, burns); err != nil {
		return errors.Wrap(err, "failed to update burned amount")
	}

	// construct RuneTransaction
	runeTx := entity.RuneTransaction{
		Hash:        tx.TxHash,
		BlockHeight: uint64(blockHeader.Height),
		Index:       tx.Index,
		Timestamp:   blockHeader.Timestamp,
		Inputs:      make([]*entity.OutPointBalance, 0),
		Outputs:     make([]*entity.OutPointBalance, 0),
		Mints:       mints,
		Burns:       burns,
		Runestone:   runestone,
	}
	for inputIndex, balances := range inputBalances {
		for runeId, amount := range balances {
			pkScript := txInputsPkScripts[inputIndex]
			runeTx.Inputs = append(runeTx.Inputs, &entity.OutPointBalance{
				PkScript:   pkScript,
				Id:         runeId,
				Amount:     amount,
				Index:      uint32(inputIndex),
				TxHash:     tx.TxIn[inputIndex].PreviousOutTxHash,
				TxOutIndex: tx.TxIn[inputIndex].PreviousOutIndex,
			})
		}
	}
	for outputIndex, balances := range allocated {
		pkScript := tx.TxOut[outputIndex].PkScript
		for runeId, amount := range balances {
			runeTx.Outputs = append(runeTx.Outputs, &entity.OutPointBalance{
				PkScript: pkScript,
				Id:       runeId,
				Amount:   amount,
				Index:    uint32(outputIndex),
			})
		}
	}
	p.newRuneTxs = append(p.newRuneTxs, &runeTx)
	logger.DebugContext(ctx, "[RunesProcessor] created RuneTransaction", slogx.Any("runeTx", runeTx))
	return nil
}

func (p *Processor) getInputBalances(ctx context.Context, txInputs []*types.TxIn) (map[int]map[runes.RuneId]uint128.Uint128, error) {
	inputBalances := make(map[int]map[runes.RuneId]uint128.Uint128)
	for i, txIn := range txInputs {
		// check balances in p.newOutPointBalances first, since it may use outpoint from txs in the same block
		balances, ok := p.newOutPointBalances[wire.OutPoint{
			Hash:  txIn.PreviousOutTxHash,
			Index: txIn.PreviousOutIndex,
		}]
		if !ok {
			var err error
			balances, err = p.runesDg.GetRunesBalancesAtOutPoint(ctx, wire.OutPoint{
				Hash:  txIn.PreviousOutTxHash,
				Index: txIn.PreviousOutIndex,
			})
			if err != nil {
				return nil, errors.Wrap(err, "failed to get runes balances at outpoint")
			}
		}
		if len(balances) > 0 {
			inputBalances[i] = balances
		}
	}
	return inputBalances, nil
}

func (p *Processor) getTxInputsPkScripts(ctx context.Context, tx *types.Transaction, inputIndexes []int) (map[int][]byte, error) {
	txInPkScripts := make(map[int][]byte)
	for _, inputIndex := range inputIndexes {
		inputIndex := inputIndex
		prevTxHash := tx.TxIn[inputIndex].PreviousOutTxHash
		prevTxOutIndex := tx.TxIn[inputIndex].PreviousOutIndex
		prevTx, err := p.bitcoinClient.GetTransaction(ctx, prevTxHash)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get transaction")
		}
		pkScript := prevTx.TxOut[prevTxOutIndex].PkScript
		txInPkScripts[inputIndex] = pkScript
	}
	return txInPkScripts, nil
}

func (p *Processor) updateNewBalances(ctx context.Context, tx *types.Transaction, txInputsPkScripts map[int][]byte, inputBalances map[int]map[runes.RuneId]uint128.Uint128, allocated map[int]map[runes.RuneId]uint128.Uint128) error {
	// getCurrentBalance returns the current balance of the pkScript and runeId since last flush
	getCurrentBalance := func(ctx context.Context, pkScript []byte, runeId runes.RuneId) (uint128.Uint128, error) {
		balance, err := p.runesDg.GetBalancesByPkScriptAndRuneId(ctx, pkScript, runeId, uint64(tx.BlockHeight-1))
		if err != nil {
			return uint128.Uint128{}, errors.Wrap(err, "failed to get balance by pk script and rune id")
		}
		return balance.Amount, nil
	}

	// deduct balances used in inputs
	for inputIndex, balances := range inputBalances {
		pkScript := txInputsPkScripts[inputIndex]
		pkScriptStr := hex.EncodeToString(pkScript)
		for runeId, amount := range balances {
			if p.newBalances[pkScriptStr][runeId].Cmp(amount) < 0 {
				// total pkScript's balance is less that balance in input. This is impossible. Something is wrong.
				return errors.Errorf("current balance is less than balance in input: %s", runeId)
			}
			if _, ok := p.newBalances[pkScriptStr]; !ok {
				p.newBalances[pkScriptStr] = make(map[runes.RuneId]uint128.Uint128)
			}
			if _, ok := p.newBalances[pkScriptStr][runeId]; !ok {
				balance, err := getCurrentBalance(ctx, pkScript, runeId)
				if err != nil {
					return errors.WithStack(err)
				}
				p.newBalances[pkScriptStr][runeId] = balance
			}
			p.newBalances[pkScriptStr][runeId] = p.newBalances[pkScriptStr][runeId].Sub(amount)
		}
	}

	// add balances allocated in outputs
	for outputIndex, balances := range allocated {
		pkScript := tx.TxOut[outputIndex].PkScript
		pkScriptStr := hex.EncodeToString(pkScript)
		for runeId, amount := range balances {
			if _, ok := p.newBalances[pkScriptStr]; !ok {
				p.newBalances[pkScriptStr] = make(map[runes.RuneId]uint128.Uint128)
			}
			if _, ok := p.newBalances[pkScriptStr][runeId]; !ok {
				balance, err := getCurrentBalance(ctx, pkScript, runeId)
				if err != nil {
					return errors.WithStack(err)
				}
				p.newBalances[pkScriptStr][runeId] = balance
			}
			p.newBalances[pkScriptStr][runeId] = p.newBalances[pkScriptStr][runeId].Add(amount)
		}
	}

	return nil
}

func (p *Processor) mint(ctx context.Context, runeId runes.RuneId, blockHeader types.BlockHeader) (uint128.Uint128, error) {
	runeEntry, err := p.runesDg.GetRuneEntryByRuneId(ctx, runeId)
	if err != nil {
		if errors.Is(err, errs.NotFound) {
			return uint128.Zero, nil
		}
		return uint128.Uint128{}, errors.Wrap(err, "failed to get rune entry by rune id")
	}

	amount, err := runeEntry.GetMintableAmount(uint64(blockHeader.Height))
	if err != nil {
		return uint128.Zero, nil
	}

	if err := p.incrementMintCount(ctx, runeId, blockHeader); err != nil {
		return uint128.Zero, errors.Wrap(err, "failed to increment mint count")
	}
	logger.DebugContext(ctx, "[RunesProcessor] Minted rune", slogx.Any("runeId", runeId), slogx.Stringer("amount", amount), slogx.Stringer("mintCount", runeEntry.Mints), slogx.Stringer("cap", runeEntry.Terms.Cap))
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

		_, err := p.runesDg.GetRuneIdFromRune(ctx, *rune)
		if err != nil && !errors.Is(err, errs.NotFound) {
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
	witness = removeAnnexFromWitness(witness)
	if len(witness) < 2 {
		return txscript.ScriptTokenizer{}, false
	}
	script := witness[len(witness)-2]

	return txscript.MakeScriptTokenizer(0, script), true
}

func removeAnnexFromWitness(witness [][]byte) [][]byte {
	if len(witness) >= 2 && len(witness[len(witness)-1]) > 0 && witness[len(witness)-1][0] == txscript.TaprootAnnexTag {
		return witness[:len(witness)-1]
	}
	return witness
}

func (p *Processor) createRuneEntry(ctx context.Context, runestone *runes.Runestone, runeId runes.RuneId, rune runes.Rune, blockHeight uint64) error {
	var runeEntry *runes.RuneEntry
	if runestone.Cenotaph {
		runeEntry = &runes.RuneEntry{
			RuneId:            runeId,
			SpacedRune:        runes.NewSpacedRune(rune, 0),
			Mints:             uint128.Zero,
			BurnedAmount:      uint128.Zero,
			Premine:           uint128.Zero,
			Symbol:            '¤',
			Divisibility:      0,
			Terms:             nil,
			Turbo:             false,
			CompletedAt:       time.Time{},
			CompletedAtHeight: nil,
		}
	} else {
		etching := runestone.Etching
		runeEntry = &runes.RuneEntry{
			RuneId:            runeId,
			SpacedRune:        runes.NewSpacedRune(rune, lo.FromPtr(etching.Spacers)),
			Mints:             uint128.Zero,
			BurnedAmount:      uint128.Zero,
			Premine:           lo.FromPtr(etching.Premine),
			Symbol:            lo.FromPtrOr(etching.Symbol, '¤'),
			Divisibility:      lo.FromPtr(etching.Divisibility),
			Terms:             etching.Terms,
			Turbo:             etching.Turbo,
			CompletedAt:       time.Time{},
			CompletedAtHeight: nil,
		}
	}
	if err := p.runesDg.CreateRuneEntry(ctx, runeEntry, blockHeight); err != nil {
		return errors.Wrap(err, "failed to set rune entry")
	}
	logger.DebugContext(ctx, "[RunesProcessor] created RuneEntry", slogx.Any("runeEntry", runeEntry))
	return nil
}

func (p *Processor) incrementMintCount(ctx context.Context, runeId runes.RuneId, blockHeader types.BlockHeader) (err error) {
	runeEntry, ok := p.newRuneEntryStates[runeId]
	if !ok {
		runeEntry, err = p.runesDg.GetRuneEntryByRuneId(ctx, runeId)
		if err != nil {
			return errors.Wrap(err, "failed to get rune entry by rune id")
		}
	}
	runeEntry.Mints = runeEntry.Mints.Add64(1)
	if runeEntry.Mints == lo.FromPtr(runeEntry.Terms.Cap) {
		runeEntry.CompletedAt = blockHeader.Timestamp
		runeEntry.CompletedAtHeight = lo.ToPtr(uint64(blockHeader.Height))
	}
	p.newRuneEntryStates[runeId] = runeEntry
	return nil
}

func (p *Processor) incrementBurnedAmount(ctx context.Context, burned map[runes.RuneId]uint128.Uint128) (err error) {
	runeEntries := make(map[runes.RuneId]*runes.RuneEntry)
	runeIdsToFetch := make([]runes.RuneId, 0)
	for runeId := range burned {
		runeEntry, ok := p.newRuneEntryStates[runeId]
		if !ok {
			runeIdsToFetch = append(runeIdsToFetch, runeId)
		} else {
			runeEntries[runeId] = runeEntry
		}
	}
	if len(runeIdsToFetch) > 0 {
		entries, err := p.runesDg.GetRuneEntryByRuneIdBatch(ctx, runeIdsToFetch)
		if err != nil {
			return errors.Wrap(err, "failed to get rune entry by rune id batch")
		}
		if len(entries) != len(runeIdsToFetch) {
			// there are missing entries, db may be corrupted
			return errors.Errorf("missing rune entries: expected %d, got %d", len(runeIdsToFetch), len(entries))
		}
		for runeId, entry := range entries {
			runeEntries[runeId] = entry
		}
	}

	// update rune entries
	for runeId, amount := range burned {
		runeEntry, ok := runeEntries[runeId]
		if !ok {
			continue
		}
		runeEntry.BurnedAmount = runeEntry.BurnedAmount.Add(amount)
		logger.DebugContext(ctx, "[RunesProcessor] burned amount incremented", slogx.Any("runeId", runeId), slogx.Any("amount", amount))
		p.newRuneEntryStates[runeId] = runeEntry
	}
	return nil
}

func (p *Processor) flushBlock(ctx context.Context, blockHeader types.BlockHeader) error {
	// createdIndexedBlock must be called before other flush methods to correctly calculate event hash
	if err := p.createIndexedBlock(ctx, blockHeader); err != nil {
		return errors.Wrap(err, "failed to create indexed block")
	}

	if err := p.flushNewRuneEntryStates(ctx, uint64(blockHeader.Height)); err != nil {
		return errors.Wrap(err, "failed to flush new rune entry states")
	}

	if err := p.flushNewOutPointBalances(ctx, uint64(blockHeader.Height)); err != nil {
		return errors.Wrap(err, "failed to flush new outpoint balances")
	}

	if err := p.flushNewSpendOutPoints(ctx, uint64(blockHeader.Height)); err != nil {
		return errors.Wrap(err, "failed to flush new spend outpoints")
	}

	if err := p.flushNewBalances(ctx, uint64(blockHeader.Height)); err != nil {
		return errors.Wrap(err, "failed to flush new balances")
	}

	if err := p.flushNewRuneTxs(ctx); err != nil {
		return errors.Wrap(err, "failed to flush new rune transactions")
	}
	logger.InfoContext(ctx, "[RunesProcessor] block flushed")
	return nil
}

func (p *Processor) flushNewRuneEntryStates(ctx context.Context, blockHeight uint64) error {
	for _, runeEntry := range p.newRuneEntryStates {
		if err := p.runesDg.CreateRuneEntryState(ctx, runeEntry, blockHeight); err != nil {
			return errors.Wrap(err, "failed to create rune entry state")
		}
	}
	p.newRuneEntryStates = make(map[runes.RuneId]*runes.RuneEntry)
	return nil
}

func (p *Processor) flushNewOutPointBalances(ctx context.Context, blockHeight uint64) error {
	for outPoint, balances := range p.newOutPointBalances {
		if err := p.runesDg.CreateOutPointBalances(ctx, outPoint, balances, blockHeight); err != nil {
			return errors.Wrap(err, "failed to create outpoint balances")
		}
	}
	p.newOutPointBalances = make(map[wire.OutPoint]map[runes.RuneId]uint128.Uint128)
	return nil
}

func (p *Processor) flushNewSpendOutPoints(ctx context.Context, blockHeight uint64) error {
	for _, outPoint := range p.newSpendOutPoints {
		if err := p.runesDg.SpendOutPointBalances(ctx, outPoint, blockHeight); err != nil {
			return errors.Wrap(err, "failed to create spend outpoint")
		}
	}
	p.newSpendOutPoints = make([]wire.OutPoint, 0)
	return nil
}

func (p *Processor) flushNewBalances(ctx context.Context, blockHeight uint64) error {
	params := make([]datagateway.CreateRuneBalancesParams, 0)
	for pkScriptStr, balances := range p.newBalances {
		pkScript, err := hex.DecodeString(pkScriptStr)
		if err != nil {
			return errors.Wrap(err, "failed to decode pk script")
		}
		for runeId, balance := range balances {
			params = append(params, datagateway.CreateRuneBalancesParams{
				PkScript:    pkScript,
				RuneId:      runeId,
				Balance:     balance,
				BlockHeight: blockHeight,
			})
		}
	}
	if err := p.runesDg.CreateRuneBalances(ctx, params); err != nil {
		return errors.Wrap(err, "failed to create balances at block")
	}
	p.newBalances = make(map[string]map[runes.RuneId]uint128.Uint128)
	return nil
}

func (p *Processor) flushNewRuneTxs(ctx context.Context) error {
	for _, runeTx := range p.newRuneTxs {
		if err := p.runesDg.CreateRuneTransaction(ctx, runeTx); err != nil {
			return errors.Wrap(err, "failed to create rune transaction")
		}
	}
	p.newRuneTxs = make([]*entity.RuneTransaction, 0)
	return nil
}

func (p *Processor) createIndexedBlock(ctx context.Context, header types.BlockHeader) error {
	eventHash, err := p.calculateEventHash(header)
	if err != nil {
		return errors.Wrap(err, "failed to calculate event hash")
	}
	prevIndexedBlock, err := p.runesDg.GetIndexedBlockByHeight(ctx, header.Height-1)
	if err != nil && errors.Is(err, errs.NotFound) && header.Height-1 == startingBlockHeader[p.network].Height {
		prevIndexedBlock = &entity.IndexedBlock{
			Height:              startingBlockHeader[p.network].Height,
			Hash:                startingBlockHeader[p.network].Hash,
			EventHash:           chainhash.Hash{},
			CumulativeEventHash: chainhash.Hash{},
		}
		err = nil
	}
	if err != nil {
		if errors.Is(err, errs.NotFound) {
			return errors.Errorf("indexed block not found for height %d. Indexed block must be created for every Bitcoin block", header.Height)
		}
		return errors.Wrap(err, "failed to get indexed block by height")
	}
	cumulativeEventHash := chainhash.DoubleHashH(append(prevIndexedBlock.CumulativeEventHash[:], eventHash[:]...))

	if err := p.runesDg.CreateIndexedBlock(ctx, &entity.IndexedBlock{
		Height:              header.Height,
		Hash:                header.Hash,
		PrevHash:            header.PrevBlock,
		EventHash:           eventHash,
		CumulativeEventHash: cumulativeEventHash,
	}); err != nil {
		return errors.Wrap(err, "failed to create indexed block")
	}
	return nil
}
