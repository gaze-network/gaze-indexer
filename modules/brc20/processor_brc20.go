package brc20

import (
	"bytes"
	"context"
	"encoding/hex"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/brc20"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/datagateway"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/ordinals"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

func (p *Processor) processBRC20States(ctx context.Context, transfers []*entity.InscriptionTransfer, blockHeader types.BlockHeader) error {
	payloads := make([]*brc20.Payload, 0)
	ticks := make(map[string]struct{})
	for _, transfer := range transfers {
		payload, err := brc20.ParsePayload(transfer)
		if err != nil {
			return errors.Wrap(err, "failed to parse payload")
		}
		payloads = append(payloads, payload)
		ticks[payload.Tick] = struct{}{}
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

	for _, payload := range payloads {
		tickEntry := tickEntries[payload.Tick]

		if payload.Transfer.SentAsFee && payload.Transfer.OldSatPoint == (ordinals.SatPoint{}) {
			logger.DebugContext(ctx, "found inscription inscribed as fee, skipping...",
				slogx.String("tick", payload.Tick),
				slogx.Stringer("inscriptionId", payload.Transfer.InscriptionId),
			)
			continue
		}

		switch payload.Op {
		case brc20.OperationDeploy:
			if payload.Transfer.TransferCount > 1 {
				logger.DebugContext(ctx, "found deploy inscription but it is already used, skipping...",
					slogx.String("tick", payload.Tick),
					slogx.Stringer("inscriptionId", payload.Transfer.InscriptionId),
					slogx.Uint32("transferCount", payload.Transfer.TransferCount),
				)
				continue
			}
			if tickEntry != nil {
				logger.DebugContext(ctx, "found deploy inscription but tick already exists, skipping...",
					slogx.String("tick", payload.Tick),
					slogx.Stringer("entryInscriptionId", tickEntry.DeployInscriptionId),
					slogx.Stringer("currentInscriptionId", payload.Transfer.InscriptionId),
				)
				continue
			}
			tickEntry := &entity.TickEntry{
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
			newTickEntries[payload.Tick] = tickEntry
			newTickEntryStates[payload.Tick] = tickEntry
			// update entries for other operations in same block
			tickEntries[payload.Tick] = tickEntry

			newEventDeploys = append(newEventDeploys, &entity.EventDeploy{
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
			})
			latestEventId++
		case brc20.OperationMint:
			if payload.Transfer.TransferCount > 1 {
				logger.DebugContext(ctx, "found mint inscription but it is already used, skipping...",
					slogx.String("tick", payload.Tick),
					slogx.Stringer("inscriptionId", payload.Transfer.InscriptionId),
					slogx.Uint32("transferCount", payload.Transfer.TransferCount),
				)
				continue
			}
			if tickEntry == nil {
				logger.DebugContext(ctx, "found mint inscription but tick does not exist, skipping...",
					slogx.String("tick", payload.Tick),
					slogx.Stringer("inscriptionId", payload.Transfer.InscriptionId),
				)
				continue
			}
			if -payload.Amt.Exponent() > int32(tickEntry.Decimals) {
				logger.DebugContext(ctx, "found mint inscription but amount has invalid decimals, skipping...",
					slogx.String("tick", payload.Tick),
					slogx.Stringer("inscriptionId", payload.Transfer.InscriptionId),
					slogx.Stringer("amount", payload.Amt),
					slogx.Uint16("entryDecimals", tickEntry.Decimals),
					slogx.Int32("payloadDecimals", -payload.Amt.Exponent()),
				)
				continue
			}
			if tickEntry.MintedAmount.GreaterThanOrEqual(tickEntry.TotalSupply) {
				logger.DebugContext(ctx, "found mint inscription but total supply is reached, skipping...",
					slogx.String("tick", payload.Tick),
					slogx.Stringer("inscriptionId", payload.Transfer.InscriptionId),
					slogx.Stringer("mintedAmount", tickEntry.MintedAmount),
					slogx.Stringer("totalSupply", tickEntry.TotalSupply),
				)
				continue
			}
			if payload.Amt.GreaterThan(tickEntry.LimitPerMint) {
				logger.DebugContext(ctx, "found mint inscription but amount exceeds limit per mint, skipping...",
					slogx.String("tick", payload.Tick),
					slogx.Stringer("inscriptionId", payload.Transfer.InscriptionId),
					slogx.Stringer("amount", payload.Amt),
					slogx.Stringer("limitPerMint", tickEntry.LimitPerMint),
				)
				continue
			}
			mintableAmount := tickEntry.TotalSupply.Sub(tickEntry.MintedAmount)
			if payload.Amt.GreaterThan(mintableAmount) {
				payload.Amt = mintableAmount
			}
			var parentId *ordinals.InscriptionId
			if tickEntry.IsSelfMint {
				parentIdValue, ok := inscriptionIdsToParent[payload.Transfer.InscriptionId]
				if !ok {
					logger.DebugContext(ctx, "found mint inscription for self mint tick, but it does not have a parent, skipping...",
						slogx.String("tick", payload.Tick),
						slogx.Stringer("inscriptionId", payload.Transfer.InscriptionId),
					)
					continue
				}
				if parentIdValue != tickEntry.DeployInscriptionId {
					logger.DebugContext(ctx, "found mint inscription for self mint tick, but parent id does not match deploy inscription id, skipping...",
						slogx.String("tick", payload.Tick),
						slogx.Stringer("inscriptionId", payload.Transfer.InscriptionId),
						slogx.Stringer("parentId", parentId),
						slogx.Stringer("deployInscriptionId", tickEntry.DeployInscriptionId),
					)
				}
				parentId = &parentIdValue
			}

			tickEntry.MintedAmount = tickEntry.MintedAmount.Add(payload.Amt)
			if tickEntry.MintedAmount.GreaterThanOrEqual(tickEntry.TotalSupply) {
				tickEntry.CompletedAt = blockHeader.Timestamp
				tickEntry.CompletedAtHeight = payload.Transfer.BlockHeight
			}

			newTickEntryStates[payload.Tick] = tickEntry
			newEventMints = append(newEventMints, &entity.EventMint{
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
			})
			latestEventId++
		case brc20.OperationTransfer:
			if payload.Transfer.TransferCount > 2 {
				logger.DebugContext(ctx, "found mint inscription but it is already used, skipping...",
					slogx.String("tick", payload.Tick),
					slogx.Stringer("inscriptionId", payload.Transfer.InscriptionId),
					slogx.Uint32("transferCount", payload.Transfer.TransferCount),
				)
				continue
			}
			if tickEntry == nil {
				logger.DebugContext(ctx, "found transfer inscription but tick does not exist, skipping...",
					slogx.String("tick", payload.Tick),
					slogx.Stringer("inscriptionId", payload.Transfer.InscriptionId),
				)
				continue
			}
			if -payload.Amt.Exponent() > int32(tickEntry.Decimals) {
				logger.DebugContext(ctx, "found transfer inscription but amount has invalid decimals, skipping...",
					slogx.String("tick", payload.Tick),
					slogx.Stringer("inscriptionId", payload.Transfer.InscriptionId),
					slogx.Stringer("amount", payload.Amt),
					slogx.Uint16("entryDecimals", tickEntry.Decimals),
					slogx.Int32("payloadDecimals", -payload.Amt.Exponent()),
				)
				continue
			}

			if payload.Transfer.OldSatPoint == (ordinals.SatPoint{}) {
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
					logger.DebugContext(ctx, "found transfer inscription but amount exceeds available balance, skipping...",
						slogx.String("tick", payload.Tick),
						slogx.Stringer("inscriptionId", payload.Transfer.InscriptionId),
						slogx.Stringer("amount", payload.Amt),
						slogx.Stringer("availableBalance", balance.AvailableBalance),
					)
					continue
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
			} else {
				// transfer transfer event
				inscribeTransfer, ok := eventInscribeTransfers[payload.Transfer.InscriptionId]
				if !ok {
					logger.DebugContext(ctx, "found transfer transfer event but inscribe transfer does not exist, skipping...",
						slogx.String("tick", payload.Tick),
						slogx.Stringer("inscriptionId", payload.Transfer.InscriptionId),
					)
					continue
				}

				if payload.Transfer.SentAsFee {
					// return balance to sender
					fromPkScriptHex := hex.EncodeToString(inscribeTransfer.PkScript)
					fromBalance, ok := balances[fromPkScriptHex][payload.Tick]
					if !ok {
						logger.DebugContext(ctx, "found transfer transfer event but from balance does not exist, skipping...",
							slogx.String("tick", payload.Tick),
							slogx.Stringer("inscriptionId", payload.Transfer.InscriptionId),
							slogx.String("pkScript", fromPkScriptHex),
						)
						continue
					}
					fromBalance.BlockHeight = uint64(blockHeader.Height)
					fromBalance.AvailableBalance = fromBalance.AvailableBalance.Sub(payload.Amt)
					if _, ok := balances[fromPkScriptHex]; !ok {
						balances[fromPkScriptHex] = make(map[string]*entity.Balance)
					}
					balances[fromPkScriptHex][payload.Tick] = fromBalance
					if _, ok := newBalances[fromPkScriptHex]; !ok {
						newBalances[fromPkScriptHex] = make(map[string]*entity.Balance)
					}
					newBalances[fromPkScriptHex][payload.Tick] = fromBalance

					newEventTransferTransfers = append(newEventTransferTransfers, &entity.EventTransferTransfer{
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
					})
				} else {
					// subtract balance from sender
					fromPkScriptHex := hex.EncodeToString(inscribeTransfer.PkScript)
					fromBalance, ok := balances[fromPkScriptHex][payload.Tick]
					if !ok {
						logger.DebugContext(ctx, "found transfer transfer event but from balance does not exist, skipping...",
							slogx.String("tick", payload.Tick),
							slogx.Stringer("inscriptionId", payload.Transfer.InscriptionId),
							slogx.String("pkScript", fromPkScriptHex),
						)
						continue
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

					newEventTransferTransfers = append(newEventTransferTransfers, &entity.EventTransferTransfer{
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
					})
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
	return nil
}
