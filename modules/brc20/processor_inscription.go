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

func (p *Processor) processInscriptionTx(ctx context.Context, tx *types.Transaction, blockHeader types.BlockHeader) error {
	ctx = logger.WithContext(ctx, slogx.String("tx_hash", tx.TxHash.String()))
	envelopes := ordinals.ParseEnvelopesFromTx(tx)
	inputOutPoints := lo.Map(tx.TxIn, func(txIn *types.TxIn, _ int) wire.OutPoint {
		return wire.OutPoint{
			Hash:  txIn.PreviousOutTxHash,
			Index: txIn.PreviousOutIndex,
		}
	})
	inscriptionIdsInOutPoints, err := p.getInscriptionIdsInOutPoints(ctx, inputOutPoints)
	if err != nil {
		return errors.Wrap(err, "failed to get inscriptions in outpoints")
	}
	// cache outpoint values for future blocks
	for outIndex, txOut := range tx.TxOut {
		p.outPointValueCache.Add(wire.OutPoint{
			Hash:  tx.TxHash,
			Index: uint32(outIndex),
		}, uint64(txOut.Value))
	}
	if len(envelopes) == 0 && len(inscriptionIdsInOutPoints) == 0 {
		// no inscription activity, skip
		return nil
	}
	logger.DebugContext(ctx, "Processing new tx",
		slogx.String("tx_hash", tx.TxHash.String()),
		slogx.Uint32("tx_index", tx.Index),
	)

	floatingInscriptions := make([]*Flotsam, 0)
	totalInputValue := uint64(0)
	totalOutputValue := lo.SumBy(tx.TxOut, func(txOut *types.TxOut) uint64 { return uint64(txOut.Value) })
	inscribeOffsets := make(map[uint64]*struct {
		inscriptionId ordinals.InscriptionId
		count         int
	})
	idCounter := uint32(0)
	inputValues, err := p.getOutPointValues(ctx, inputOutPoints)
	if err != nil {
		return errors.Wrap(err, "failed to get outpoint values")
	}
	for i, input := range tx.TxIn {
		inputOutPoint := wire.OutPoint{
			Hash:  input.PreviousOutTxHash,
			Index: input.PreviousOutIndex,
		}
		inputValue := inputValues[inputOutPoint]
		// skip coinbase inputs since there can't be an inscription in coinbase
		if input.PreviousOutTxHash.IsEqual(&chainhash.Hash{}) {
			totalInputValue += p.getBlockSubsidy(uint64(tx.BlockHeight))
			continue
		}

		inscriptionIdsInOutPoint := inscriptionIdsInOutPoints[inputOutPoint]
		for satPoint, inscriptionIds := range inscriptionIdsInOutPoint {
			offset := totalInputValue + satPoint.Offset
			for _, inscriptionId := range inscriptionIds {
				floatingInscriptions = append(floatingInscriptions, &Flotsam{
					Offset:        offset,
					InscriptionId: inscriptionId,
					Tx:            tx,
					OriginOld: &OriginOld{
						OldSatPoint: satPoint,
					},
				})
				if _, ok := inscribeOffsets[offset]; !ok {
					inscribeOffsets[offset] = &struct {
						inscriptionId ordinals.InscriptionId
						count         int
					}{inscriptionId, 0}
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
					initialInscriptionEntry, err := p.brc20Dg.GetInscriptionEntryById(ctx, initial.inscriptionId)
					if err != nil {
						return errors.Wrap(err, "failed to get inscription entry")
					}
					if !initialInscriptionEntry.Cursed {
						cursed = true // reinscription curse if initial inscription is not cursed
					}
					if initialInscriptionEntry.CursedForBRC20 {
						cursedForBRC20 = true
					}
				}
			}
			// inscriptions are no longer cursed after jubilee, but BRC20 still considers them as cursed
			if cursed && uint64(tx.BlockHeight) > ordinals.GetJubileeHeight(p.network) {
				cursed = false
			}

			unbound := inputValue == 0 || envelope.UnrecognizedEvenField
			if envelope.Inscription.Pointer != nil && *envelope.Inscription.Pointer < totalOutputValue {
				offset = *envelope.Inscription.Pointer
			}

			floatingInscriptions = append(floatingInscriptions, &Flotsam{
				Offset:        offset,
				InscriptionId: inscriptionId,
				Tx:            tx,
				OriginNew: &OriginNew{
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
	isCoinbase := tx.TxIn[0].PreviousOutTxHash.IsEqual(&chainhash.Hash{})
	if isCoinbase {
		floatingInscriptions = append(floatingInscriptions, p.flotsamsSentAsFee...)
	}
	slices.SortFunc(floatingInscriptions, func(i, j *Flotsam) int {
		return int(i.Offset) - int(j.Offset)
	})

	outputValue := uint64(0)
	curIncrIdx := 0
	// newLocations := make(map[ordinals.SatPoint][]*Flotsam)
	type location struct {
		satPoint  ordinals.SatPoint
		flotsam   *Flotsam
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
		if err := p.updateInscriptionLocation(ctx, satPoint, flotsam, sentAsFee, tx, blockHeader); err != nil {
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
			if err := p.updateInscriptionLocation(ctx, newSatPoint, flotsam, false, tx, blockHeader); err != nil {
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

func (p *Processor) updateInscriptionLocation(ctx context.Context, newSatPoint ordinals.SatPoint, flotsam *Flotsam, sentAsFee bool, tx *types.Transaction, blockHeader types.BlockHeader) error {
	txOut := tx.TxOut[newSatPoint.OutPoint.Index]
	if flotsam.OriginOld != nil {
		transfer := &entity.InscriptionTransfer{
			InscriptionId:  flotsam.InscriptionId,
			BlockHeight:    uint64(tx.BlockHeight),
			OldSatPoint:    flotsam.OriginOld.OldSatPoint,
			NewSatPoint:    newSatPoint,
			NewPkScript:    txOut.PkScript,
			NewOutputValue: uint64(txOut.Value),
			SentAsFee:      sentAsFee,
		}
		entry, err := p.getInscriptionEntryById(ctx, flotsam.InscriptionId)
		if err != nil {
			// skip inscriptions without entry (likely non-brc20 inscriptions)
			if errors.Is(err, errs.NotFound) {
				return nil
			}
			return errors.Wrap(err, "failed to get inscription entry")
		}
		entry.TransferCount++

		// dont track transfers that exceed limit
		if entry.TransferCount <= p.transferCountLimit {
			p.newInscriptionTransfers = append(p.newInscriptionTransfers, transfer)
			p.newInscriptionEntryStates[entry.Id] = entry
		}
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
			BlockHeight:    uint64(tx.BlockHeight),
			OldSatPoint:    ordinals.SatPoint{},
			NewSatPoint:    newSatPoint,
			NewPkScript:    txOut.PkScript,
			NewOutputValue: uint64(txOut.Value),
			SentAsFee:      sentAsFee,
		}
		entry := &ordinals.InscriptionEntry{
			Id:              flotsam.InscriptionId,
			Number:          inscriptionNumber,
			SequenceNumber:  sequenceNumber,
			Cursed:          origin.Cursed,
			CursedForBRC20:  origin.CursedForBRC20,
			CreatedAt:       blockHeader.Timestamp,
			CreatedAtHeight: uint64(tx.BlockHeight),
			Inscription:     origin.Inscription,
			TransferCount:   1, // count inscription as first transfer
		}
		p.newInscriptionTransfers = append(p.newInscriptionTransfers, transfer)
		p.newInscriptionEntries[entry.Id] = entry
		p.newInscriptionEntryStates[entry.Id] = entry

		return nil
	}
	panic("unreachable")
}

type brc20Inscription struct {
	P string `json:"p"`
}

func isBRC20Inscription(inscription ordinals.Inscription) bool {
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
				return errors.Wrap(err, "failed to get transaction outputs")
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
		if result[outPoints[i]] == 0 {
			result[outPoints[i]] = uint64(txOutsByHash[outPoints[i].Hash][outPoints[i].Index].Value)
		}
	}
	return result, nil
}

func (p *Processor) getInscriptionIdsInOutPoints(ctx context.Context, outPoints []wire.OutPoint) (map[wire.OutPoint]map[ordinals.SatPoint][]ordinals.InscriptionId, error) {
	inscriptionIds, err := p.brc20Dg.GetInscriptionIdsInOutPoints(ctx, outPoints)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get inscriptions by outpoint")
	}
	result := make(map[wire.OutPoint]map[ordinals.SatPoint][]ordinals.InscriptionId)
	for satPoint, inscriptionIds := range inscriptionIds {
		if _, ok := result[satPoint.OutPoint]; !ok {
			result[satPoint.OutPoint] = make(map[ordinals.SatPoint][]ordinals.InscriptionId)
		}
		result[satPoint.OutPoint][satPoint] = append(result[satPoint.OutPoint][satPoint], inscriptionIds...)
	}
	return result, nil
}

func (p *Processor) getInscriptionEntryById(ctx context.Context, inscriptionId ordinals.InscriptionId) (*ordinals.InscriptionEntry, error) {
	if inscriptionEntry, ok := p.newInscriptionEntryStates[inscriptionId]; ok {
		return inscriptionEntry, nil
	}

	inscription, err := p.brc20Dg.GetInscriptionEntryById(ctx, inscriptionId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get inscriptions by outpoint")
	}
	return inscription, nil
}

func (p *Processor) getBlockSubsidy(blockHeight uint64) uint64 {
	return uint64(blockchain.CalcBlockSubsidy(int32(blockHeight), p.network.ChainParams()))
}
