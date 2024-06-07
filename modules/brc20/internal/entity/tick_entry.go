package entity

import (
	"time"

	"github.com/gaze-network/indexer-network/modules/brc20/internal/ordinals"
	"github.com/shopspring/decimal"
)

type TickEntry struct {
	Tick                string
	OriginalTick        string
	TotalSupply         decimal.Decimal
	Decimals            uint16
	LimitPerMint        decimal.Decimal
	IsSelfMint          bool
	DeployInscriptionId ordinals.InscriptionId
	DeployedAt          time.Time
	DeployedAtHeight    uint64

	MintedAmount      decimal.Decimal
	BurnedAmount      decimal.Decimal
	CompletedAt       time.Time
	CompletedAtHeight uint64
}
