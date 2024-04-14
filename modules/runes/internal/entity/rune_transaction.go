package entity

import (
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
	"github.com/gaze-network/uint128"
)

type OutPointBalance struct {
	PkScript       []byte
	Id             runes.RuneId
	Amount         uint128.Uint128
	Index          uint32
	PrevTxHash     chainhash.Hash
	PrevTxOutIndex uint32
}

type RuneTransaction struct {
	Hash        chainhash.Hash
	BlockHeight uint64
	Timestamp   time.Time
	Inputs      []*OutPointBalance
	Outputs     []*OutPointBalance
	Mints       map[runes.RuneId]uint128.Uint128
	Burns       map[runes.RuneId]uint128.Uint128
	Runestone   *runes.Runestone
}
