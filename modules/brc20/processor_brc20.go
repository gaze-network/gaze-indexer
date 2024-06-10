package brc20

import (
	"bytes"
	"context"
	"encoding/hex"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/brc20"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/datagateway"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/ordinals"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

func (p *Processor) processBRC20States(ctx context.Context, transfers []*entity.InscriptionTransfer, blockHeader types.BlockHeader) error {
	payloads := make([]*brc20.Payload, 0)
	ticks := make(map[string]struct{})
	for _, transfer := range transfers {
		if transfer.Content == nil {
			// skip empty content
			continue
		}
		payload, err := brc20.ParsePayload(transfer)
		if err != nil {
			// skip invalid payloads
			continue
		}
		payloads = append(payloads, payload)
		ticks[payload.Tick] = struct{}{}
	}
	if len(payloads) == 0 {
		// skip if no valid payloads
		return nil
	}
	// TODO: concurrently fetch from db to optimize speed
	tickEntries, err := p.brc20Dg.GetTickEntriesByTicks(ctx, lo.Keys(ticks))
	if err != nil {
		return errors.Wrap(err, "failed to get inscription entries by ids")
	}

	// preload required data to reduce individual data fetching during process
	inscriptionIds := make([]ordinals.InscriptionId, 0)
	inscriptionIdsToFetchParent := make([]ordinals.InscriptionId, 0)
	inscriptionIdsToFetchEventInscribeTransfer := make([]ordinals.InscriptionId, 0)
	balancesToFetch := make([]datagateway.GetBalancesBatchAtHeightQuery, 0) // pkscript -> tick -> struct{}
	for _, payload := range payloads {
		inscriptionIds = append(inscriptionIds, payload.Transfer.InscriptionId)
		if payload.Op == brc20.OperationMint {
			// preload parent id to validate mint events with self mint
			if entry := tickEntries[payload.Tick]; entry.IsSelfMint {
				inscriptionIdsToFetchParent = append(inscriptionIdsToFetchParent, payload.Transfer.InscriptionId)
			}
		}
		if payload.Op == brc20.OperationTransfer {
			if payload.Transfer.OldSatPoint == (ordinals.SatPoint{}) {
				// preload balance to validate inscribe transfer event
				balancesToFetch = append(balancesToFetch, datagateway.GetBalancesBatchAtHeightQuery{
					PkScriptHex: hex.EncodeToString(payload.Transfer.NewPkScript),
					Tick:        payload.Tick,
				})
			} else {
				// preload inscribe-transfer events to validate transfer-transfer event
				inscriptionIdsToFetchEventInscribeTransfer = append(inscriptionIdsToFetchEventInscribeTransfer, payload.Transfer.InscriptionId)
			}
		}
	}
	inscriptionIdsToNumber, err := p.getInscriptionNumbersByIds(ctx, lo.Uniq(inscriptionIds))
	if err != nil {
		return errors.Wrap(err, "failed to get inscription numbers by ids")
	}
	inscriptionIdsToParent, err := p.getInscriptionParentsByIds(ctx, lo.Uniq(inscriptionIdsToFetchParent))
	if err != nil {
		return errors.Wrap(err, "failed to get inscription parents by ids")
	}
	latestEventId, err := p.brc20Dg.GetLatestEventId(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get latest event id")
	}
	// pkscript -> tick -> balance
	balances, err := p.brc20Dg.GetBalancesBatchAtHeight(ctx, uint64(blockHeader.Height-1), balancesToFetch)
	if err != nil {
		return errors.Wrap(err, "failed to get balances batch at height")
	}
	eventInscribeTransfers, err := p.brc20Dg.GetEventInscribeTransfersByInscriptionIds(ctx, lo.Uniq(inscriptionIdsToFetchEventInscribeTransfer))
	if err != nil {
		return errors.Wrap(err, "failed to get event inscribe transfers by inscription ids")
	}

	newTickEntries := make(map[string]*entity.TickEntry)
	newTickEntryStates := make(map[string]*entity.TickEntry)
	newEventDeploys := make([]*entity.EventDeploy, 0)
	newEventMints := make([]*entity.EventMint, 0)
	newEventInscribeTransfers := make([]*entity.EventInscribeTransfer, 0)
	newEventTransferTransfers := make([]*entity.EventTransferTransfer, 0)
	newBalances := make(map[string]map[string]*entity.Balance)
	var eventHashBuilder strings.Builder

	handleEventDeploy := func(payload *brc20.Payload, tickEntry *entity.TickEntry) {
		if payload.Transfer.TransferCount > 1 {
			// skip used deploy inscriptions
			return
		}
		if tickEntry != nil {
			// skip deploy inscriptions for duplicate ticks
			return
		}
		newEntry := &entity.TickEntry{
			Tick:                payload.Tick,
			OriginalTick:        payload.OriginalTick,
			TotalSupply:         payload.Max,
			Decimals:            payload.Dec,
			LimitPerMint:        payload.Lim,
			IsSelfMint:          payload.SelfMint,
			DeployInscriptionId: payload.Transfer.InscriptionId,
			DeployedAt:          blockHeader.Timestamp,
			DeployedAtHeight:    payload.Transfer.BlockHeight,
			MintedAmount:        decimal.Zero,
			BurnedAmount:        decimal.Zero,
			CompletedAt:         time.Time{},
			CompletedAtHeight:   0,
		}
		newTickEntries[payload.Tick] = newEntry
		newTickEntryStates[payload.Tick] = newEntry
		// update entries for other operations in same block
		tickEntries[payload.Tick] = newEntry

		event := &entity.EventDeploy{
			Id:                latestEventId + 1,
			InscriptionId:     payload.Transfer.InscriptionId,
			InscriptionNumber: inscriptionIdsToNumber[payload.Transfer.InscriptionId],
			Tick:              payload.Tick,
			OriginalTick:      payload.OriginalTick,
			TxHash:            payload.Transfer.TxHash,
			BlockHeight:       payload.Transfer.BlockHeight,
			TxIndex:           payload.Transfer.TxIndex,
			Timestamp:         blockHeader.Timestamp,
			PkScript:          payload.Transfer.NewPkScript,
			SatPoint:          payload.Transfer.NewSatPoint,
			TotalSupply:       payload.Max,
			Decimals:          payload.Dec,
			LimitPerMint:      payload.Lim,
			IsSelfMint:        payload.SelfMint,
		}
		newEventDeploys = append(newEventDeploys, event)
		latestEventId++

		eventHashBuilder.WriteString(getEventDeployString(event) + eventHashSeparator)
	}
	handleEventMint := func(payload *brc20.Payload, tickEntry *entity.TickEntry) {
		if payload.Transfer.TransferCount > 1 {
			// skip used mint inscriptions that are already used
			return
		}
		if tickEntry == nil {
			// skip mint inscriptions for non-existent ticks
			return
		}
		if -payload.Amt.Exponent() > int32(tickEntry.Decimals) {
			// skip mint inscriptions with decimals greater than allowed
			return
		}
		if tickEntry.MintedAmount.GreaterThanOrEqual(tickEntry.TotalSupply) {
			// skip mint inscriptions for ticks with completed mints
			return
		}
		if payload.Amt.GreaterThan(tickEntry.LimitPerMint) {
			// skip mint inscriptions with amount greater than limit per mint
			return
		}
		mintableAmount := tickEntry.TotalSupply.Sub(tickEntry.MintedAmount)
		if payload.Amt.GreaterThan(mintableAmount) {
			payload.Amt = mintableAmount
		}
		var parentId *ordinals.InscriptionId
		if tickEntry.IsSelfMint {
			parentIdValue, ok := inscriptionIdsToParent[payload.Transfer.InscriptionId]
			if !ok {
				// skip mint inscriptions for self mint ticks without parent inscription
				return
			}
			if parentIdValue != tickEntry.DeployInscriptionId {
				// skip mint inscriptions for self mint ticks with invalid parent inscription
				return
			}
			parentId = &parentIdValue
		}

		tickEntry.MintedAmount = tickEntry.MintedAmount.Add(payload.Amt)
		// mark as completed if this mint completes the total supply
		if tickEntry.MintedAmount.GreaterThanOrEqual(tickEntry.TotalSupply) {
			tickEntry.CompletedAt = blockHeader.Timestamp
			tickEntry.CompletedAtHeight = payload.Transfer.BlockHeight
		}

		newTickEntryStates[payload.Tick] = tickEntry
		event := &entity.EventMint{
			Id:                latestEventId + 1,
			InscriptionId:     payload.Transfer.InscriptionId,
			InscriptionNumber: inscriptionIdsToNumber[payload.Transfer.InscriptionId],
			Tick:              payload.Tick,
			OriginalTick:      payload.OriginalTick,
			TxHash:            payload.Transfer.TxHash,
			BlockHeight:       payload.Transfer.BlockHeight,
			TxIndex:           payload.Transfer.TxIndex,
			Timestamp:         blockHeader.Timestamp,
			PkScript:          payload.Transfer.NewPkScript,
			SatPoint:          payload.Transfer.NewSatPoint,
			Amount:            payload.Amt,
			ParentId:          parentId,
		}
		newEventMints = append(newEventMints, event)
		latestEventId++

		eventHashBuilder.WriteString(getEventMintString(event, tickEntry.Decimals) + eventHashSeparator)
	}
	handleEventInscribeTransfer := func(payload *brc20.Payload, tickEntry *entity.TickEntry) {
		// inscribe transfer event
		pkScriptHex := hex.EncodeToString(payload.Transfer.NewPkScript)
		balance, ok := balances[pkScriptHex][payload.Tick]
		if !ok {
			balance = &entity.Balance{
				PkScript:         payload.Transfer.NewPkScript,
				Tick:             payload.Tick,
				BlockHeight:      uint64(blockHeader.Height - 1),
				OverallBalance:   decimal.Zero, // defaults balance to zero if not found
				AvailableBalance: decimal.Zero,
			}
		}
		if payload.Amt.GreaterThan(balance.AvailableBalance) {
			// skip inscribe transfer event if amount exceeds available balance
			return
		}
		// update balance state
		balance.BlockHeight = uint64(blockHeader.Height)
		balance.AvailableBalance = balance.AvailableBalance.Sub(payload.Amt)
		if _, ok := balances[pkScriptHex]; !ok {
			balances[pkScriptHex] = make(map[string]*entity.Balance)
		}
		balances[pkScriptHex][payload.Tick] = balance
		if _, ok := newBalances[pkScriptHex]; !ok {
			newBalances[pkScriptHex] = make(map[string]*entity.Balance)
		}
		newBalances[pkScriptHex][payload.Tick] = &entity.Balance{}

		event := &entity.EventInscribeTransfer{
			Id:                latestEventId + 1,
			InscriptionId:     payload.Transfer.InscriptionId,
			InscriptionNumber: inscriptionIdsToNumber[payload.Transfer.InscriptionId],
			Tick:              payload.Tick,
			OriginalTick:      payload.OriginalTick,
			TxHash:            payload.Transfer.TxHash,
			BlockHeight:       payload.Transfer.BlockHeight,
			TxIndex:           payload.Transfer.TxIndex,
			Timestamp:         blockHeader.Timestamp,
			PkScript:          payload.Transfer.NewPkScript,
			SatPoint:          payload.Transfer.NewSatPoint,
			OutputIndex:       payload.Transfer.NewSatPoint.OutPoint.Index,
			SatsAmount:        payload.Transfer.NewOutputValue,
			Amount:            payload.Amt,
		}
		latestEventId++
		eventInscribeTransfers[payload.Transfer.InscriptionId] = event
		newEventInscribeTransfers = append(newEventInscribeTransfers, event)

		eventHashBuilder.WriteString(getEventInscribeTransferString(event, tickEntry.Decimals) + eventHashSeparator)
	}
	handleEventTransferTransferAsFee := func(payload *brc20.Payload, tickEntry *entity.TickEntry, inscribeTransfer *entity.EventInscribeTransfer) {
		// return balance to sender
		fromPkScriptHex := hex.EncodeToString(inscribeTransfer.PkScript)
		fromBalance, ok := balances[fromPkScriptHex][payload.Tick]
		if !ok {
			fromBalance = &entity.Balance{
				PkScript:         inscribeTransfer.PkScript,
				Tick:             payload.Tick,
				BlockHeight:      uint64(blockHeader.Height),
				OverallBalance:   decimal.Zero, // defaults balance to zero if not found
				AvailableBalance: decimal.Zero,
			}
		}
		fromBalance.BlockHeight = uint64(blockHeader.Height)
		fromBalance.AvailableBalance = fromBalance.AvailableBalance.Add(payload.Amt)
		if _, ok := balances[fromPkScriptHex]; !ok {
			balances[fromPkScriptHex] = make(map[string]*entity.Balance)
		}
		balances[fromPkScriptHex][payload.Tick] = fromBalance
		if _, ok := newBalances[fromPkScriptHex]; !ok {
			newBalances[fromPkScriptHex] = make(map[string]*entity.Balance)
		}
		newBalances[fromPkScriptHex][payload.Tick] = fromBalance

		event := &entity.EventTransferTransfer{
			Id:                latestEventId + 1,
			InscriptionId:     payload.Transfer.InscriptionId,
			InscriptionNumber: inscriptionIdsToNumber[payload.Transfer.InscriptionId],
			Tick:              payload.Tick,
			OriginalTick:      payload.OriginalTick,
			TxHash:            payload.Transfer.TxHash,
			BlockHeight:       payload.Transfer.BlockHeight,
			TxIndex:           payload.Transfer.TxIndex,
			Timestamp:         blockHeader.Timestamp,
			FromPkScript:      inscribeTransfer.PkScript,
			FromSatPoint:      inscribeTransfer.SatPoint,
			FromInputIndex:    payload.Transfer.FromInputIndex,
			ToPkScript:        payload.Transfer.NewPkScript,
			ToSatPoint:        payload.Transfer.NewSatPoint,
			ToOutputIndex:     payload.Transfer.NewSatPoint.OutPoint.Index,
			SpentAsFee:        true,
			Amount:            payload.Amt,
		}
		newEventTransferTransfers = append(newEventTransferTransfers, event)

		eventHashBuilder.WriteString(getEventTransferTransferString(event, tickEntry.Decimals) + eventHashSeparator)
	}
	handleEventTransferTransferNormal := func(payload *brc20.Payload, tickEntry *entity.TickEntry, inscribeTransfer *entity.EventInscribeTransfer) {
		// subtract balance from sender
		fromPkScriptHex := hex.EncodeToString(inscribeTransfer.PkScript)
		fromBalance, ok := balances[fromPkScriptHex][payload.Tick]
		if !ok {
			// skip transfer transfer event if from balance does not exist
			return
		}
		fromBalance.BlockHeight = uint64(blockHeader.Height)
		fromBalance.OverallBalance = fromBalance.OverallBalance.Sub(payload.Amt)
		if _, ok := balances[fromPkScriptHex]; !ok {
			balances[fromPkScriptHex] = make(map[string]*entity.Balance)
		}
		balances[fromPkScriptHex][payload.Tick] = fromBalance
		if _, ok := newBalances[fromPkScriptHex]; !ok {
			newBalances[fromPkScriptHex] = make(map[string]*entity.Balance)
		}
		newBalances[fromPkScriptHex][payload.Tick] = fromBalance

		// add balance to receiver
		if bytes.Equal(payload.Transfer.NewPkScript, []byte{0x6a}) {
			// burn if sent to OP_RETURN
			tickEntry.BurnedAmount = tickEntry.BurnedAmount.Add(payload.Amt)
			tickEntries[payload.Tick] = tickEntry
			newTickEntryStates[payload.Tick] = tickEntry
		} else {
			toPkScriptHex := hex.EncodeToString(payload.Transfer.NewPkScript)
			toBalance, ok := balances[toPkScriptHex][payload.Tick]
			if !ok {
				toBalance = &entity.Balance{
					PkScript:         payload.Transfer.NewPkScript,
					Tick:             payload.Tick,
					BlockHeight:      uint64(blockHeader.Height),
					OverallBalance:   decimal.Zero, // defaults balance to zero if not found
					AvailableBalance: decimal.Zero,
				}
			}
			toBalance.BlockHeight = uint64(blockHeader.Height)
			toBalance.OverallBalance = toBalance.OverallBalance.Add(payload.Amt)
			toBalance.AvailableBalance = toBalance.AvailableBalance.Add(payload.Amt)
			if _, ok := balances[toPkScriptHex]; !ok {
				balances[toPkScriptHex] = make(map[string]*entity.Balance)
			}
			balances[toPkScriptHex][payload.Tick] = toBalance
			if _, ok := newBalances[toPkScriptHex]; !ok {
				newBalances[toPkScriptHex] = make(map[string]*entity.Balance)
			}
			newBalances[toPkScriptHex][payload.Tick] = toBalance
		}

		event := &entity.EventTransferTransfer{
			Id:                latestEventId + 1,
			InscriptionId:     payload.Transfer.InscriptionId,
			InscriptionNumber: inscriptionIdsToNumber[payload.Transfer.InscriptionId],
			Tick:              payload.Tick,
			OriginalTick:      payload.OriginalTick,
			TxHash:            payload.Transfer.TxHash,
			BlockHeight:       payload.Transfer.BlockHeight,
			TxIndex:           payload.Transfer.TxIndex,
			Timestamp:         blockHeader.Timestamp,
			FromPkScript:      inscribeTransfer.PkScript,
			FromSatPoint:      inscribeTransfer.SatPoint,
			FromInputIndex:    payload.Transfer.FromInputIndex,
			ToPkScript:        payload.Transfer.NewPkScript,
			ToSatPoint:        payload.Transfer.NewSatPoint,
			ToOutputIndex:     payload.Transfer.NewSatPoint.OutPoint.Index,
			SpentAsFee:        false,
			Amount:            payload.Amt,
		}
		newEventTransferTransfers = append(newEventTransferTransfers, event)

		eventHashBuilder.WriteString(getEventTransferTransferString(event, tickEntry.Decimals) + eventHashSeparator)
	}

	for _, payload := range payloads {
		tickEntry := tickEntries[payload.Tick]

		if payload.Transfer.SentAsFee && payload.Transfer.OldSatPoint == (ordinals.SatPoint{}) {
			// skip inscriptions inscribed as fee
			continue
		}

		switch payload.Op {
		case brc20.OperationDeploy:
			handleEventDeploy(payload, tickEntry)
		case brc20.OperationMint:
			handleEventMint(payload, tickEntry)
		case brc20.OperationTransfer:
			if payload.Transfer.TransferCount > 2 {
				// skip used transfer inscriptions
				continue
			}
			if tickEntry == nil {
				// skip transfer inscriptions for non-existent ticks
				continue
			}
			if -payload.Amt.Exponent() > int32(tickEntry.Decimals) {
				// skip transfer inscriptions with decimals greater than allowed
				continue
			}

			if payload.Transfer.OldSatPoint == (ordinals.SatPoint{}) {
				handleEventInscribeTransfer(payload, tickEntry)
			} else {
				// transfer transfer event
				inscribeTransfer, ok := eventInscribeTransfers[payload.Transfer.InscriptionId]
				if !ok {
					// skip transfer transfer event if prior inscribe transfer event does not exist
					continue
				}

				if payload.Transfer.SentAsFee {
					handleEventTransferTransferAsFee(payload, tickEntry, inscribeTransfer)
				} else {
					handleEventTransferTransferNormal(payload, tickEntry, inscribeTransfer)
				}
			}
		}
	}

	p.newTickEntries = newTickEntries
	p.newTickEntryStates = newTickEntryStates
	p.newEventDeploys = newEventDeploys
	p.newEventMints = newEventMints
	p.newEventInscribeTransfers = newEventInscribeTransfers
	p.newEventTransferTransfers = newEventTransferTransfers
	p.newBalances = newBalances
	p.eventHashString = eventHashBuilder.String()
	return nil
}
