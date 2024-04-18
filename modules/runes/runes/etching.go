package runes

import (
	"github.com/Cleverse/go-utilities/utils"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/uint128"
	"github.com/samber/lo"
)

type Terms struct {
	// Amount of the rune to be minted per transaction
	Amount *uint128.Uint128
	// Number of allowed mints
	Cap *uint128.Uint128
	// Block height at which the rune can start being minted. If both HeightStart and OffsetStart are set, use the higher value.
	HeightStart *uint64
	// Block height at which the rune can no longer be minted. If both HeightEnd and OffsetEnd are set, use the lower value.
	HeightEnd *uint64
	// Offset from etched block at which the rune can start being minted. If both HeightStart and OffsetStart are set, use the higher value.
	OffsetStart *uint64
	// Offset from etched block at which the rune can no longer be minted. If both HeightEnd and OffsetEnd are set, use the lower value.
	OffsetEnd *uint64
}

type Etching struct {
	// Number of decimals when displaying the rune
	Divisibility *uint8
	// Number of runes to be minted during etching
	Premine *uint128.Uint128
	// Rune name
	Rune *Rune
	// Bitmap of spacers to be displayed between each letter of the rune name
	Spacers *uint32
	// Single Unicode codepoint to represent the rune
	Symbol *rune
	// Minting terms. If not provided, the rune is not mintable.
	Terms *Terms
	// Whether to opt-in to future protocol changes, whatever they may be
	Turbo bool
}

const (
	maxDivisibility uint8  = 38
	maxSpacers      uint32 = 0b00000111_11111111_11111111_11111111
)

func (e Etching) Supply() (uint128.Uint128, error) {
	terms := utils.Default(e.Terms, &Terms{})

	amount := lo.FromPtr(terms.Amount)
	cap := lo.FromPtr(terms.Cap)
	premine := lo.FromPtr(e.Premine)

	result, overflow := amount.MulOverflow(cap)
	if overflow {
		return uint128.Uint128{}, errors.WithStack(errs.OverflowUint128)
	}
	result, overflow = result.AddOverflow(premine)
	if overflow {
		return uint128.Uint128{}, errors.WithStack(errs.OverflowUint128)
	}
	return result, nil
}
