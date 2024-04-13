package runes

import (
	"math"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/uint128"
)

type RuneEntry struct {
	RuneId       RuneId
	Divisibility uint8
	// Premine is the amount of the rune that was premined.
	Premine    uint128.Uint128
	SpacedRune SpacedRune
	Symbol     rune
	// Terms is the minting terms of the rune.
	Terms *Terms
	Turbo bool
	// Mints is the number of times that this rune has been minted.
	Mints        uint128.Uint128
	BurnedAmount uint128.Uint128
	// CompletedAt is the time when the rune was fully minted.
	CompletedAt time.Time
	// CompletedAtHeight is the block height when the rune was fully minted.
	CompletedAtHeight *uint64
}

var (
	ErrUnmintable      = errors.New("rune is not mintable")
	ErrMintCapReached  = errors.New("rune mint cap reached")
	ErrMintBeforeStart = errors.New("rune minting has not started")
	ErrMintAfterEnd    = errors.New("rune minting has ended")
)

func (e *RuneEntry) GetMintableAmount(height uint64) (uint128.Uint128, error) {
	if e.Terms == nil {
		return uint128.Uint128{}, ErrUnmintable
	}
	if !e.IsMintStarted(height) {
		return uint128.Uint128{}, ErrMintBeforeStart
	}
	if e.IsMintEnded(height) {
		return uint128.Uint128{}, ErrMintAfterEnd
	}
	var cap uint128.Uint128
	if e.Terms.Cap != nil {
		cap = *e.Terms.Cap
	}
	if e.Mints.Cmp(cap) > 0 {
		return uint128.Uint128{}, ErrMintCapReached
	}
	var amount uint128.Uint128
	if e.Terms.Amount != nil {
		amount = *e.Terms.Amount
	}
	return amount, nil
}

func (e *RuneEntry) IsMintStarted(height uint64) bool {
	if e.Terms == nil {
		return false
	}

	var relative, absolute uint64
	if e.Terms.OffsetStart != nil {
		relative = e.RuneId.BlockHeight + *e.Terms.OffsetStart
	}
	if e.Terms.HeightStart != nil {
		absolute = *e.Terms.HeightStart
	}

	return height >= max(relative, absolute)
}

func (e *RuneEntry) IsMintEnded(height uint64) bool {
	if e.Terms == nil {
		return false
	}

	var relative, absolute uint64 = math.MaxUint64, math.MaxUint64
	if e.Terms.OffsetEnd != nil {
		relative = e.RuneId.BlockHeight + *e.Terms.OffsetEnd
	}
	if e.Terms.HeightEnd != nil {
		absolute = *e.Terms.HeightEnd
	}

	return height >= min(relative, absolute)
}
