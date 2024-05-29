package brc20

import (
	"context"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/ordinals"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

// Process implements indexer.Processor.
func (p *Processor) Process(ctx context.Context, blocks []*types.Block) error {
	for _, block := range blocks {
		ctx = logger.WithContext(ctx, slogx.Uint64("height", uint64(block.Header.Height)))
		logger.DebugContext(ctx, "Processing new block")
		p.blockReward = p.getBlockSubsidy(uint64(block.Header.Height))
		p.flotsamsSentAsFee = make([]*Flotsam, 0)
		for _, tx := range block.Transactions {
			if err := p.processInscriptionTx(ctx, tx, block.Header); err != nil {
				return errors.Wrap(err, "failed to process tx")
			}
		}
		// TODO: add brc20 processing
		if err := p.flushBlock(ctx, block.Header); err != nil {
			return errors.Wrap(err, "failed to flush block")
		}

		logger.DebugContext(ctx, "Inserted new block")
	}
	return nil
}

func (p *Processor) flushBlock(ctx context.Context, blockHeader types.BlockHeader) error {
	brc20DgTx, err := p.brc20Dg.BeginBRC20Tx(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer func() {
		if err := brc20DgTx.Rollback(ctx); err != nil {
			logger.WarnContext(ctx, "failed to rollback transaction",
				slogx.Error(err),
				slogx.String("event", "rollback_brc20_insertion"),
			)
		}
	}()

	blockHeight := uint64(blockHeader.Height)

	// CreateIndexedBlock must be performed before other flush methods to correctly calculate event hash
	// TODO: calculate event hash
	if err := brc20DgTx.CreateIndexedBlock(ctx, &entity.IndexedBlock{
		Height:              blockHeight,
		Hash:                blockHeader.Hash,
		EventHash:           chainhash.Hash{},
		CumulativeEventHash: chainhash.Hash{},
	}); err != nil {
		return errors.Wrap(err, "failed to create indexed block")
	}

	// flush new inscription entries
	{
		newInscriptionEntries := lo.Values(p.newInscriptionEntries)
		if err := brc20DgTx.CreateInscriptionEntries(ctx, blockHeight, newInscriptionEntries); err != nil {
			return errors.Wrap(err, "failed to create inscription entries")
		}
		p.newInscriptionEntries = make(map[ordinals.InscriptionId]*ordinals.InscriptionEntry)
	}

	// flush new inscription entry states
	{
		newInscriptionEntryStates := lo.Values(p.newInscriptionEntryStates)
		if err := brc20DgTx.CreateInscriptionEntryStates(ctx, blockHeight, newInscriptionEntryStates); err != nil {
			return errors.Wrap(err, "failed to create inscription entry states")
		}
		p.newInscriptionEntryStates = make(map[ordinals.InscriptionId]*ordinals.InscriptionEntry)
	}

	// flush new inscription entry states
	{
		if err := brc20DgTx.CreateInscriptionTransfers(ctx, p.newInscriptionTransfers); err != nil {
			return errors.Wrap(err, "failed to create inscription transfers")
		}
		p.newInscriptionTransfers = make([]*entity.InscriptionTransfer, 0)
	}

	// flush processor stats
	{
		stats := &entity.ProcessorStats{
			BlockHeight:             blockHeight,
			CursedInscriptionCount:  p.cursedInscriptionCount,
			BlessedInscriptionCount: p.blessedInscriptionCount,
			LostSats:                p.lostSats,
		}
		if err := brc20DgTx.CreateProcessorStats(ctx, stats); err != nil {
			return errors.Wrap(err, "failed to create processor stats")
		}
	}

	if err := brc20DgTx.Commit(ctx); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}
	return nil
}
