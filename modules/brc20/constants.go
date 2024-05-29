package brc20

import (
	"github.com/Cleverse/go-utilities/utils"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/core/types"
)

const (
	ClientVersion    = "v0.0.1"
	DBVersion        = 1
	EventHashVersion = 1
)

var startingBlockHeader = map[common.Network]types.BlockHeader{
	common.NetworkMainnet: {
		Height: 767430,
		Hash:   *utils.Must(chainhash.NewHashFromStr("00000000000000000003f079883a81997d3238b287ea53904f1ecd3d1f225209")),
		// PrevBlock: *utils.Must(chainhash.NewHashFromStr("00000000000000000001dcce6ce7c8a45872cafd1fb04732b447a14a91832591")),
	},
	common.NetworkTestnet: {
		Height: 2413343,
		Hash:   *utils.Must(chainhash.NewHashFromStr("00000000000022e97030b143af785de812f836dd0651b6ac2b7dd9e90dc9abf9")),
		// PrevBlock: *utils.Must(chainhash.NewHashFromStr("00000000000668f3bafac992f53424774515440cb47e1cb9e73af3f496139e28")),
	},
}
