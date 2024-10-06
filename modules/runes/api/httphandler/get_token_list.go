package httphandler

import (
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
)

const (
	getTokenListMaxLimit = 1000
)

type GetTokenListScope string

const (
	GetTokenListScopeAll     GetTokenListScope = "all"
	GetTokenListScopeMinting GetTokenListScope = "minting"
)

func (s GetTokenListScope) IsValid() bool {
	switch s {
	case GetTokenListScopeAll, GetTokenListScopeMinting:
		return true
	}
	return false
}

type getTokenListRequest struct {
	paginationRequest
	Search      string            `query:"search"`
	BlockHeight uint64            `query:"blockHeight"`
	Scope       GetTokenListScope `query:"scope"`
}

func (req getTokenListRequest) Validate() error {
	var errList []error
	if err := req.paginationRequest.Validate(); err != nil {
		errList = append(errList, err)
	}
	if req.Limit > getTokenListMaxLimit {
		errList = append(errList, errors.Errorf("limit must be less than or equal to 1000"))
	}
	if req.Scope != "" && !req.Scope.IsValid() {
		errList = append(errList, errors.Errorf("invalid scope"))
	}
	return errs.WithPublicMessage(errors.Join(errList...), "validation error")
}

func (req *getTokenListRequest) ParseDefault() error {
	if err := req.paginationRequest.ParseDefault(); err != nil {
		return errors.WithStack(err)
	}
	if req.Scope == "" {
		req.Scope = GetTokenListScopeAll
	}
	return nil
}

type getTokenListResult struct {
	List []getTokenInfoResult `json:"list"`
}

type getTokenListResponse = HttpResponse[getTokenListResult]

func (h *HttpHandler) GetTokenList(ctx *fiber.Ctx) (err error) {
	var req getTokenListRequest
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
	case GetTokenListScopeAll:
		entries, err = h.usecase.GetRuneEntries(ctx.UserContext(), search, blockHeight, req.Limit, req.Offset)
		if err != nil {
			return errors.Wrap(err, "error during GetRuneEntryList")
		}
	case GetTokenListScopeMinting:
		entries, err = h.usecase.GetMintingRuneEntries(ctx.UserContext(), search, blockHeight, req.Limit, req.Offset)
		if err != nil {
			return errors.Wrap(err, "error during GetRuneEntryList")
		}
	default:
		panic("invalid scope")
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
