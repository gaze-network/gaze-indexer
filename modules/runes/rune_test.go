package runes

import (
	"fmt"
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/Cleverse/go-utilities/utils"
	"github.com/gaze-network/indexer-network/common"
	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	test := func(rune *Rune, encoded string) {
		t.Run(encoded, func(t *testing.T) {
			t.Parallel()
			actualEncoded := rune.String()
			assert.Equal(t, encoded, actualEncoded)

			actualRune, err := NewRuneFromBase26(encoded)
			assert.NoError(t, err)
			assert.Equal(t, rune, actualRune)
		})
	}

	test(NewRune(0), "A")
	test(NewRune(1), "B")
	test(NewRune(2), "C")
	test(NewRune(3), "D")
	test(NewRune(4), "E")
	test(NewRune(5), "F")
	test(NewRune(6), "G")
	test(NewRune(7), "H")
	test(NewRune(8), "I")
	test(NewRune(9), "J")
	test(NewRune(10), "K")
	test(NewRune(11), "L")
	test(NewRune(12), "M")
	test(NewRune(13), "N")
	test(NewRune(14), "O")
	test(NewRune(15), "P")
	test(NewRune(16), "Q")
	test(NewRune(17), "R")
	test(NewRune(18), "S")
	test(NewRune(19), "T")
	test(NewRune(20), "U")
	test(NewRune(21), "V")
	test(NewRune(22), "W")
	test(NewRune(23), "X")
	test(NewRune(24), "Y")
	test(NewRune(25), "Z")
	test(NewRune(26), "AA")
	test(NewRune(27), "AB")
	test(NewRune(51), "AZ")
	test(NewRune(52), "BA")
	test(NewRune(53), "BB")
	test(utils.Must(NewRuneFromString("2055900680524219742")), "UNCOMMONGOODS")
	maxUint128 := new(big.Int).Lsh(big.NewInt(1), 128)
	maxUint128 = maxUint128.Sub(maxUint128, big.NewInt(1))
	test(NewRuneFromBigInt(new(big.Int).Sub(maxUint128, big.NewInt(2))), "BCGDENLQRQWDSLRUGSNLBTMFIJAT")
	test(NewRuneFromBigInt(new(big.Int).Sub(maxUint128, big.NewInt(1))), "BCGDENLQRQWDSLRUGSNLBTMFIJAU")
	test(NewRuneFromBigInt(maxUint128), "BCGDENLQRQWDSLRUGSNLBTMFIJAV")
}

func TestNewRuneFromBase26Error(t *testing.T) {
	_, err := NewRuneFromBase26("?")
	assert.ErrorIs(t, err, ErrInvalidBase26)
}

func TestFirstRuneHeight(t *testing.T) {
	test := func(network common.Network, expected uint64) {
		t.Run(network.String(), func(t *testing.T) {
			t.Parallel()
			actual := FirstRuneHeight(network)
			assert.Equal(t, expected, actual)
		})
	}

	test(common.NetworkMainnet, 840_000)
	test(common.NetworkTestnet, 2_520_000)
}

func TestMinimumRuneAtHeightMainnet(t *testing.T) {
	test := func(height uint64, encoded string) {
		t.Run(fmt.Sprintf("%d", height), func(t *testing.T) {
			t.Parallel()
			rune, err := NewRuneFromBase26(encoded)
			assert.NoError(t, err)
			actual := MinimumRuneAtHeight(common.NetworkMainnet, height)
			assert.Equal(t, (*big.Int)(rune).String(), (*big.Int)(actual).String())
		})
	}

	start := FirstRuneHeight(common.NetworkMainnet)
	end := start + common.HalvingInterval
	interval := uint64(common.HalvingInterval / 12)

	test(0, "AAAAAAAAAAAAA")
	test(start/2, "AAAAAAAAAAAAA")
	test(start, "ZZYZXBRKWXVA")
	test(start+1, "ZZXZUDIVTVQA")
	test(end-1, "A")
	test(end, "A")
	test(end+1, "A")
	test(math.MaxUint32, "A")

	test(start+interval*0-1, "AAAAAAAAAAAAA")
	test(start+interval*0, "ZZYZXBRKWXVA")
	test(start+interval*0+1, "ZZXZUDIVTVQA")

	test(start+interval*1-1, "AAAAAAAAAAAA")
	test(start+interval*1, "ZZYZXBRKWXV")
	test(start+interval*1+1, "ZZXZUDIVTVQ")

	test(start+interval*2-1, "AAAAAAAAAAA")
	test(start+interval*2, "ZZYZXBRKWY")
	test(start+interval*2+1, "ZZXZUDIVTW")

	test(start+interval*3-1, "AAAAAAAAAA")
	test(start+interval*3, "ZZYZXBRKX")
	test(start+interval*3+1, "ZZXZUDIVU")

	test(start+interval*4-1, "AAAAAAAAA")
	test(start+interval*4, "ZZYZXBRL")
	test(start+interval*4+1, "ZZXZUDIW")

	test(start+interval*5-1, "AAAAAAAA")
	test(start+interval*5, "ZZYZXBS")
	test(start+interval*5+1, "ZZXZUDJ")

	test(start+interval*6-1, "AAAAAAA")
	test(start+interval*6, "ZZYZXC")
	test(start+interval*6+1, "ZZXZUE")

	test(start+interval*7-1, "AAAAAA")
	test(start+interval*7, "ZZYZY")
	test(start+interval*7+1, "ZZXZV")

	test(start+interval*8-1, "AAAAA")
	test(start+interval*8, "ZZZA")
	test(start+interval*8+1, "ZZYA")

	test(start+interval*9-1, "AAAA")
	test(start+interval*9, "ZZZ")
	test(start+interval*9+1, "ZZY")

	test(start+interval*10-2, "AAC")
	test(start+interval*10-1, "AAA")
	test(start+interval*10, "AAA")
	test(start+interval*10+1, "AAA")

	test(start+interval*10+interval/2, "NA")

	test(start+interval*11-2, "AB")
	test(start+interval*11-1, "AA")
	test(start+interval*11, "AA")
	test(start+interval*11+1, "AA")

	test(start+interval*11+interval/2, "N")

	test(start+interval*12-2, "B")
	test(start+interval*12-1, "A")
	test(start+interval*12, "A")
	test(start+interval*12+1, "A")
}

func TestMinimumRuneAtHeightTestnet(t *testing.T) {
	test := func(height uint64, runeStr string) {
		t.Run(fmt.Sprintf("%d", height), func(t *testing.T) {
			t.Parallel()
			rune, err := NewRuneFromBase26(runeStr)
			assert.NoError(t, err)
			actual := MinimumRuneAtHeight(common.NetworkTestnet, height)
			assert.Equal(t, rune, actual)
		})
	}

	start := FirstRuneHeight(common.NetworkTestnet)

	test(start-1, "AAAAAAAAAAAAA")
	test(start, "ZZYZXBRKWXVA")
	test(start+1, "ZZXZUDIVTVQA")
}

func TestIsReserved(t *testing.T) {
	test := func(runeStr string, expected bool) {
		t.Run(runeStr, func(t *testing.T) {
			t.Parallel()
			rune, err := NewRuneFromBase26(runeStr)
			assert.NoError(t, err)
			actual := rune.IsReserved()
			assert.Equal(t, expected, actual)
		})
	}

	test("A", false)
	test("B", false)
	test("ZZZZZZZZZZZZZZZZZZZZZZZZZZ", false)
	test("AAAAAAAAAAAAAAAAAAAAAAAAAAA", true)
	test("AAAAAAAAAAAAAAAAAAAAAAAAAAB", true)
	test("BCGDENLQRQWDSLRUGSNLBTMFIJAV", true)
}

func TestGetReservedRune(t *testing.T) {
	test := func(blockHeight uint64, txIndex uint64, expected *Rune) {
		t.Run(fmt.Sprintf("blockHeight_%d_txIndex_%d", blockHeight, txIndex), func(t *testing.T) {
			t.Parallel()
			rune := GetReservedRune(blockHeight, txIndex)
			assert.Equal(t, expected, rune)
		})
	}

	test(0, 0, firstReservedRune)
	test(0, 1, firstReservedRune.Add(big.NewInt(1)))
	test(0, 2, firstReservedRune.Add(big.NewInt(2)))
	test(1, 0, firstReservedRune.Add(new(big.Int).Lsh(big.NewInt(1), 32)))
	test(1, 1, firstReservedRune.Add(new(big.Int).Lsh(big.NewInt(1), 32)).Add(big.NewInt(1)))
	test(1, 2, firstReservedRune.Add(new(big.Int).Lsh(big.NewInt(1), 32)).Add(big.NewInt(2)))
	test(2, 0, firstReservedRune.Add(new(big.Int).Lsh(big.NewInt(2), 32)))
	test(2, 1, firstReservedRune.Add(new(big.Int).Lsh(big.NewInt(2), 32)).Add(big.NewInt(1)))
	test(2, 2, firstReservedRune.Add(new(big.Int).Lsh(big.NewInt(2), 32)).Add(big.NewInt(2)))
	test(math.MaxInt64, 0, firstReservedRune.Add(new(big.Int).Lsh(big.NewInt(math.MaxInt64), 32)))
	test(math.MaxInt64, math.MaxInt64, firstReservedRune.Add(new(big.Int).Lsh(big.NewInt(math.MaxInt64), 32)).Add(big.NewInt(math.MaxInt64)))
}

func TestUnlockSteps(t *testing.T) {
	for i := 0; i < len(unlockSteps); i++ {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			encoded := (*Rune)(unlockSteps[i]).String()
			expected := strings.Repeat("A", i+1)
			assert.Equal(t, expected, encoded)
		})
	}
}

func TestCommitment(t *testing.T) {
	test := func(rune *Rune, expected []byte) {
		t.Run((*big.Int)(rune).String(), func(t *testing.T) {
			t.Parallel()
			actual := rune.Commitment()
			assert.Equal(t, expected, actual)
		})
	}

	test(NewRune(0), []byte{})
	test(NewRune(1), []byte{1})
	test(NewRune(2), []byte{2})
	test(NewRune(255), []byte{255})
	test(NewRune(256), []byte{0, 1})
	test(NewRune(257), []byte{1, 1})
	test(NewRune(65535), []byte{255, 255})
	test(NewRune(65536), []byte{0, 0, 1})
}
