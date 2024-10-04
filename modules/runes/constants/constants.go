package constants

import (
	"fmt"
	"time"

	"github.com/Cleverse/go-utilities/utils"
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

var StartingBlockHeader = map[common.Network]types.BlockHeader{
	common.NetworkMainnet: {
		Height: 839999,
		Hash:   *utils.Must(chainhash.NewHashFromStr("0000000000000000000172014ba58d66455762add0512355ad651207918494ab")),
	},
	common.NetworkTestnet: {
		Height: 2519999,
		Hash:   *utils.Must(chainhash.NewHashFromStr("000000000006f45c16402f05d9075db49d3571cf5273cf4cbeaa2aa295f7c833")),
	},
	common.NetworkFractalMainnet: {
		Height: 83999,
		Hash:   *utils.Must(chainhash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")), // TODO: Update this to match real hash
	},
	common.NetworkFractalTestnet: {
		Height: 83999,
		Hash:   *utils.Must(chainhash.NewHashFromStr("00000000000000613ddfbdd1778b17cea3818febcbbf82762eafaa9461038343")),
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
	EtchingBlock  uint64
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
		EtchingBlock:  1,
		EtchingTxHash: chainhash.Hash{},
		EtchedAt:      time.Time{},
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
		EtchingBlock:  1,
		EtchingTxHash: chainhash.Hash{},
		EtchedAt:      time.Time{},
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
		EtchingBlock:  1,
		EtchingTxHash: chainhash.Hash{},
		EtchedAt:      time.Time{},
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
