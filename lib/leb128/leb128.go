package leb128

import (
	"math/big"

	"github.com/gaze-network/indexer-network/common/errs"
)

const ErrUnterminated = errs.ErrorKind("leb128: unterminated byte sequence")

func EncodeLEB128(input *big.Int) []byte {
	n := new(big.Int).Set(input)
	bytes := make([]byte, 0)
	// for n >> 7 > 0
	for new(big.Int).Rsh(n, 7).Sign() > 0 {
		last_7_bits := byte(new(big.Int).And(n, big.NewInt(0b0111_1111)).Int64())
		bytes = append(bytes, last_7_bits|0b1000_0000)
		n = n.Rsh(n, 7)
	}
	last_byte := byte(n.Int64())
	bytes = append(bytes, last_byte)
	return bytes
}

func DecodeLEB128(data []byte) (n *big.Int, length int, err error) {
	n = big.NewInt(0)

	for i, b := range data {
		value := big.NewInt(int64(b & 0b0111_1111))
		n = n.Or(n, value.Lsh(value, uint(7*i)))
		// if the high bit is not set, then this is the last byte
		if b&0b1000_0000 == 0 {
			return n, i + 1, nil
		}
	}
	return nil, 0, ErrUnterminated
}
