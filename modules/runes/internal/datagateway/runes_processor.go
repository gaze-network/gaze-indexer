package datagateway

import (
	"context"

	"github.com/btcsuite/btcd/wire"
	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
	"github.com/gaze-network/uint128"
)

type RunesProcessorDataGateway interface {
	GetRunesBalancesInOutPoint(ctx context.Context, outPoint wire.OutPoint) (map[runes.RuneId]uint128.Uint128, error)
	GetRuneIdByRune(ctx context.Context, rune runes.Rune) (runes.RuneId, error)
	GetRuneEntryByRuneId(ctx context.Context, runeId runes.RuneId) (*runes.RuneEntry, error)

	SetRunesBalancesInOutPoint(ctx context.Context, outPoint wire.OutPoint, balances map[runes.RuneId]uint128.Uint128) error
	SetRuneIdByRune(ctx context.Context, rune runes.Rune, runeId runes.RuneId) error
	SetRuneEntryByRuneId(ctx context.Context, runeId runes.RuneId, entry *runes.RuneEntry) error
}
