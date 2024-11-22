package bip322

// This package is forked from https://github.com/unisat-wallet/libbrc20-indexer/blob/v1.1.0/utils/bip322/verify.go,
// with a few modifications to make the interface more friendly with Gaze types.

import (
	"crypto/sha256"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/pkg/btcutils"
)

func GetSha256(data []byte) (hash []byte) {
	sha := sha256.New()
	sha.Write(data[:])
	hash = sha.Sum(nil)
	return
}

func GetTagSha256(data []byte) (hash []byte) {
	tag := []byte("BIP0322-signed-message")
	hashTag := GetSha256(tag)
	var msg []byte
	msg = append(msg, hashTag...)
	msg = append(msg, hashTag...)
	msg = append(msg, data...)
	return GetSha256(msg)
}

func PrepareTx(pkScript []byte, message string) (toSign *wire.MsgTx, err error) {
	// Create a new transaction to spend
	toSpend := wire.NewMsgTx(0)

	// Decode the message hash
	messageHash := GetTagSha256([]byte(message))

	// Create the script for to_spend
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_0)
	builder.AddData(messageHash)
	scriptSig, err := builder.Script()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Create a TxIn with the outpoint 000...000:FFFFFFFF
	prevOutHash, _ := chainhash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	prevOut := wire.NewOutPoint(prevOutHash, wire.MaxPrevOutIndex)
	txIn := wire.NewTxIn(prevOut, scriptSig, nil)
	txIn.Sequence = 0

	toSpend.AddTxIn(txIn)
	toSpend.AddTxOut(wire.NewTxOut(0, pkScript))

	// Create a transaction for to_sign
	toSign = wire.NewMsgTx(0)
	hash := toSpend.TxHash()

	prevOutSpend := wire.NewOutPoint((*chainhash.Hash)(hash.CloneBytes()), 0)

	txSignIn := wire.NewTxIn(prevOutSpend, nil, nil)
	txSignIn.Sequence = 0
	toSign.AddTxIn(txSignIn)

	// Create the script for to_sign
	builderPk := txscript.NewScriptBuilder()
	builderPk.AddOp(txscript.OP_RETURN)
	scriptPk, err := builderPk.Script()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	toSign.AddTxOut(wire.NewTxOut(0, scriptPk))
	return toSign, nil
}

func VerifyMessage(address *btcutils.Address, signature []byte, message string) bool {
	if len(signature) == 0 {
		// empty signature is invalid
		return false
	}

	// BIP322 signature format is the serialized witness of the toSign transaction.
	// [0x02] [SIGNATURE_LEN, ...(signature that go into witness[0])] [PUBLIC_KEY_LEN, ...(public key that was used to sign the message, go to witness[1])]
	witness, err := DeserializeWitnessSignature(signature)
	if err != nil {
		// invalid signature
		return false
	}

	return verifySignatureWitness(witness, address.ScriptPubKey(), message)
}

// verifySignatureWitness
// signature: 64B, pkScript: 33B, message: any
func verifySignatureWitness(witness wire.TxWitness, pkScript []byte, message string) bool {
	toSign, err := PrepareTx(pkScript, message)
	if err != nil {
		return false
	}

	toSign.TxIn[0].Witness = witness
	prevFetcher := txscript.NewCannedPrevOutputFetcher(
		pkScript, 0,
	)
	hashCache := txscript.NewTxSigHashes(toSign, prevFetcher)
	vm, err := txscript.NewEngine(pkScript, toSign, 0, txscript.StandardVerifyFlags, nil, hashCache, 0, prevFetcher)
	if err != nil {
		return false
	}
	if err := vm.Execute(); err != nil {
		return false
	}
	return true
}

func SignMessage(privateKey *btcec.PrivateKey, address *btcutils.Address, message string) ([]byte, error) {
	var witness wire.TxWitness
	var err error
	switch address.Type() {
	case btcutils.AddressP2TR:
		witness, _, err = SignSignatureTaproot(privateKey, message)
	case btcutils.AddressP2WPKH:
		witness, _, err = SignSignatureP2WPKH(privateKey, message)
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}
	signature, err := SerializeWitnessSignature(witness)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return signature, nil
}

func SignSignatureTaproot(privKey *btcec.PrivateKey, message string) (witness wire.TxWitness, pkScript []byte, err error) {
	pubKey := txscript.ComputeTaprootKeyNoScript(privKey.PubKey())

	pkScript, err = PayToTaprootScript(pubKey)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	toSign, err := PrepareTx(pkScript, message)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	prevFetcher := txscript.NewCannedPrevOutputFetcher(
		pkScript, 0,
	)
	sigHashes := txscript.NewTxSigHashes(toSign, prevFetcher)

	witness, err = txscript.TaprootWitnessSignature(
		toSign, sigHashes, 0, 0, pkScript,
		txscript.SigHashDefault, privKey,
	)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	return witness, pkScript, nil
}

func SignSignatureP2WPKH(privKey *btcec.PrivateKey, message string) (witness wire.TxWitness, pkScript []byte, err error) {
	pubKey := privKey.PubKey()
	pkScript, err = PayToWitnessScript(pubKey)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	toSign, err := PrepareTx(pkScript, message)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	prevFetcher := txscript.NewCannedPrevOutputFetcher(
		pkScript, 0,
	)
	sigHashes := txscript.NewTxSigHashes(toSign, prevFetcher)

	witness, err = txscript.WitnessSignature(toSign, sigHashes,
		0, 0, pkScript, txscript.SigHashAll,
		privKey, true)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	return witness, pkScript, nil
}
