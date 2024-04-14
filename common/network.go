package common

type Network string

const (
	NetworkMainnet Network = "mainnet"
	NetworkTestnet Network = "testnet"
)

var supportedNetworks = map[Network]struct{}{
	NetworkMainnet: {},
	NetworkTestnet: {},
}

func (n Network) IsSupported() bool {
	_, ok := supportedNetworks[n]
	return ok
}

func (n Network) String() string {
	return string(n)
}

// HalvingInterval is the number of blocks between each halving event.
const HalvingInterval = 210_000
