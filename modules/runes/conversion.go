package runes

import (
	"github.com/gaze-network/uint128"
)

func unwrapUint128OrDefault(value *uint128.Uint128) uint128.Uint128 {
	if value == nil {
		return uint128.Uint128{}
	}
	return *value
}
