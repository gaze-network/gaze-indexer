package ordinals

import (
	"testing"

	"github.com/Cleverse/go-utilities/utils"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"
)

func TestNewSatPointFromString(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    SatPoint
		shouldError bool
	}{
		{
			name:  "valid sat point",
			input: "1111111111111111111111111111111111111111111111111111111111111111:1:2",
			expected: SatPoint{
				OutPoint: wire.OutPoint{
					Hash:  *utils.Must(chainhash.NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")),
					Index: 1,
				},
				Offset: 2,
			},
		},
		{
			name:        "error no separator",
			input:       "abc",
			shouldError: true,
		},
		{
			name:        "error invalid output index",
			input:       "abc:xyz",
			shouldError: true,
		},
		{
			name:        "error no offset",
			input:       "1111111111111111111111111111111111111111111111111111111111111111:1",
			shouldError: true,
		},
		{
			name:        "error invalid offset",
			input:       "1111111111111111111111111111111111111111111111111111111111111111:1:foo",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := NewSatPointFromString(tt.input)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}

func TestSatPointString(t *testing.T) {
	tests := []struct {
		name     string
		input    SatPoint
		expected string
	}{
		{
			name: "valid sat point",
			input: SatPoint{
				OutPoint: wire.OutPoint{
					Hash:  *utils.Must(chainhash.NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")),
					Index: 1,
				},
				Offset: 2,
			},
			expected: "1111111111111111111111111111111111111111111111111111111111111111:1:2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.input.String())
		})
	}
}
