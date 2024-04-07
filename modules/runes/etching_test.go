package runes

import (
	"fmt"
	"testing"

	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/uint128"
	"github.com/stretchr/testify/assert"
)

// TODO: add maxSpacers test

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
		Premine: uint128.From64(0),
		Terms:   nil,
	}, uint128.From64(0))

	test(Etching{
		Premine: uint128.From64(1),
		Terms:   nil,
	}, uint128.From64(1))

	test(Etching{
		Premine: uint128.From64(1),
		Terms: &Terms{
			Amount: uint128.From64(0),
			Cap:    uint128.From64(0),
		},
	}, uint128.From64(1))

	test(Etching{
		Premine: uint128.From64(1000),
		Terms: &Terms{
			Amount: uint128.From64(100),
			Cap:    uint128.From64(10),
		},
	}, uint128.From64(2000))

	test(Etching{
		Premine: uint128.From64(0),
		Terms: &Terms{
			Amount: uint128.From64(100),
			Cap:    uint128.From64(10),
		},
	}, uint128.From64(1000))

	test(Etching{
		Premine: uint128.From64(1000),
		Terms: &Terms{
			Amount: uint128.From64(100),
			Cap:    uint128.From64(0),
		},
	}, uint128.From64(1000))

	test(Etching{
		Premine: uint128.From64(1000),
		Terms: &Terms{
			Amount: uint128.From64(0),
			Cap:    uint128.From64(10),
		},
	}, uint128.From64(1000))

	test(Etching{
		Premine: uint128.Max.Div64(2).Add64(1),
		Terms: &Terms{
			Amount: uint128.From64(1),
			Cap:    uint128.Max.Div64(2),
		},
	}, uint128.Max)

	test(Etching{
		Premine: uint128.From64(0),
		Terms: &Terms{
			Amount: uint128.From64(1),
			Cap:    uint128.Max,
		},
	}, uint128.Max)

	testError(Etching{
		Premine: uint128.Max,
		Terms: &Terms{
			Amount: uint128.From64(1),
			Cap:    uint128.From64(1),
		},
	}, errs.OverflowUint128)
}
