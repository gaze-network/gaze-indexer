package entity

import (
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/ordinals"
)

type OriginOld struct {
	OldSatPoint ordinals.SatPoint
}
type OriginNew struct {
	Cursed         bool
	CursedForBRC20 bool
	Fee            uint64
	Hidden         bool
	Parent         *ordinals.InscriptionId
	Pointer        *uint64
	Reinscription  bool
	Unbound        bool
	Inscription    ordinals.Inscription
}

type Flotsam struct {
	Offset        uint64
	InscriptionId ordinals.InscriptionId
	Tx            *types.Transaction
	OriginOld     *OriginOld // OriginOld and OriginNew are mutually exclusive
	OriginNew     *OriginNew // OriginOld and OriginNew are mutually exclusive
}
