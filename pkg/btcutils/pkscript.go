package btcutils

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/pkg/errors"
)

// ToPkScript converts a string of address or pkscript to bytes of pkscript
func ToPkScript(network common.Network, from string) ([]byte, error) {
	if from == "" {
		return nil, errors.Wrap(errs.InvalidArgument, "empty input")
	}

	defaultNet, err := func() (*chaincfg.Params, error) {
		switch network {
		case common.NetworkMainnet:
			return &chaincfg.MainNetParams, nil
		case common.NetworkTestnet:
			return &chaincfg.TestNet3Params, nil
		default:
			return nil, errors.Wrap(errs.InvalidArgument, "invalid network")
		}
	}()
	if err != nil {
		return nil, err
	}

	// attempt to parse as address
	address, err := btcutil.DecodeAddress(from, defaultNet)
	if err == nil {
		pkScript, err := txscript.PayToAddrScript(address)
		if err != nil {
			return nil, errors.Wrap(err, "error converting address to pkscript")
		}
		return pkScript, nil
	}

	// attempt to parse as pkscript
	pkScript, err := hex.DecodeString(from)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding pkscript")
	}

	return pkScript, nil
}

// PkScriptToAddress returns the address from the given pkScript. If the pkScript is invalid or not standard, it returns empty string.
func PkScriptToAddress(pkScript []byte, network common.Network) (string, error) {
	_, addrs, _, err := txscript.ExtractPkScriptAddrs(pkScript, network.ChainParams())
	if err != nil {
		return "", errors.Wrap(err, "error extracting addresses from pkscript")
	}
	if len(addrs) != 1 {
		return "", errors.New("invalid number of addresses extracted from pkscript")
	}
	return addrs[0].EncodeAddress(), nil
}
