package datagateway

import (
	"context"

	"github.com/btcsuite/btcd/wire"
	"github.com/gaze-network/indexer-network/core/types"
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
	GetLatestBlock(ctx context.Context) (types.BlockHeader, error)
	GetIndexedBlockByHeight(ctx context.Context, height int64) (*entity.IndexedBlock, error)
	GetProcessorStats(ctx context.Context) (*entity.ProcessorStats, error)
	GetInscriptionTransfersInOutPoints(ctx context.Context, outPoints []wire.OutPoint) (map[ordinals.SatPoint][]*entity.InscriptionTransfer, error)
	GetInscriptionEntriesByIds(ctx context.Context, ids []ordinals.InscriptionId) (map[ordinals.InscriptionId]*ordinals.InscriptionEntry, error)
	GetTickEntriesByTicks(ctx context.Context, ticks []string) (map[string]*entity.TickEntry, error)
}

type BRC20WriterDataGateway interface {
	CreateIndexedBlock(ctx context.Context, block *entity.IndexedBlock) error
	CreateProcessorStats(ctx context.Context, stats *entity.ProcessorStats) error
	CreateTickEntries(ctx context.Context, blockHeight uint64, entries []*entity.TickEntry) error
	CreateTickEntryStates(ctx context.Context, blockHeight uint64, entryStates []*entity.TickEntry) error
	CreateInscriptionEntries(ctx context.Context, blockHeight uint64, entries []*ordinals.InscriptionEntry) error
	CreateInscriptionEntryStates(ctx context.Context, blockHeight uint64, entryStates []*ordinals.InscriptionEntry) error
	CreateInscriptionTransfers(ctx context.Context, transfers []*entity.InscriptionTransfer) error
	CreateEventDeploys(ctx context.Context, events []*entity.EventDeploy) error
	CreateEventMints(ctx context.Context, events []*entity.EventMint) error
	CreateEventInscribeTransfers(ctx context.Context, events []*entity.EventInscribeTransfer) error
	CreateEventTransferTransfers(ctx context.Context, events []*entity.EventTransferTransfer) error

	// used for revert data
	DeleteIndexedBlocksSinceHeight(ctx context.Context, height uint64) error
	DeleteProcessorStatsSinceHeight(ctx context.Context, height uint64) error
	DeleteTickEntriesSinceHeight(ctx context.Context, height uint64) error
	DeleteTickEntryStatesSinceHeight(ctx context.Context, height uint64) error
	DeleteEventDeploysSinceHeight(ctx context.Context, height uint64) error
	DeleteEventMintsSinceHeight(ctx context.Context, height uint64) error
	DeleteEventInscribeTransfersSinceHeight(ctx context.Context, height uint64) error
	DeleteEventTransferTransfersSinceHeight(ctx context.Context, height uint64) error
	DeleteBalancesSinceHeight(ctx context.Context, height uint64) error
	DeleteInscriptionEntriesSinceHeight(ctx context.Context, height uint64) error
	DeleteInscriptionEntryStatesSinceHeight(ctx context.Context, height uint64) error
	DeleteInscriptionTransfersSinceHeight(ctx context.Context, height uint64) error
}
