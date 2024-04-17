package common

import "github.com/btcsuite/btcd/chaincfg"

type Network string

const (
	NetworkMainnet Network = "mainnet"
	NetworkTestnet Network = "testnet"
)

var supportedNetworks = map[Network]struct{}{
	NetworkMainnet: {},
	NetworkTestnet: {},
}

var chainParams = map[Network]*chaincfg.Params{
	NetworkMainnet: &chaincfg.MainNetParams,
	NetworkTestnet: &chaincfg.TestNet3Params,
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
