package runes

import (
	"math/big"
	"slices"

	"github.com/Cleverse/go-utilities/utils"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/common/errs"
)

type Rune big.Int

func NewRune(value int64) *Rune {
	return (*Rune)(big.NewInt(value))
}

func NewRuneFromBigInt(value *big.Int) *Rune {
	return (*Rune)(new(big.Int).Set(value))
}

var ErrInvalidBase10 = errs.ErrorKind("invalid base-10 character: must be in the range [0-9]")

// NewRuneFromString creates a new Rune from a string of base-10 integer
func NewRuneFromString(value string) (*Rune, error) {
	bi, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return nil, ErrInvalidBase10
	}
	return (*Rune)(bi), nil
}

var ErrInvalidBase26 = errs.ErrorKind("invalid base-26 character: must be in the range [A-Z]")

// NewRuneFromBase26 creates a new Rune from a string of modified base-26 integer
func NewRuneFromBase26(value string) (*Rune, error) {
	x := big.NewInt(0)
	one := big.NewInt(1)
	int26 := big.NewInt(26)
	for i, char := range value {
		if i > 0 {
			x = x.Add(x, one)
		}
		x = x.Mul(x, int26)
		if char < 'A' || char > 'Z' {
			return nil, ErrInvalidBase26
		}
		x = x.Add(x, big.NewInt(int64(char-'A')))
	}
	return (*Rune)(x), nil
}

var firstReservedRune = utils.Must(NewRuneFromString("6402364363415443603228541259936211926"))

var unlockSteps = []*big.Int{
	utils.Must(new(big.Int).SetString("0", 10)),                                       // A
	utils.Must(new(big.Int).SetString("26", 10)),                                      // AA
	utils.Must(new(big.Int).SetString("702", 10)),                                     // AAA
	utils.Must(new(big.Int).SetString("18278", 10)),                                   // AAAA
	utils.Must(new(big.Int).SetString("475254", 10)),                                  // AAAAA
	utils.Must(new(big.Int).SetString("12356630", 10)),                                // AAAAAA
	utils.Must(new(big.Int).SetString("321272406", 10)),                               // AAAAAAA
	utils.Must(new(big.Int).SetString("8353082582", 10)),                              // AAAAAAAA
	utils.Must(new(big.Int).SetString("217180147158", 10)),                            // AAAAAAAAA
	utils.Must(new(big.Int).SetString("5646683826134", 10)),                           // AAAAAAAAAA
	utils.Must(new(big.Int).SetString("146813779479510", 10)),                         // AAAAAAAAAAA
	utils.Must(new(big.Int).SetString("3817158266467286", 10)),                        // AAAAAAAAAAAA
	utils.Must(new(big.Int).SetString("99246114928149462", 10)),                       // AAAAAAAAAAAAA
	utils.Must(new(big.Int).SetString("2580398988131886038", 10)),                     // AAAAAAAAAAAAAA
	utils.Must(new(big.Int).SetString("67090373691429037014", 10)),                    // AAAAAAAAAAAAAAA
	utils.Must(new(big.Int).SetString("1744349715977154962390", 10)),                  // AAAAAAAAAAAAAAAA
	utils.Must(new(big.Int).SetString("45353092615406029022166", 10)),                 // AAAAAAAAAAAAAAAAA
	utils.Must(new(big.Int).SetString("1179180408000556754576342", 10)),               // AAAAAAAAAAAAAAAAAA
	utils.Must(new(big.Int).SetString("30658690608014475618984918", 10)),              // AAAAAAAAAAAAAAAAAAA
	utils.Must(new(big.Int).SetString("797125955808376366093607894", 10)),             // AAAAAAAAAAAAAAAAAAAA
	utils.Must(new(big.Int).SetString("20725274851017785518433805270", 10)),           // AAAAAAAAAAAAAAAAAAAAA
	utils.Must(new(big.Int).SetString("538857146126462423479278937046", 10)),          // AAAAAAAAAAAAAAAAAAAAAA
	utils.Must(new(big.Int).SetString("14010285799288023010461252363222", 10)),        // AAAAAAAAAAAAAAAAAAAAAAA
	utils.Must(new(big.Int).SetString("364267430781488598271992561443798", 10)),       // AAAAAAAAAAAAAAAAAAAAAAAA
	utils.Must(new(big.Int).SetString("9470953200318703555071806597538774", 10)),      // AAAAAAAAAAAAAAAAAAAAAAAAA
	utils.Must(new(big.Int).SetString("246244783208286292431866971536008150", 10)),    // AAAAAAAAAAAAAAAAAAAAAAAAAA
	utils.Must(new(big.Int).SetString("6402364363415443603228541259936211926", 10)),   // AAAAAAAAAAAAAAAAAAAAAAAAAAA
	utils.Must(new(big.Int).SetString("166461473448801533683942072758341510102", 10)), // AAAAAAAAAAAAAAAAAAAAAAAAAAAA
}

func (r Rune) IsReserved() bool {
	return (*big.Int)(&r).Cmp((*big.Int)(firstReservedRune)) >= 0
}

func (r Rune) Add(value *big.Int) *Rune {
	bi := (*big.Int)(&r)
	return (*Rune)(bi.Add(bi, value))
}

// Commitment returns the commitment of the rune. The commitment is the little-endian encoding of the rune.
func (r Rune) Commitment() []byte {
	bytes := (*big.Int)(&r).Bytes()
	slices.Reverse(bytes)
	return bytes
}

// String returns the string representation of the rune in modified base-26 integer
func (r Rune) String() string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	// value = r + 1
	value := new(big.Int).Add((*big.Int)(&r), big.NewInt(1))
	var encoded []byte
	for value.Sign() > 0 {
		// idx = (value - 1) % 26
		idx := new(big.Int).Mod(new(big.Int).Sub(value, big.NewInt(1)), big.NewInt(26)).Int64()
		encoded = append(encoded, chars[idx])
		// value = (value - 1) / 26
		value = new(big.Int).Div(new(big.Int).Sub(value, big.NewInt(1)), big.NewInt(26))
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

func MinimumRuneAtHeight(network common.Network, height uint64) *Rune {
	offset := height + 1
	interval := common.HalvingInterval / 12

	// runes are gradually unlocked from rune activation height until the next halving
	start := FirstRuneHeight(network)
	end := start + common.HalvingInterval

	if offset < start {
		return (*Rune)(unlockSteps[12])
	}
	if offset >= end {
		return (*Rune)(unlockSteps[0])
	}
	progress := offset - start
	length := 12 - progress/uint64(interval)

	startRune := unlockSteps[length]
	endRune := unlockSteps[length-1] // length cannot be 0 because we checked that offset < end
	remainder := big.NewInt(int64(progress) % int64(interval))

	runeRange := new(big.Int).Sub(startRune, endRune)
	result := new(big.Int).Mul(runeRange, remainder)
	result = result.Div(result, big.NewInt(int64(interval)))
	result = result.Sub(startRune, result)
	return (*Rune)(result)
}

func GetReservedRune(blockHeight uint64, txIndex uint64) *Rune {
	// firstReservedRune + ((blockHeight << 32) | txIndex)
	increment := big.NewInt(int64(blockHeight))
	increment = increment.Lsh(increment, 32)
	increment = increment.Or(increment, big.NewInt(int64(txIndex)))
	return firstReservedRune.Add(increment)
}
