package runes

import (
	"math/big"
	"slices"

	"github.com/Cleverse/go-utilities/utils"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/uint128"
)

type Rune uint128.Uint128

func (r Rune) Uint128() uint128.Uint128 {
	return uint128.Uint128(r)
}

func NewRune(value uint64) Rune {
	return Rune(uint128.From64(value))
}

func NewRuneFromUint128(value uint128.Uint128) Rune {
	return Rune(value)
}

func NewRuneFromBigInt(value *big.Int) (Rune, error) {
	rune, err := uint128.FromBig(value)
	if err != nil {
		return Rune{}, errors.WithStack(err)
	}
	return Rune(rune), nil
}

const ErrInvalidBase10 = errs.ErrorKind("invalid base-10 character: must be in the range [0-9]")

// NewRuneFromString creates a new Rune from a string of base-10 integer
func NewRuneFromString(value string) (Rune, error) {
	rune, err := uint128.FromString(value)
	if err != nil {
		return Rune{}, errors.WithStack(err)
	}
	return Rune(rune), nil
}

const ErrInvalidBase26 = errs.ErrorKind("invalid base-26 character: must be in the range [A-Z]")

// NewRuneFromBase26 creates a new Rune from a string of modified base-26 integer
func NewRuneFromBase26(value string) (Rune, error) {
	n := uint128.From64(0)
	for i, char := range value {
		if i > 0 {
			n = n.Add(uint128.From64(1))
		}
		n = n.Mul(uint128.From64(26))
		if char < 'A' || char > 'Z' {
			return Rune{}, ErrInvalidBase26
		}
		n = n.Add(uint128.From64(uint64(char - 'A')))
	}
	return Rune(n), nil
}

var firstReservedRune = utils.Must(NewRuneFromString("6402364363415443603228541259936211926"))

var unlockSteps = []uint128.Uint128{
	utils.Must(uint128.FromString("0")),                                       // A
	utils.Must(uint128.FromString("26")),                                      // AA
	utils.Must(uint128.FromString("702")),                                     // AAA
	utils.Must(uint128.FromString("18278")),                                   // AAAA
	utils.Must(uint128.FromString("475254")),                                  // AAAAA
	utils.Must(uint128.FromString("12356630")),                                // AAAAAA
	utils.Must(uint128.FromString("321272406")),                               // AAAAAAA
	utils.Must(uint128.FromString("8353082582")),                              // AAAAAAAA
	utils.Must(uint128.FromString("217180147158")),                            // AAAAAAAAA
	utils.Must(uint128.FromString("5646683826134")),                           // AAAAAAAAAA
	utils.Must(uint128.FromString("146813779479510")),                         // AAAAAAAAAAA
	utils.Must(uint128.FromString("3817158266467286")),                        // AAAAAAAAAAAA
	utils.Must(uint128.FromString("99246114928149462")),                       // AAAAAAAAAAAAA
	utils.Must(uint128.FromString("2580398988131886038")),                     // AAAAAAAAAAAAAA
	utils.Must(uint128.FromString("67090373691429037014")),                    // AAAAAAAAAAAAAAA
	utils.Must(uint128.FromString("1744349715977154962390")),                  // AAAAAAAAAAAAAAAA
	utils.Must(uint128.FromString("45353092615406029022166")),                 // AAAAAAAAAAAAAAAAA
	utils.Must(uint128.FromString("1179180408000556754576342")),               // AAAAAAAAAAAAAAAAAA
	utils.Must(uint128.FromString("30658690608014475618984918")),              // AAAAAAAAAAAAAAAAAAA
	utils.Must(uint128.FromString("797125955808376366093607894")),             // AAAAAAAAAAAAAAAAAAAA
	utils.Must(uint128.FromString("20725274851017785518433805270")),           // AAAAAAAAAAAAAAAAAAAAA
	utils.Must(uint128.FromString("538857146126462423479278937046")),          // AAAAAAAAAAAAAAAAAAAAAA
	utils.Must(uint128.FromString("14010285799288023010461252363222")),        // AAAAAAAAAAAAAAAAAAAAAAA
	utils.Must(uint128.FromString("364267430781488598271992561443798")),       // AAAAAAAAAAAAAAAAAAAAAAAA
	utils.Must(uint128.FromString("9470953200318703555071806597538774")),      // AAAAAAAAAAAAAAAAAAAAAAAAA
	utils.Must(uint128.FromString("246244783208286292431866971536008150")),    // AAAAAAAAAAAAAAAAAAAAAAAAAA
	utils.Must(uint128.FromString("6402364363415443603228541259936211926")),   // AAAAAAAAAAAAAAAAAAAAAAAAAAA
	utils.Must(uint128.FromString("166461473448801533683942072758341510102")), // AAAAAAAAAAAAAAAAAAAAAAAAAAAA
}

func (r Rune) IsReserved() bool {
	return r.Uint128().Cmp(firstReservedRune.Uint128()) >= 0
}

// Commitment returns the commitment of the rune. The commitment is the little-endian encoding of the rune.
func (r Rune) Commitment() []byte {
	bytes := make([]byte, 16)
	r.Uint128().PutBytes(bytes)
	end := len(bytes)
	for end > 0 && bytes[end-1] == 0 {
		end--
	}
	return bytes[:end]
}

// String returns the string representation of the rune in modified base-26 integer
func (r Rune) String() string {
	if r.Uint128() == uint128.Max {
		return "BCGDENLQRQWDSLRUGSNLBTMFIJAV"
	}

	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	value := r.Uint128().Add64(1)
	var encoded []byte
	for !value.IsZero() {
		idx := value.Sub64(1).Mod64(26)
		encoded = append(encoded, chars[idx])
		value = value.Sub64(1).Div64(26)
	}
	slices.Reverse(encoded)
	return string(encoded)
}

func FirstRuneHeight(network common.Network) uint64 {
	switch network {
	case common.NetworkMainnet:
		return common.HalvingInterval * 4
	case common.NetworkTestnet:
		return common.HalvingInterval * 12
	}
	panic("invalid network")
}

func MinimumRuneAtHeight(network common.Network, height uint64) Rune {
	offset := height + 1
	interval := common.HalvingInterval / 12

	// runes are gradually unlocked from rune activation height until the next halving
	start := FirstRuneHeight(network)
	end := start + common.HalvingInterval

	if offset < start {
		return (Rune)(unlockSteps[12])
	}
	if offset >= end {
		return (Rune)(unlockSteps[0])
	}
	progress := offset - start
	length := 12 - progress/uint64(interval)

	startRune := unlockSteps[length]
	endRune := unlockSteps[length-1] // length cannot be 0 because we checked that offset < end
	remainder := progress % uint64(interval)

	// result = startRune - ((startRune - endRune) * remainder / interval)
	result := startRune.Sub(startRune.Sub(endRune).Mul64(remainder).Div64(uint64(interval)))
	return Rune(result)
}

func GetReservedRune(blockHeight uint64, txIndex uint32) Rune {
	// firstReservedRune + ((blockHeight << 32) | txIndex)
	delta := uint128.From64(blockHeight).Lsh(32).Or64(uint64(txIndex))
	return Rune(firstReservedRune.Uint128().Add(delta))
}
