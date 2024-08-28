package common

import "github.com/btcsuite/btcd/chaincfg"

type Network string

const (
	NetworkMainnet        Network = "mainnet"
	NetworkTestnet        Network = "testnet"
	NetworkFractalMainnet Network = "fractal-mainnet"
	NetworkFractalTestnet Network = "fractal-testnet"
)

var supportedNetworks = map[Network]struct{}{
	NetworkMainnet:        {},
	NetworkTestnet:        {},
	NetworkFractalMainnet: {},
	NetworkFractalTestnet: {},
}

var chainParams = map[Network]*chaincfg.Params{
	NetworkMainnet:        &chaincfg.MainNetParams,
	NetworkTestnet:        &chaincfg.TestNet3Params,
	NetworkFractalMainnet: &chaincfg.MainNetParams,
	NetworkFractalTestnet: &chaincfg.MainNetParams,
}

func (n Network) IsSupported() bool {
	_, ok := supportedNetworks[n]
	return ok
}

func (n Network) ChainParams() *chaincfg.Params {
	return chainParams[n]
}

func (n Network) String() string {
	return string(n)
}
