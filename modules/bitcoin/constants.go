package bitcoin

import (
	"github.com/Cleverse/go-utilities/utils"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gaze-network/indexer-network/core/types"
)

const (
	Version   = "v0.0.1"
	DBVersion = 1
)

// DefaultCurrentBlockHeight is the default value for the current block height for first time indexing
var defaultCurrentBlock = types.BlockHeader{
	Hash:   *utils.Must(chainhash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")),
	Height: -1,
}
