package runes

type FlawFlag int

const (
	FlawFlagEdictOutput FlawFlag = iota
	FlawFlagEdictRuneId
	FlawFlagInvalidScript
	FlawFlagOpCode
	FlawFlagTrailingIntegers
	FlawFlagTruncatedField
	FlawFlagUnrecognizedEvenTag
	FlawFlagUnrecognizedFlag
	FlawFlagVarInt
)

func (f FlawFlag) Mask() Flaws {
	return 1 << f
}

// Flaws is a bitmask of flaws that caused a runestone to be a cenotaph.
type Flaws uint32
