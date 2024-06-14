package psbtutils

import (
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/samber/lo"
)

func IsReadyPSBT(pC *psbt.Packet, feeRate int64) (bool, error) {
	// if input = output + fee then it's ready

	// Calculate tx fee
	fee, err := TxFee(feeRate, pC)
	if err != nil {
		return false, errors.Wrap(err, "calculate fee")
	}

	// sum total input and output
	totalInputValue := lo.SumBy(pC.Inputs, func(input psbt.PInput) int64 { return input.WitnessUtxo.Value })
	totalOutputValue := lo.SumBy(pC.UnsignedTx.TxOut, func(txout *wire.TxOut) int64 { return txout.Value }) + fee

	// it's perfect match
	if totalInputValue == totalOutputValue {
		return true, nil
	}

	// if input is more than output + fee but not more than 1000 satoshi,
	// then it's ready
	if totalInputValue > totalOutputValue && totalInputValue-totalOutputValue < 1000 {
		return true, nil
	}

	return false, nil
}
