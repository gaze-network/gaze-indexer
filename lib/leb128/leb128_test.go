package leb128

import (
	"testing"

	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/uint128"
	"github.com/stretchr/testify/assert"
)

func TestRoundTrip(t *testing.T) {
	test := func(n uint128.Uint128) {
		t.Run(n.String(), func(t *testing.T) {
			t.Parallel()
			encoded := EncodeLEB128(n)
			decoded, length, err := DecodeLEB128(encoded)
			assert.NoError(t, err)
			assert.Equal(t, n, decoded)
			assert.Equal(t, len(encoded), length)
		})
	}

	test(uint128.Zero)
	// powers of two
	for i := 0; i < 128; i++ {
		n := uint128.From64(1)
		n = n.Lsh(uint(i))
		test(n)
	}

	// alternating bits
	n := uint128.Zero
	for i := 0; i < 128; i++ {
		n = n.Lsh(1).Or(uint128.From64(uint64(i % 2)))
		test(n)
	}
}

func TestDecodeError(t *testing.T) {
	testError := func(name string, bytes []byte, expectedError error) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, _, err := DecodeLEB128(bytes)
			if expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, expectedError)
			}
		})
	}

	testError("empty", []byte{}, ErrEmpty)
	testError("unterminated", []byte{0b1000_0000}, ErrUnterminated)

	// may not be longer than 19 bytes
	testError("valid 18 bytes", []byte{
		128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 0,
	}, nil)
	testError("overflow 19 bytes", []byte{
		128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128,
		128, 0,
	}, errs.OverflowUint128)

	// may not overflow uint128
	testError("overflow 1", []byte{
		128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 64,
	}, errs.OverflowUint128)
	testError("overflow 2", []byte{
		128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 32,
	}, errs.OverflowUint128)
	testError("overflow 3", []byte{
		128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 16,
	}, errs.OverflowUint128)
	testError("overflow 4", []byte{
		128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 8,
	}, errs.OverflowUint128)
	testError("overflow 5", []byte{
		128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 4,
	}, errs.OverflowUint128)
	testError("not overflow", []byte{
		128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 128, 2,
	}, nil)
}
