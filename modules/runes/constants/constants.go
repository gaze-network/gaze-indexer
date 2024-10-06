package constants

import (
	"fmt"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/core/types"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/uint128"
	"github.com/samber/lo"
)

const (
	Version          = "v0.0.1"
	DBVersion        = 1
	EventHashVersion = 1
)

// starting block heights and hashes should be 1 block before activation block, as indexer will start from the block after this value
var StartingBlockHeader = map[common.Network]types.BlockHeader{
	common.NetworkMainnet: {
		Height: 839999,
	},
	common.NetworkTestnet: {
		Height: 2519999,
	},
	common.NetworkFractalMainnet: {
		Height: 83999,
	},
	common.NetworkFractalTestnet: {
		Height: 83999,
	},
}

type GenesisRuneConfig struct {
	RuneId        runes.RuneId
	Name          string
	Number        uint64
	Divisibility  uint8
	Premine       uint128.Uint128
	SpacedRune    runes.SpacedRune
	Symbol        rune
	Terms         *runes.Terms
	Turbo         bool
	EtchingTxHash chainhash.Hash
	EtchedAt      time.Time
}

var GenesisRuneConfigMap = map[common.Network]GenesisRuneConfig{
	common.NetworkMainnet: {
		RuneId:       runes.RuneId{BlockHeight: 1, TxIndex: 0},
		Number:       0,
		Divisibility: 0,
		Premine:      uint128.Zero,
		SpacedRune:   runes.NewSpacedRune(runes.NewRune(2055900680524219742), 0b10000000),
		Symbol:       '\u29c9',
		Terms: &runes.Terms{
			Amount:      lo.ToPtr(uint128.From64(1)),
			Cap:         &uint128.Max,
			HeightStart: lo.ToPtr(uint64(840000)),
			HeightEnd:   lo.ToPtr(uint64(1050000)),
			OffsetStart: nil,
			OffsetEnd:   nil,
		},
		Turbo:         true,
		EtchingTxHash: chainhash.Hash{},
		EtchedAt:      time.Unix(0, 0),
	},
	common.NetworkFractalMainnet: {
		RuneId:       runes.RuneId{BlockHeight: 1, TxIndex: 0},
		Number:       0,
		Divisibility: 0,
		Premine:      uint128.Zero,
		SpacedRune:   runes.NewSpacedRune(runes.NewRune(2055900680524219742), 0b10000000),
		Symbol:       '\u29c9',
		Terms: &runes.Terms{
			Amount:      lo.ToPtr(uint128.From64(1)),
			Cap:         &uint128.Max,
			HeightStart: lo.ToPtr(uint64(84000)),
			HeightEnd:   lo.ToPtr(uint64(2184000)),
			OffsetStart: nil,
			OffsetEnd:   nil,
		},
		Turbo:         true,
		EtchingTxHash: chainhash.Hash{},
		EtchedAt:      time.Unix(0, 0),
	},
	common.NetworkFractalTestnet: {
		RuneId:       runes.RuneId{BlockHeight: 1, TxIndex: 0},
		Number:       0,
		Divisibility: 0,
		Premine:      uint128.Zero,
		SpacedRune:   runes.NewSpacedRune(runes.NewRune(2055900680524219742), 0b10000000),
		Symbol:       '\u29c9',
		Terms: &runes.Terms{
			Amount:      lo.ToPtr(uint128.From64(1)),
			Cap:         &uint128.Max,
			HeightStart: lo.ToPtr(uint64(84000)),
			HeightEnd:   lo.ToPtr(uint64(2184000)),
			OffsetStart: nil,
			OffsetEnd:   nil,
		},
		Turbo:         true,
		EtchingTxHash: chainhash.Hash{},
		EtchedAt:      time.Unix(0, 0),
	},
}

func NetworkHasGenesisRune(network common.Network) bool {
	switch network {
	case common.NetworkMainnet, common.NetworkFractalMainnet, common.NetworkFractalTestnet:
		return true
	case common.NetworkTestnet:
		return false
	default:
		logger.Panic(fmt.Sprintf("unsupported network: %s", network))
		return false
	}
}
