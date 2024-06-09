package decimals

import (
	"math/big"
	"reflect"

	"github.com/Cleverse/go-utilities/utils"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/gaze-network/uint128"
	"github.com/holiman/uint256"
	"github.com/shopspring/decimal"
)

const (
	DefaultDivPrecision = 36
)

func init() {
	decimal.DivisionPrecision = DefaultDivPrecision
}

// MustFromString convert string to decimal.Decimal. Panic if error
// string must be a valid number, not NaN, Inf or empty string.
func MustFromString(s string) decimal.Decimal {
	return utils.Must(decimal.NewFromString(s))
}

// ToDecimal convert any type to decimal.Decimal (safety floating point)
func ToDecimal(ivalue any, decimals uint16) decimal.Decimal {
	value := new(big.Int)
	switch v := ivalue.(type) {
	case string:
		value.SetString(v, 10)
	case *big.Int:
		value = v
	case int64:
		value = big.NewInt(v)
	case int, int8, int16, int32:
		rValue := reflect.ValueOf(v)
		value.SetInt64(rValue.Int())
	case uint64:
		value = big.NewInt(0).SetUint64(v)
	case uint, uint8, uint16, uint32:
		rValue := reflect.ValueOf(v)
		value.SetUint64(rValue.Uint())
	case []byte:
		value.SetBytes(v)
	case uint128.Uint128:
		value = v.Big()
	case uint256.Int:
		value = v.ToBig()
	case *uint256.Int:
		value = v.ToBig()
	}
	return decimal.NewFromBigInt(value, -int32(decimals))
}

// ToBigInt convert any type to *big.Int
func ToBigInt(iamount any, decimals uint16) *big.Int {
	amount := decimal.NewFromFloat(0)
	switch v := iamount.(type) {
	case string:
		amount, _ = decimal.NewFromString(v)
	case float64:
		amount = decimal.NewFromFloat(v)
	case float32:
		amount = decimal.NewFromFloat32(v)
	case int64:
		amount = decimal.NewFromInt(v)
	case int, int8, int16, int32:
		rValue := reflect.ValueOf(v)
		amount = decimal.NewFromInt(rValue.Int())
	case decimal.Decimal:
		amount = v
	case *decimal.Decimal:
		amount = *v
	case big.Float:
		amount, _ = decimal.NewFromString(v.String())
	case *big.Float:
		amount, _ = decimal.NewFromString(v.String())
	}
	return amount.Mul(PowerOfTen(decimals)).BigInt()
}

// ToUint256 convert any type to *uint256.Int
func ToUint256(iamount any, decimals uint16) *uint256.Int {
	result := new(uint256.Int)
	if overflow := result.SetFromBig(ToBigInt(iamount, decimals)); overflow {
		logger.Panic("ToUint256 overflow", slogx.Any("amount", iamount), slogx.Uint16("decimals", decimals))
	}
	return result
}
