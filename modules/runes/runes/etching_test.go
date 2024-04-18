package runes

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/uint128"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestMaxSpacers(t *testing.T) {
	maxRune := Rune(uint128.Max)
	var sb strings.Builder
	for i, c := range maxRune.String() {
		if i > 0 {
			sb.WriteRune('•')
		}
		sb.WriteRune(c)
	}
	spacedRune, err := NewSpacedRuneFromString(sb.String())
	assert.NoError(t, err)
	assert.Equal(t, maxSpacers, spacedRune.Spacers)
}

func TestSupply(t *testing.T) {
	testNumber := 0
	test := func(e Etching, expectedSupply uint128.Uint128) {
		t.Run(fmt.Sprintf("case_%d", testNumber), func(t *testing.T) {
			t.Parallel()
			actualSupply, err := e.Supply()
			assert.NoError(t, err)
			assert.Equal(t, expectedSupply, actualSupply)
		})
		testNumber++
	}
	testError := func(e Etching, expectedError error) {
		t.Run(fmt.Sprintf("case_%d", testNumber), func(t *testing.T) {
			t.Parallel()
			_, err := e.Supply()
			assert.ErrorIs(t, err, expectedError)
		})
		testNumber++
	}

	test(Etching{}, uint128.From64(0))

	test(Etching{
		Premine: lo.ToPtr(uint128.From64(0)),
		Terms:   nil,
	}, uint128.From64(0))

	test(Etching{
		Premine: lo.ToPtr(uint128.From64(1)),
		Terms:   nil,
	}, uint128.From64(1))

	test(Etching{
		Premine: lo.ToPtr(uint128.From64(1)),
		Terms: &Terms{
			Amount: lo.ToPtr(uint128.From64(0)),
			Cap:    lo.ToPtr(uint128.From64(0)),
		},
	}, uint128.From64(1))

	test(Etching{
		Premine: lo.ToPtr(uint128.From64(1000)),
		Terms: &Terms{
			Amount: lo.ToPtr(uint128.From64(100)),
			Cap:    lo.ToPtr(uint128.From64(10)),
		},
	}, uint128.From64(2000))

	test(Etching{
		Premine: lo.ToPtr(uint128.From64(0)),
		Terms: &Terms{
			Amount: lo.ToPtr(uint128.From64(100)),
			Cap:    lo.ToPtr(uint128.From64(10)),
		},
	}, uint128.From64(1000))

	test(Etching{
		Premine: lo.ToPtr(uint128.From64(1000)),
		Terms: &Terms{
			Amount: lo.ToPtr(uint128.From64(100)),
			Cap:    lo.ToPtr(uint128.From64(0)),
		},
	}, uint128.From64(1000))

	test(Etching{
		Premine: lo.ToPtr(uint128.From64(1000)),
		Terms: &Terms{
			Amount: lo.ToPtr(uint128.From64(0)),
			Cap:    lo.ToPtr(uint128.From64(10)),
		},
	}, uint128.From64(1000))

	test(Etching{
		Premine: lo.ToPtr(uint128.Max.Div64(2).Add64(1)),
		Terms: &Terms{
			Amount: lo.ToPtr(uint128.From64(1)),
			Cap:    lo.ToPtr(uint128.Max.Div64(2)),
		},
	}, uint128.Max)

	test(Etching{
		Premine: lo.ToPtr(uint128.From64(0)),
		Terms: &Terms{
			Amount: lo.ToPtr(uint128.From64(1)),
			Cap:    lo.ToPtr(uint128.Max),
		},
	}, uint128.Max)

	testError(Etching{
		Premine: lo.ToPtr(uint128.Max),
		Terms: &Terms{
			Amount: lo.ToPtr(uint128.From64(1)),
			Cap:    lo.ToPtr(uint128.From64(1)),
		},
	}, errs.OverflowUint128)
}
