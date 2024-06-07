package entity

import (
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/ordinals"
	"github.com/gaze-network/uint128"
)

type EventDeploy struct {
	Id                uint64
	InscriptionId     ordinals.InscriptionId
	InscriptionNumber uint64
	Tick              string
	OriginalTick      string
	TxHash            chainhash.Hash
	BlockHeight       uint64
	TxIndex           uint32
	Timestamp         time.Time

	PkScript     []byte
	SatPoint     ordinals.SatPoint
	TotalSupply  uint128.Uint128
	Decimals     uint16
	LimitPerMint uint128.Uint128
	IsSelfMint   bool
}
