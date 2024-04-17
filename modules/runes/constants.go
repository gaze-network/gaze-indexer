package runes

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

var startingBlockHeader = map[common.Network]types.BlockHeader{
	// TODO: add starting block header for mainnet after block 840,000 is mined
	common.NetworkMainnet: {},
	common.NetworkTestnet: {
		Height:    2583200,
		Hash:      *utils.Must(chainhash.NewHashFromStr("000000000006c5f0dfcd9e0e81f27f97a87aef82087ffe69cd3c390325bb6541")),
		PrevBlock: *utils.Must(chainhash.NewHashFromStr("00000000000668f3bafac992f53424774515440cb47e1cb9e73af3f496139e28")),
	},
}
