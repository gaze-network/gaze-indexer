package brc20

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"slices"

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
)

// Process implements indexer.Processor.
func (p *Processor) Process(ctx context.Context, blocks []*types.Block) error {
	for _, block := range blocks {
		ctx = logger.WithContext(ctx, slogx.Uint64("height", uint64(block.Header.Height)))
		logger.DebugContext(ctx, "Processing new block")
		p.blockReward = p.getBlockSubsidy(uint64(block.Header.Height))
		p.flotsamsSentAsFee = make([]*entity.Flotsam, 0)

		var inputOutPoints []wire.OutPoint
		for _, tx := range block.Transactions {
			for _, txIn := range tx.TxIn {
				if txIn.PreviousOutTxHash == (chainhash.Hash{}) {
					// skip coinbase input
					continue
				}
				inputOutPoints = append(inputOutPoints, wire.OutPoint{
					Hash:  txIn.PreviousOutTxHash,
					Index: txIn.PreviousOutIndex,
				})
			}
		}
		transfersInOutPoints, err := p.getInscriptionTransfersInOutPoints(ctx, inputOutPoints)
		if err != nil {
			return errors.Wrap(err, "failed to get inscriptions in outpoints")
		}
		logger.DebugContext(ctx, "Got inscriptions in outpoints", slogx.Int("countOutPoints", len(transfersInOutPoints)))

		// put coinbase tx (first tx) at the end of block
		transactions := append(block.Transactions[1:], block.Transactions[0])
		for _, tx := range transactions {
			if err := p.processInscriptionTx(ctx, tx, block.Header, transfersInOutPoints); err != nil {
				return errors.Wrap(err, "failed to process tx")
			}
		}

		// sort transfers by tx index, output index, output sat offset
		// NOTE: ord indexes inscription transfers spent as fee at the end of the block, but brc20 indexes them as soon as they are sent
		slices.SortFunc(p.newInscriptionTransfers, func(t1, t2 *entity.InscriptionTransfer) int {
			if t1.TxIndex != t2.TxIndex {
				return int(t1.TxIndex) - int(t2.TxIndex)
			}
			if t1.SentAsFee != t2.SentAsFee {
				// transfers sent as fee should be ordered after non-fees
				if t1.SentAsFee {
					return 1
				}
				return -1
			}
			if t1.NewSatPoint.OutPoint.Index != t2.NewSatPoint.OutPoint.Index {
				return int(t1.NewSatPoint.OutPoint.Index) - int(t2.NewSatPoint.OutPoint.Index)
			}
			return int(t1.NewSatPoint.Offset) - int(t2.NewSatPoint.Offset)
		})

		if err := p.processBRC20States(ctx, p.newInscriptionTransfers, block.Header); err != nil {
			return errors.Wrap(err, "failed to process brc20 states")
		}

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

	// calculate event hash
	{
		eventHashString := p.eventHashString
		if len(eventHashString) > 0 && eventHashString[len(eventHashString)-1:] == eventHashSeparator {
			eventHashString = eventHashString[:len(eventHashString)-1]
		}
		eventHash := sha256.Sum256([]byte(eventHashString))
		prevIndexedBlock, err := brc20DgTx.GetIndexedBlockByHeight(ctx, blockHeader.Height-1)
		if err != nil && errors.Is(err, errs.NotFound) && blockHeader.Height-1 == startingBlockHeader[p.network].Height {
			prevIndexedBlock = &entity.IndexedBlock{
				Height:              uint64(startingBlockHeader[p.network].Height),
				Hash:                startingBlockHeader[p.network].Hash,
				EventHash:           []byte{},
				CumulativeEventHash: []byte{},
			}
			err = nil
		}
		if err != nil {
			return errors.Wrap(err, "failed to get previous indexed block")
		}
		var cumulativeEventHash [32]byte
		if len(prevIndexedBlock.CumulativeEventHash) == 0 {
			cumulativeEventHash = eventHash
		} else {
			cumulativeEventHash = sha256.Sum256([]byte(hex.EncodeToString(prevIndexedBlock.CumulativeEventHash[:]) + hex.EncodeToString(eventHash[:])))
		}
		if err := brc20DgTx.CreateIndexedBlock(ctx, &entity.IndexedBlock{
			Height:              blockHeight,
			Hash:                blockHeader.Hash,
			EventHash:           eventHash[:],
			CumulativeEventHash: cumulativeEventHash[:],
		}); err != nil {
			return errors.Wrap(err, "failed to create indexed block")
		}
		p.eventHashString = ""
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
	// newTickEntries            map[string]*entity.TickEntry
	// newTickEntryStates        map[string]*entity.TickEntry
	// newEventDeploys           []*entity.EventDeploy
	// newEventMints             []*entity.EventMint
	// newEventInscribeTransfers []*entity.EventInscribeTransfer
	// newEventTransferTransfers []*entity.EventTransferTransfer
	// newBalances               map[string]map[string]*entity.Balance

	// flush new tick entries
	{
		newTickEntries := lo.Values(p.newTickEntries)
		if err := brc20DgTx.CreateTickEntries(ctx, blockHeight, newTickEntries); err != nil {
			return errors.Wrap(err, "failed to create tick entries")
		}
		p.newTickEntries = make(map[string]*entity.TickEntry)
	}

	// flush new tick entry states
	{
		newTickEntryStates := lo.Values(p.newTickEntryStates)
		if err := brc20DgTx.CreateTickEntryStates(ctx, blockHeight, newTickEntryStates); err != nil {
			return errors.Wrap(err, "failed to create tick entry states")
		}
		p.newTickEntryStates = make(map[string]*entity.TickEntry)
	}

	// flush new events
	{
		if err := brc20DgTx.CreateEventDeploys(ctx, p.newEventDeploys); err != nil {
			return errors.Wrap(err, "failed to create event deploys")
		}
		if err := brc20DgTx.CreateEventMints(ctx, p.newEventMints); err != nil {
			return errors.Wrap(err, "failed to create event mints")
		}
		if err := brc20DgTx.CreateEventInscribeTransfers(ctx, p.newEventInscribeTransfers); err != nil {
			return errors.Wrap(err, "failed to create event inscribe transfers")
		}
		if err := brc20DgTx.CreateEventTransferTransfers(ctx, p.newEventTransferTransfers); err != nil {
			return errors.Wrap(err, "failed to create event transfer transfers")
		}
		p.newEventDeploys = make([]*entity.EventDeploy, 0)
		p.newEventMints = make([]*entity.EventMint, 0)
		p.newEventInscribeTransfers = make([]*entity.EventInscribeTransfer, 0)
		p.newEventTransferTransfers = make([]*entity.EventTransferTransfer, 0)
	}

	// flush new balances
	{
		newBalances := make([]*entity.Balance, 0)
		for _, tickBalances := range p.newBalances {
			for _, balance := range tickBalances {
				newBalances = append(newBalances, balance)
			}
		}
		if err := brc20DgTx.CreateBalances(ctx, newBalances); err != nil {
			return errors.Wrap(err, "failed to create balances")
		}
		p.newBalances = make(map[string]map[string]*entity.Balance)
	}

	if err := brc20DgTx.Commit(ctx); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}
	return nil
}
