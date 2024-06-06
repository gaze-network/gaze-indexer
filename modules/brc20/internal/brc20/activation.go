package brc20

import "github.com/gaze-network/indexer-network/common"

var selfMintActivationHeights = map[common.Network]uint64{
	common.NetworkMainnet: 837090,
	common.NetworkTestnet: 837090,
}

func isSelfMintActivated(height uint64, network common.Network) bool {
	activationHeight, ok := selfMintActivationHeights[network]
	if !ok {
		return false
	}
	return height >= activationHeight
}
