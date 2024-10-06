package httphandler

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
)

const (
	getTokensMaxLimit = 1000
)

type GetTokensScope string

const (
	GetTokensScopeAll     GetTokensScope = "all"
	GetTokensScopeOngoing GetTokensScope = "ongoing"
)

func (s GetTokensScope) IsValid() bool {
	switch s {
	case GetTokensScopeAll, GetTokensScopeOngoing:
		return true
	}
	return false
}

type getTokensRequest struct {
	paginationRequest
	Search      string         `query:"search"`
	BlockHeight uint64         `query:"blockHeight"`
	Scope       GetTokensScope `query:"scope"`
}

func (req getTokensRequest) Validate() error {
	var errList []error
	if err := req.paginationRequest.Validate(); err != nil {
		errList = append(errList, err)
	}
	if req.Limit > getTokensMaxLimit {
		errList = append(errList, errors.Errorf("limit must be less than or equal to 1000"))
	}
	if req.Scope != "" && !req.Scope.IsValid() {
		errList = append(errList, errors.Errorf("invalid scope: %s", req.Scope))
	}
	return errs.WithPublicMessage(errors.Join(errList...), "validation error")
}

func (req *getTokensRequest) ParseDefault() error {
	if err := req.paginationRequest.ParseDefault(); err != nil {
		return errors.WithStack(err)
	}
	if req.Scope == "" {
		req.Scope = GetTokensScopeAll
	}
	return nil
}

type getTokensResult struct {
	List []getTokenInfoResult `json:"list"`
}

type getTokensResponse = HttpResponse[getTokensResult]

func (h *HttpHandler) GetTokens(ctx *fiber.Ctx) (err error) {
	var req getTokensRequest
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
			if errors.Is(err, errs.NotFound) {
				return errs.NewPublicError("latest block not found")
			}
			return errors.Wrap(err, "error during GetLatestBlock")
		}
		blockHeight = uint64(blockHeader.Height)
	}

	// remove spacers
	search := strings.Replace(strings.Replace(req.Search, "•", "", -1), ".", "", -1)

	var entries []*runes.RuneEntry
	switch req.Scope {
	case GetTokensScopeAll:
		entries, err = h.usecase.GetRuneEntries(ctx.UserContext(), search, blockHeight, req.Limit, req.Offset)
		if err != nil {
			return errors.Wrap(err, "error during GetRuneEntryList")
		}
	case GetTokensScopeOngoing:
		entries, err = h.usecase.GetOngoingRuneEntries(ctx.UserContext(), search, blockHeight, req.Limit, req.Offset)
		if err != nil {
			return errors.Wrap(err, "error during GetRuneEntryList")
		}
	default:
		return errs.NewPublicError(fmt.Sprintf("invalid scope: %s", req.Scope))
	}

	runeIds := lo.Map(entries, func(item *runes.RuneEntry, _ int) runes.RuneId { return item.RuneId })
	totalHolders, err := h.usecase.GetTotalHoldersByRuneIds(ctx.UserContext(), runeIds, blockHeight)
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
			DeployedAt:        ent.EtchedAt.Unix(),
			DeployedAtHeight:  ent.EtchingBlock,
			CompletedAt:       lo.Ternary(ent.CompletedAt.IsZero(), nil, lo.ToPtr(ent.CompletedAt.Unix())),
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

	return errors.WithStack(ctx.JSON(getTokensResponse{
		Result: &getTokensResult{
			List: result,
		},
	}))
}
