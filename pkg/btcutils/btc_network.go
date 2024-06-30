package btcutils

import (
	"github.com/btcsuite/btcd/chaincfg"
)

var supportedNetworks = map[string]*chaincfg.Params{
	"mainnet": &chaincfg.MainNetParams,
	"testnet": &chaincfg.TestNet3Params,
}

// IsSupportedNetwork returns true if the given network is supported.
//
// TODO: create enum for network
func IsSupportedNetwork(network string) bool {
	_, ok := supportedNetworks[network]
	return ok
}

// GetNetParams returns the *chaincfg.Params for the given network.
func GetNetParams(network string) *chaincfg.Params {
	return supportedNetworks[network]
}
