package postgres

import (
	"testing"

	"github.com/gaze-network/uint128"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestUint128FromNumeric(t *testing.T) {
	numeric := pgtype.Numeric{}
	numeric.ScanInt64(pgtype.Int8{
		Int64: 1000,
		Valid: true,
	})

	result, err := uint128FromNumeric(numeric)
	assert.NoError(t, err)
	assert.Equal(t, uint128.From64(1000), result)
}
