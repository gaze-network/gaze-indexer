package httphandler

import (
	"slices"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/uint128"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
)

type getBalancesByAddressRequest struct {
	Wallet      string `params:"wallet"`
	Id          string `query:"id"`
	BlockHeight uint64 `query:"blockHeight"`
}

func (r getBalancesByAddressRequest) Validate() error {
	var errList []error
	if r.Wallet == "" {
		errList = append(errList, errors.New("'wallet' is required"))
	}
	if r.Id != "" && !isRuneIdOrRuneName(r.Id) {
		errList = append(errList, errors.New("'id' is not valid rune id or rune name"))
	}
	return errs.WithPublicMessage(errors.Join(errList...), "validation error")
}

type balance struct {
	Amount   uint128.Uint128  `json:"amount"`
	Id       runes.RuneId     `json:"id"`
	Name     runes.SpacedRune `json:"name"`
	Symbol   string           `json:"symbol"`
	Decimals uint8            `json:"decimals"`
}

type getBalancesByAddressResult struct {
	List        []balance `json:"list"`
	BlockHeight uint64    `json:"blockHeight"`
}

type getBalancesByAddressResponse = HttpResponse[getBalancesByAddressResult]

func (h *HttpHandler) GetBalancesByAddress(ctx *fiber.Ctx) (err error) {
	var req getBalancesByAddressRequest
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

	balances, err := h.usecase.GetBalancesByPkScript(ctx.UserContext(), pkScript, blockHeight)
	if err != nil {
		return errors.Wrap(err, "error during GetBalancesByPkScript")
	}

	runeId, ok := h.resolveRuneId(ctx.UserContext(), req.Id)
	if ok {
		// filter out balances that don't match the requested rune id
		for key := range balances {
			if key != runeId {
				delete(balances, key)
			}
		}
	}

	balanceRuneIds := lo.Keys(balances)
	runeEntries, err := h.usecase.GetRuneEntryByRuneIdBatch(ctx.UserContext(), balanceRuneIds)
	if err != nil {
		return errors.Wrap(err, "error during GetRuneEntryByRuneIdBatch")
	}

	balanceList := make([]balance, 0, len(balances))
	for id, b := range balances {
		runeEntry := runeEntries[id]
		balanceList = append(balanceList, balance{
			Amount:   b.Amount,
			Id:       id,
			Name:     runeEntry.SpacedRune,
			Symbol:   string(runeEntry.Symbol),
			Decimals: runeEntry.Divisibility,
		})
	}
	slices.SortFunc(balanceList, func(i, j balance) int {
		return j.Amount.Cmp(i.Amount)
	})

	resp := getBalancesByAddressResponse{
		Result: &getBalancesByAddressResult{
			BlockHeight: blockHeight,
			List:        balanceList,
		},
	}

	return errors.WithStack(ctx.JSON(resp))
}
