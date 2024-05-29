package brc20

import (
	"github.com/Cleverse/go-utilities/utils"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
)

const (
	ClientVersion    = "v0.0.1"
	DBVersion        = 1
	EventHashVersion = 1
)

type StartingBlockData struct {
	Height int64
	Hash   chainhash.Hash
	Stats  entity.ProcessorStats
}

var startingBlockData = map[common.Network]StartingBlockData{
	common.NetworkMainnet: {
		Height: 779831, // first brc20 inscription
		Hash:   *utils.Must(chainhash.NewHashFromStr("00000000000000000003f079883a81997d3238b287ea53904f1ecd3d1f225209")),
		Stats: entity.ProcessorStats{ // need to seed these stats to keep sequence number the same
			CursedInscriptionCount:  155,
			BlessedInscriptionCount: 348020,
			LostSats:                0, // TODO: need to check lost sats at block 779831
		},
	},
	common.NetworkTestnet: {
		Height: 2413342,
		Hash:   *utils.Must(chainhash.NewHashFromStr("00000000000022e97030b143af785de812f836dd0651b6ac2b7dd9e90dc9abf9")),
		Stats: entity.ProcessorStats{
			CursedInscriptionCount:  0,
			BlessedInscriptionCount: 0,
			LostSats:                0,
		},
	},
}
