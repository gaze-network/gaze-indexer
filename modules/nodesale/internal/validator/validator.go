package validator

import (
	"bytes"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2"
)

type Validator struct {
	Valid  bool
	Reason string
}

func New() *Validator {
	return &Validator{
		Valid: true,
	}
}

func (v *Validator) EqualXonlyPublicKey(target string, expected *btcec.PublicKey) bool {
	if !v.Valid {
		return false
	}
	targetBytes, err := hex.DecodeString(target)
	if err != nil {
		v.Valid = false
		v.Reason = "cannot decode publickey hexstring"
	}

	targetPubKey, err := btcec.ParsePubKey(targetBytes)
	if err != nil {
		v.Valid = false
		v.Reason = "cannot parse public key"
	}
	xOnlyTargetPubKey := btcec.ToSerialized(targetPubKey).SchnorrSerialized()
	xOnlyExpectedPubKey := btcec.ToSerialized(expected).SchnorrSerialized()

	v.Valid = bytes.Equal(xOnlyTargetPubKey[:], xOnlyExpectedPubKey[:])
	if !v.Valid {
		v.Reason = "Invalid public key"
	}
	return v.Valid
}
