package runes

import (
	"math/big"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/uint128"
)

// Tags represent data fields in a runestone. Unrecognized odd tags are ignored. Unrecognized even tags produce a cenotaph.
type Tag uint128.Uint128

func (t Tag) Uint128() uint128.Uint128 {
	return uint128.Uint128(t)
}

var (
	TagBody        = Tag(uint128.From64(0))
	TagFlags       = Tag(uint128.From64(2))
	TagRune        = Tag(uint128.From64(4))
	TagPremine     = Tag(uint128.From64(6))
	TagCap         = Tag(uint128.From64(8))
	TagAmount      = Tag(uint128.From64(10))
	TagHeightStart = Tag(uint128.From64(12))
	TagHeightEnd   = Tag(uint128.From64(14))
	TagOffsetStart = Tag(uint128.From64(16))
	TagOffsetEnd   = Tag(uint128.From64(18))
	TagMint        = Tag(uint128.From64(20))
	TagPointer     = Tag(uint128.From64(22))
	// TagCenotaph is unrecognized
	TagCenotaph = Tag(uint128.From64(126))

	TagDivisibility = Tag(uint128.From64(1))
	TagSpacers      = Tag(uint128.From64(3))
	TagSymbol       = Tag(uint128.From64(5))
	// TagNop is unrecognized
	TagNop = Tag(uint128.From64(127))
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
	case uint128.Uint128:
		return Tag(input), nil
	case int:
		return Tag(uint128.From64(uint64(input))), nil
	case int8:
		return Tag(uint128.From64(uint64(input))), nil
	case int16:
		return Tag(uint128.From64(uint64(input))), nil
	case int32:
		return Tag(uint128.From64(uint64(input))), nil
	case int64:
		return Tag(uint128.From64(uint64(input))), nil
	case uint:
		return Tag(uint128.From64(uint64(input))), nil
	case uint8:
		return Tag(uint128.From64(uint64(input))), nil
	case uint16:
		return Tag(uint128.From64(uint64(input))), nil
	case uint32:
		return Tag(uint128.From64(uint64(input))), nil
	case uint64:
		return Tag(uint128.From64(input)), nil
	case big.Int:
		u128, err := uint128.FromBig(&input)
		if err != nil {
			return Tag{}, errors.Join(err, errs.OverflowUint128)
		}
		return Tag(u128), nil
	case *big.Int:
		u128, err := uint128.FromBig(input)
		if err != nil {
			return Tag{}, errors.Join(err, errs.OverflowUint128)
		}
		return Tag(u128), nil
	default:
		logger.Panic("invalid tag input type")
		return Tag{}, nil
	}
}
