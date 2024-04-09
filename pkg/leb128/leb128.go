package leb128

import (
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/uint128"
)

const (
	ErrEmpty        = errs.ErrorKind("leb128: empty byte sequence")
	ErrUnterminated = errs.ErrorKind("leb128: unterminated byte sequence")
)

func EncodeUint128(input uint128.Uint128) []byte {
	bytes := make([]byte, 0)
	// for n >> 7 > 0
	for !input.Rsh(7).IsZero() {
		last_7_bits := input.And64(0b0111_1111).Uint8()
		bytes = append(bytes, last_7_bits|0b1000_0000)
		input = input.Rsh(7)
	}
	last_byte := input.Uint8()
	bytes = append(bytes, last_byte)
	return bytes
}

func DecodeUint128(data []byte) (n uint128.Uint128, length int, err error) {
	if len(data) == 0 {
		return uint128.Uint128{}, 0, ErrEmpty
	}
	n = uint128.From64(0)

	for i, b := range data {
		if i > 18 {
			return uint128.Uint128{}, 0, errs.OverflowUint128
		}
		value := uint128.New(uint64(b&0b0111_1111), 0)
		if i == 18 && !value.And64(0b0111_1100).IsZero() {
			return uint128.Uint128{}, 0, errs.OverflowUint128
		}
		n = n.Or(value.Lsh(uint(7 * i)))
		// if the high bit is not set, then this is the last byte
		if b&0b1000_0000 == 0 {
			return n, i + 1, nil
		}
	}
	return uint128.Uint128{}, 0, ErrUnterminated
}
