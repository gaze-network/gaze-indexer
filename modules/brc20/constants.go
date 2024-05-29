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
		Height: 767429,
		Hash:   *utils.Must(chainhash.NewHashFromStr("00000000000000000002b35aef66eb15cd2b232a800f75a2f25cedca4cfe52c4")),
	},
	common.NetworkTestnet: {
		Height: 2413342,
		Hash:   *utils.Must(chainhash.NewHashFromStr("00000000000022e97030b143af785de812f836dd0651b6ac2b7dd9e90dc9abf9")),
	},
}
