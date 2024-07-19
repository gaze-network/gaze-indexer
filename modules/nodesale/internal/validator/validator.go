package validator

import (
	"bytes"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/cockroachdb/errors"
)

type Validator struct {
	Valid bool
}

func New() *Validator {
	return &Validator{
		Valid: true,
	}
}

func (v *Validator) EqualXonlyPublicKey(target string, expected *btcec.PublicKey) (bool, error) {
	if !v.Valid {
		return false, nil
	}
	targetBytes, err := hex.DecodeString(target)
	if err != nil {
		v.Valid = false
		return v.Valid, errors.Wrap(err, "cannot decode hexstring")
	}

	targetPubKey, err := btcec.ParsePubKey(targetBytes)
	if err != nil {
		v.Valid = false
		return v.Valid, errors.Wrap(err, "cannot parse public key")
	}
	xOnlyTargetPubKey := btcec.ToSerialized(targetPubKey).SchnorrSerialized()
	xOnlyExpectedPubKey := btcec.ToSerialized(expected).SchnorrSerialized()

	v.Valid = bytes.Equal(xOnlyTargetPubKey[:], xOnlyExpectedPubKey[:])
	return v.Valid, nil
}
