package entity

import (
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/uint128"
)

type Balance struct {
	PkScript []byte
	Amount   uint128.Uint128
	RuneId   runes.RuneId
	// BlockHeight last updated block height
	BlockHeight uint64
}
