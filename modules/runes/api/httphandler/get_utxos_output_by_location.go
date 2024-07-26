package httphandler

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/indexer-network/modules/runes/usecase"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
)

type getUTXOsOutputByTxIdRequest struct {
	TxHash      string `params:"txHash"`
	outputIndex int32  `query:"outputIndex"`
}

func (r getUTXOsOutputByTxIdRequest) Validate() error {
	var errList []error
	if r.TxHash == "" {
		errList = append(errList, errors.New("'txHash' is required"))
	}
	if r.outputIndex < 0 {
		errList = append(errList, errors.New("'outputIndex' must be non-negative"))
	}
	return errs.WithPublicMessage(errors.Join(errList...), "validation error")
}

type getUTXOsOutputByTxIdResponse = HttpResponse[utxoItem]

func (h *HttpHandler) GetUTXOsOutputByLocation(ctx *fiber.Ctx) (err error) {
	var req getUTXOsOutputByTxIdRequest
	if err := ctx.ParamsParser(&req); err != nil {
		return errors.WithStack(err)
	}
	if err := ctx.QueryParser(&req); err != nil {
		return errors.WithStack(err)
	}
	if err := req.Validate(); err != nil {
		return errors.WithStack(err)
	}

	txHash, err := chainhash.NewHashFromStr(req.TxHash)
	if err != nil {
		return errs.NewPublicError("unable to resolve txHash")
	}

	utxo, err := h.usecase.GetUTXOsOutputByLocation(ctx.UserContext(), *txHash, uint32(req.outputIndex))
	if err != nil {
		if errors.Is(err, usecase.ErrUTXONotFound) {
			return errs.NewPublicError("utxo not found")
		}
		return errors.WithStack(err)
	}

	runeIds := make(map[runes.RuneId]struct{}, 0)
	for _, balance := range utxo.RuneBalances {
		runeIds[balance.RuneId] = struct{}{}
	}
	runeIdsList := lo.Keys(runeIds)
	runeEntries, err := h.usecase.GetRuneEntryByRuneIdBatch(ctx.UserContext(), runeIdsList)
	if err != nil {
		if errors.Is(err, errs.NotFound) {
			return errs.NewPublicError("rune entries not found")
		}
		return errors.Wrap(err, "error during GetRuneEntryByRuneIdBatch")
	}

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

	resp := getUTXOsOutputByTxIdResponse{
		Result: &utxoItem{
			TxHash:      utxo.OutPoint.Hash,
			OutputIndex: utxo.OutPoint.Index,
			Sats:        utxo.Sats,
			Extend: utxoExtend{
				Runes: runeBalances,
			},
		},
	}
	return errors.WithStack(ctx.JSON(resp))
}
