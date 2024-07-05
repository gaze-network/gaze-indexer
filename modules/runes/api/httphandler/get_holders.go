package httphandler

import (
	"encoding/hex"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/uint128"
	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"
)

type getHoldersRequest struct {
	Id          string `params:"id"`
	BlockHeight uint64 `query:"blockHeight"`
	Limit       int32  `json:"limit"`
	Offset      int32  `json:"offset"`
}

const getHoldersMaxLimit = 1000

func (r getHoldersRequest) Validate() error {
	var errList []error
	if !isRuneIdOrRuneName(r.Id) {
		errList = append(errList, errors.New("'id' is not valid rune id or rune name"))
	}
	if r.Limit < 0 {
		errList = append(errList, errors.New("'limit' must be non-negative"))
	}
	if r.Limit > getHoldersMaxLimit {
		errList = append(errList, errors.Errorf("'limit' cannot exceed %d", getHoldersMaxLimit))
	}
	return errs.WithPublicMessage(errors.Join(errList...), "validation error")
}

type holdingBalance struct {
	Address  string          `json:"address"`
	PkScript string          `json:"pkScript"`
	Amount   uint128.Uint128 `json:"amount"`
	Percent  float64         `json:"percent"`
}

type getHoldersResult struct {
	BlockHeight  uint64           `json:"blockHeight"`
	TotalSupply  uint128.Uint128  `json:"totalSupply"`
	MintedAmount uint128.Uint128  `json:"mintedAmount"`
	List         []holdingBalance `json:"list"`
}

type getHoldersResponse = HttpResponse[getHoldersResult]

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

	if req.Limit == 0 {
		req.Limit = getHoldersMaxLimit
	}

	var runeId runes.RuneId
	if req.Id != "" {
		var ok bool
		runeId, ok = h.resolveRuneId(ctx.UserContext(), req.Id)
		if !ok {
			return errs.NewPublicError("unable to resolve rune id from \"id\"")
		}
	}

	runeEntry, err := h.usecase.GetRuneEntryByRuneIdAndHeight(ctx.UserContext(), runeId, blockHeight)
	if err != nil {
		return errors.Wrap(err, "error during GetHoldersByHeight")
	}
	holdingBalances, err := h.usecase.GetBalancesByRuneId(ctx.UserContext(), runeId, blockHeight, req.Limit, req.Offset)
	if err != nil {
		return errors.Wrap(err, "error during GetBalancesByRuneId")
	}

	totalSupply, err := runeEntry.Supply()
	if err != nil {
		return errors.Wrap(err, "cannot get total supply of rune")
	}
	mintedAmount, err := runeEntry.MintedAmount()
	if err != nil {
		return errors.Wrap(err, "cannot get minted amount of rune")
	}

	list := make([]holdingBalance, 0, len(holdingBalances))
	for _, balance := range holdingBalances {
		address := addressFromPkScript(balance.PkScript, h.network)
		amount := decimal.NewFromBigInt(balance.Amount.Big(), 0)
		percent := amount.Div(decimal.NewFromBigInt(totalSupply.Big(), 0))
		list = append(list, holdingBalance{
			Address:  address,
			PkScript: hex.EncodeToString(balance.PkScript),
			Amount:   balance.Amount,
			Percent:  percent.InexactFloat64(),
		})
	}

	resp := getHoldersResponse{
		Result: &getHoldersResult{
			BlockHeight:  blockHeight,
			TotalSupply:  totalSupply,
			MintedAmount: mintedAmount,
			List:         list,
		},
	}

	return errors.WithStack(ctx.JSON(resp))
}
