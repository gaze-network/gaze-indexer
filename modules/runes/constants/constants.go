package constants

import (
	"fmt"

	"github.com/Cleverse/go-utilities/utils"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/core/types"
)

const (
	Version          = "v0.0.1"
	DBVersion        = 1
	EventHashVersion = 1
)

var StartingBlockHeader = map[common.Network]types.BlockHeader{
	common.NetworkMainnet: {
		Height: 839999,
		Hash:   *utils.Must(chainhash.NewHashFromStr("0000000000000000000172014ba58d66455762add0512355ad651207918494ab")),
	},
	common.NetworkTestnet: {
		Height: 2519999,
		Hash:   *utils.Must(chainhash.NewHashFromStr("000000000006f45c16402f05d9075db49d3571cf5273cf4cbeaa2aa295f7c833")),
	},
	common.NetworkFractalMainnet: {
		Height: 83999,
		Hash:   *utils.Must(chainhash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")), // TODO: Update this to match real hash
	},
	common.NetworkFractalTestnet: {
		Height: 83999,
		Hash:   *utils.Must(chainhash.NewHashFromStr("00000000000000613ddfbdd1778b17cea3818febcbbf82762eafaa9461038343")),
	},
}

func NetworkHasGenesisRune(network common.Network) bool {
	switch network {
	case common.NetworkMainnet, common.NetworkFractalMainnet, common.NetworkFractalTestnet:
		return true
	case common.NetworkTestnet:
		return false
	default:
		panic(fmt.Sprintf("unsupported network: %s", network))
	}
}
