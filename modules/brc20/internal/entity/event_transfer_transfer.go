package entity

import (
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/ordinals"
	"github.com/shopspring/decimal"
)

type EventTransferTransfer struct {
	Id                int64
	InscriptionId     ordinals.InscriptionId
	InscriptionNumber int64
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
	SpentAsFee     bool
	Amount         decimal.Decimal
}