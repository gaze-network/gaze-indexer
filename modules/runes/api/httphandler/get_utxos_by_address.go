package httphandler

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/uint128"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
)

type getUTXOsRequest struct {
	paginationRequest
	Wallet      string `params:"wallet"`
	Id          string `query:"id"`
	BlockHeight uint64 `query:"blockHeight"`
}

const (
	getUTXOsMaxLimit = 3000
)

func (r getUTXOsRequest) Validate() error {
	var errList []error
	if r.Wallet == "" {
		errList = append(errList, errors.New("'wallet' is required"))
	}
	if r.Id != "" && !isRuneIdOrRuneName(r.Id) {
		errList = append(errList, errors.New("'id' is not valid rune id or rune name"))
	}
	if r.Limit < 0 {
		errList = append(errList, errors.New("'limit' must be non-negative"))
	}
	if r.Limit > getUTXOsMaxLimit {
		errList = append(errList, errors.Errorf("'limit' cannot exceed %d", getUTXOsMaxLimit))
	}
	return errs.WithPublicMessage(errors.Join(errList...), "validation error")
}

type runeBalance struct {
	RuneId       runes.RuneId     `json:"runeId"`
	Rune         runes.SpacedRune `json:"rune"`
	Symbol       string           `json:"symbol"`
	Amount       uint128.Uint128  `json:"amount"`
	Divisibility uint8            `json:"divisibility"`
}

type utxoExtend struct {
	Runes []runeBalance `json:"runes"`
}

type utxoItem struct {
	TxHash      chainhash.Hash `json:"txHash"`
	OutputIndex uint32         `json:"outputIndex"`
	Sats        int64          `json:"sats"`
	Extend      utxoExtend     `json:"extend"`
}

type getUTXOsResult struct {
	List        []utxoItem `json:"list"`
	BlockHeight uint64     `json:"blockHeight"`
}

type getUTXOsResponse = HttpResponse[getUTXOsResult]

func (h *HttpHandler) GetUTXOs(ctx *fiber.Ctx) (err error) {
	var req getUTXOsRequest
	if err := ctx.ParamsParser(&req); err != nil {
		return errors.WithStack(err)
	}
	if err := ctx.QueryParser(&req); err != nil {
		return errors.WithStack(err)
	}
	if err := req.Validate(); err != nil {
		return errors.WithStack(err)
	}
	if err := req.ParseDefault(); err != nil {
		return errors.WithStack(err)
	}

	pkScript, ok := resolvePkScript(h.network, req.Wallet)
	if !ok {
		return errs.NewPublicError("unable to resolve pkscript from \"wallet\"")
	}

	blockHeight := req.BlockHeight
	if blockHeight == 0 {
		blockHeader, err := h.usecase.GetLatestBlock(ctx.UserContext())
		if err != nil {
			if errors.Is(err, errs.NotFound) {
				return errs.NewPublicError("latest block not found")
			}
			return errors.Wrap(err, "error during GetLatestBlock")
		}
		blockHeight = uint64(blockHeader.Height)
	}

	var utxos []*entity.RunesUTXOWithSats
	if runeId, ok := h.resolveRuneId(ctx.UserContext(), req.Id); ok {
		utxos, err = h.usecase.GetRunesUTXOsByRuneIdAndPkScript(ctx.UserContext(), runeId, pkScript, blockHeight, req.Limit, req.Offset)
		if err != nil {
			if errors.Is(err, errs.NotFound) {
				return errs.NewPublicError("utxos not found")
			}
			return errors.Wrap(err, "error during GetBalancesByPkScript")
		}
	} else {
		utxos, err = h.usecase.GetRunesUTXOsByPkScript(ctx.UserContext(), pkScript, blockHeight, req.Limit, req.Offset)
		if err != nil {
			if errors.Is(err, errs.NotFound) {
				return errs.NewPublicError("utxos not found")
			}
			return errors.Wrap(err, "error during GetBalancesByPkScript")
		}
	}

	runeIds := make(map[runes.RuneId]struct{}, 0)
	for _, utxo := range utxos {
		for _, balance := range utxo.RuneBalances {
			runeIds[balance.RuneId] = struct{}{}
		}
	}
	runeIdsList := lo.Keys(runeIds)
	runeEntries, err := h.usecase.GetRuneEntryByRuneIdBatch(ctx.UserContext(), runeIdsList)
	if err != nil {
		if errors.Is(err, errs.NotFound) {
			return errs.NewPublicError("rune entries not found")
		}
		return errors.Wrap(err, "error during GetRuneEntryByRuneIdBatch")
	}

	utxoRespList := make([]utxoItem, 0, len(utxos))
	for _, utxo := range utxos {
		runeBalances := make([]runeBalance, 0, len(utxo.RuneBalances))
		for _, balance := range utxo.RuneBalances {
			runeEntry := runeEntries[balance.RuneId]
			runeBalances = append(runeBalances, runeBalance{
				RuneId:       balance.RuneId,
				Rune:         runeEntry.SpacedRune,
				Symbol:       string(runeEntry.Symbol),
				Amount:       balance.Amount,
				Divisibility: runeEntry.Divisibility,
			})
		}

		utxoRespList = append(utxoRespList, utxoItem{
			TxHash:      utxo.OutPoint.Hash,
			OutputIndex: utxo.OutPoint.Index,
			Sats:        utxo.Sats,
			Extend: utxoExtend{
				Runes: runeBalances,
			},
		})
	}

	resp := getUTXOsResponse{
		Result: &getUTXOsResult{
			BlockHeight: blockHeight,
			List:        utxoRespList,
		},
	}

	return errors.WithStack(ctx.JSON(resp))
}
