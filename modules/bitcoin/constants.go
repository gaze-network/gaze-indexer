package bitcoin

import (
	"github.com/Cleverse/go-utilities/utils"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/core/types"
)

const (
	Version   = "v0.0.1"
	DBVersion = 1
)

var (
	// defaultCurrentBlockHeight is the default value for the current block height for first time indexing
	defaultCurrentBlock = types.BlockHeader{
		Hash:   common.ZeroHash,
		Height: -1,
	}

	lastV1Block = types.BlockHeader{
		Hash:   *utils.Must(chainhash.NewHashFromStr("00000000000001aa077d7aa84c532a4d69bdbff519609d1da0835261b7a74eb6")),
		Height: 227835,
	}
)
