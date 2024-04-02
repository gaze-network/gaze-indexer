package runes

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO: add maxSpacers test

func TestSupply(t *testing.T) {
	testNumber := 0
	test := func(e Etching, expectedSupply *big.Int) {
		t.Run(fmt.Sprintf("case_%d", testNumber), func(t *testing.T) {
			t.Parallel()
			actualSupply := e.Supply()
			assert.Equal(t, expectedSupply, actualSupply)
		})
		testNumber++
	}

	test(Etching{}, big.NewInt(0))

	test(Etching{
		Premine: big.NewInt(0),
		Terms:   nil,
	}, big.NewInt(0))

	test(Etching{
		Premine: big.NewInt(1),
		Terms:   nil,
	}, big.NewInt(1))

	test(Etching{
		Premine: big.NewInt(1),
		Terms: &Terms{
			Amount: nil,
			Cap:    nil,
		},
	}, big.NewInt(1))

	test(Etching{
		Premine: big.NewInt(1000),
		Terms: &Terms{
			Amount: big.NewInt(100),
			Cap:    big.NewInt(10),
		},
	}, big.NewInt(2000))

	test(Etching{
		Premine: nil,
		Terms: &Terms{
			Amount: big.NewInt(100),
			Cap:    big.NewInt(10),
		},
	}, big.NewInt(1000))

	test(Etching{
		Premine: big.NewInt(1000),
		Terms: &Terms{
			Amount: big.NewInt(100),
			Cap:    nil,
		},
	}, big.NewInt(1000))

	test(Etching{
		Premine: big.NewInt(1000),
		Terms: &Terms{
			Amount: nil,
			Cap:    big.NewInt(10),
		},
	}, big.NewInt(1000))
}
