package decimals

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/gaze-network/uint128"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

func TestToDecimal(t *testing.T) {
	t.Run("overflow_decimals", func(t *testing.T) {
		assert.NotPanics(t, func() { ToDecimal(1, math.MaxInt32-1) }, "in-range decimals shouldn't panic")
		assert.NotPanics(t, func() { ToDecimal(1, math.MinInt32+1) }, "in-range decimals shouldn't panic")
		assert.Panics(t, func() { ToDecimal(1, math.MaxInt32+1) }, "out of range decimals should panic")
		assert.Panics(t, func() { ToDecimal(1, math.MinInt32) }, "out of range decimals should panic")
	})
	t.Run("check_supported_types", func(t *testing.T) {
		testcases := []struct {
			decimals uint16
			value    uint64
			expected string
		}{
			{0, 1, "1"},
			{1, 1, "0.1"},
			{2, 1, "0.01"},
			{3, 1, "0.001"},
			{18, 1, "0.000000000000000001"},
			{36, 1, "0.000000000000000000000000000000000001"},
		}
		typesConv := []func(uint64) any{
			func(i uint64) any { return int(i) },
			func(i uint64) any { return int8(i) },
			func(i uint64) any { return int16(i) },
			func(i uint64) any { return int32(i) },
			func(i uint64) any { return int64(i) },
			func(i uint64) any { return uint(i) },
			func(i uint64) any { return uint8(i) },
			func(i uint64) any { return uint16(i) },
			func(i uint64) any { return uint32(i) },
			func(i uint64) any { return uint64(i) },
			func(i uint64) any { return fmt.Sprint(i) },
			func(i uint64) any { return new(big.Int).SetUint64(i) },
			func(i uint64) any { return new(uint128.Uint128).Add64(i) },
			func(i uint64) any { return uint256.NewInt(i) },
		}
		for _, tc := range testcases {
			t.Run(fmt.Sprintf("%d_%d", tc.decimals, tc.value), func(t *testing.T) {
				for _, conv := range typesConv {
					input := conv(tc.value)
					t.Run(fmt.Sprintf("%T", input), func(t *testing.T) {
						actual := ToDecimal(input, tc.decimals)
						assert.Equal(t, tc.expected, actual.String())
					})
				}
			})
		}
	})

	testcases := []struct {
		decimals uint16
		value    interface{}
		expected string
	}{
		{0, uint64(math.MaxUint64), "18446744073709551615"},
		{18, uint64(math.MaxUint64), "18.446744073709551615"},
		{36, uint64(math.MaxUint64), "0.000000000000000018446744073709551615"},
		/* max uint128 */
		{0, uint128.Max, "340282366920938463463374607431768211455"},
		{18, uint128.Max, "340282366920938463463.374607431768211455"},
		{36, uint128.Max, "340.282366920938463463374607431768211455"},
		/* max uint256 */
		{0, new(uint256.Int).SetAllOne(), "115792089237316195423570985008687907853269984665640564039457584007913129639935"},
		{18, new(uint256.Int).SetAllOne(), "115792089237316195423570985008687907853269984665640564039457.584007913129639935"},
		{36, new(uint256.Int).SetAllOne(), "115792089237316195423570985008687907853269.984665640564039457584007913129639935"},
	}
	for _, tc := range testcases {
		t.Run(fmt.Sprintf("%d_%s", tc.decimals, tc.value), func(t *testing.T) {
			actual := ToDecimal(tc.value, tc.decimals)
			assert.Equal(t, tc.expected, actual.String())
		})
	}
}
