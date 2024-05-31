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
	GetInscriptionEntryById(ctx context.Context, id ordinals.InscriptionId) (*ordinals.InscriptionEntry, error)
}

type BRC20WriterDataGateway interface {
	CreateIndexedBlock(ctx context.Context, block *entity.IndexedBlock) error
	CreateProcessorStats(ctx context.Context, stats *entity.ProcessorStats) error
	CreateInscriptionEntries(ctx context.Context, blockHeight uint64, entries []*ordinals.InscriptionEntry) error
	CreateInscriptionEntryStates(ctx context.Context, blockHeight uint64, entryStates []*ordinals.InscriptionEntry) error
	CreateInscriptionTransfers(ctx context.Context, transfers []*entity.InscriptionTransfer) error

	// used for revert data
	DeleteIndexedBlocksSinceHeight(ctx context.Context, height uint64) error
	DeleteProcessorStatsSinceHeight(ctx context.Context, height uint64) error
	DeleteTicksSinceHeight(ctx context.Context, height uint64) error
	DeleteTickStatesSinceHeight(ctx context.Context, height uint64) error
	DeleteDeployEventsSinceHeight(ctx context.Context, height uint64) error
	DeleteMintEventsSinceHeight(ctx context.Context, height uint64) error
	DeleteTransferEventsSinceHeight(ctx context.Context, height uint64) error
	DeleteBalancesSinceHeight(ctx context.Context, height uint64) error
	DeleteInscriptionEntriesSinceHeight(ctx context.Context, height uint64) error
	DeleteInscriptionEntryStatesSinceHeight(ctx context.Context, height uint64) error
	DeleteInscriptionTransfersSinceHeight(ctx context.Context, height uint64) error
}
