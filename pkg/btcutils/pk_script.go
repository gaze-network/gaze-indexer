package btcutils

import (
	"github.com/Cleverse/go-utilities/utils"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/cockroachdb/errors"
)

// NewPkScript creates a pubkey script(or witness program) from the given address string
//
// see: https://en.bitcoin.it/wiki/Script
func NewPkScript(address string, defaultNet ...*chaincfg.Params) ([]byte, error) {
	net := utils.DefaultOptional(defaultNet, &chaincfg.MainNetParams)
	decoded, _, err := parseAddress(address, net)
	if err != nil {
		return nil, errors.Wrap(err, "can't parse address")
	}
	scriptPubkey, err := txscript.PayToAddrScript(decoded)
	if err != nil {
		return nil, errors.Wrap(err, "can't get script pubkey")
	}
	return scriptPubkey, nil
}

// GetAddressTypeFromPkScript returns the address type from the given pubkey script/script pubkey.
func GetAddressTypeFromPkScript(pkScript []byte, defaultNet ...*chaincfg.Params) (AddressType, error) {
	net := utils.DefaultOptional(defaultNet, &chaincfg.MainNetParams)
	scriptClass, _, _, err := txscript.ExtractPkScriptAddrs(pkScript, net)
	if err != nil {
		return txscript.NonStandardTy, errors.Wrap(err, "can't parse pkScript")
	}
	return scriptClass, nil
}

// ExtractAddressFromPkScript extracts address from the given pubkey script/script pubkey.
// multi-signature script not supported
func ExtractAddressFromPkScript(pkScript []byte, defaultNet ...*chaincfg.Params) (Address, error) {
	net := utils.DefaultOptional(defaultNet, &chaincfg.MainNetParams)
	addrType, addrs, _, err := txscript.ExtractPkScriptAddrs(pkScript, net)
	if err != nil {
		return Address{}, errors.Wrap(err, "can't parse pkScript")
	}
	if len(addrs) == 0 {
		return Address{}, errors.New("can't extract address from pkScript")
	}
	return Address{
		decoded:      addrs[0],
		net:          net,
		encoded:      addrs[0].EncodeAddress(),
		encodedType:  addrType,
		scriptPubKey: pkScript,
	}, nil
}
