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
			t.Parallel()
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

func TestRuneIdMarshal(t *testing.T) {
	runeId := RuneId{
		BlockHeight: 1,
		TxIndex:     2,
	}
	bytes, err := runeId.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, []byte(`"1:2"`), bytes)
}

func TestRuneIdUnmarshal(t *testing.T) {
	str := `"1:2"`
	var runeId RuneId
	err := runeId.UnmarshalJSON([]byte(str))
	assert.NoError(t, err)
	assert.Equal(t, RuneId{
		BlockHeight: 1,
		TxIndex:     2,
	}, runeId)

	str = `1`
	err = runeId.UnmarshalJSON([]byte(str))
	assert.Error(t, err)
}
