package runes

import (
	"testing"

	"github.com/Cleverse/go-utilities/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewSpacedRuneFromString(t *testing.T) {
	test := func(str string, expected SpacedRune) {
		t.Run(str, func(t *testing.T) {
			t.Parallel()
			actual, err := NewSpacedRuneFromString(str)
			assert.NoError(t, err)
			assert.Equal(t, expected, actual)
		})
	}
	testError := func(str string, expectedError error) {
		t.Run(str, func(t *testing.T) {
			t.Parallel()
			_, err := NewSpacedRuneFromString(str)
			assert.ErrorIs(t, err, expectedError)
		})
	}

	test("A", SpacedRune{Rune: utils.Must(NewRuneFromString("A")), Spacers: 0b0})
	test("AB", SpacedRune{Rune: utils.Must(NewRuneFromString("AB")), Spacers: 0b0})
	test("A•B", SpacedRune{Rune: utils.Must(NewRuneFromString("AB")), Spacers: 0b1})
	test("A•B•C", SpacedRune{Rune: utils.Must(NewRuneFromString("ABC")), Spacers: 0b11})
	test("A•BC", SpacedRune{Rune: utils.Must(NewRuneFromString("ABC")), Spacers: 0b01})
	test("A.B", SpacedRune{Rune: utils.Must(NewRuneFromString("AB")), Spacers: 0b1})
	test("A.B.C", SpacedRune{Rune: utils.Must(NewRuneFromString("ABC")), Spacers: 0b11})
	test("A.BC", SpacedRune{Rune: utils.Must(NewRuneFromString("ABC")), Spacers: 0b01})
	test("UNCOMMON•GOODS", SpacedRune{Rune: utils.Must(NewRuneFromString("UNCOMMONGOODS")), Spacers: 0b10000000})

	testError("•A", ErrLeadingSpacer)
	testError("A•", ErrTrailingSpacer)
	testError("A••B", ErrDoubleSpacer)
	testError("A•B+", ErrInvalidSpacedRuneCharacter)
}

func TestSpacedRuneString(t *testing.T) {
	test := func(spacedRune SpacedRune, expected string) {
		t.Run(expected, func(t *testing.T) {
			t.Parallel()
			actual := spacedRune.String()
			assert.Equal(t, expected, actual)
		})
	}

	test(SpacedRune{Rune: utils.Must(NewRuneFromString("A")), Spacers: 0b0}, "A")
	test(SpacedRune{Rune: utils.Must(NewRuneFromString("AB")), Spacers: 0b0}, "AB")
	test(SpacedRune{Rune: utils.Must(NewRuneFromString("AB")), Spacers: 0b1}, "A•B")
	test(SpacedRune{Rune: utils.Must(NewRuneFromString("ABC")), Spacers: 0b11}, "A•B•C")
	test(SpacedRune{Rune: utils.Must(NewRuneFromString("ABC")), Spacers: 0b01}, "A•BC")
	test(SpacedRune{Rune: utils.Must(NewRuneFromString("UNCOMMONGOODS")), Spacers: 0b10000000}, "UNCOMMON•GOODS")
}
