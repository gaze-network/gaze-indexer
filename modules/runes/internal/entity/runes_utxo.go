package entity

import (
	"github.com/btcsuite/btcd/wire"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/uint128"
)

type RunesUTXOBalance struct {
	RuneId runes.RuneId
	Amount uint128.Uint128
}

type RunesUTXO struct {
	PkScript     []byte
	OutPoint     wire.OutPoint
	RuneBalances []RunesUTXOBalance
}

type RunesUTXOWithSats struct {
	RunesUTXO
	Sats int64
}
