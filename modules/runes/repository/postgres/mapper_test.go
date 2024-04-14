package postgres

import (
	"testing"

	"github.com/gaze-network/uint128"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestUint128FromNumeric(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		numeric := pgtype.Numeric{}
		numeric.ScanInt64(pgtype.Int8{
			Int64: 1000,
			Valid: true,
		})

		expected := uint128.From64(1000)

		result, err := uint128FromNumeric(numeric)
		assert.NoError(t, err)
		assert.Equal(t, &expected, result)
	})
	t.Run("nil", func(t *testing.T) {
		numeric := pgtype.Numeric{}
		numeric.ScanInt64(pgtype.Int8{
			Valid: false,
		})

		result, err := uint128FromNumeric(numeric)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestNumericFromUint128(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		u128 := uint128.From64(1)

		expected := pgtype.Numeric{}
		expected.ScanInt64(pgtype.Int8{
			Int64: 1,
			Valid: true,
		})

		result, err := numericFromUint128(&u128)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})
	t.Run("nil", func(t *testing.T) {
		expected := pgtype.Numeric{}
		expected.ScanInt64(pgtype.Int8{
			Valid: false,
		})

		result, err := numericFromUint128(nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})
}
