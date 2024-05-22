package btcutils

import (
	"encoding/hex"
	"strings"

	"github.com/Cleverse/go-utilities/utils"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
)

const (
	witnessSeparator = " "
)

// CoinbaseWitness is the witness data for a coinbase transaction.
var CoinbaseWitness = utils.Must(WitnessFromHex([]string{"0000000000000000000000000000000000000000000000000000000000000000"}))

// WitnessToHex formats the passed witness stack as a slice of hex-encoded strings.
func WitnessToHex(witness wire.TxWitness) []string {
	if len(witness) == 0 {
		return nil
	}

	result := make([]string, 0, len(witness))
	for _, wit := range witness {
		result = append(result, hex.EncodeToString(wit))
	}

	return result
}

// WitnessToString formats the passed witness stack as a space-separated string of hex-encoded strings.
func WitnessToString(witness wire.TxWitness) string {
	w := WitnessToHex(witness)
	if w == nil {
		return ""
	}
	return strings.Join(w, witnessSeparator)
}

// WitnessFromHex parses the passed slice of hex-encoded strings into a witness stack.
func WitnessFromHex(witnesses []string) (wire.TxWitness, error) {
	if len(witnesses) == 0 {
		return nil, nil
	}

	result := make(wire.TxWitness, 0, len(witnesses))
	for _, wit := range witnesses {
		decoded, err := hex.DecodeString(wit)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		result = append(result, decoded)
	}

	return result, nil
}

// WitnessFromString parses the passed space-separated string of hex-encoded strings into a witness stack.
func WitnessFromString(witnesses string) (wire.TxWitness, error) {
	return WitnessFromHex(strings.Split(witnesses, witnessSeparator))
}
