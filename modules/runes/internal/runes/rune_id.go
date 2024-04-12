package runes

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
)

type RuneId struct {
	BlockHeight uint64
	TxIndex     uint32
}

var ErrRuneIdZeroBlockNonZeroTxIndex = errors.New("rune id cannot be zero block height and non-zero tx index")

func NewRuneId(blockHeight uint64, txIndex uint32) (RuneId, error) {
	if blockHeight == 0 && txIndex != 0 {
		return RuneId{}, errors.WithStack(ErrRuneIdZeroBlockNonZeroTxIndex)
	}
	return RuneId{
		BlockHeight: blockHeight,
		TxIndex:     txIndex,
	}, nil
}

var (
	ErrInvalidSeparator       = errors.New("invalid rune id: must contain exactly one separator")
	ErrCannotParseBlockHeight = errors.New("invalid rune id: cannot parse block height")
	ErrCannotParseTxIndex     = errors.New("invalid rune id: cannot parse tx index")
)

func NewRuneIdFromString(str string) (RuneId, error) {
	strs := strings.Split(str, ":")
	if len(strs) != 2 {
		return RuneId{}, ErrInvalidSeparator
	}
	blockHeightStr, txIndexStr := strs[0], strs[1]
	blockHeight, err := strconv.ParseUint(blockHeightStr, 10, 64)
	if err != nil {
		return RuneId{}, errors.WithStack(errors.Join(err, ErrCannotParseBlockHeight))
	}
	txIndex, err := strconv.ParseUint(txIndexStr, 10, 32)
	if err != nil {
		return RuneId{}, errors.WithStack(errors.Join(err, ErrCannotParseTxIndex))
	}
	return RuneId{
		BlockHeight: blockHeight,
		TxIndex:     uint32(txIndex),
	}, nil
}

func (r RuneId) String() string {
	return fmt.Sprintf("%d:%d", r.BlockHeight, r.TxIndex)
}

// Delta calculates the delta encoding between two RuneIds. If the two RuneIds are in the same block, then the block delta is 0 and the tx index delta is the difference between the two tx indices.
// If the two RuneIds are in different blocks, then the block delta is the difference between the two block indices and the tx index delta is the tx index in the other block.
func (r RuneId) Delta(next RuneId) (uint64, uint32) {
	blockDelta := next.BlockHeight - r.BlockHeight
	// if the block is the same, then tx index is the difference between the two
	if blockDelta == 0 {
		return 0, next.TxIndex - r.TxIndex
	}
	// otherwise, tx index is the tx index in the next block
	return blockDelta, next.TxIndex
}

// Next calculates the next RuneId given a block delta and tx index delta.
func (r RuneId) Next(blockDelta uint64, txIndexDelta uint32) (RuneId, error) {
	if blockDelta == 0 {
		return NewRuneId(
			r.BlockHeight,
			r.TxIndex+txIndexDelta,
		)
	}
	return NewRuneId(
		r.BlockHeight+blockDelta,
		txIndexDelta,
	)
}

// MarshalJSON implements json.Marshaler
func (r RuneId) MarshalJSON() ([]byte, error) {
	return []byte(`"` + r.String() + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (r *RuneId) UnmarshalJSON(data []byte) error {
	// data must be quoted
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return errors.New("must be string")
	}
	data = data[1 : len(data)-1]
	parsed, err := NewRuneIdFromString(string(data))
	if err != nil {
		return errors.WithStack(err)
	}
	*r = parsed
	return nil
}
