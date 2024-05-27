package processors

import (
	"context"

	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/datagateway"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/ordinals"
	"github.com/gaze-network/indexer-network/pkg/btcclient"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/gaze-network/indexer-network/pkg/lru"
	"github.com/samber/lo"
)

type Processor struct {
	brc20Dg            datagateway.BRC20DataGateway
	btcClient          btcclient.Contract
	network            common.Network
	transferCountLimit uint32 // number of transfers to track per inscription

	// block states
	flotsamsSentAsFee []*Flotsam
	blockReward       uint64

	// processor stats
	cursedInscriptionCount  uint64
	blessedInscriptionCount uint64
	lostSats                uint64

	// cache
	outPointValueCache *lru.Cache[wire.OutPoint, uint64]

	// flush buffers
	newInscriptionTransfers   []*entity.InscriptionTransfer
	newInscriptionEntries     map[ordinals.InscriptionId]*ordinals.InscriptionEntry
	newInscriptionEntryStates map[ordinals.InscriptionId]*ordinals.InscriptionEntry
}

// TODO: move this to config
const outPointValueCacheSize = 100000

func NewInscriptionProcessor(ctx context.Context, brc20Dg datagateway.BRC20DataGateway, btcClient btcclient.Contract, network common.Network, transferCountLimit uint32) (*Processor, error) {
	outPointValueCache, err := lru.New[wire.OutPoint, uint64](outPointValueCacheSize)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create outPointValueCache")
	}
	stats, err := brc20Dg.GetProcessorStats(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to count cursed inscriptions")
	}

	return &Processor{
		brc20Dg:            brc20Dg,
		btcClient:          btcClient,
		network:            network,
		transferCountLimit: transferCountLimit,

		flotsamsSentAsFee: make([]*Flotsam, 0),
		blockReward:       0,

		cursedInscriptionCount:  stats.CursedInscriptionCount,
		blessedInscriptionCount: stats.BlessedInscriptionCount,
		lostSats:                stats.LostSats,
		outPointValueCache:      outPointValueCache,

		newInscriptionTransfers:   make([]*entity.InscriptionTransfer, 0),
		newInscriptionEntries:     make(map[ordinals.InscriptionId]*ordinals.InscriptionEntry),
		newInscriptionEntryStates: make(map[ordinals.InscriptionId]*ordinals.InscriptionEntry),
	}, nil
}

func (p *Processor) Process(ctx context.Context, blocks []*types.Block) error {
	// collect all inputs to prefetch outpoints values
	inputs := make([]wire.OutPoint, 0)
	for _, block := range blocks {
		for _, tx := range block.Transactions {
			for _, txIn := range tx.TxIn {
				inputs = append(inputs, wire.OutPoint{
					Hash:  txIn.PreviousOutTxHash,
					Index: txIn.PreviousOutIndex,
				})
			}
		}
	}
	// prefetch outpoint values to cache
	_, err := p.getOutPointValues(ctx, inputs)
	if err != nil {
		return errors.Wrap(err, "failed to prefetch outpoint values")
	}

	for _, block := range blocks {
		p.blockReward = p.getBlockSubsidy(uint64(block.Header.Height))
		p.flotsamsSentAsFee = make([]*Flotsam, 0)
		for _, tx := range block.Transactions {
			if err := p.processTx(ctx, tx, block.Header); err != nil {
				return errors.Wrap(err, "failed to process tx")
			}
		}
		// TODO: add brc20 processing
		if err := p.flushBlock(ctx, block.Header); err != nil {
			return errors.Wrap(err, "failed to flush block")
		}
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
	return nil
}
