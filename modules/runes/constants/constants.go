package constants

import (
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
		Height: 2583200,
		Hash:   *utils.Must(chainhash.NewHashFromStr("000000000006c5f0dfcd9e0e81f27f97a87aef82087ffe69cd3c390325bb6541")),
	},
	common.NetworkFractalMainnet: {
		Height: 84000,
		Hash:   *utils.Must(chainhash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")), // TODO: Update this to match real hash
	},
	common.NetworkFractalTestnet: {
		Height: 84000,
		Hash:   *utils.Must(chainhash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")),
	},
}
