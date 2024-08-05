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
		v.Reason = INVALID_PUBKEY_FORMAT
	}

	targetPubKey, err := btcec.ParsePubKey(targetBytes)
	if err != nil {
		v.Valid = false
		v.Reason = INVALID_PUBKEY_FORMAT
	}
	xOnlyTargetPubKey := btcec.ToSerialized(targetPubKey).SchnorrSerialized()
	xOnlyExpectedPubKey := btcec.ToSerialized(expected).SchnorrSerialized()

	v.Valid = bytes.Equal(xOnlyTargetPubKey[:], xOnlyExpectedPubKey[:])
	if !v.Valid {
		v.Reason = INVALID_PUBKEY
	}
	return v.Valid
}
