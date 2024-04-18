package runes

type FlawFlag int

const (
	FlawFlagEdictOutput FlawFlag = iota
	FlawFlagEdictRuneId
	FlawFlagInvalidScript
	FlawFlagOpCode
	FlawFlagSupplyOverflow
	FlawFlagTrailingIntegers
	FlawFlagTruncatedField
	FlawFlagUnrecognizedEvenTag
	FlawFlagUnrecognizedFlag
	FlawFlagVarInt
)

func (f FlawFlag) Mask() Flaws {
	return 1 << f
}

var flawMessages = map[FlawFlag]string{
	FlawFlagEdictOutput:         "edict output greater than transaction output count",
	FlawFlagEdictRuneId:         "invalid runeId in edict",
	FlawFlagInvalidScript:       "invalid script in OP_RETURN",
	FlawFlagOpCode:              "non-pushdata opcode in OP_RETURN",
	FlawFlagSupplyOverflow:      "supply overflows uint128",
	FlawFlagTrailingIntegers:    "trailing integers in body",
	FlawFlagTruncatedField:      "field with missing value",
	FlawFlagUnrecognizedEvenTag: "unrecognized even tag",
	FlawFlagUnrecognizedFlag:    "unrecognized field",
	FlawFlagVarInt:              "invalid varint",
}

func (f FlawFlag) String() string {
	return flawMessages[f]
}

// Flaws is a bitmask of flaws that caused a runestone to be a cenotaph.
type Flaws uint32

func (f Flaws) Collect() []FlawFlag {
	var flags []FlawFlag
	// collect from list of all flags
	for flag := range flawMessages {
		if f&flag.Mask() != 0 {
			flags = append(flags, flag)
		}
	}
	return flags
}

func (f Flaws) CollectAsString() []string {
	flawFlags := f.Collect()
	flawMsgs := make([]string, 0, len(flawFlags))
	for _, flag := range flawFlags {
		flawMsgs = append(flawMsgs, flag.String())
	}
	return flawMsgs
}
