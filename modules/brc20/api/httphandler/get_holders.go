package httphandler

import (
	"encoding/hex"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/pkg/btcutils"
	"github.com/gaze-network/indexer-network/pkg/decimals"
	"github.com/gofiber/fiber/v2"
	"github.com/holiman/uint256"
)

type getHoldersRequest struct {
	Id          string `params:"id"`
	BlockHeight uint64 `query:"blockHeight"`
}

func (r getHoldersRequest) Validate() error {
	var errList []error
	return errs.WithPublicMessage(errors.Join(errList...), "validation error")
}

type holdingBalanceExtend struct {
	Transferable *uint256.Int `json:"transferable"`
	Available    *uint256.Int `json:"available"`
}

type holdingBalance struct {
	Address  string               `json:"address"`
	PkScript string               `json:"pkScript"`
	Amount   *uint256.Int         `json:"amount"`
	Percent  float64              `json:"percent"`
	Extend   holdingBalanceExtend `json:"extend"`
}

type getHoldersResult struct {
	BlockHeight  uint64           `json:"blockHeight"`
	TotalSupply  *uint256.Int     `json:"totalSupply"`
	MintedAmount *uint256.Int     `json:"mintedAmount"`
	Decimals     uint16           `json:"decimals"`
	List         []holdingBalance `json:"list"`
}

type getHoldersResponse = common.HttpResponse[getHoldersResult]

func (h *HttpHandler) GetHolders(ctx *fiber.Ctx) (err error) {
	var req getHoldersRequest
	if err := ctx.ParamsParser(&req); err != nil {
		return errors.WithStack(err)
	}
	if err := ctx.QueryParser(&req); err != nil {
		return errors.WithStack(err)
	}
	if err := req.Validate(); err != nil {
		return errors.WithStack(err)
	}

	blockHeight := req.BlockHeight
	if blockHeight == 0 {
		blockHeader, err := h.usecase.GetLatestBlock(ctx.UserContext())
		if err != nil {
			return errors.Wrap(err, "error during GetLatestBlock")
		}
		blockHeight = uint64(blockHeader.Height)
	}

	entry, err := h.usecase.GetTickEntryByTickAndHeight(ctx.UserContext(), req.Id, blockHeight)
	if err != nil {
		return errors.Wrap(err, "error during GetTickEntryByTickAndHeight")
	}
	holdingBalances, err := h.usecase.GetBalancesByTick(ctx.UserContext(), req.Id, blockHeight)
	if err != nil {
		return errors.Wrap(err, "error during GetBalancesByTick")
	}

	list := make([]holdingBalance, 0, len(holdingBalances))
	for _, balance := range holdingBalances {
		address, err := btcutils.PkScriptToAddress(balance.PkScript, h.network)
		if err != nil {
			return errors.Wrapf(err, "can't convert pkscript(%x) to address", balance.PkScript)
		}
		percent := balance.OverallBalance.Div(entry.TotalSupply)
		list = append(list, holdingBalance{
			Address:  address,
			PkScript: hex.EncodeToString(balance.PkScript),
			Amount:   decimals.ToUint256(balance.OverallBalance, entry.Decimals),
			Percent:  percent.InexactFloat64(),
			Extend: holdingBalanceExtend{
				Transferable: decimals.ToUint256(balance.OverallBalance.Sub(balance.AvailableBalance), entry.Decimals),
				Available:    decimals.ToUint256(balance.AvailableBalance, entry.Decimals),
			},
		})
	}

	resp := getHoldersResponse{
		Result: &getHoldersResult{
			BlockHeight:  blockHeight,
			TotalSupply:  decimals.ToUint256(entry.TotalSupply, entry.Decimals),  // TODO: convert to wei
			MintedAmount: decimals.ToUint256(entry.MintedAmount, entry.Decimals), // TODO: convert to wei
			List:         list,
		},
	}

	return errors.WithStack(ctx.JSON(resp))
}
