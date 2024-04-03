package runes

import "math/big"

func uint64OrDefault(value *big.Int) (result uint64) {
	if value == nil {
		return
	}
	if value.IsUint64() {
		return value.Uint64()
	}
	return
}
