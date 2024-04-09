package processor

import (
	"time"

	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
	"github.com/gaze-network/uint128"
)

type RuneEntry struct {
	// EtchingTime is the time when the rune was created.
	EtchingTime time.Time
	// Terms is the minting terms of the rune.
	Terms *runes.Terms
	// EtchingTxHash is the hash of the transaction that created the rune.
	EtchingTxHash string
	SpacedRune    runes.SpacedRune
	BurnedAmount  uint128.Uint128
	// Mints is the number of times that this rune has been minted.
	Mints uint128.Uint128
	// Premine is the amount of the rune that was premined.
	Premine uint128.Uint128
	// EtchingBlock is the block height where the rune was created.
	EtchingBlock uint64
	// Number is the sequential number of the rune, starting from one.
	Number       uint64
	Symbol       rune
	Divisibility uint8
}
