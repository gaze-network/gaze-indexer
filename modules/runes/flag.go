package runes

import (
	"math/big"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
)

// Flag represents a single flag that can be set on a runestone.
type Flag uint8

const (
	FlagEtching = Flag(0)
	FlagTerms   = Flag(1)
)

func (f Flag) Mask() Flags {
	return 1 << f
}

// Flags is a bitmask of flags that can be set on a runestone.
type Flags uint64

func ParseFlags(input interface{}) (Flags, error) {
	switch input := input.(type) {
	case Flags:
		return input, nil
	case int, int8, int16, int32, int64:
		return Flags(input.(int64)), nil
	case uint, uint8, uint16, uint32, uint64:
		return Flags(input.(uint64)), nil
	case big.Int:
		if !input.IsUint64() {
			return 0, errors.WithStack(errs.OverflowUint64)
		}
		return Flags(input.Uint64()), nil
	case *big.Int:
		if !input.IsUint64() {
			return 0, errors.WithStack(errs.OverflowUint64)
		}
		return Flags(input.Uint64()), nil
	default:
		panic("invalid flags input type")
	}
}

func (f Flags) Take(flag Flag) bool {
	found := f&flag.Mask() != 0
	f &= ^flag.Mask()
	return found
}
