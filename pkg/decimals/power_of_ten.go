package decimals

import (
	"github.com/shopspring/decimal"
	"golang.org/x/exp/constraints"
)

// max precision is 36
const (
	minPowerOfTen = -DefaultDivPrecision
	maxPowerOfTen = DefaultDivPrecision
)

var powerOfTen = map[int64]decimal.Decimal{
	minPowerOfTen: MustFromString("0.000000000000000000000000000000000001"),
	-35:           MustFromString("0.00000000000000000000000000000000001"),
	-34:           MustFromString("0.0000000000000000000000000000000001"),
	-33:           MustFromString("0.000000000000000000000000000000001"),
	-32:           MustFromString("0.00000000000000000000000000000001"),
	-31:           MustFromString("0.0000000000000000000000000000001"),
	-30:           MustFromString("0.000000000000000000000000000001"),
	-29:           MustFromString("0.00000000000000000000000000001"),
	-28:           MustFromString("0.0000000000000000000000000001"),
	-27:           MustFromString("0.000000000000000000000000001"),
	-26:           MustFromString("0.00000000000000000000000001"),
	-25:           MustFromString("0.0000000000000000000000001"),
	-24:           MustFromString("0.000000000000000000000001"),
	-23:           MustFromString("0.00000000000000000000001"),
	-22:           MustFromString("0.0000000000000000000001"),
	-21:           MustFromString("0.000000000000000000001"),
	-20:           MustFromString("0.00000000000000000001"),
	-19:           MustFromString("0.0000000000000000001"),
	-18:           MustFromString("0.000000000000000001"),
	-17:           MustFromString("0.00000000000000001"),
	-16:           MustFromString("0.0000000000000001"),
	-15:           MustFromString("0.000000000000001"),
	-14:           MustFromString("0.00000000000001"),
	-13:           MustFromString("0.0000000000001"),
	-12:           MustFromString("0.000000000001"),
	-11:           MustFromString("0.00000000001"),
	-10:           MustFromString("0.0000000001"),
	-9:            MustFromString("0.000000001"),
	-8:            MustFromString("0.00000001"),
	-7:            MustFromString("0.0000001"),
	-6:            MustFromString("0.000001"),
	-5:            MustFromString("0.00001"),
	-4:            MustFromString("0.0001"),
	-3:            MustFromString("0.001"),
	-2:            MustFromString("0.01"),
	-1:            MustFromString("0.1"),
	0:             MustFromString("1"),
	1:             MustFromString("10"),
	2:             MustFromString("100"),
	3:             MustFromString("1000"),
	4:             MustFromString("10000"),
	5:             MustFromString("100000"),
	6:             MustFromString("1000000"),
	7:             MustFromString("10000000"),
	8:             MustFromString("100000000"),
	9:             MustFromString("1000000000"),
	10:            MustFromString("10000000000"),
	11:            MustFromString("100000000000"),
	12:            MustFromString("1000000000000"),
	13:            MustFromString("10000000000000"),
	14:            MustFromString("100000000000000"),
	15:            MustFromString("1000000000000000"),
	16:            MustFromString("10000000000000000"),
	17:            MustFromString("100000000000000000"),
	18:            MustFromString("1000000000000000000"),
	19:            MustFromString("10000000000000000000"),
	20:            MustFromString("100000000000000000000"),
	21:            MustFromString("1000000000000000000000"),
	22:            MustFromString("10000000000000000000000"),
	23:            MustFromString("100000000000000000000000"),
	24:            MustFromString("1000000000000000000000000"),
	25:            MustFromString("10000000000000000000000000"),
	26:            MustFromString("100000000000000000000000000"),
	27:            MustFromString("1000000000000000000000000000"),
	28:            MustFromString("10000000000000000000000000000"),
	29:            MustFromString("100000000000000000000000000000"),
	30:            MustFromString("1000000000000000000000000000000"),
	31:            MustFromString("10000000000000000000000000000000"),
	32:            MustFromString("100000000000000000000000000000000"),
	33:            MustFromString("1000000000000000000000000000000000"),
	34:            MustFromString("10000000000000000000000000000000000"),
	35:            MustFromString("100000000000000000000000000000000000"),
	maxPowerOfTen: MustFromString("1000000000000000000000000000000000000"),
}

// PowerOfTen optimized arithmetic performance for 10^n.
func PowerOfTen[T constraints.Integer](n T) decimal.Decimal {
	nInt64 := int64(n)
	if val, ok := powerOfTen[nInt64]; ok {
		return val
	}
	return powerOfTen[1].Pow(decimal.NewFromInt(nInt64))
}
