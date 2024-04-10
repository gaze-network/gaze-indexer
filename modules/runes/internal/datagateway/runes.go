package datagateway

import (
	"context"

	"github.com/btcsuite/btcd/wire"
	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
	"github.com/gaze-network/uint128"
)

type RunesDataGateway interface {
	RunesReaderDataGateway
	RunesWriterDataGateway
}

type RunesReaderDataGateway interface {
	GetRunesBalancesAtOutPoint(ctx context.Context, outPoint wire.OutPoint) (map[runes.RuneId]uint128.Uint128, error)
	GetRuneIdByRune(ctx context.Context, rune runes.Rune) (runes.RuneId, error)
	GetRuneEntryByRuneId(ctx context.Context, runeId runes.RuneId) (*runes.RuneEntry, error)
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

	SetRuneEntry(ctx context.Context, entry *runes.RuneEntry) error
	CreateRuneBalancesAtOutPoint(ctx context.Context, outPoint wire.OutPoint, balances map[runes.RuneId]uint128.Uint128) error
	CreateRuneBalancesAtBlock(ctx context.Context, params []CreateRuneBalancesAtBlockParams) error
	UpdateLatestBlockHeight(ctx context.Context, blockHeight uint64) error
}

type CreateRuneBalancesAtBlockParams struct {
	PkScript    []byte
	RuneId      runes.RuneId
	Balance     uint128.Uint128
	BlockHeight uint64
}
