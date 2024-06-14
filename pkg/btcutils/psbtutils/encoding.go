package psbtutils

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"

	"github.com/Cleverse/go-utilities/utils"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
)

const (
	// default psbt encoding is hex
	DefaultEncoding = EncodingHex
)

type Encoding string

const (
	EncodingBase64 Encoding = "base64"
	EncodingHex    Encoding = "hex"
)

// DecodeString decodes a psbt hex/base64 string into a psbt.Packet
//
// encoding is optional, default is EncodingHex
func DecodeString(psbtStr string, encoding ...Encoding) (*psbt.Packet, error) {
	pC, err := Decode([]byte(psbtStr), encoding...)
	return pC, errors.WithStack(err)
}

// Decode decodes a psbt hex/base64 byte into a psbt.Packet
//
// encoding is optional, default is EncodingHex
func Decode(psbtB []byte, encoding ...Encoding) (*psbt.Packet, error) {
	enc, ok := utils.Optional(encoding)
	if !ok {
		enc = DefaultEncoding
	}

	var (
		psbtBytes []byte
		err       error
	)

	switch enc {
	case EncodingBase64, "b64":
		psbtBytes = make([]byte, base64.StdEncoding.DecodedLen(len(psbtB)))
		_, err = base64.StdEncoding.Decode(psbtBytes, psbtB)
	case EncodingHex:
		psbtBytes = make([]byte, hex.DecodedLen(len(psbtB)))
		_, err = hex.Decode(psbtBytes, psbtB)
	default:
		return nil, errors.Wrap(errs.Unsupported, "invalid encoding")
	}
	if err != nil {
		return nil, errors.Wrap(err, "can't decode psbt string")
	}

	pC, err := psbt.NewFromRawBytes(bytes.NewReader(psbtBytes), false)
	if err != nil {
		return nil, errors.Wrap(err, "can't create psbt from given psbt")
	}

	return pC, nil
}

// EncodeToString encodes a psbt.Packet into a psbt hex/base64 string
//
// encoding is optional, default is EncodingHex
func EncodeToString(pC *psbt.Packet, encoding ...Encoding) (string, error) {
	enc, ok := utils.Optional(encoding)
	if !ok {
		enc = DefaultEncoding
	}

	var buf bytes.Buffer
	if err := pC.Serialize(&buf); err != nil {
		return "", errors.Wrap(err, "can't serialize psbt")
	}

	switch enc {
	case EncodingBase64, "b64":
		return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
	case EncodingHex:
		return hex.EncodeToString(buf.Bytes()), nil
	default:
		return "", errors.Wrap(errs.Unsupported, "invalid encoding")
	}
}
