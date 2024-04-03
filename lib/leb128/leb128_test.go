package leb128

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoundTrip(t *testing.T) {
	test := func(n *big.Int) {
		t.Run(n.String(), func(t *testing.T) {
			t.Parallel()
			encoded := EncodeLEB128(n)
			decoded, length, err := DecodeLEB128(encoded)
			assert.NoError(t, err)
			assert.Equal(t, n, decoded)
			assert.Equal(t, len(encoded), length)
		})
	}

	test(big.NewInt(0))
	// powers of two
	for i := 0; i < 128; i++ {
		n := big.NewInt(1)
		n = n.Lsh(n, uint(i))
		test(n)
	}
	// alternating bits
	n := big.NewInt(0)
	for i := 0; i < 128; i++ {
		n = n.Lsh(n, 1)
		n = n.Or(n, big.NewInt(int64(i%2)))
		test(n)
	}
}

func TestDecodeUnterminated(t *testing.T) {
	_, _, err := DecodeLEB128([]byte{0b1000_0000})
	assert.ErrorIs(t, err, ErrUnterminated)
}
