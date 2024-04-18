package datagateway

import (
	"context"

	"github.com/btcsuite/btcd/wire"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/uint128"
)

type RunesDataGateway interface {
	RunesReaderDataGateway
	RunesWriterDataGateway

	// BeginRunesTx returns a new RunesDataGateway with transaction enabled. All write operations performed in this datagateway must be committed to persist changes.
	BeginRunesTx(ctx context.Context) (RunesDataGatewayWithTx, error)
}

type RunesDataGatewayWithTx interface {
	RunesDataGateway
	Tx
}

type RunesReaderDataGateway interface {
	GetLatestBlock(ctx context.Context) (types.BlockHeader, error)
	GetIndexedBlockByHeight(ctx context.Context, height int64) (*entity.IndexedBlock, error)
	GetRuneTransactionsByHeight(ctx context.Context, height uint64) ([]*entity.RuneTransaction, error)

	GetRunesBalancesAtOutPoint(ctx context.Context, outPoint wire.OutPoint) (map[runes.RuneId]*entity.OutPointBalance, error)
	GetUnspentOutPointBalancesByPkScript(ctx context.Context, pkScript []byte, blockHeight uint64) ([]*entity.OutPointBalance, error)
	// GetRuneIdFromRune returns the RuneId for the given rune. Returns errs.NotFound if the rune entry is not found.
	GetRuneIdFromRune(ctx context.Context, rune runes.Rune) (runes.RuneId, error)
	// GetRuneEntryByRuneId returns the RuneEntry for the given runeId. Returns errs.NotFound if the rune entry is not found.
	GetRuneEntryByRuneId(ctx context.Context, runeId runes.RuneId) (*runes.RuneEntry, error)
	// GetRuneEntryByRuneIdBatch returns the RuneEntries for the given runeIds.
	GetRuneEntryByRuneIdBatch(ctx context.Context, runeIds []runes.RuneId) (map[runes.RuneId]*runes.RuneEntry, error)
	// GetRuneEntryByRuneIdAndHeight returns the RuneEntry for the given runeId and block height. Returns errs.NotFound if the rune entry is not found.
	GetRuneEntryByRuneIdAndHeight(ctx context.Context, runeId runes.RuneId, blockHeight uint64) (*runes.RuneEntry, error)
	// GetRuneEntryByRuneIdAndHeightBatch returns the RuneEntries for the given runeIds and block height.
	GetRuneEntryByRuneIdAndHeightBatch(ctx context.Context, runeIds []runes.RuneId, blockHeight uint64) (map[runes.RuneId]*runes.RuneEntry, error)
	// CountRuneEntries returns the number of existing rune entries.
	CountRuneEntries(ctx context.Context) (uint64, error)

	// GetBalancesByPkScript returns the balances for the given pkScript at the given blockHeight.
	GetBalancesByPkScript(ctx context.Context, pkScript []byte, blockHeight uint64) (map[runes.RuneId]*entity.Balance, error)
	// GetBalancesByRuneId returns the balances for the given runeId at the given blockHeight.
	// Cannot use []byte as map key, so we're returning as slice.
	GetBalancesByRuneId(ctx context.Context, runeId runes.RuneId, blockHeight uint64) ([]*entity.Balance, error)
	// GetBalancesByPkScriptAndRuneId returns the balance for the given pkScript and runeId at the given blockHeight.
	GetBalanceByPkScriptAndRuneId(ctx context.Context, pkScript []byte, runeId runes.RuneId, blockHeight uint64) (*entity.Balance, error)
}

type RunesWriterDataGateway interface {
	CreateRuneEntry(ctx context.Context, entry *runes.RuneEntry, blockHeight uint64) error
	CreateRuneEntryState(ctx context.Context, entry *runes.RuneEntry, blockHeight uint64) error
	CreateOutPointBalances(ctx context.Context, outPointBalances []*entity.OutPointBalance) error
	SpendOutPointBalances(ctx context.Context, outPoint wire.OutPoint, blockHeight uint64) error
	CreateRuneBalances(ctx context.Context, params []CreateRuneBalancesParams) error
	CreateRuneTransaction(ctx context.Context, tx *entity.RuneTransaction) error
	CreateIndexedBlock(ctx context.Context, block *entity.IndexedBlock) error

	// TODO: collapse these into a single function (ResetStateToHeight)?
	DeleteIndexedBlockSinceHeight(ctx context.Context, height uint64) error
	DeleteRuneEntriesSinceHeight(ctx context.Context, height uint64) error
	DeleteRuneEntryStatesSinceHeight(ctx context.Context, height uint64) error
	DeleteRuneTransactionsSinceHeight(ctx context.Context, height uint64) error
	DeleteRunestonesSinceHeight(ctx context.Context, height uint64) error
	DeleteOutPointBalancesSinceHeight(ctx context.Context, height uint64) error
	UnspendOutPointBalancesSinceHeight(ctx context.Context, height uint64) error
	DeleteRuneBalancesSinceHeight(ctx context.Context, height uint64) error
}

type CreateRuneBalancesParams struct {
	PkScript    []byte
	RuneId      runes.RuneId
	Balance     uint128.Uint128
	BlockHeight uint64
}
