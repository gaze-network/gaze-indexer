package httphandler

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
)

type getTransactionByHashRequest struct {
	Hash string `params:"hash"`
}

func (r getTransactionByHashRequest) Validate() error {
	var errList []error
	if len(r.Hash) == 0 {
		errList = append(errList, errs.NewPublicError("hash is required"))
	}
	if len(r.Hash) > chainhash.MaxHashStringSize {
		errList = append(errList, errs.NewPublicError(fmt.Sprintf("hash length must be less than or equal to %d bytes", chainhash.MaxHashStringSize)))
	}
	if len(errList) == 0 {
		return nil
	}
	return errs.WithPublicMessage(errors.Join(errList...), "validation error")
}

type getTransactionByHashResponse = HttpResponse[transaction]

func (h *HttpHandler) GetTransactionByHash(ctx *fiber.Ctx) (err error) {
	var req getTransactionByHashRequest
	if err := ctx.ParamsParser(&req); err != nil {
		return errors.WithStack(err)
	}
	if err := req.Validate(); err != nil {
		return errors.WithStack(err)
	}

	hash, err := chainhash.NewHashFromStr(req.Hash)
	if err != nil {
		return errs.NewPublicError("invalid transaction hash")
	}

	tx, err := h.usecase.GetRuneTransaction(ctx.UserContext(), *hash)
	if err != nil {
		if errors.Is(err, errs.NotFound) {
			return fiber.NewError(fiber.StatusNotFound, "transaction not found")
		}
		return errors.Wrap(err, "error during GetRuneTransaction")
	}

	allRuneIds := make(map[runes.RuneId]struct{})
	for id := range tx.Mints {
		allRuneIds[id] = struct{}{}
	}
	for id := range tx.Burns {
		allRuneIds[id] = struct{}{}
	}
	for _, input := range tx.Inputs {
		allRuneIds[input.RuneId] = struct{}{}
	}
	for _, output := range tx.Outputs {
		allRuneIds[output.RuneId] = struct{}{}
	}

	runeEntries, err := h.usecase.GetRuneEntryByRuneIdBatch(ctx.UserContext(), lo.Keys(allRuneIds))
	if err != nil {
		return errors.Wrap(err, "error during GetRuneEntryByRuneIdBatch")
	}

	respTx := &transaction{
		TxHash:      tx.Hash,
		BlockHeight: tx.BlockHeight,
		Index:       tx.Index,
		Timestamp:   tx.Timestamp.Unix(),
		Inputs:      make([]txInputOutput, 0, len(tx.Inputs)),
		Outputs:     make([]txInputOutput, 0, len(tx.Outputs)),
		Mints:       make(map[string]amountWithDecimal, len(tx.Mints)),
		Burns:       make(map[string]amountWithDecimal, len(tx.Burns)),
		Extend: runeTransactionExtend{
			RuneEtched: tx.RuneEtched,
			Runestone:  nil,
		},
	}
	for _, input := range tx.Inputs {
		address := addressFromPkScript(input.PkScript, h.network)
		respTx.Inputs = append(respTx.Inputs, txInputOutput{
			PkScript: hex.EncodeToString(input.PkScript),
			Address:  address,
			Id:       input.RuneId,
			Amount:   input.Amount,
			Decimals: runeEntries[input.RuneId].Divisibility,
			Index:    input.Index,
		})
	}
	for _, output := range tx.Outputs {
		address := addressFromPkScript(output.PkScript, h.network)
		respTx.Outputs = append(respTx.Outputs, txInputOutput{
			PkScript: hex.EncodeToString(output.PkScript),
			Address:  address,
			Id:       output.RuneId,
			Amount:   output.Amount,
			Decimals: runeEntries[output.RuneId].Divisibility,
			Index:    output.Index,
		})
	}
	for id, amount := range tx.Mints {
		respTx.Mints[id.String()] = amountWithDecimal{
			Amount:   amount,
			Decimals: runeEntries[id].Divisibility,
		}
	}
	for id, amount := range tx.Burns {
		respTx.Burns[id.String()] = amountWithDecimal{
			Amount:   amount,
			Decimals: runeEntries[id].Divisibility,
		}
	}
	if tx.Runestone != nil {
		var e *etching
		if tx.Runestone.Etching != nil {
			var symbol *string
			if tx.Runestone.Etching.Symbol != nil {
				symbol = lo.ToPtr(string(*tx.Runestone.Etching.Symbol))
			}
			var t *terms
			if tx.Runestone.Etching.Terms != nil {
				t = &terms{
					Amount:      tx.Runestone.Etching.Terms.Amount,
					Cap:         tx.Runestone.Etching.Terms.Cap,
					HeightStart: tx.Runestone.Etching.Terms.HeightStart,
					HeightEnd:   tx.Runestone.Etching.Terms.HeightEnd,
					OffsetStart: tx.Runestone.Etching.Terms.OffsetStart,
					OffsetEnd:   tx.Runestone.Etching.Terms.OffsetEnd,
				}
			}
			e = &etching{
				Divisibility: tx.Runestone.Etching.Divisibility,
				Premine:      tx.Runestone.Etching.Premine,
				Rune:         tx.Runestone.Etching.Rune,
				Spacers:      tx.Runestone.Etching.Spacers,
				Symbol:       symbol,
				Terms:        t,
				Turbo:        tx.Runestone.Etching.Turbo,
			}
		}
		respTx.Extend.Runestone = &runestone{
			Cenotaph: tx.Runestone.Cenotaph,
			Flaws:    lo.Ternary(tx.Runestone.Cenotaph, tx.Runestone.Flaws.CollectAsString(), nil),
			Etching:  e,
			Edicts: lo.Map(tx.Runestone.Edicts, func(ed runes.Edict, _ int) edict {
				return edict{
					Id:     ed.Id,
					Amount: ed.Amount,
					Output: ed.Output,
				}
			}),
			Mint:    tx.Runestone.Mint,
			Pointer: tx.Runestone.Pointer,
		}
	}

	return errors.WithStack(ctx.JSON(getTransactionByHashResponse{
		Result: respTx,
	}))
}
