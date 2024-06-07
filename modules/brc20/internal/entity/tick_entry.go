package entity

import (
	"time"

	"github.com/gaze-network/indexer-network/modules/brc20/internal/ordinals"
	"github.com/gaze-network/uint128"
)

type TickEntry struct {
	Tick                string
	OriginalTick        string
	TotalSupply         uint128.Uint128
	Decimals            uint16
	LimitPerMint        uint128.Uint128
	IsSelfMint          bool
	DeployInscriptionId ordinals.InscriptionId
	DeployedAt          time.Time
	DeployedAtHeight    uint64

	MintedAmount      uint128.Uint128
	BurnedAmount      uint128.Uint128
	CompletedAt       time.Time
	CompletedAtHeight uint64
}
