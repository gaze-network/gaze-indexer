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
	// Begin starts a DB transaction. All write operations done after this call must call Commit() to persist the changes.
	Begin(ctx context.Context) error
	// Commit commits the DB transaction. All changes made after Begin() will be persisted.
	Commit(ctx context.Context) error
	// Rollback rolls back the DB transaction. All changes made after Begin() will be discarded.
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
