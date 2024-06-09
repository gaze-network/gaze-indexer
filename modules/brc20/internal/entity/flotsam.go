package entity

import (
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/ordinals"
)

type OriginOld struct {
	Content     []byte
	OldSatPoint ordinals.SatPoint
	InputIndex  uint32
}
type OriginNew struct {
	Inscription    ordinals.Inscription
	Parent         *ordinals.InscriptionId
	Pointer        *uint64
	Fee            uint64
	Cursed         bool
	CursedForBRC20 bool
	Hidden         bool
	Reinscription  bool
	Unbound        bool
}

type Flotsam struct {
	Tx            *types.Transaction
	OriginOld     *OriginOld // OriginOld and OriginNew are mutually exclusive
	OriginNew     *OriginNew // OriginOld and OriginNew are mutually exclusive
	Offset        uint64
	InscriptionId ordinals.InscriptionId
}
