package common

import (
	"github.com/Cleverse/go-utilities/utils"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Zero value of chainhash.Hash
var (
	ZeroHash = *utils.Must(chainhash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000"))
	NullHash = ZeroHash
)
