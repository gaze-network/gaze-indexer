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
	"github.com/gaze-network/indexer-network/modules/runes/constants"
	"github.com/gaze-network/indexer-network/modules/runes/datagateway"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/gaze-network/indexer-network/pkg/reportingclient"
	"github.com/gaze-network/uint128"
	"github.com/samber/lo"
)

func (p *Processor) Process(ctx context.Context, blocks []*types.Block) error {
	for _, block := range blocks {
		ctx := logger.WithContext(ctx, slog.Int64("height", block.Header.Height))
		logger.InfoContext(ctx, "Processing new block",
			slogx.String("event", "runes_processor_processing_block"),
			slog.Int("txs", len(block.Transactions)),
		)

		start := time.Now()
		for _, tx := range block.Transactions {
			if err := p.processTx(ctx, tx, block.Header); err != nil {
				return errors.Wrap(err, "failed to process tx")
			}
		}
		timeTakenToProcess := time.Since(start)
		logger.InfoContext(ctx, "Processed block",
			slogx.String("event", "runes_processor_processed_block"),
			slog.Duration("time_taken", timeTakenToProcess),
		)

		if err := p.flushBlock(ctx, block.Header); err != nil {
			return errors.Wrap(err, "failed to flush block")
		}
	}
	return nil
}

func (p *Processor) processTx(ctx context.Context, tx *types.Transaction, blockHeader types.BlockHeader) error {
	if tx.BlockHeight < int64(runes.FirstRuneHeight(p.network)) {
		// prevent processing txs before the activation height
		return nil
	}
	runestone, err := runes.DecipherRunestone(tx)
	if err != nil {
		return errors.Wrap(err, "failed to decipher runestone")
	}

	inputBalances, err := p.getInputBalances(ctx, tx.TxIn)
	if err != nil {
		return errors.Wrap(err, "failed to get input balances")
	}

	if runestone == nil && len(inputBalances) == 0 {
		// no runes involved in this tx
		return nil
	}

	unallocated := make(map[runes.RuneId]uint128.Uint128)
	allocated := make(map[int]map[runes.RuneId]uint128.Uint128)
	for _, balances := range inputBalances {
		for runeId, balance := range balances {
			unallocated[runeId] = unallocated[runeId].Add(balance.Amount)
			p.newSpendOutPoints = append(p.newSpendOutPoints, balance.OutPoint)
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
	}

	mints := make(map[runes.RuneId]uint128.Uint128)
	var runeEtched bool
	if runestone != nil {
		if runestone.Mint != nil {
			mintRuneId := *runestone.Mint
			amount, err := p.mint(ctx, mintRuneId, blockHeader)
			if err != nil {
				return errors.Wrap(err, "error during mint")
			}
			if !amount.IsZero() {
				unallocated[mintRuneId] = unallocated[mintRuneId].Add(amount)
				mints[mintRuneId] = amount
			}
		}

		etching, etchedRuneId, etchedRune, err := p.getEtchedRune(ctx, tx, runestone)
		if err != nil {
			return errors.Wrap(err, "error during getting etched rune")
		}
		if etching != nil {
			runeEtched = true
		}

		if !runestone.Cenotaph {
			// include premine in unallocated, if exists
			if etching != nil {
				premine := lo.FromPtr(etching.Premine)
				if !premine.IsZero() {
					unallocated[etchedRuneId] = unallocated[etchedRuneId].Add(premine)
					mints[etchedRuneId] = mints[etchedRuneId].Add(premine)
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
						if !txOut.IsOpReturn() {
							destinations = append(destinations, i)
						}
					}

					if len(destinations) > 0 {
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
			if err := p.createRuneEntry(ctx, runestone, etchedRuneId, etchedRune, tx, blockHeader); err != nil {
				return errors.Wrap(err, "failed to create rune entry")
			}
		}
	}

	burns := make(map[runes.RuneId]uint128.Uint128)
	if runestone != nil && runestone.Cenotaph {
		// all input runes and minted runes in a tx with cenotaph are burned
		for runeId, amount := range unallocated {
			burns[runeId] = burns[runeId].Add(amount)
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
				if !txOut.IsOpReturn() {
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
			}
		}
	}

	// update outpoint balances
	for output, balances := range allocated {
		if tx.TxOut[output].IsOpReturn() {
			// burn all allocated runes to OP_RETURN outputs
			for runeId, amount := range balances {
				burns[runeId] = burns[runeId].Add(amount)
			}
			continue
		}

		outPoint := wire.OutPoint{
			Hash:  tx.TxHash,
			Index: uint32(output),
		}
		for runeId, amount := range balances {
			p.newOutPointBalances[outPoint] = append(p.newOutPointBalances[outPoint], &entity.OutPointBalance{
				RuneId:      runeId,
				PkScript:    tx.TxOut[output].PkScript,
				OutPoint:    outPoint,
				Amount:      amount,
				BlockHeight: uint64(tx.BlockHeight),
				SpentHeight: nil,
			})
		}
	}

	if err := p.updateNewBalances(ctx, tx, inputBalances, allocated); err != nil {
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
		Inputs:      make([]*entity.TxInputOutput, 0),
		Outputs:     make([]*entity.TxInputOutput, 0),
		Mints:       mints,
		Burns:       burns,
		Runestone:   runestone,
		RuneEtched:  runeEtched,
	}
	for inputIndex, balances := range inputBalances {
		for runeId, balance := range balances {
			runeTx.Inputs = append(runeTx.Inputs, &entity.TxInputOutput{
				PkScript:   balance.PkScript,
				RuneId:     runeId,
				Amount:     balance.Amount,
				Index:      uint32(inputIndex),
				TxHash:     tx.TxIn[inputIndex].PreviousOutTxHash,
				TxOutIndex: tx.TxIn[inputIndex].PreviousOutIndex,
			})
		}
	}
	for outputIndex, balances := range allocated {
		pkScript := tx.TxOut[outputIndex].PkScript
		for runeId, amount := range balances {
			runeTx.Outputs = append(runeTx.Outputs, &entity.TxInputOutput{
				PkScript:   pkScript,
				RuneId:     runeId,
				Amount:     amount,
				Index:      uint32(outputIndex),
				TxHash:     tx.TxHash,
				TxOutIndex: uint32(outputIndex),
			})
		}
	}
	p.newRuneTxs = append(p.newRuneTxs, &runeTx)
	return nil
}

func (p *Processor) getInputBalances(ctx context.Context, txInputs []*types.TxIn) (map[int]map[runes.RuneId]*entity.OutPointBalance, error) {
	inputBalances := make(map[int]map[runes.RuneId]*entity.OutPointBalance)
	for i, txIn := range txInputs {
		balances, err := p.getRunesBalancesAtOutPoint(ctx, wire.OutPoint{
			Hash:  txIn.PreviousOutTxHash,
			Index: txIn.PreviousOutIndex,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to get runes balances at outpoint")
		}

		if len(balances) > 0 {
			inputBalances[i] = balances
		}
	}
	return inputBalances, nil
}

func (p *Processor) updateNewBalances(ctx context.Context, tx *types.Transaction, inputBalances map[int]map[runes.RuneId]*entity.OutPointBalance, allocated map[int]map[runes.RuneId]uint128.Uint128) error {
	// getBalanceFromDg returns the current balance of the pkScript and runeId since last flush
	getBalanceFromDg := func(ctx context.Context, pkScript []byte, runeId runes.RuneId) (uint128.Uint128, error) {
		balance, err := p.runesDg.GetBalanceByPkScriptAndRuneId(ctx, pkScript, runeId, uint64(tx.BlockHeight-1))
		if err != nil {
			if errors.Is(err, errs.NotFound) {
				return uint128.Zero, nil
			}
			return uint128.Uint128{}, errors.Wrap(err, "failed to get balance by pk script and rune id")
		}
		return balance.Amount, nil
	}

	// deduct balances used in inputs
	for _, balances := range inputBalances {
		for runeId, balance := range balances {
			pkScript := balance.PkScript
			pkScriptStr := hex.EncodeToString(pkScript)
			if _, ok := p.newBalances[pkScriptStr]; !ok {
				p.newBalances[pkScriptStr] = make(map[runes.RuneId]uint128.Uint128)
			}
			if _, ok := p.newBalances[pkScriptStr][runeId]; !ok {
				balance, err := getBalanceFromDg(ctx, pkScript, runeId)
				if err != nil {
					return errors.WithStack(err)
				}
				p.newBalances[pkScriptStr][runeId] = balance
			}
			if p.newBalances[pkScriptStr][runeId].Cmp(balance.Amount) < 0 {
				// total pkScript's balance is less that balance in input. This is impossible. Something is wrong.
				return errors.Errorf("current balance is less than balance in input: %s", runeId)
			}
			p.newBalances[pkScriptStr][runeId] = p.newBalances[pkScriptStr][runeId].Sub(balance.Amount)
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
				balance, err := getBalanceFromDg(ctx, pkScript, runeId)
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
	runeEntry, err := p.getRuneEntryByRuneId(ctx, runeId)
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
	return amount, nil
}

func (p *Processor) getEtchedRune(ctx context.Context, tx *types.Transaction, runestone *runes.Runestone) (*runes.Etching, runes.RuneId, runes.Rune, error) {
	if runestone.Etching == nil {
		return nil, runes.RuneId{}, runes.Rune{}, nil
	}
	rune := runestone.Etching.Rune
	if rune != nil {
		minimumRune := runes.MinimumRuneAtHeight(p.network, uint64(tx.BlockHeight))
		if rune.Cmp(minimumRune) < 0 {
			return nil, runes.RuneId{}, runes.Rune{}, nil
		}
		if rune.IsReserved() {
			return nil, runes.RuneId{}, runes.Rune{}, nil
		}

		ok, err := p.isRuneExists(ctx, *rune)
		if err != nil {
			return nil, runes.RuneId{}, runes.Rune{}, errors.Wrap(err, "error during check rune existence")
		}
		if ok {
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
	for i, txIn := range tx.TxIn {
		tapscript, ok := extractTapScript(txIn.Witness)
		if !ok {
			continue
		}
		for tapscript.Next() {
			// ignore errors and continue to next input
			if tapscript.Err() != nil {
				break
			}
			// check opcode is valid
			if !runes.IsDataPushOpCode(tapscript.Opcode()) {
				continue
			}

			// tapscript must contain commitment of the rune
			if !bytes.Equal(tapscript.Data(), commitment) {
				continue
			}

			// It is impossible to verify that input utxo is a P2TR output with just the input.
			// Need to verify with utxo's pk script.

			prevTx, blockHeight, err := p.bitcoinClient.GetRawTransactionAndHeightByTxHash(ctx, txIn.PreviousOutTxHash)
			if err != nil && errors.Is(err, errs.NotFound) {
				continue
			}
			if err != nil {
				return false, errors.Wrapf(err, "can't get previous txout for txin `%v:%v`", tx.TxHash.String(), i)
			}
			pkScript := prevTx.TxOut[txIn.PreviousOutIndex].PkScript
			// input utxo must be P2TR
			if !txscript.IsPayToTaproot(pkScript) {
				break
			}
			// input must be mature enough
			confirmations := tx.BlockHeight - blockHeight + 1
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

func (p *Processor) createRuneEntry(ctx context.Context, runestone *runes.Runestone, runeId runes.RuneId, rune runes.Rune, tx *types.Transaction, blockHeader types.BlockHeader) error {
	count, err := p.countRuneEntries(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to count rune entries")
	}

	var runeEntry *runes.RuneEntry
	if runestone.Cenotaph {
		runeEntry = &runes.RuneEntry{
			RuneId:            runeId,
			Number:            count,
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
			EtchingBlock:      uint64(tx.BlockHeight),
			EtchingTxHash:     tx.TxHash,
			EtchedAt:          blockHeader.Timestamp,
		}
	} else {
		etching := runestone.Etching
		runeEntry = &runes.RuneEntry{
			RuneId:            runeId,
			Number:            count,
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
			EtchingBlock:      uint64(tx.BlockHeight),
			EtchingTxHash:     tx.TxHash,
			EtchedAt:          blockHeader.Timestamp,
		}
	}
	p.newRuneEntries[runeId] = runeEntry
	p.newRuneEntryStates[runeId] = runeEntry
	return nil
}

func (p *Processor) incrementMintCount(ctx context.Context, runeId runes.RuneId, blockHeader types.BlockHeader) (err error) {
	runeEntry, err := p.getRuneEntryByRuneId(ctx, runeId)
	if err != nil {
		return errors.Wrap(err, "failed to get rune entry by rune id")
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
	for runeId, amount := range burned {
		if amount.IsZero() {
			// ignore zero burn amount
			continue
		}
		runeEntry, ok := p.newRuneEntryStates[runeId]
		if !ok {
			runeIdsToFetch = append(runeIdsToFetch, runeId)
		} else {
			runeEntries[runeId] = runeEntry
		}
	}
	if len(runeIdsToFetch) > 0 {
		for _, runeId := range runeIdsToFetch {
			runeEntry, err := p.getRuneEntryByRuneId(ctx, runeId)
			if err != nil {
				if errors.Is(err, errs.NotFound) {
					return errors.Wrap(err, "rune entry not found")
				}
				return errors.Wrap(err, "failed to get rune entry by rune id")
			}
			runeEntries[runeId] = runeEntry
		}
	}

	// update rune entries
	for runeId, amount := range burned {
		runeEntry, ok := runeEntries[runeId]
		if !ok {
			continue
		}
		runeEntry.BurnedAmount = runeEntry.BurnedAmount.Add(amount)
		p.newRuneEntryStates[runeId] = runeEntry
	}
	return nil
}

func (p *Processor) countRuneEntries(ctx context.Context) (uint64, error) {
	runeCountInDB, err := p.runesDg.CountRuneEntries(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to count rune entries in db")
	}
	return runeCountInDB + uint64(len(p.newRuneEntries)), nil
}

func (p *Processor) getRuneEntryByRuneId(ctx context.Context, runeId runes.RuneId) (*runes.RuneEntry, error) {
	runeEntry, ok := p.newRuneEntryStates[runeId]
	if ok {
		return runeEntry, nil
	}
	// not checking from p.newRuneEntries since new rune entries add to p.newRuneEntryStates as well

	runeEntry, err := p.runesDg.GetRuneEntryByRuneId(ctx, runeId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rune entry by rune id")
	}
	return runeEntry, nil
}

func (p *Processor) isRuneExists(ctx context.Context, rune runes.Rune) (bool, error) {
	for _, runeEntry := range p.newRuneEntries {
		if runeEntry.SpacedRune.Rune == rune {
			return true, nil
		}
	}

	_, err := p.runesDg.GetRuneIdFromRune(ctx, rune)
	if err != nil {
		if errors.Is(err, errs.NotFound) {
			return false, nil
		}
		return false, errors.Wrap(err, "failed to get rune id from rune")
	}
	return true, nil
}

func (p *Processor) getRunesBalancesAtOutPoint(ctx context.Context, outPoint wire.OutPoint) (map[runes.RuneId]*entity.OutPointBalance, error) {
	if outPointBalances, ok := p.newOutPointBalances[outPoint]; ok {
		balances := make(map[runes.RuneId]*entity.OutPointBalance)
		for _, outPointBalance := range outPointBalances {
			balances[outPointBalance.RuneId] = outPointBalance
		}
		return balances, nil
	}

	balances, err := p.runesDg.GetRunesBalancesAtOutPoint(ctx, outPoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get runes balances at outpoint")
	}
	return balances, nil
}

func (p *Processor) flushBlock(ctx context.Context, blockHeader types.BlockHeader) error {
	start := time.Now()
	runesDgTx, err := p.runesDg.BeginRunesTx(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to begin runes tx")
	}
	defer func() {
		if err := runesDgTx.Rollback(ctx); err != nil {
			logger.WarnContext(ctx, "failed to rollback transaction",
				slogx.Error(err),
				slogx.String("event", "rollback_runes_insertion"),
			)
		}
	}()

	// CreateIndexedBlock must be performed before other flush methods to correctly calculate event hash
	eventHash, err := p.calculateEventHash(blockHeader)
	if err != nil {
		return errors.Wrap(err, "failed to calculate event hash")
	}
	prevIndexedBlock, err := runesDgTx.GetIndexedBlockByHeight(ctx, blockHeader.Height-1)
	if err != nil && errors.Is(err, errs.NotFound) && blockHeader.Height-1 == constants.StartingBlockHeader[p.network].Height {
		prevIndexedBlock = &entity.IndexedBlock{
			Height:              constants.StartingBlockHeader[p.network].Height,
			Hash:                chainhash.Hash{},
			EventHash:           chainhash.Hash{},
			CumulativeEventHash: chainhash.Hash{},
		}
		err = nil
	}
	if err != nil {
		if errors.Is(err, errs.NotFound) {
			return errors.Errorf("indexed block not found for height %d. Indexed block must be created for every Bitcoin block", blockHeader.Height)
		}
		return errors.Wrap(err, "failed to get indexed block by height")
	}
	cumulativeEventHash := chainhash.DoubleHashH(append(prevIndexedBlock.CumulativeEventHash[:], eventHash[:]...))

	if err := runesDgTx.CreateIndexedBlock(ctx, &entity.IndexedBlock{
		Height:              blockHeader.Height,
		Hash:                blockHeader.Hash,
		PrevHash:            blockHeader.PrevBlock,
		EventHash:           eventHash,
		CumulativeEventHash: cumulativeEventHash,
	}); err != nil {
		return errors.Wrap(err, "failed to create indexed block")
	}
	// flush new rune entries
	newRuneEntries := lo.Values(p.newRuneEntries)
	if err := runesDgTx.CreateRuneEntries(ctx, newRuneEntries, uint64(blockHeader.Height)); err != nil {
		return errors.Wrap(err, "failed to create rune entry")
	}
	p.newRuneEntries = make(map[runes.RuneId]*runes.RuneEntry)

	// flush new rune entry states
	newRuneEntryStates := lo.Values(p.newRuneEntryStates)
	if err := runesDgTx.CreateRuneEntryStates(ctx, newRuneEntryStates, uint64(blockHeader.Height)); err != nil {
		return errors.Wrap(err, "failed to create rune entry state")
	}
	p.newRuneEntryStates = make(map[runes.RuneId]*runes.RuneEntry)

	// flush new outpoint balances
	newBalances := make([]*entity.OutPointBalance, 0)
	for _, balances := range p.newOutPointBalances {
		newBalances = append(newBalances, balances...)
	}
	if err := runesDgTx.CreateOutPointBalances(ctx, newBalances); err != nil {
		return errors.Wrap(err, "failed to create outpoint balances")
	}
	p.newOutPointBalances = make(map[wire.OutPoint][]*entity.OutPointBalance)

	// flush new spend outpoints
	newSpendOutPoints := p.newSpendOutPoints
	if err := runesDgTx.SpendOutPointBalancesBatch(ctx, newSpendOutPoints, uint64(blockHeader.Height)); err != nil {
		return errors.Wrap(err, "failed to create spend outpoint")
	}
	p.newSpendOutPoints = make([]wire.OutPoint, 0)

	// flush new balances
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
				BlockHeight: uint64(blockHeader.Height),
			})
		}
	}
	if err := runesDgTx.CreateRuneBalances(ctx, params); err != nil {
		return errors.Wrap(err, "failed to create balances at block")
	}
	p.newBalances = make(map[string]map[runes.RuneId]uint128.Uint128)

	// flush new rune transactions
	newRuneTxs := p.newRuneTxs
	if err := runesDgTx.CreateRuneTransactions(ctx, newRuneTxs); err != nil {
		return errors.Wrap(err, "failed to create rune transaction")
	}
	p.newRuneTxs = make([]*entity.RuneTransaction, 0)

	if err := runesDgTx.Commit(ctx); err != nil {
		return errors.Wrap(err, "failed to commit runes tx")
	}
	timeTaken := time.Since(start)
	logger.InfoContext(ctx, "Flushed block",
		slogx.String("event", "runes_processor_flushed_block"),
		slog.Int64("height", blockHeader.Height),
		slog.String("hash", blockHeader.Hash.String()),
		slog.String("event_hash", hex.EncodeToString(eventHash[:])),
		slog.String("cumulative_event_hash", hex.EncodeToString(cumulativeEventHash[:])),
		slog.Int("new_rune_entries", len(newRuneEntries)),
		slog.Int("new_rune_entry_states", len(newRuneEntryStates)),
		slog.Int("new_outpoint_balances", len(newBalances)),
		slog.Int("new_spend_outpoints", len(newSpendOutPoints)),
		slog.Int("new_balances", len(params)),
		slog.Int("new_rune_txs", len(newRuneTxs)),
		slogx.Duration("time_taken", timeTaken),
	)

	// submit event to reporting system
	if p.reportingClient != nil {
		if err := p.reportingClient.SubmitBlockReport(ctx, reportingclient.SubmitBlockReportPayload{
			Type:                "runes",
			ClientVersion:       constants.Version,
			DBVersion:           constants.DBVersion,
			EventHashVersion:    constants.EventHashVersion,
			Network:             p.network,
			BlockHeight:         uint64(blockHeader.Height),
			BlockHash:           blockHeader.Hash,
			EventHash:           eventHash,
			CumulativeEventHash: cumulativeEventHash,
		}); err != nil {
			return errors.Wrap(err, "failed to submit block report")
		}
	}
	return nil
}
