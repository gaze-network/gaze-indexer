package btcutils

import (
	"github.com/Cleverse/go-utilities/utils"
	verifier "github.com/bitonicnl/verify-signed-message/pkg"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
)

func VerifySignature(address string, message string, sigBase64 string, defaultNet ...*chaincfg.Params) error {
	net := utils.DefaultOptional(defaultNet, &chaincfg.MainNetParams)
	_, err := verifier.VerifyWithChain(verifier.SignedMessage{
		Address:   address,
		Message:   message,
		Signature: sigBase64,
	}, net)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func SignTxInput(tx *wire.MsgTx, privateKey *btcec.PrivateKey, prevTxOut *wire.TxOut, inputIndex int) (*wire.MsgTx, error) {
	if privateKey == nil {
		return nil, errors.Wrap(errs.InvalidArgument, "PrivateKey is required")
	}
	if tx == nil {
		return nil, errors.Wrap(errs.InvalidArgument, "Tx is required")
	}
	if prevTxOut == nil {
		return nil, errors.Wrap(errs.InvalidArgument, "PrevTxOut is required")
	}

	prevOutFetcher := txscript.NewCannedPrevOutputFetcher(prevTxOut.PkScript, prevTxOut.Value)
	sigHashes := txscript.NewTxSigHashes(tx, prevOutFetcher)
	if len(tx.TxIn) <= inputIndex {
		return nil, errors.Errorf("input to sign (%d) is out of range", inputIndex)
	}
	address, err := ExtractAddressFromPkScript(prevTxOut.PkScript)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract address")
	}

	switch address.Type() {
	case AddressP2TR:
		witness, err := txscript.TaprootWitnessSignature(
			tx,
			sigHashes,
			inputIndex,
			prevTxOut.Value,
			prevTxOut.PkScript,
			txscript.SigHashAll|txscript.SigHashAnyOneCanPay,
			privateKey)
		if err != nil {
			return nil, errors.Wrap(err, "failed to sign")
		}
		tx.TxIn[inputIndex].Witness = witness
	case AddressP2WPKH:
		witness, err := txscript.WitnessSignature(
			tx,
			sigHashes,
			inputIndex,
			prevTxOut.Value,
			prevTxOut.PkScript,
			txscript.SigHashAll|txscript.SigHashAnyOneCanPay,
			privateKey,
			true,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to sign")
		}
		tx.TxIn[inputIndex].Witness = witness
	case AddressP2PKH:
		sigScript, err := txscript.SignatureScript(
			tx,
			inputIndex,
			prevTxOut.PkScript,
			txscript.SigHashAll|txscript.SigHashAnyOneCanPay,
			privateKey,
			true,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to sign")
		}
		tx.TxIn[inputIndex].SignatureScript = sigScript
	default:
		return nil, errors.Wrapf(errs.NotSupported, "unsupported input address type %s", address.Type())
	}

	return tx, nil
}

func SignTxInputTapScript(tx *wire.MsgTx, privateKey *btcec.PrivateKey, prevTxOut *wire.TxOut, inputIndex int) (*wire.MsgTx, error) {
	if privateKey == nil {
		return nil, errors.Wrap(errs.InvalidArgument, "PrivateKey is required")
	}
	if tx == nil {
		return nil, errors.Wrap(errs.InvalidArgument, "Tx is required")
	}
	if prevTxOut == nil {
		return nil, errors.Wrap(errs.InvalidArgument, "PrevTxOut is required")
	}

	prevOutFetcher := txscript.NewCannedPrevOutputFetcher(prevTxOut.PkScript, prevTxOut.Value)
	sigHashes := txscript.NewTxSigHashes(tx, prevOutFetcher)
	if len(tx.TxIn) <= inputIndex {
		return nil, errors.Errorf("input to sign (%d) is out of range", inputIndex)
	}
	address, err := ExtractAddressFromPkScript(prevTxOut.PkScript)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract address")
	}

	if address.Type() != AddressTaproot {
		return nil, errors.Errorf("input type must be %s", AddressTaproot)
	}

	witness := tx.TxIn[inputIndex].Witness
	if len(witness) != 3 {
		return nil, errors.Wrapf(errs.InvalidArgument, "invalid witness length: expected 3, got %d", len(witness))
	}

	tapLeaf := txscript.NewBaseTapLeaf(witness[1])
	signature, err := txscript.RawTxInTapscriptSignature(
		tx,
		sigHashes,
		inputIndex,
		prevTxOut.Value,
		prevTxOut.PkScript,
		tapLeaf,
		txscript.SigHashAll|txscript.SigHashAnyOneCanPay,
		privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign")
	}
	tx.TxIn[inputIndex].Witness[0] = signature

	return tx, nil
}
