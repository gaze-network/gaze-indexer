package httphandler

import (
	"slices"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/pkg/btcutils"
	"github.com/gaze-network/indexer-network/pkg/decimals"
	"github.com/gofiber/fiber/v2"
	"github.com/holiman/uint256"
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
	return errs.WithPublicMessage(errors.Join(errList...), "validation error")
}

type balanceExtend struct {
	Transferable *uint256.Int `json:"transferable"`
	Available    *uint256.Int `json:"available"`
}

type balance struct {
	Amount   *uint256.Int  `json:"amount"`
	Id       string        `json:"id"`
	Name     string        `json:"name"`
	Symbol   string        `json:"symbol"`
	Decimals uint16        `json:"decimals"`
	Extend   balanceExtend `json:"extend"`
}

type getBalancesByAddressResult struct {
	List        []balance `json:"list"`
	BlockHeight uint64    `json:"blockHeight"`
}

type getBalancesByAddressResponse = common.HttpResponse[getBalancesByAddressResult]

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

	pkScript, err := btcutils.ToPkScript(h.network, req.Wallet)
	if err != nil {
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

	balanceRuneIds := lo.Keys(balances)
	entries, err := h.usecase.GetTickEntryByTickBatch(ctx.UserContext(), balanceRuneIds)
	if err != nil {
		return errors.Wrap(err, "error during GetTickEntryByTickBatch")
	}

	balanceList := make([]balance, 0, len(balances))
	for id, b := range balances {
		entry := entries[id]
		balanceList = append(balanceList, balance{
			Amount:   decimals.ToUint256(b.OverallBalance, entry.Decimals),
			Id:       id,
			Name:     entry.OriginalTick,
			Symbol:   entry.Tick,
			Decimals: entry.Decimals,
			Extend: balanceExtend{
				Transferable: decimals.ToUint256(b.OverallBalance.Sub(b.AvailableBalance), entry.Decimals),
				Available:    decimals.ToUint256(b.AvailableBalance, entry.Decimals),
			},
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
