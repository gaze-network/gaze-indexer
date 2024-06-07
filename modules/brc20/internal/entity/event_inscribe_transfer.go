package entity

import (
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/ordinals"
	"github.com/shopspring/decimal"
)

type EventInscribeTransfer struct {
	Id                uint64
	InscriptionId     ordinals.InscriptionId
	InscriptionNumber uint64
	Tick              string
	OriginalTick      string
	TxHash            chainhash.Hash
	BlockHeight       uint64
	TxIndex           uint32
	Timestamp         time.Time

	PkScript    []byte
	SatPoint    ordinals.SatPoint
	OutputIndex uint32
	SatsAmount  uint64
	Amount      decimal.Decimal
}
