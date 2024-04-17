package bitcoin

import (
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/core/types"
)

const (
	Version   = "v0.0.1"
	DBVersion = 1
)

// DefaultCurrentBlockHeight is the default value for the current block height for first time indexing
var defaultCurrentBlock = types.BlockHeader{
	Hash:   common.ZeroHash,
	Height: -1,
}
