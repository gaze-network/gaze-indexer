package btcutils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSatoshiConversion(t *testing.T) {
	testcases := []struct {
		sats float64
		btc  int64
	}{
		{2.29980951, 229980951},
		{1.29609085, 129609085},
		{1.2768897, 127688970},
		{0.62518296, 62518296},
		{0.29998462, 29998462},
		{0.1251, 12510000},
		{0.02016011, 2016011},
		{0.0198473, 1984730},
		{0.0051711, 517110},
		{0.0012, 120000},
		{7e-05, 7000},
		{3.835e-05, 3835},
		{1.962e-05, 1962},
	}
	for _, testcase := range testcases {
		t.Run(fmt.Sprintf("BtcToSats/%v", testcase.sats), func(t *testing.T) {
			require.NotEqual(t, testcase.btc, int64(testcase.sats*1e8), "Testcase value should have precision error")
			assert.Equal(t, testcase.btc, BitcoinToSatoshi(testcase.sats))
		})
		t.Run(fmt.Sprintf("SatsToBtc/%v", testcase.sats), func(t *testing.T) {
			assert.Equal(t, testcase.sats, SatoshiToBitcoin(testcase.btc))
		})
	}
}
