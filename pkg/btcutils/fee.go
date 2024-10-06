package btcutils

import (
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
)

// EstimateSignedTxNetworkFee estimates the network fee for the given transaction. "prevTxOuts" should be list of all outputs used as inputs in the transaction.
// If the transaction has unsigned inputs, the fee will be calculated as if those inputs were signed.
func EstimateSignedTxNetworkFee(tx *wire.MsgTx, prevTxOuts []*wire.TxOut, feeRate int64) (int64, error) {
	if len(tx.TxIn) != len(prevTxOuts) {
		return 0, errors.Wrapf(errs.InvalidArgument, "tx.TxIn length (%d) must match prevTxOuts length (%d)", len(tx.TxIn), len(prevTxOuts))
	}
	tx = tx.Copy()
	mockPrivateKey, _ := btcec.NewPrivateKey()
	for i := range tx.TxIn {
		if len(tx.TxIn[i].SignatureScript) > 0 || (len(tx.TxIn[i].Witness) > 0 && len(tx.TxIn[i].Witness[0]) > 0) {
			// already signed, skip
			continue
		}
		address, err := ExtractAddressFromPkScript(prevTxOuts[i].PkScript)
		if err != nil {
			return 0, errors.Wrapf(err, "failed to extract address from pkScript %d", i)
		}
		// if the input is a taproot script-path spend, we need to sign it with the tapscript
		if address.Type() == AddressTaproot && len(tx.TxIn[i].Witness) == 3 {
			tx, err = SignTxInputTapScript(tx, mockPrivateKey, prevTxOuts[i], i)
			if err != nil {
				return 0, errors.Wrapf(err, "failed to sign tx input %d (tapscript)", i)
			}
		} else {
			tx, err = SignTxInput(tx, mockPrivateKey, prevTxOuts[i], i)
			if err != nil {
				return 0, errors.Wrapf(err, "failed to sign tx input %d", i)
			}
		}
	}
	txWeight := blockchain.GetTransactionWeight(btcutil.NewTx(tx))
	txVBytes := calVBytes(txWeight)

	fee := txVBytes * feeRate
	return fee, nil
}

func calVBytes(txWeight int64) int64 {
	// VBytes = txWeight/4, a fraction of Vbyte uses 1 Vbyte.
	txVBytes := txWeight / 4
	if txWeight%4 > 0 {
		txVBytes += 1
	}
	return txVBytes
}
