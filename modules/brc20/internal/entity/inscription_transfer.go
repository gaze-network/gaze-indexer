package entity

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/ordinals"
)

type InscriptionTransfer struct {
	InscriptionId  ordinals.InscriptionId
	BlockHeight    uint64
	TxIndex        uint32
	TxHash         chainhash.Hash
	Content        []byte
	FromInputIndex uint32
	OldSatPoint    ordinals.SatPoint
	NewSatPoint    ordinals.SatPoint
	NewPkScript    []byte
	NewOutputValue uint64
	SentAsFee      bool
	TransferCount  uint32
}
