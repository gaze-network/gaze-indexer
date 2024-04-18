package httphandler

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/uint128"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
)

type getUTXOsByAddressRequest struct {
	Wallet      string `params:"wallet"`
	Id          string `query:"id"`
	BlockHeight uint64 `query:"blockHeight"`
}

func (r getUTXOsByAddressRequest) Validate() error {
	var errList []error
	if r.Wallet == "" {
		errList = append(errList, errors.New("'wallet' is required"))
	}
	if r.Id != "" && !isRuneIdOrRuneName(r.Id) {
		errList = append(errList, errors.New("'id' is not valid rune id or rune name"))
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

type utxo struct {
	TxHash      chainhash.Hash `json:"txHash"`
	OutputIndex uint32         `json:"outputIndex"`
	Extend      utxoExtend     `json:"extend"`
}

type getUTXOsByAddressResult struct {
	List        []utxo `json:"list"`
	BlockHeight uint64 `json:"blockHeight"`
}

type getUTXOsByAddressResponse = HttpResponse[getUTXOsByAddressResult]

func (h *HttpHandler) GetUTXOsByAddress(ctx *fiber.Ctx) (err error) {
	var req getUTXOsByAddressRequest
	if err := ctx.ParamsParser(&req); err != nil {
		return errors.WithStack(err)
	}
	if err := ctx.QueryParser(&req); err != nil {
		return errors.WithStack(err)
	}
	if err := req.Validate(); err != nil {
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
			return errors.Wrap(err, "error during GetLatestBlock")
		}
		blockHeight = uint64(blockHeader.Height)
	}

	outPointBalances, err := h.usecase.GetUnspentOutPointBalancesByPkScript(ctx.UserContext(), pkScript, blockHeight)
	if err != nil {
		return errors.Wrap(err, "error during GetBalancesByPkScript")
	}

	outPointBalanceRuneIds := lo.Map(outPointBalances, func(outPointBalance *entity.OutPointBalance, _ int) runes.RuneId {
		return outPointBalance.RuneId
	})
	runeEntries, err := h.usecase.GetRuneEntryByRuneIdBatch(ctx.UserContext(), outPointBalanceRuneIds)
	if err != nil {
		return errors.Wrap(err, "error during GetRuneEntryByRuneIdBatch")
	}

	groupedBalances := lo.GroupBy(outPointBalances, func(outPointBalance *entity.OutPointBalance) wire.OutPoint {
		return outPointBalance.OutPoint
	})

	utxoList := make([]utxo, 0, len(groupedBalances))
	for outPoint, balances := range groupedBalances {
		runeBalances := make([]runeBalance, 0, len(balances))
		for _, balance := range balances {
			runeEntry := runeEntries[balance.RuneId]
			runeBalances = append(runeBalances, runeBalance{
				RuneId:       balance.RuneId,
				Rune:         runeEntry.SpacedRune,
				Symbol:       string(runeEntry.Symbol),
				Amount:       balance.Amount,
				Divisibility: runeEntry.Divisibility,
			})
		}

		utxoList = append(utxoList, utxo{
			TxHash:      outPoint.Hash,
			OutputIndex: outPoint.Index,
			Extend: utxoExtend{
				Runes: runeBalances,
			},
		})
	}

	// filter by req.Id if exists
	{
		runeId, ok := h.resolveRuneId(ctx.UserContext(), req.Id)
		if ok {
			utxoList = lo.Filter(utxoList, func(u utxo, _ int) bool {
				for _, runeBalance := range u.Extend.Runes {
					if runeBalance.RuneId == runeId {
						return true
					}
				}
				return false
			})
		}
	}

	resp := getUTXOsByAddressResponse{
		Result: &getUTXOsByAddressResult{
			BlockHeight: blockHeight,
			List:        utxoList,
		},
	}

	return errors.WithStack(ctx.JSON(resp))
}
