package brc20

import (
	"context"

	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
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

func NewProcessor(ctx context.Context, brc20Dg datagateway.BRC20DataGateway, btcClient btcclient.Contract, network common.Network, transferCountLimit uint32) (*Processor, error) {
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

// VerifyStates implements indexer.Processor.
func (p *Processor) VerifyStates(ctx context.Context) error {
	panic("unimplemented")
}

// CurrentBlock implements indexer.Processor.
func (p *Processor) CurrentBlock(ctx context.Context) (types.BlockHeader, error) {
	panic("unimplemented")
}

// GetIndexedBlock implements indexer.Processor.
func (p *Processor) GetIndexedBlock(ctx context.Context, height int64) (types.BlockHeader, error) {
	panic("unimplemented")
}

// Name implements indexer.Processor.
func (p *Processor) Name() string {
	panic("unimplemented")
}

// RevertData implements indexer.Processor.
func (p *Processor) RevertData(ctx context.Context, from int64) error {
	panic("unimplemented")
}
