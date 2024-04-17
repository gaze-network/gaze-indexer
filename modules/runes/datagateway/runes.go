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
}

type RunesReaderDataGateway interface {
	GetLatestBlock(ctx context.Context) (types.BlockHeader, error)
	GetIndexedBlockByHeight(ctx context.Context, height int64) (*entity.IndexedBlock, error)
	GetRuneTransactionsByHeight(ctx context.Context, height uint64) ([]*entity.RuneTransaction, error)

	GetRunesBalancesAtOutPoint(ctx context.Context, outPoint wire.OutPoint) (map[runes.RuneId]uint128.Uint128, error)
	// GetRuneIdFromRune returns the RuneId for the given rune. Returns errs.NotFound if the rune entry is not found.
	GetRuneIdFromRune(ctx context.Context, rune runes.Rune) (runes.RuneId, error)
	// GetRuneEntryByRuneId returns the RuneEntry for the given runeId. Returns errs.NotFound if the rune entry is not found.
	GetRuneEntryByRuneId(ctx context.Context, runeId runes.RuneId) (*runes.RuneEntry, error)
	// GetRuneEntryByRuneId returns the RuneEntry for the given runeId. Returns errs.NotFound if the rune entry is not found.
	GetRuneEntryByRuneIdBatch(ctx context.Context, runeIds []runes.RuneId) (map[runes.RuneId]*runes.RuneEntry, error)
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
	// Begin starts a DB transaction. All write operations done after this call must be followed by Commit() to persist those changes, or Rollback() to discard them.
	Begin(ctx context.Context) error
	// Commit commits the DB transaction. All changes made after Begin() will be persisted. Calling Commit() will close the current transaction.
	// If Commit() is called without a prior Begin(), it must be a no-op.
	Commit(ctx context.Context) error
	// Rollback rolls back the DB transaction. All changes made after Begin() will be discarded.
	// Rollback() must be safe to call even if no transaction is active. Hence, a defer Rollback() is safe, even if Commit() was called prior with non-error conditions.
	Rollback(ctx context.Context) error

	CreateRuneEntry(ctx context.Context, entry *runes.RuneEntry, blockHeight uint64) error
	CreateRuneEntryState(ctx context.Context, entry *runes.RuneEntry, blockHeight uint64) error
	CreateOutPointBalances(ctx context.Context, outPoint wire.OutPoint, balances map[runes.RuneId]uint128.Uint128, blockHeight uint64) error
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
