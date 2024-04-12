package entity

import (
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
	"github.com/gaze-network/uint128"
)

type OutPointBalance struct {
	PkScript []byte
	Id       runes.RuneId
	Value    uint128.Uint128
	Index    uint32
}

type RuneTransaction struct {
	Hash        chainhash.Hash
	BlockHeight uint64
	Timestamp   time.Time
	Inputs      []*OutPointBalance
	Outputs     []*OutPointBalance
	Mints       []*OutPointBalance
	Burns       map[runes.RuneId]uint128.Uint128
	Runestone   *runes.Runestone
}
