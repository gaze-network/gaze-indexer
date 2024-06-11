package brc20

import (
	"context"
	"encoding/json"
	"slices"
	"sync"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/ordinals"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

func (p *Processor) processInscriptionTx(ctx context.Context, tx *types.Transaction, blockHeader types.BlockHeader, transfersInOutPoints map[wire.OutPoint]map[ordinals.SatPoint][]*entity.InscriptionTransfer, outpointValues map[wire.OutPoint]uint64) error {
	ctx = logger.WithContext(ctx, slogx.String("tx_hash", tx.TxHash.String()))
	envelopes := ordinals.ParseEnvelopesFromTx(tx)
	inputOutPoints := lo.Map(tx.TxIn, func(txIn *types.TxIn, _ int) wire.OutPoint {
		return wire.OutPoint{
			Hash:  txIn.PreviousOutTxHash,
			Index: txIn.PreviousOutIndex,
		}
	})
	// cache outpoint values for future blocks
	for outIndex, txOut := range tx.TxOut {
		outPoint := wire.OutPoint{
			Hash:  tx.TxHash,
			Index: uint32(outIndex),
		}
		p.outPointValueCache.Add(outPoint, uint64(txOut.Value))
		outpointValues[outPoint] = uint64(txOut.Value)
	}
	outPointsWithTransfers := lo.Keys(transfersInOutPoints)
	txContainsTransfers := len(lo.Intersect(inputOutPoints, outPointsWithTransfers)) > 0
	isCoinbase := tx.TxIn[0].PreviousOutTxHash.IsEqual(&chainhash.Hash{})
	if len(envelopes) == 0 && !txContainsTransfers && !isCoinbase {
		// no inscription activity, skip
		return nil
	}

	if err := p.ensureOutPointValues(ctx, outpointValues, inputOutPoints); err != nil {
		return errors.Wrap(err, "failed to ensure outpoint values")
	}

	floatingInscriptions := make([]*entity.Flotsam, 0)
	totalInputValue := uint64(0)
	totalOutputValue := lo.SumBy(tx.TxOut, func(txOut *types.TxOut) uint64 { return uint64(txOut.Value) })
	inscribeOffsets := make(map[uint64]*struct {
		inscriptionId ordinals.InscriptionId
		count         int
	})
	idCounter := uint32(0)
	for i, input := range tx.TxIn {
		// skip coinbase inputs since there can't be an inscription in coinbase
		if input.PreviousOutTxHash.IsEqual(&chainhash.Hash{}) {
			totalInputValue += p.getBlockSubsidy(uint64(tx.BlockHeight))
			continue
		}
		inputOutPoint := wire.OutPoint{
			Hash:  input.PreviousOutTxHash,
			Index: input.PreviousOutIndex,
		}
		inputValue, ok := outpointValues[inputOutPoint]
		if !ok {
			return errors.Wrapf(errs.NotFound, "outpoint value not found for %s", inputOutPoint.String())
		}

		transfersInOutPoint := transfersInOutPoints[inputOutPoint]
		for satPoint, transfers := range transfersInOutPoint {
			offset := totalInputValue + satPoint.Offset
			for _, transfer := range transfers {
				floatingInscriptions = append(floatingInscriptions, &entity.Flotsam{
					Offset:        offset,
					InscriptionId: transfer.InscriptionId,
					Tx:            tx,
					OriginOld: &entity.OriginOld{
						OldSatPoint: satPoint,
						Content:     transfer.Content,
						InputIndex:  uint32(i),
					},
				})
				if _, ok := inscribeOffsets[offset]; !ok {
					inscribeOffsets[offset] = &struct {
						inscriptionId ordinals.InscriptionId
						count         int
					}{transfer.InscriptionId, 0}
				}
				inscribeOffsets[offset].count++
			}
		}
		// offset on output to inscribe new inscriptions from this input
		offset := totalInputValue
		totalInputValue += inputValue

		envelopesInInput := lo.Filter(envelopes, func(envelope *ordinals.Envelope, _ int) bool {
			return envelope.InputIndex == uint32(i)
		})
		for _, envelope := range envelopesInInput {
			inscriptionId := ordinals.InscriptionId{
				TxHash: tx.TxHash,
				Index:  idCounter,
			}
			var cursed, cursedForBRC20 bool
			if envelope.UnrecognizedEvenField || // unrecognized even field
				envelope.DuplicateField || // duplicate field
				envelope.IncompleteField || // incomplete field
				envelope.InputIndex != 0 || // not first input
				envelope.Offset != 0 || // not first envelope in input
				envelope.Inscription.Pointer != nil || // contains pointer
				envelope.PushNum || // contains pushnum opcodes
				envelope.Stutter { // contains stuttering curse structure
				cursed = true
				cursedForBRC20 = true
			}
			if initial, ok := inscribeOffsets[offset]; !cursed && ok {
				if initial.count > 1 {
					cursed = true // reinscription
					cursedForBRC20 = true
				} else {
					initialInscriptionEntry, err := p.getInscriptionEntryById(ctx, initial.inscriptionId)
					if err != nil {
						return errors.Wrapf(err, "failed to get inscription entry id %s", initial.inscriptionId)
					}
					if !initialInscriptionEntry.Cursed {
						cursed = true // reinscription curse if initial inscription is not cursed
					}
					if !initialInscriptionEntry.CursedForBRC20 {
						cursedForBRC20 = true
					}
				}
			}
			// inscriptions are no longer cursed after jubilee, but BRC20 still considers them as cursed
			if cursed && uint64(tx.BlockHeight) >= ordinals.GetJubileeHeight(p.network) {
				cursed = false
			}

			unbound := inputValue == 0 || envelope.UnrecognizedEvenField
			if envelope.Inscription.Pointer != nil && *envelope.Inscription.Pointer < totalOutputValue {
				offset = *envelope.Inscription.Pointer
			}

			floatingInscriptions = append(floatingInscriptions, &entity.Flotsam{
				Offset:        offset,
				InscriptionId: inscriptionId,
				Tx:            tx,
				OriginNew: &entity.OriginNew{
					Reinscription:  inscribeOffsets[offset] != nil,
					Cursed:         cursed,
					CursedForBRC20: cursedForBRC20,
					Fee:            0,
					Hidden:         false, // we don't care about this field for brc20
					Parent:         envelope.Inscription.Parent,
					Pointer:        envelope.Inscription.Pointer,
					Unbound:        unbound,
					Inscription:    envelope.Inscription,
				},
			})

			if _, ok := inscribeOffsets[offset]; !ok {
				inscribeOffsets[offset] = &struct {
					inscriptionId ordinals.InscriptionId
					count         int
				}{inscriptionId, 0}
			}
			inscribeOffsets[offset].count++
			idCounter++
		}
	}

	// parents must exist in floatingInscriptions to be valid
	potentialParents := make(map[ordinals.InscriptionId]struct{})
	for _, flotsam := range floatingInscriptions {
		potentialParents[flotsam.InscriptionId] = struct{}{}
	}
	for _, flotsam := range floatingInscriptions {
		if flotsam.OriginNew != nil && flotsam.OriginNew.Parent != nil {
			if _, ok := potentialParents[*flotsam.OriginNew.Parent]; !ok {
				// parent not found, ignore parent
				flotsam.OriginNew.Parent = nil
			}
		}
	}

	// calculate fee for each new inscription
	for _, flotsam := range floatingInscriptions {
		if flotsam.OriginNew != nil {
			flotsam.OriginNew.Fee = (totalInputValue - totalOutputValue) / uint64(idCounter)
		}
	}

	// if tx is coinbase, add inscriptions sent as fee to outputs of this tx
	ownInscriptionCount := len(floatingInscriptions)
	if isCoinbase {
		floatingInscriptions = append(floatingInscriptions, p.flotsamsSentAsFee...)
	}
	// sort floatingInscriptions by offset
	slices.SortFunc(floatingInscriptions, func(i, j *entity.Flotsam) int {
		return int(i.Offset) - int(j.Offset)
	})

	outputValue := uint64(0)
	curIncrIdx := 0
	// newLocations := make(map[ordinals.SatPoint][]*Flotsam)
	type location struct {
		satPoint  ordinals.SatPoint
		flotsam   *entity.Flotsam
		sentAsFee bool
	}
	newLocations := make([]*location, 0)
	outputToSumValue := make([]uint64, 0, len(tx.TxOut))
	for outIndex, txOut := range tx.TxOut {
		end := outputValue + uint64(txOut.Value)

		// process all inscriptions that are supposed to be inscribed in this output
		for curIncrIdx < len(floatingInscriptions) && floatingInscriptions[curIncrIdx].Offset < end {
			newSatPoint := ordinals.SatPoint{
				OutPoint: wire.OutPoint{
					Hash:  tx.TxHash,
					Index: uint32(outIndex),
				},
				Offset: floatingInscriptions[curIncrIdx].Offset - outputValue,
			}
			// newLocations[newSatPoint] = append(newLocations[newSatPoint], floatingInscriptions[curIncrIdx])
			newLocations = append(newLocations, &location{
				satPoint:  newSatPoint,
				flotsam:   floatingInscriptions[curIncrIdx],
				sentAsFee: isCoinbase && curIncrIdx >= ownInscriptionCount, // if curIncrIdx >= ownInscriptionCount, then current inscription came from p.flotSamsSentAsFee
			})
			curIncrIdx++
		}

		outputValue = end
		outputToSumValue = append(outputToSumValue, outputValue)
	}

	for _, loc := range newLocations {
		satPoint := loc.satPoint
		flotsam := loc.flotsam
		sentAsFee := loc.sentAsFee
		// TODO: not sure if we still need to handle pointer here, it's already handled above.
		if flotsam.OriginNew != nil && flotsam.OriginNew.Pointer != nil {
			pointer := *flotsam.OriginNew.Pointer
			for outIndex, outputValue := range outputToSumValue {
				start := uint64(0)
				if outIndex > 0 {
					start = outputToSumValue[outIndex-1]
				}
				end := outputValue
				if start <= pointer && pointer < end {
					satPoint.Offset = pointer - start
					break
				}
			}
		}
		if err := p.updateInscriptionLocation(ctx, satPoint, flotsam, sentAsFee, tx, blockHeader, transfersInOutPoints); err != nil {
			return errors.Wrap(err, "failed to update inscription location")
		}
	}

	// handle leftover flotsams (flotsams with offset over total output value) )
	if isCoinbase {
		// if there are leftover inscriptions in coinbase, they are lost permanently
		for _, flotsam := range floatingInscriptions[curIncrIdx:] {
			newSatPoint := ordinals.SatPoint{
				OutPoint: wire.OutPoint{},
				Offset:   p.lostSats + flotsam.Offset - totalOutputValue,
			}
			if err := p.updateInscriptionLocation(ctx, newSatPoint, flotsam, false, tx, blockHeader, transfersInOutPoints); err != nil {
				return errors.Wrap(err, "failed to update inscription location")
			}
		}
		p.lostSats += p.blockReward - totalOutputValue
	} else {
		// if there are leftover inscriptions in non-coinbase tx, they are stored in p.flotsamsSentAsFee for processing in this block's coinbase tx
		for _, flotsam := range floatingInscriptions[curIncrIdx:] {
			flotsam.Offset = p.blockReward + flotsam.Offset - totalOutputValue
			p.flotsamsSentAsFee = append(p.flotsamsSentAsFee, flotsam)
		}
		// add fees to block reward
		p.blockReward = totalInputValue - totalOutputValue
	}
	return nil
}

func (p *Processor) updateInscriptionLocation(ctx context.Context, newSatPoint ordinals.SatPoint, flotsam *entity.Flotsam, sentAsFee bool, tx *types.Transaction, blockHeader types.BlockHeader, transfersInOutPoints map[wire.OutPoint]map[ordinals.SatPoint][]*entity.InscriptionTransfer) error {
	txOut := tx.TxOut[newSatPoint.OutPoint.Index]
	if flotsam.OriginOld != nil {
		entry, err := p.getInscriptionEntryById(ctx, flotsam.InscriptionId)
		if err != nil {
			return errors.Wrapf(err, "failed to get inscription entry id %s", flotsam.InscriptionId)
		}
		entry.TransferCount++
		transfer := &entity.InscriptionTransfer{
			InscriptionId:  flotsam.InscriptionId,
			BlockHeight:    uint64(flotsam.Tx.BlockHeight), // use flotsam's tx to track tx that initiated the transfer
			TxIndex:        flotsam.Tx.Index,               // use flotsam's tx to track tx that initiated the transfer
			TxHash:         flotsam.Tx.TxHash,
			Content:        flotsam.OriginOld.Content,
			FromInputIndex: flotsam.OriginOld.InputIndex,
			OldSatPoint:    flotsam.OriginOld.OldSatPoint,
			NewSatPoint:    newSatPoint,
			NewPkScript:    txOut.PkScript,
			NewOutputValue: uint64(txOut.Value),
			SentAsFee:      sentAsFee,
			TransferCount:  entry.TransferCount,
		}

		// track transfers even if transfer count exceeds 2 (because we need to check for reinscriptions)
		p.newInscriptionTransfers = append(p.newInscriptionTransfers, transfer)
		p.newInscriptionEntryStates[entry.Id] = entry

		// add new transfer to transfersInOutPoints cache
		if _, ok := transfersInOutPoints[newSatPoint.OutPoint]; !ok {
			transfersInOutPoints[newSatPoint.OutPoint] = make(map[ordinals.SatPoint][]*entity.InscriptionTransfer)
		}
		transfersInOutPoints[newSatPoint.OutPoint][newSatPoint] = append(transfersInOutPoints[newSatPoint.OutPoint][newSatPoint], transfer)
		return nil
	}

	if flotsam.OriginNew != nil {
		origin := flotsam.OriginNew
		var inscriptionNumber int64
		sequenceNumber := p.cursedInscriptionCount + p.blessedInscriptionCount
		if origin.Cursed {
			inscriptionNumber = -int64(p.cursedInscriptionCount + 1)
			p.cursedInscriptionCount++
		} else {
			inscriptionNumber = int64(p.blessedInscriptionCount)
			p.blessedInscriptionCount++
		}
		// if not valid brc20 inscription, delete content to save space
		if !isBRC20Inscription(origin.Inscription) {
			origin.Inscription.Content = nil
			origin.Inscription.ContentType = ""
			origin.Inscription.ContentEncoding = ""
		}
		transfer := &entity.InscriptionTransfer{
			InscriptionId:  flotsam.InscriptionId,
			BlockHeight:    uint64(flotsam.Tx.BlockHeight), // use flotsam's tx to track tx that initiated the transfer
			TxIndex:        flotsam.Tx.Index,               // use flotsam's tx to track tx that initiated the transfer
			TxHash:         flotsam.Tx.TxHash,
			Content:        origin.Inscription.Content,
			FromInputIndex: 0, // unused
			OldSatPoint:    ordinals.SatPoint{},
			NewSatPoint:    newSatPoint,
			NewPkScript:    txOut.PkScript,
			NewOutputValue: uint64(txOut.Value),
			SentAsFee:      sentAsFee,
			TransferCount:  1, // count inscription as first transfer
		}
		entry := &ordinals.InscriptionEntry{
			Id:              flotsam.InscriptionId,
			Number:          inscriptionNumber,
			SequenceNumber:  sequenceNumber,
			Cursed:          origin.Cursed,
			CursedForBRC20:  origin.CursedForBRC20,
			CreatedAt:       blockHeader.Timestamp,
			CreatedAtHeight: uint64(blockHeader.Height),
			Inscription:     origin.Inscription,
			TransferCount:   1, // count inscription as first transfer
		}
		p.newInscriptionTransfers = append(p.newInscriptionTransfers, transfer)
		p.newInscriptionEntries[entry.Id] = entry
		p.newInscriptionEntryStates[entry.Id] = entry

		// add new transfer to transfersInOutPoints cache
		if _, ok := transfersInOutPoints[newSatPoint.OutPoint]; !ok {
			transfersInOutPoints[newSatPoint.OutPoint] = make(map[ordinals.SatPoint][]*entity.InscriptionTransfer)
		}
		transfersInOutPoints[newSatPoint.OutPoint][newSatPoint] = append(transfersInOutPoints[newSatPoint.OutPoint][newSatPoint], transfer)
		return nil
	}
	panic("unreachable")
}

func (p *Processor) ensureOutPointValues(ctx context.Context, outPointValues map[wire.OutPoint]uint64, outPoints []wire.OutPoint) error {
	missingOutPoints := make([]wire.OutPoint, 0)
	for _, outPoint := range outPoints {
		if _, ok := outPointValues[outPoint]; !ok {
			missingOutPoints = append(missingOutPoints, outPoint)
		}
	}
	if len(missingOutPoints) == 0 {
		return nil
	}
	missingOutPointValues, err := p.getOutPointValues(ctx, missingOutPoints)
	if err != nil {
		return errors.Wrap(err, "failed to get outpoint values")
	}
	for outPoint, value := range missingOutPointValues {
		outPointValues[outPoint] = value
	}
	return nil
}

type brc20Inscription struct {
	P string `json:"p"`
}

func isBRC20Inscription(inscription ordinals.Inscription) bool {
	if inscription.ContentType != "application/json" && inscription.ContentType != "text/plain" {
		return false
	}

	// attempt to parse content as json
	if inscription.Content == nil {
		return false
	}
	var parsed brc20Inscription
	if err := json.Unmarshal(inscription.Content, &parsed); err != nil {
		return false
	}
	if parsed.P != "brc-20" {
		return false
	}
	return true
}

func (p *Processor) getOutPointValues(ctx context.Context, outPoints []wire.OutPoint) (map[wire.OutPoint]uint64, error) {
	// try to get from cache if exists
	cacheValues := p.outPointValueCache.MGet(outPoints)
	result := make(map[wire.OutPoint]uint64)

	outPointsToFetch := make([]wire.OutPoint, 0)
	for i, outPoint := range outPoints {
		if outPoint.Hash == (chainhash.Hash{}) {
			// skip coinbase input
			continue
		}
		if cacheValues[i] != 0 {
			result[outPoint] = cacheValues[i]
		} else {
			outPointsToFetch = append(outPointsToFetch, outPoint)
		}
	}
	eg, ectx := errgroup.WithContext(ctx)
	txHashes := make(map[chainhash.Hash]struct{})
	for _, outPoint := range outPointsToFetch {
		txHashes[outPoint.Hash] = struct{}{}
	}
	txOutsByHash := make(map[chainhash.Hash][]*types.TxOut)
	var mutex sync.Mutex
	for txHash := range txHashes {
		txHash := txHash
		eg.Go(func() error {
			txOuts, err := p.btcClient.GetTransactionOutputs(ectx, txHash)
			if err != nil {
				return errors.Wrapf(err, "failed to get transaction outputs for hash %s", txHash)
			}

			// update cache
			mutex.Lock()
			defer mutex.Unlock()

			txOutsByHash[txHash] = txOuts
			for i, txOut := range txOuts {
				p.outPointValueCache.Add(wire.OutPoint{Hash: txHash, Index: uint32(i)}, uint64(txOut.Value))
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, errors.WithStack(err)
	}
	for i := range outPoints {
		if outPoints[i].Hash == (chainhash.Hash{}) {
			// skip coinbase input
			continue
		}
		if result[outPoints[i]] == 0 {
			result[outPoints[i]] = uint64(txOutsByHash[outPoints[i].Hash][outPoints[i].Index].Value)
		}
	}
	return result, nil
}

func (p *Processor) getInscriptionTransfersInOutPoints(ctx context.Context, outPoints []wire.OutPoint) (map[wire.OutPoint]map[ordinals.SatPoint][]*entity.InscriptionTransfer, error) {
	outPoints = lo.Uniq(outPoints)
	// try to get from flush buffer if exists
	result := make(map[wire.OutPoint]map[ordinals.SatPoint][]*entity.InscriptionTransfer)

	outPointsToFetch := make([]wire.OutPoint, 0)
	for _, outPoint := range outPoints {
		var found bool
		for _, transfer := range p.newInscriptionTransfers {
			if transfer.NewSatPoint.OutPoint == outPoint {
				found = true
				if _, ok := result[outPoint]; !ok {
					result[outPoint] = make(map[ordinals.SatPoint][]*entity.InscriptionTransfer)
				}
				result[outPoint][transfer.NewSatPoint] = append(result[outPoint][transfer.NewSatPoint], transfer)
			}
		}
		if !found {
			outPointsToFetch = append(outPointsToFetch, outPoint)
		}
	}

	transfers, err := p.brc20Dg.GetInscriptionTransfersInOutPoints(ctx, outPointsToFetch)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get inscriptions by outpoint")
	}

	for satPoint, transferList := range transfers {
		if _, ok := result[satPoint.OutPoint]; !ok {
			result[satPoint.OutPoint] = make(map[ordinals.SatPoint][]*entity.InscriptionTransfer)
		}
		result[satPoint.OutPoint][satPoint] = append(result[satPoint.OutPoint][satPoint], transferList...)
	}
	return result, nil
}

func (p *Processor) getInscriptionEntryById(ctx context.Context, id ordinals.InscriptionId) (*ordinals.InscriptionEntry, error) {
	inscriptions, err := p.getInscriptionEntriesByIds(ctx, []ordinals.InscriptionId{id})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get inscriptions by outpoint")
	}
	inscription, ok := inscriptions[id]
	if !ok {
		return nil, errors.Wrap(errs.NotFound, "inscription not found")
	}
	return inscription, nil
}

func (p *Processor) getInscriptionEntriesByIds(ctx context.Context, ids []ordinals.InscriptionId) (map[ordinals.InscriptionId]*ordinals.InscriptionEntry, error) {
	// try to get from cache if exists
	result := make(map[ordinals.InscriptionId]*ordinals.InscriptionEntry)

	idsToFetch := make([]ordinals.InscriptionId, 0)
	for _, id := range ids {
		if inscriptionEntry, ok := p.newInscriptionEntryStates[id]; ok {
			result[id] = inscriptionEntry
		} else {
			idsToFetch = append(idsToFetch, id)
		}
	}

	if len(idsToFetch) > 0 {
		inscriptions, err := p.brc20Dg.GetInscriptionEntriesByIds(ctx, idsToFetch)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get inscriptions by outpoint")
		}
		for id, inscription := range inscriptions {
			result[id] = inscription
		}
	}
	return result, nil
}

func (p *Processor) getInscriptionNumbersByIds(ctx context.Context, ids []ordinals.InscriptionId) (map[ordinals.InscriptionId]int64, error) {
	// try to get from cache if exists
	result := make(map[ordinals.InscriptionId]int64)

	idsToFetch := make([]ordinals.InscriptionId, 0)
	for _, id := range ids {
		if entry, ok := p.newInscriptionEntryStates[id]; ok {
			result[id] = int64(entry.Number)
		} else {
			idsToFetch = append(idsToFetch, id)
		}
	}

	if len(idsToFetch) > 0 {
		inscriptions, err := p.brc20Dg.GetInscriptionNumbersByIds(ctx, idsToFetch)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get inscriptions by outpoint")
		}
		for id, number := range inscriptions {
			result[id] = number
		}
	}
	return result, nil
}

func (p *Processor) getInscriptionParentsByIds(ctx context.Context, ids []ordinals.InscriptionId) (map[ordinals.InscriptionId]ordinals.InscriptionId, error) {
	// try to get from cache if exists
	result := make(map[ordinals.InscriptionId]ordinals.InscriptionId)

	idsToFetch := make([]ordinals.InscriptionId, 0)
	for _, id := range ids {
		if entry, ok := p.newInscriptionEntryStates[id]; ok {
			if entry.Inscription.Parent != nil {
				result[id] = *entry.Inscription.Parent
			}
		} else {
			idsToFetch = append(idsToFetch, id)
		}
	}

	if len(idsToFetch) > 0 {
		inscriptions, err := p.brc20Dg.GetInscriptionParentsByIds(ctx, idsToFetch)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get inscriptions by outpoint")
		}
		for id, parent := range inscriptions {
			result[id] = parent
		}
	}
	return result, nil
}

func (p *Processor) getBlockSubsidy(blockHeight uint64) uint64 {
	return uint64(blockchain.CalcBlockSubsidy(int32(blockHeight), p.network.ChainParams()))
}
