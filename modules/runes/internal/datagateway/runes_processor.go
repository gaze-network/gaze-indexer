package datagateway

import (
	"context"

	"github.com/btcsuite/btcd/wire"
	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
	"github.com/gaze-network/uint128"
)

type FlushData struct {
	RuneEntries      map[runes.RuneId]*runes.RuneEntry
	OutPointBalances map[wire.OutPoint]map[runes.RuneId]uint128.Uint128
	Balances         map[string]map[uint64]map[runes.RuneId]uint128.Uint128
	BlockHeight      uint64
}

type RunesProcessorDataGateway interface {
	GetRunesBalancesAtOutPoint(ctx context.Context, outPoint wire.OutPoint) (map[runes.RuneId]uint128.Uint128, error)
	GetRuneIdByRune(ctx context.Context, rune runes.Rune) (runes.RuneId, error)
	GetRuneEntryByRuneId(ctx context.Context, runeId runes.RuneId) (*runes.RuneEntry, error)

	FlushStates(ctx context.Context, data FlushData) error
}
