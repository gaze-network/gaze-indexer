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

	fistV2Block = types.BlockHeader{
		Hash:   *utils.Must(chainhash.NewHashFromStr("00000000000000d0dfd4c9d588d325dce4f32c1b31b7c0064cba7025a9b9adcc")),
		Height: 227836,
	}
)
