package entity

import (
	"github.com/btcsuite/btcd/wire"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/uint128"
)

type OutPointBalance struct {
	RuneId      runes.RuneId
	PkScript    []byte
	OutPoint    wire.OutPoint
	Amount      uint128.Uint128
	BlockHeight uint64
	SpentHeight *uint64
}
