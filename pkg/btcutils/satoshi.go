package btcutils

import "github.com/shopspring/decimal"

const (
	BitcoinDecimals = 8
)

// satsUnit is 10^8
var satsUnit = decimal.New(1, BitcoinDecimals)

// BitcoinToSatoshi converts a amount in Bitcoin format to Satoshi format.
func BitcoinToSatoshi(v float64) int64 {
	amount := decimal.NewFromFloat(v)
	return amount.Mul(satsUnit).IntPart()
}

// SatoshiToBitcoin converts a amount in Satoshi format to Bitcoin format.
func SatoshiToBitcoin(v int64) float64 {
	return decimal.New(v, -BitcoinDecimals).InexactFloat64()
}
