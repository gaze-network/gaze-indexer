package httphandler

import (
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/uint128"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
)

type getBalancesRequest struct {
	Wallet      string `params:"wallet"`
	Id          string `query:"id"`
	BlockHeight uint64 `query:"blockHeight"`
	Limit       int32  `query:"limit"`
	Offset      int32  `query:"offset"`
}

const (
	getBalancesMaxLimit     = 5000
	getBalancesDefaultLimit = 100
)

func (r getBalancesRequest) Validate() error {
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
	if r.Limit > getBalancesMaxLimit {
		errList = append(errList, errors.Errorf("'limit' cannot exceed %d", getBalancesMaxLimit))
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

type getBalancesResult struct {
	List        []balance `json:"list"`
	BlockHeight uint64    `json:"blockHeight"`
}

type getBalancesResponse = HttpResponse[getBalancesResult]

func (h *HttpHandler) GetBalances(ctx *fiber.Ctx) (err error) {
	var req getBalancesRequest
	if err := ctx.ParamsParser(&req); err != nil {
		return errors.WithStack(err)
	}
	if err := ctx.QueryParser(&req); err != nil {
		return errors.WithStack(err)
	}
	if err := req.Validate(); err != nil {
		return errors.WithStack(err)
	}
	if req.Limit == 0 {
		req.Limit = getBalancesDefaultLimit
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

	balances, err := h.usecase.GetBalancesByPkScript(ctx.UserContext(), pkScript, blockHeight, req.Limit, req.Offset)
	if err != nil {
		if errors.Is(err, errs.NotFound) {
			return errs.NewPublicError("balances not found")
		}
		return errors.Wrap(err, "error during GetBalancesByPkScript")
	}

	runeId, ok := h.resolveRuneId(ctx.UserContext(), req.Id)
	if ok {
		// filter out balances that don't match the requested rune id
		balances = lo.Filter(balances, func(b *entity.Balance, _ int) bool {
			return b.RuneId == runeId
		})
	}

	balanceRuneIds := lo.Map(balances, func(b *entity.Balance, _ int) runes.RuneId {
		return b.RuneId
	})
	runeEntries, err := h.usecase.GetRuneEntryByRuneIdBatch(ctx.UserContext(), balanceRuneIds)
	if err != nil {
		return errors.Wrap(err, "error during GetRuneEntryByRuneIdBatch")
	}

	balanceList := make([]balance, 0, len(balances))
	for _, b := range balances {
		runeEntry := runeEntries[b.RuneId]
		balanceList = append(balanceList, balance{
			Amount:   b.Amount,
			Id:       b.RuneId,
			Name:     runeEntry.SpacedRune,
			Symbol:   string(runeEntry.Symbol),
			Decimals: runeEntry.Divisibility,
		})
	}

	resp := getBalancesResponse{
		Result: &getBalancesResult{
			BlockHeight: blockHeight,
			List:        balanceList,
		},
	}

	return errors.WithStack(ctx.JSON(resp))
}
