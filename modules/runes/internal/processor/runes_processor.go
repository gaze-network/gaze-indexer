package processor

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gaze-network/indexer-network/core/indexers"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
	"github.com/gaze-network/uint128"
)

var _ indexers.BitcoinProcessor = (*runesProcessor)(nil)

func toLocationId(txHash *chainhash.Hash, vout uint32) string {
	return fmt.Sprintf("%si%d", txHash.String(), vout)
}

// TODO: flush internal states to database periodically to avoid memory leak
type runesProcessor struct {
	runeIdToEntry          map[runes.RuneId]*RuneEntry
	outPointToRuneBalances map[string]uint128.Uint128
	runeToRuneId           map[runes.Rune]runes.RuneId
}

func NewRunesProcessor() *runesProcessor {
	return &runesProcessor{
		runeIdToEntry:          make(map[runes.RuneId]*RuneEntry),
		outPointToRuneBalances: make(map[string]uint128.Uint128),
		runeToRuneId:           make(map[runes.Rune]runes.RuneId),
	}
}

func (r *runesProcessor) CurrentBlock() (types.BlockHeader, error) {
	panic("implement me")
}

func (r *runesProcessor) PrepareData(ctx context.Context, from, to int64) ([]types.Block, error) {
	panic("implement me")
}

func (r *runesProcessor) RevertData(ctx context.Context, from types.BlockHeader) error {
	panic("implement me")
}
