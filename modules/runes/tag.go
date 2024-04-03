package runes

import (
	"math/big"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
)

// Tags represent data fields in a runestone. Unrecognized odd tags are ignored. Unrecognized even tags produce a cenotaph.
type Tag uint8

const (
	TagBody        = Tag(0)
	TagFlags       = Tag(2)
	TagRune        = Tag(4)
	TagPremine     = Tag(6)
	TagCap         = Tag(8)
	TagAmount      = Tag(10)
	TagHeightStart = Tag(12)
	TagHeightEnd   = Tag(14)
	TagOffsetStart = Tag(16)
	TagOffsetEnd   = Tag(18)
	TagMint        = Tag(20)
	TagPointer     = Tag(22)
	// TagCenotaph is unrecognized
	TagCenotaph = Tag(126)
	TagInvalid  = Tag(128)

	TagDivisibility = Tag(1)
	TagSpacers      = Tag(3)
	TagSymbol       = Tag(5)
	// TagNop is unrecognized
	TagNop = Tag(127)
)

var allTags = map[Tag]struct{}{
	TagBody:         {},
	TagFlags:        {},
	TagRune:         {},
	TagPremine:      {},
	TagCap:          {},
	TagAmount:       {},
	TagHeightStart:  {},
	TagHeightEnd:    {},
	TagOffsetStart:  {},
	TagOffsetEnd:    {},
	TagMint:         {},
	TagPointer:      {},
	TagCenotaph:     {},
	TagDivisibility: {},
	TagSpacers:      {},
	TagSymbol:       {},
	TagNop:          {},
}

func (t Tag) IsValid() bool {
	_, ok := allTags[t]
	return ok
}

func ParseTag(input interface{}) (Tag, error) {
	switch input := input.(type) {
	case Tag:
		return input, nil
	case int, int8, int16, int32, int64:
		return Tag(input.(int64)), nil
	case uint, uint8, uint16, uint32, uint64:
		return Tag(input.(uint64)), nil
	case big.Int:
		if !input.IsUint64() {
			return 0, errors.WithStack(errs.OverflowUint64)
		}
		return Tag(input.Uint64()), nil
	case *big.Int:
		if !input.IsUint64() {
			return 0, errors.WithStack(errs.OverflowUint64)
		}
		return Tag(input.Uint64()), nil
	default:
		panic("invalid tag input type")
	}
}
