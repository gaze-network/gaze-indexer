package runes

import (
	"math/big"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/uint128"
)

// Flag represents a single flag that can be set on a runestone.
type Flag uint8

const (
	FlagEtching  = Flag(0)
	FlagTerms    = Flag(1)
	FlagTurbo    = Flag(2)
	FlagCenotaph = Flag(127)
)

func (f Flag) Mask() Flags {
	return Flags(uint128.From64(1).Lsh(uint(f)))
}

// Flags is a bitmask of flags that can be set on a runestone.
type Flags uint128.Uint128

func (f Flags) Uint128() uint128.Uint128 {
	return uint128.Uint128(f)
}

func (f Flags) And(other Flags) Flags {
	return Flags(f.Uint128().And(other.Uint128()))
}

func (f Flags) Or(other Flags) Flags {
	return Flags(f.Uint128().Or(other.Uint128()))
}

func ParseFlags(input interface{}) (Flags, error) {
	switch input := input.(type) {
	case Flags:
		return input, nil
	case uint128.Uint128:
		return Flags(input), nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return Flags(uint128.From64(input.(uint64))), nil
	case big.Int:
		u128, err := uint128.FromBig(&input)
		if err != nil {
			return Flags{}, errors.Join(err, errs.OverflowUint128)
		}
		return Flags(u128), nil
	case *big.Int:
		u128, err := uint128.FromBig(input)
		if err != nil {
			return Flags{}, errors.Join(err, errs.OverflowUint128)
		}
		return Flags(u128), nil
	default:
		logger.Panic("invalid flags input type")
		return Flags{}, nil
	}
}

func (f *Flags) Take(flag Flag) bool {
	found := !f.And(flag.Mask()).Uint128().Equals64(0)
	if found {
		// f = f - (1 << flag)
		*f = Flags(f.Uint128().Sub(flag.Mask().Uint128()))
	}
	return found
}

func (f *Flags) Set(flag Flag) {
	// f = f | (1 << flag)
	*f = Flags(f.Uint128().Or(flag.Mask().Uint128()))
}
