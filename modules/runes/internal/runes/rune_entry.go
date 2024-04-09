package runes

import (
	"time"

	"github.com/gaze-network/uint128"
)

type RuneEntry struct {
	// CompletionTime is the time when the rune was fully minted.
	CompletionTime time.Time
	// Terms is the minting terms of the rune.
	Terms        *Terms
	SpacedRune   SpacedRune
	RuneId       RuneId
	BurnedAmount uint128.Uint128
	// Mints is the number of times that this rune has been minted.
	Mints uint128.Uint128
	// Premine is the amount of the rune that was premined.
	Premine      uint128.Uint128
	Symbol       rune
	Divisibility uint8
}
