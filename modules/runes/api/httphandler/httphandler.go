package httphandler

import (
	"context"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/indexer-network/modules/runes/usecase"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
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

type paginationRequest struct {
	Offset int32 `query:"offset"`
	Limit  int32 `query:"limit"`

	// OrderBy string `query:"orderBy"` // ASC or DESC
	// SortBy  string `query:"sortBy"`  // column name
}

func (req paginationRequest) Validate() error {
	var errList []error

	// this just safeguard for limit,
	// each path should have own validation.
	if req.Limit > 10000 {
		errList = append(errList, errors.Errorf("too large limit"))
	}
	if req.Limit < 0 {
		errList = append(errList, errors.Errorf("limit must be greater than or equal to 0"))
	}
	if req.Offset < 0 {
		errList = append(errList, errors.Errorf("offset must be greater than or equal to 0"))
	}

	// TODO:
	// if req.OrderBy != "" && req.OrderBy != "ASC" && req.OrderBy != "DESC" {
	// 	errList = append(errList, errors.Errorf("invalid orderBy value, must be `ASC` or `DESC`"))
	// }

	return errs.WithPublicMessage(errors.Join(errList...), "pagination validation error")
}

func (req *paginationRequest) ParseDefault() error {
	if req == nil {
		return nil
	}

	if req.Limit == 0 {
		req.Limit = 100
	}

	// TODO:
	// if req.OrderBy == "" {
	// 	req.OrderBy = "ASC"
	// }
	return nil
}

func resolvePkScript(network common.Network, wallet string) ([]byte, bool) {
	if wallet == "" {
		return nil, false
	}
	defaultNet := func() *chaincfg.Params {
		switch network {
		case common.NetworkMainnet:
			return &chaincfg.MainNetParams
		case common.NetworkTestnet:
			return &chaincfg.TestNet3Params
		case common.NetworkFractalMainnet:
			return &chaincfg.MainNetParams
		case common.NetworkFractalTestnet:
			return &chaincfg.MainNetParams
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

// TODO: extract this function somewhere else
// addressFromPkScript returns the address from the given pkScript. If the pkScript is invalid or not standard, it returns empty string.
func addressFromPkScript(pkScript []byte, network common.Network) string {
	_, addrs, _, err := txscript.ExtractPkScriptAddrs(pkScript, network.ChainParams())
	if err != nil {
		logger.Debug("unable to extract address from pkscript", slogx.Error(err))
		return ""
	}
	if len(addrs) != 1 {
		logger.Debug("invalid number of addresses extracted from pkscript. Expected only 1.", slogx.Int("numAddresses", len(addrs)))
		return ""
	}
	return addrs[0].EncodeAddress()
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
		runeId, err := h.usecase.GetRuneIdFromRune(ctx, rune)
		if err != nil {
			return runes.RuneId{}, false
		}
		return runeId, true
	}

	return runes.RuneId{}, false
}

func isRuneIdOrRuneName(id string) bool {
	if _, err := runes.NewRuneIdFromString(id); err == nil {
		return true
	}
	if _, err := runes.NewRuneFromString(id); err == nil {
		return true
	}
	return false
}
