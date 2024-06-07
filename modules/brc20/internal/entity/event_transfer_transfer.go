package entity

import (
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/ordinals"
	"github.com/gaze-network/uint128"
)

type EventTransferTransfer struct {
	Id                uint64
	InscriptionId     ordinals.InscriptionId
	InscriptionNumber uint64
	Tick              string
	OriginalTick      string
	TxHash            chainhash.Hash
	BlockHeight       uint64
	TxIndex           uint32
	Timestamp         time.Time

	FromPkScript   []byte
	FromSatPoint   ordinals.SatPoint
	FromInputIndex uint32
	ToPkScript     []byte
	ToSatPoint     ordinals.SatPoint
	ToOutputIndex  uint32
	Amount         uint128.Uint128
}
