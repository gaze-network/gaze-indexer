package decimals

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPowerOfTen(t *testing.T) {
	for n := int64(-36); n <= 36; n++ {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			expected := powerOfTenString(n)
			actual := PowerOfTen(n)
			assert.Equal(t, expected, actual.String())
		})
	}
	t.Run("constants", func(t *testing.T) {
		for n, p := range powerOfTen {
			t.Run(p.String(), func(t *testing.T) {
				require.False(t, p.IsZero(), "power of ten must not be zero")
				actual := PowerOfTen(n)
				assert.Equal(t, p, actual)
			})
		}
	})
}

// powerOfTenString add zero padding to power of ten string
func powerOfTenString(n int64) string {
	s := "1"
	if n < 0 {
		for i := int64(0); i < -n-1; i++ {
			s = "0" + s
		}
		s = "0." + s
	} else {
		for i := int64(0); i < n; i++ {
			s = s + "0"
		}
	}
	return s
}
