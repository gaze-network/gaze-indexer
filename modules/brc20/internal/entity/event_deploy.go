package entity

import (
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/ordinals"
	"github.com/shopspring/decimal"
)

type EventDeploy struct {
	Id                int64
	InscriptionId     ordinals.InscriptionId
	InscriptionNumber int64
	Tick              string
	OriginalTick      string
	TxHash            chainhash.Hash
	BlockHeight       uint64
	TxIndex           uint32
	Timestamp         time.Time

	PkScript     []byte
	SatPoint     ordinals.SatPoint
	TotalSupply  decimal.Decimal
	Decimals     uint16
	LimitPerMint decimal.Decimal
	IsSelfMint   bool
}
