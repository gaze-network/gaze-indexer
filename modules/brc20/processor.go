package brc20

import (
	"context"

	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/core/indexer"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/datagateway"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/ordinals"
	"github.com/gaze-network/indexer-network/pkg/btcclient"
	"github.com/gaze-network/indexer-network/pkg/lru"
)

// Make sure to implement the Bitcoin Processor interface
var _ indexer.Processor[*types.Block] = (*Processor)(nil)

type Processor struct {
	brc20Dg            datagateway.BRC20DataGateway
	indexerInfoDg      datagateway.IndexerInfoDataGateway
	btcClient          btcclient.Contract
	network            common.Network
	transferCountLimit uint32 // number of transfers to track per inscription
	cleanupFuncs       []func(context.Context) error

	// block states
	flotsamsSentAsFee []*entity.Flotsam
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

func NewProcessor(brc20Dg datagateway.BRC20DataGateway, indexerInfoDg datagateway.IndexerInfoDataGateway, btcClient btcclient.Contract, network common.Network, transferCountLimit uint32, cleanupFuncs []func(context.Context) error) (*Processor, error) {
	outPointValueCache, err := lru.New[wire.OutPoint, uint64](outPointValueCacheSize)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create outPointValueCache")
	}

	return &Processor{
		brc20Dg:            brc20Dg,
		indexerInfoDg:      indexerInfoDg,
		btcClient:          btcClient,
		network:            network,
		transferCountLimit: transferCountLimit,
		cleanupFuncs:       cleanupFuncs,

		flotsamsSentAsFee: make([]*entity.Flotsam, 0),
		blockReward:       0,

		cursedInscriptionCount:  0, // to be initialized by p.VerifyStates()
		blessedInscriptionCount: 0, // to be initialized by p.VerifyStates()
		lostSats:                0, // to be initialized by p.VerifyStates()
		outPointValueCache:      outPointValueCache,

		newInscriptionTransfers:   make([]*entity.InscriptionTransfer, 0),
		newInscriptionEntries:     make(map[ordinals.InscriptionId]*ordinals.InscriptionEntry),
		newInscriptionEntryStates: make(map[ordinals.InscriptionId]*ordinals.InscriptionEntry),
	}, nil
}

// VerifyStates implements indexer.Processor.
func (p *Processor) VerifyStates(ctx context.Context) error {
	indexerState, err := p.indexerInfoDg.GetLatestIndexerState(ctx)
	if err != nil && !errors.Is(err, errs.NotFound) {
		return errors.Wrap(err, "failed to get latest indexer state")
	}
	// if not found, create indexer state
	if errors.Is(err, errs.NotFound) {
		if err := p.indexerInfoDg.CreateIndexerState(ctx, entity.IndexerState{
			ClientVersion:    ClientVersion,
			DBVersion:        DBVersion,
			EventHashVersion: EventHashVersion,
			Network:          p.network,
		}); err != nil {
			return errors.Wrap(err, "failed to set indexer state")
		}
	} else {
		if indexerState.DBVersion != DBVersion {
			return errors.Wrapf(errs.ConflictSetting, "db version mismatch: current version is %d. Please upgrade to version %d", indexerState.DBVersion, DBVersion)
		}
		if indexerState.EventHashVersion != EventHashVersion {
			return errors.Wrapf(errs.ConflictSetting, "event version mismatch: current version is %d. Please reset rune's db first.", indexerState.EventHashVersion, EventHashVersion)
		}
		if indexerState.Network != p.network {
			return errors.Wrapf(errs.ConflictSetting, "network mismatch: latest indexed network is %d, configured network is %d. If you want to change the network, please reset the database", indexerState.Network, p.network)
		}
	}

	stats, err := p.brc20Dg.GetProcessorStats(ctx)
	if err != nil {
		if !errors.Is(err, errs.NotFound) {
			return errors.Wrap(err, "failed to count cursed inscriptions")
		}
		stats = &entity.ProcessorStats{
			BlockHeight:             uint64(startingBlockHeader[p.network].Height),
			CursedInscriptionCount:  0,
			BlessedInscriptionCount: 0,
			LostSats:                0,
		}
	}
	p.cursedInscriptionCount = stats.CursedInscriptionCount
	p.blessedInscriptionCount = stats.BlessedInscriptionCount
	p.lostSats = stats.LostSats
	return nil
}

// CurrentBlock implements indexer.Processor.
func (p *Processor) CurrentBlock(ctx context.Context) (types.BlockHeader, error) {
	blockHeader, err := p.brc20Dg.GetLatestBlock(ctx)
	if err != nil {
		if errors.Is(err, errs.NotFound) {
			return startingBlockHeader[p.network], nil
		}
		return types.BlockHeader{}, errors.Wrap(err, "failed to get latest block")
	}
	return blockHeader, nil
}

// GetIndexedBlock implements indexer.Processor.
func (p *Processor) GetIndexedBlock(ctx context.Context, height int64) (types.BlockHeader, error) {
	block, err := p.brc20Dg.GetIndexedBlockByHeight(ctx, height)
	if err != nil {
		return types.BlockHeader{}, errors.Wrap(err, "failed to get indexed block")
	}
	return types.BlockHeader{
		Height: int64(block.Height),
		Hash:   block.Hash,
	}, nil
}

// Name implements indexer.Processor.
func (p *Processor) Name() string {
	return "brc20"
}

// RevertData implements indexer.Processor.
func (p *Processor) RevertData(ctx context.Context, from int64) error {
	panic("unimplemented")
}

func (p *Processor) Shutdown(ctx context.Context) error {
	var errs []error
	for _, cleanup := range p.cleanupFuncs {
		if err := cleanup(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.WithStack(errors.Join(errs...))
}
