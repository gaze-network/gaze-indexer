package brc20

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/brc20"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/gaze-network/uint128"
	"github.com/samber/lo"
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
	entries, err := p.getTickEntriesByTicks(ctx, lo.Keys(ticks))
	if err != nil {
		return errors.Wrap(err, "failed to get inscription entries by ids")
	}

	newTickEntries := make(map[string]*entity.TickEntry)
	newTickEntryStates := make(map[string]*entity.TickEntry)
	// newDeployEvents := make([]*entity.EventDeploy, 0)
	// newMintEvents := make([]*entity.EventMint, 0)
	// newTransferEvents := make([]*entity.EventTransfer, 0)

	for _, payload := range payloads {
		entry := entries[payload.Tick]

		switch payload.Op {
		case brc20.OperationDeploy:
			if entry != nil {
				logger.DebugContext(ctx, "found deploy inscription but tick already exists, skipping...",
					slogx.String("tick", payload.Tick),
					slogx.Stringer("entryInscriptionId", entry.DeployInscriptionId),
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
				DeployedAtHeight:    uint64(blockHeader.Height),
				MintedAmount:        uint128.Zero,
				BurnedAmount:        uint128.Zero,
				CompletedAt:         time.Time{},
				CompletedAtHeight:   0,
			}
			newTickEntries[payload.Tick] = tickEntry
			newTickEntryStates[payload.Tick] = tickEntry
			// update entries for other operations in same block
			entries[payload.Tick] = tickEntry

			// TODO: handle deploy action
		case brc20.OperationMint:
			if entry == nil {
				logger.DebugContext(ctx, "found mint inscription but tick does not exist, skipping...",
					slogx.String("tick", payload.Tick),
					slogx.Stringer("inscriptionId", payload.Transfer.InscriptionId),
				)
				continue
			}
			if brc20.IsAmountWithinDecimals(payload.Amt, entry.Decimals) {
				logger.DebugContext(ctx, "found mint inscription but amount has invalid decimals, skipping...",
					slogx.String("tick", payload.Tick),
					slogx.Stringer("inscriptionId", payload.Transfer.InscriptionId),
					slogx.Stringer("amount", payload.Amt),
					slogx.Uint16("decimals", entry.Decimals),
				)
				continue
			}
			// TODO: handle mint action
		case brc20.OperationTransfer:
			if entry == nil {
				logger.DebugContext(ctx, "found transfer inscription but tick does not exist, skipping...",
					slogx.String("tick", payload.Tick),
					slogx.Stringer("inscriptionId", payload.Transfer.InscriptionId),
				)
				continue
			}
			if brc20.IsAmountWithinDecimals(payload.Amt, entry.Decimals) {
				logger.DebugContext(ctx, "found transfer inscription but amount has invalid decimals, skipping...",
					slogx.String("tick", payload.Tick),
					slogx.Stringer("inscriptionId", payload.Transfer.InscriptionId),
					slogx.Stringer("amount", payload.Amt),
					slogx.Uint16("decimals", entry.Decimals),
				)
				continue
			}
			// TODO: handle transfer action
		}
	}
	return nil
}

func (p *Processor) getTickEntriesByTicks(ctx context.Context, ticks []string) (map[string]*entity.TickEntry, error) {
	// TODO: get from buffer if exists
	result, err := p.brc20Dg.GetTickEntriesByTicks(ctx, ticks)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get tick entries by ticks")
	}
	return result, nil
}
