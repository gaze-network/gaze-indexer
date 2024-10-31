package httphandler

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/url"
	"slices"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/uint128"
	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"
)

type getHoldersRequest struct {
	paginationRequest
	Id          string `params:"id"`
	BlockHeight uint64 `query:"blockHeight"`
}

const (
	getHoldersMaxLimit = 1000
)

func (r *getHoldersRequest) Validate() error {
	var errList []error
	id, err := url.QueryUnescape(r.Id)
	if err != nil {
		return errors.WithStack(err)
	}
	r.Id = id
	if !isRuneIdOrRuneName(r.Id) {
		errList = append(errList, errors.Errorf("id '%s' is not valid rune id or rune name", r.Id))
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
	Decimals     uint8            `json:"decimals"`
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
	if err := req.ParseDefault(); err != nil {
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

	var runeId runes.RuneId
	if req.Id != "" {
		var ok bool
		runeId, ok = h.resolveRuneId(ctx.UserContext(), req.Id)
		if !ok {
			return errs.NewPublicError(fmt.Sprintf("unable to resolve rune id \"%s\" from \"id\"", req.Id))
		}
	}

	runeEntry, err := h.usecase.GetRuneEntryByRuneIdAndHeight(ctx.UserContext(), runeId, blockHeight)
	if err != nil {
		if errors.Is(err, errs.NotFound) {
			return errs.NewPublicError("rune not found")
		}
		return errors.Wrap(err, "error during GetRuneEntryByRuneIdAndHeight")
	}
	holdingBalances, err := h.usecase.GetBalancesByRuneId(ctx.UserContext(), runeId, blockHeight, req.Limit, req.Offset)
	if err != nil {
		if errors.Is(err, errs.NotFound) {
			return errs.NewPublicError("balances not found")
		}
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

	// sort by amount descending, then pk script ascending
	slices.SortFunc(holdingBalances, func(b1, b2 *entity.Balance) int {
		if b1.Amount.Cmp(b2.Amount) == 0 {
			return bytes.Compare(b1.PkScript, b2.PkScript)
		}
		return b2.Amount.Cmp(b1.Amount)
	})

	resp := getHoldersResponse{
		Result: &getHoldersResult{
			BlockHeight:  blockHeight,
			TotalSupply:  totalSupply,
			MintedAmount: mintedAmount,
			Decimals:     runeEntry.Divisibility,
			List:         list,
		},
	}

	return errors.WithStack(ctx.JSON(resp))
}
