package datagateway

import (
	"context"

	"github.com/btcsuite/btcd/wire"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/ordinals"
)

type BRC20DataGateway interface {
	BRC20ReaderDataGateway
	BRC20WriterDataGateway

	// BeginBRC20Tx returns a new BRC20DataGateway with transaction enabled. All write operations performed in this datagateway must be committed to persist changes.
	BeginBRC20Tx(ctx context.Context) (BRC20DataGatewayWithTx, error)
}

type BRC20DataGatewayWithTx interface {
	BRC20DataGateway
	Tx
}

type BRC20ReaderDataGateway interface {
	GetProcessorStats(ctx context.Context) (*entity.ProcessorStats, error)

	GetInscriptionsInOutPoint(ctx context.Context, outPoint wire.OutPoint) (map[ordinals.SatPoint]ordinals.InscriptionId, error)
	GetInscriptionEntryById(ctx context.Context, id ordinals.InscriptionId) (*ordinals.InscriptionEntry, error)
}

type BRC20WriterDataGateway interface {
	CreateInscriptionEntries(ctx context.Context, blockHeight uint64, entries []*ordinals.InscriptionEntry) error
	CreateInscriptionEntryStates(ctx context.Context, blockHeight uint64, entryStates []*ordinals.InscriptionEntry) error
	CreateInscriptionTransfers(ctx context.Context, transfers []*entity.InscriptionTransfer) error
	CreateProcessorStats(ctx context.Context, stats *entity.ProcessorStats) error
}
