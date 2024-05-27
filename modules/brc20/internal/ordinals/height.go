package ordinals

import "github.com/gaze-network/indexer-network/common"

func GetJubileeHeight(network common.Network) uint64 {
	switch network {
	case common.NetworkMainnet:
		return 824544
	case common.NetworkTestnet:
		return 2544192
	}
	panic("unsupported network")
}
