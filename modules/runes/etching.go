package runes

import (
	"math/big"

	"github.com/Cleverse/go-utilities/utils"
)

type Terms struct {
	// Amount of the rune to be minted per transaction
	Amount *big.Int
	// Number of allowed mints
	Cap *big.Int
	// Block height at which the rune can start being minted. If both HeightStart and OffsetStart are set, use the higher value.
	HeightStart uint64
	// Block height at which the rune can no longer be minted. If both HeightEnd and OffsetEnd are set, use the lower value.
	HeightEnd uint64
	// Offset from etched block at which the rune can start being minted. If both HeightStart and OffsetStart are set, use the higher value.
	OffsetStart uint64
	// Offset from etched block at which the rune can no longer be minted. If both HeightEnd and OffsetEnd are set, use the lower value.
	OffsetEnd uint64
}

type Etching struct {
	// Number of runes to be minted during etching
	Premine *big.Int
	// Rune name
	Rune *Rune
	// Minting terms. If not provided, the rune is not mintable.
	Terms *Terms
	// Bitmap of spacers to be displayed between each letter of the rune name
	Spacers uint32
	// Single Unicode codepoint to represent the rune
	Symbol rune
	// Number of decimals when displaying the rune
	Divisibility uint8
}

const (
	maxDivisibility uint8  = 38
	maxSpacers      uint32 = 0b00000111_11111111_11111111_11111111
)

func (e Etching) Supply() *big.Int {
	terms := utils.Default(e.Terms, &Terms{})

	amount := utils.Default(terms.Amount, big.NewInt(0))
	cap := utils.Default(terms.Cap, big.NewInt(0))
	premine := utils.Default(e.Premine, big.NewInt(0))

	result := new(big.Int).Mul(amount, cap)
	return result.Add(result, premine)
}
