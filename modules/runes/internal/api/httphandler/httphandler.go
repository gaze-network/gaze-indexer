package httphandler

import (
	"context"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
	"github.com/gaze-network/indexer-network/modules/runes/internal/usecase"
)

type HttpHandler struct {
	usecase *usecase.Usecase
	network common.Network
}

func New(network common.Network, usecase *usecase.Usecase) *HttpHandler {
	return &HttpHandler{
		usecase: usecase,
		network: network,
	}
}

type HttpResponse[T any] struct {
	Error  *string `json:"error"`
	Result *T      `json:"result,omitempty"`
}

func (h *HttpHandler) resolvePkScript(network common.Network, wallet string) ([]byte, bool) {
	if wallet == "" {
		return nil, false
	}
	defaultNet := func() *chaincfg.Params {
		switch network {
		case common.NetworkMainnet:
			return &chaincfg.MainNetParams
		case common.NetworkTestnet:
			return &chaincfg.TestNet3Params
		}
		panic("invalid network")
	}()

	// attempt to parse as address
	address, err := btcutil.DecodeAddress(wallet, defaultNet)
	if err == nil {
		pkScript, err := txscript.PayToAddrScript(address)
		if err != nil {
			return nil, false
		}
		return pkScript, true
	}

	// attempt to parse as pkscript
	pkScript, err := hex.DecodeString(wallet)
	if err != nil {
		return nil, false
	}

	return pkScript, true
}

func (h *HttpHandler) resolveRuneId(ctx context.Context, id string) (runes.RuneId, bool) {
	if id == "" {
		return runes.RuneId{}, false
	}

	// attempt to parse as rune id
	runeId, err := runes.NewRuneIdFromString(id)
	if err == nil {
		return runeId, true
	}

	// attempt to parse as rune
	rune, err := runes.NewRuneFromString(id)
	if err == nil {
		runeEntry, err := h.usecase.GetRuneEntryByRune(ctx, rune)
		if err != nil {
			return runes.RuneId{}, false
		}
		return runeEntry.RuneId, true
	}

	return runes.RuneId{}, false
}
