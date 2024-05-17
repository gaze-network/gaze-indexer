package ordinals

import (
	"testing"

	"github.com/Cleverse/go-utilities/utils"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/assert"
)

func TestNewInscriptionIdFromString(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    InscriptionId
		shouldError bool
	}{
		{
			name:  "valid inscription id 1",
			input: "1111111111111111111111111111111111111111111111111111111111111111i0",
			expected: InscriptionId{
				TxHash: *utils.Must(chainhash.NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")),
				Index:  0,
			},
		},
		{
			name:  "valid inscription id 2",
			input: "1111111111111111111111111111111111111111111111111111111111111111i1",
			expected: InscriptionId{
				TxHash: *utils.Must(chainhash.NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")),
				Index:  1,
			},
		},
		{
			name:  "valid inscription id 3",
			input: "1111111111111111111111111111111111111111111111111111111111111111i4294967295",
			expected: InscriptionId{
				TxHash: *utils.Must(chainhash.NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")),
				Index:  4294967295,
			},
		},
		{
			name:        "error no separator",
			input:       "abc",
			shouldError: true,
		},
		{
			name:        "error invalid index",
			input:       "xyzixyz",
			shouldError: true,
		},
		{
			name:        "error invalid index",
			input:       "abcixyz",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := NewInscriptionIdFromString(tt.input)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}

func TestInscriptionIdString(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		input    InscriptionId
	}{
		{
			name:     "valid inscription id 1",
			expected: "1111111111111111111111111111111111111111111111111111111111111111i0",
			input: InscriptionId{
				TxHash: *utils.Must(chainhash.NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")),
				Index:  0,
			},
		},
		{
			name:     "valid inscription id 2",
			expected: "1111111111111111111111111111111111111111111111111111111111111111i1",
			input: InscriptionId{
				TxHash: *utils.Must(chainhash.NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")),
				Index:  1,
			},
		},
		{
			name:     "valid inscription id 3",
			expected: "1111111111111111111111111111111111111111111111111111111111111111i4294967295",
			input: InscriptionId{
				TxHash: *utils.Must(chainhash.NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")),
				Index:  4294967295,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.input.String())
		})
	}
}
