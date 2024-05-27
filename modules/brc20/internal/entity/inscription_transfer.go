package entity

import "github.com/gaze-network/indexer-network/modules/brc20/internal/ordinals"

type InscriptionTransfer struct {
	InscriptionId  ordinals.InscriptionId
	BlockHeight    uint64
	OldSatPoint    ordinals.SatPoint
	NewSatPoint    ordinals.SatPoint
	NewPkScript    []byte
	NewOutputValue uint64
	SentAsFee      bool
}
