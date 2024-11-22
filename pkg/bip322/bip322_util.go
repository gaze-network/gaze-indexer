package bip322

import (
	"bytes"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
)

func SerializeWitnessSignature(witness wire.TxWitness) ([]byte, error) {
	result := new(bytes.Buffer)
	buf := make([]byte, 8)

	if err := wire.WriteVarIntBuf(result, 0, uint64(len(witness)), buf); err != nil {
		return nil, errors.WithStack(err)
	}
	for _, item := range witness {
		if err := wire.WriteVarBytesBuf(result, 0, item, buf); err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return result.Bytes(), nil
}

func DeserializeWitnessSignature(serialized []byte) (wire.TxWitness, error) {
	if len(serialized) == 0 {
		return nil, errors.Wrap(errs.ArgumentRequired, "serialized witness is required")
	}
	witness := make(wire.TxWitness, 0)

	current := 0
	witnessLen := int(serialized[current])
	current++
	for i := 0; i < witnessLen; i++ {
		if current >= len(serialized) {
			return nil, errors.Wrap(errs.InvalidArgument, "invalid serialized witness data: not enough bytes")
		}
		witnessItemLen := int(serialized[current])
		current++
		if current+witnessItemLen > len(serialized) {
			return nil, errors.Wrap(errs.InvalidArgument, "invalid serialized witness data: not enough bytes")
		}
		witnessItem := serialized[current : current+witnessItemLen]
		current += witnessItemLen
		witness = append(witness, witnessItem)
	}
	return witness, nil
}

// PayToTaprootScript creates a pk script for a pay-to-taproot output key.
func PayToTaprootScript(taprootKey *btcec.PublicKey) ([]byte, error) {
	script, err := txscript.NewScriptBuilder().
		AddOp(txscript.OP_1).
		AddData(schnorr.SerializePubKey(taprootKey)).
		Script()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return script, nil
}

// PayToWitnessScript creates a pk script for a pay-to-wpkh output key.
func PayToWitnessScript(pubkey *btcec.PublicKey) ([]byte, error) {
	script, err := txscript.NewScriptBuilder().
		AddOp(txscript.OP_0).
		AddData(btcutil.Hash160(pubkey.SerializeCompressed())).
		Script()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return script, nil
}
