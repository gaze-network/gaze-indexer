package runes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRuneIdFromString(t *testing.T) {
	type testcase struct {
		name           string
		input          string
		expectedOutput RuneId
		shouldError    bool
	}
	// TODO: test error instance match expected errors
	testcases := []testcase{
		{
			name:  "valid rune id",
			input: "1:2",
			expectedOutput: RuneId{
				BlockHeight: 1,
				TxIndex:     2,
			},
			shouldError: false,
		},
		{
			name:           "too many separators",
			input:          "1:2:3",
			expectedOutput: RuneId{},
			shouldError:    true,
		},
		{
			name:           "too few separators",
			input:          "1",
			expectedOutput: RuneId{},
			shouldError:    true,
		},
		{
			name:           "invalid tx index",
			input:          "1:a",
			expectedOutput: RuneId{},
			shouldError:    true,
		},
		{
			name:           "invalid block",
			input:          "a:1",
			expectedOutput: RuneId{},
			shouldError:    true,
		},
		{
			name:           "empty tx index",
			input:          "1:",
			expectedOutput: RuneId{},
			shouldError:    true,
		},
		{
			name:           "empty block",
			input:          ":1",
			expectedOutput: RuneId{},
			shouldError:    true,
		},
		{
			name:           "empty block and tx index",
			input:          ":",
			expectedOutput: RuneId{},
			shouldError:    true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			runeId, err := NewRuneIdFromString(tc.input)
			if tc.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedOutput, runeId)
			}
		})
	}
}
