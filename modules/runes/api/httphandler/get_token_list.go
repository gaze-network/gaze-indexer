package httphandler

import (
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
)

const (
	getTokenListMaxLimit = 1000
)

type getTokenListRequest struct {
	paginationRequest
}

func (req getTokenListRequest) Validate() error {
	var errList []error
	if err := req.paginationRequest.Validate(); err != nil {
		errList = append(errList, err)
	}
	if req.Limit > getTokenListMaxLimit {
		errList = append(errList, errors.Errorf("limit must be less than or equal to 1000"))
	}
	return errs.WithPublicMessage(errors.Join(errList...), "validation error")
}

type getTokenListResult struct {
	List []getTokenInfoResult `json:"list"`
}

type getTokenListResponse = HttpResponse[getTokenListResult]

func (h *HttpHandler) GetTokenList(ctx *fiber.Ctx) (err error) {
	var req getTokenListRequest
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

	entries, err := h.usecase.GetRuneEntryList(ctx.UserContext(), req.Limit, req.Offset)
	if err != nil {
		return errors.Wrap(err, "error during GetRuneEntryList")
	}

	runeIds := lo.Map(entries, func(item *runes.RuneEntry, _ int) runes.RuneId { return item.RuneId })
	totalHolders, err := h.usecase.GetTotalHoldersByRuneIds(ctx.UserContext(), runeIds)
	if err != nil {
		return errors.Wrap(err, "error during GetTotalHoldersByRuneIds")
	}

	result := make([]getTokenInfoResult, 0, len(entries))
	for _, ent := range entries {
		totalSupply, err := ent.Supply()
		if err != nil {
			return errors.Wrap(err, "cannot get total supply of rune")
		}
		mintedAmount, err := ent.MintedAmount()
		if err != nil {
			return errors.Wrap(err, "cannot get minted amount of rune")
		}
		circulatingSupply := mintedAmount.Sub(ent.BurnedAmount)

		terms := lo.FromPtr(ent.Terms)
		result = append(result, getTokenInfoResult{
			Id:                ent.RuneId,
			Name:              ent.SpacedRune,
			Symbol:            string(ent.Symbol),
			TotalSupply:       totalSupply,
			CirculatingSupply: circulatingSupply,
			MintedAmount:      mintedAmount,
			BurnedAmount:      ent.BurnedAmount,
			Decimals:          ent.Divisibility,
			DeployedAt:        uint64(ent.EtchedAt.Unix()),
			DeployedAtHeight:  ent.EtchingBlock,
			CompletedAt:       lo.Ternary(ent.CompletedAt.IsZero(), nil, lo.ToPtr(uint64(ent.CompletedAt.Unix()))),
			CompletedAtHeight: ent.CompletedAtHeight,
			HoldersCount:      int(totalHolders[ent.RuneId]),
			Extend: tokenInfoExtend{
				Entry: entry{
					Divisibility: ent.Divisibility,
					Premine:      ent.Premine,
					Rune:         ent.SpacedRune.Rune,
					Spacers:      ent.SpacedRune.Spacers,
					Symbol:       string(ent.Symbol),
					Terms: entryTerms{
						Amount:      lo.FromPtr(terms.Amount),
						Cap:         lo.FromPtr(terms.Cap),
						HeightStart: terms.HeightStart,
						HeightEnd:   terms.HeightEnd,
						OffsetStart: terms.OffsetStart,
						OffsetEnd:   terms.OffsetEnd,
					},
					Turbo: ent.Turbo,
				},
			},
		})
	}

	return errors.WithStack(ctx.JSON(getTokenListResponse{
		Result: &getTokenListResult{
			List: result,
		},
	}))
}
