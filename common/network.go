package common

type Network string

const (
	NetworkMainnet Network = "mainnet"
	NetworkTestnet Network = "testnet"
)

func (n Network) String() string {
	return string(n)
}

// HalvingInterval is the number of blocks between each halving event.
const HalvingInterval = 210_000
