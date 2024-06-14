package psbtutils

import (
	"math"

	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/pkg/btcutils"
)

// TxFee returns satoshis fee of a transaction given the fee rate (sat/vB)
// and the number of inputs and outputs.
func TxFee(feeRate int64, p *psbt.Packet) (int64, error) {
	size, err := PSBTSize(p)
	if err != nil {
		return 0, errors.Wrap(err, "psbt size")
	}
	return int64(math.Ceil(size * float64(feeRate))), nil
}

func PredictTxFee(feeRate int64, inputs, outputs int) int64 {
	/**
	TODO: handle edge cases like:
		1. when we predict that we need to use unnecessary UTXOs
		2. when we predict that we need to use more value than user have, but user do have enough for the actual transaction

	Idea for solving this:
		- When trying to find the best UTXOs to use, we:
			- Will not reject when user's balance is not enough, instead we will return all UTXOs even if it's not enough.
			- Will be okay returning excessive UTXOs (say we predict we need 10K satoshis, but actually we only need 5K satoshis, then we will return UTXOs enough for 10K satoshis)
		- And then we:
			- Construct the actual PSBT, then select UTXOs to use accordingly,
			- If the user's balance is not enough, then we will return an error,
			- Or if when we predict we expect to use more UTXOs than the actual transaction, then we will just use what's needed.
	*/
	size := defaultOverhead + 148*float64(inputs) + 43*float64(outputs)
	return int64(math.Ceil(size * float64(feeRate)))
}

type txSize struct {
	Overhead float64
	Inputs   float64
	Outputs  float64
}

const defaultOverhead = 10.5

// Transaction Virtual Sizes Bytes
//
// Reference: https://bitcoinops.org/en/tools/calc-size/
var txSizes = map[btcutils.TransactionType]txSize{
	btcutils.TransactionP2WPKH: {
		Inputs:  68,
		Outputs: 31,
	},
	btcutils.TransactionP2TR: {
		Inputs:  57.5,
		Outputs: 43,
	},
	btcutils.TransactionP2SH: {
		Inputs:  91,
		Outputs: 32,
	},
	btcutils.TransactionP2PKH: {
		Inputs:  148,
		Outputs: 34,
	},
	btcutils.TransactionP2WSH: {
		Inputs:  104.5,
		Outputs: 43,
	},
}

func PSBTSize(psbt *psbt.Packet) (float64, error) {
	if err := psbt.SanityCheck(); err != nil {
		return 0, errors.Wrap(errors.Join(err, errs.InvalidArgument), "psbt sanity check")
	}

	inputs := map[btcutils.TransactionType]int{}
	outputs := map[btcutils.TransactionType]int{}

	for _, input := range psbt.Inputs {
		addrType, err := btcutils.GetAddressTypeFromPkScript(input.WitnessUtxo.PkScript)
		if err != nil {
			return 0, errors.Wrap(err, "get address type from pk script")
		}
		inputs[addrType]++
	}

	for _, output := range psbt.UnsignedTx.TxOut {
		addrType, err := btcutils.GetAddressTypeFromPkScript(output.PkScript)
		if err != nil {
			return 0, errors.Wrap(err, "get address type from pk script")
		}
		outputs[addrType]++
	}

	totalSize := defaultOverhead
	for txType, txSizeData := range txSizes {
		if inputCount, ok := inputs[txType]; ok {
			totalSize += txSizeData.Inputs * float64(inputCount)
		}
		if outputCount, ok := outputs[txType]; ok {
			totalSize += txSizeData.Outputs * float64(outputCount)
		}
	}

	return totalSize, nil
}
