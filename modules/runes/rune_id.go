package runes

import (
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
)

type RuneId struct {
	BlockHeight uint64
	TxIndex     uint64
}

func NewRuneId(blockHeight uint64, txIndex uint64) RuneId {
	return RuneId{
		BlockHeight: blockHeight,
		TxIndex:     txIndex,
	}
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
	txIndex, err := strconv.ParseUint(txIndexStr, 10, 64)
	if err != nil {
		return RuneId{}, errors.WithStack(errors.Join(err, ErrCannotParseTxIndex))
	}
	return RuneId{
		BlockHeight: blockHeight,
		TxIndex:     txIndex,
	}, nil
}

// Delta calculates the delta encoding between two RuneIds. If the two RuneIds are in the same block, then the block delta is 0 and the tx index delta is the difference between the two tx indices.
// If the two RuneIds are in different blocks, then the block delta is the difference between the two block indices and the tx index delta is the tx index in the other block.
func (r RuneId) Delta(next RuneId) (uint64, uint64) {
	blockDelta := next.BlockHeight - r.BlockHeight
	// if the block is the same, then tx index is the difference between the two
	if blockDelta == 0 {
		return 0, next.TxIndex - r.TxIndex
	}
	// otherwise, tx index is the tx index in the next block
	return blockDelta, next.TxIndex
}

// Next calculates the next RuneId given a block delta and tx index delta.
func (r RuneId) Next(blockDelta uint64, txIndexDelta uint64) RuneId {
	if blockDelta == 0 {
		return RuneId{
			BlockHeight: r.BlockHeight,
			TxIndex:     r.TxIndex + txIndexDelta,
		}
	}
	return RuneId{
		BlockHeight: r.BlockHeight + blockDelta,
		TxIndex:     txIndexDelta,
	}
}
