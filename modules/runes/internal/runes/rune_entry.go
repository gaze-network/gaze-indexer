package runes

import (
	"math"
	"time"

	"github.com/gaze-network/indexer-network/common/errs"
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

var (
	ErrUnmintable      = errs.ErrorKind("rune is not mintable")
	ErrMintCapReached  = errs.ErrorKind("rune mint cap reached")
	ErrMintBeforeStart = errs.ErrorKind("rune minting has not started")
	ErrMintAfterEnd    = errs.ErrorKind("rune minting has ended")
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
