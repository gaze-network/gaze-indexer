package httphandler

import (
	"net/url"
	"slices"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/uint128"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
)

type getTokenInfoRequest struct {
	Id          string `params:"id"`
	BlockHeight uint64 `query:"blockHeight"`
}

func (r *getTokenInfoRequest) Validate() error {
	var errList []error
	id, err := url.QueryUnescape(r.Id)
	if err != nil {
		return errors.WithStack(err)
	}
	r.Id = id
	if !isRuneIdOrRuneName(r.Id) {
		errList = append(errList, errors.Errorf("id '%s' is not valid rune id or rune name", r.Id))
	}
	return errs.WithPublicMessage(errors.Join(errList...), "validation error")
}

type entryTerms struct {
	Amount      uint128.Uint128 `json:"amount"`
	Cap         uint128.Uint128 `json:"cap"`
	HeightStart *uint64         `json:"heightStart"`
	HeightEnd   *uint64         `json:"heightEnd"`
	OffsetStart *uint64         `json:"offsetStart"`
	OffsetEnd   *uint64         `json:"offsetEnd"`
}

type entry struct {
	Divisibility uint8           `json:"divisibility"`
	Premine      uint128.Uint128 `json:"premine"`
	Rune         runes.Rune      `json:"rune"`
	Spacers      uint32          `json:"spacers"`
	Symbol       string          `json:"symbol"`
	Terms        entryTerms      `json:"terms"`
	Turbo        bool            `json:"turbo"`
}

type tokenInfoExtend struct {
	Entry entry `json:"entry"`
}

type getTokenInfoResult struct {
	Id                runes.RuneId     `json:"id"`
	Name              runes.SpacedRune `json:"name"` // rune name
	Symbol            string           `json:"symbol"`
	TotalSupply       uint128.Uint128  `json:"totalSupply"`
	CirculatingSupply uint128.Uint128  `json:"circulatingSupply"`
	MintedAmount      uint128.Uint128  `json:"mintedAmount"`
	BurnedAmount      uint128.Uint128  `json:"burnedAmount"`
	Decimals          uint8            `json:"decimals"`
	DeployedAt        int64            `json:"deployedAt"` // unix timestamp
	DeployedAtHeight  uint64           `json:"deployedAtHeight"`
	CompletedAt       *int64           `json:"completedAt"` // unix timestamp
	CompletedAtHeight *uint64          `json:"completedAtHeight"`
	HoldersCount      int              `json:"holdersCount"`
	Extend            tokenInfoExtend  `json:"extend"`
}

type getTokenInfoResponse = HttpResponse[getTokenInfoResult]

func (h *HttpHandler) GetTokenInfo(ctx *fiber.Ctx) (err error) {
	var req getTokenInfoRequest
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
			if errors.Is(err, errs.NotFound) {
				return errs.NewPublicError("latest block not found")
			}
			return errors.Wrap(err, "error during GetLatestBlock")
		}
		blockHeight = uint64(blockHeader.Height)
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
		if errors.Is(err, errs.NotFound) {
			return errs.NewPublicError("rune not found")
		}
		return errors.Wrap(err, "error during GetTokenInfoByHeight")
	}
	holdingBalances, err := h.usecase.GetBalancesByRuneId(ctx.UserContext(), runeId, blockHeight, -1, 0) // get all balances
	if err != nil {
		if errors.Is(err, errs.NotFound) {
			return errs.NewPublicError("rune not found")
		}
		return errors.Wrap(err, "error during GetBalancesByRuneId")
	}

	holdingBalances = lo.Filter(holdingBalances, func(b *entity.Balance, _ int) bool {
		return !b.Amount.IsZero()
	})
	// sort by amount descending
	slices.SortFunc(holdingBalances, func(i, j *entity.Balance) int {
		return j.Amount.Cmp(i.Amount)
	})

	totalSupply, err := runeEntry.Supply()
	if err != nil {
		return errors.Wrap(err, "cannot get total supply of rune")
	}
	mintedAmount, err := runeEntry.MintedAmount()
	if err != nil {
		return errors.Wrap(err, "cannot get minted amount of rune")
	}
	circulatingSupply := mintedAmount.Sub(runeEntry.BurnedAmount)

	terms := lo.FromPtr(runeEntry.Terms)
	resp := getTokenInfoResponse{
		Result: &getTokenInfoResult{
			Id:                runeId,
			Name:              runeEntry.SpacedRune,
			Symbol:            string(runeEntry.Symbol),
			TotalSupply:       totalSupply,
			CirculatingSupply: circulatingSupply,
			MintedAmount:      mintedAmount,
			BurnedAmount:      runeEntry.BurnedAmount,
			Decimals:          runeEntry.Divisibility,
			DeployedAt:        runeEntry.EtchedAt.Unix(),
			DeployedAtHeight:  runeEntry.EtchingBlock,
			CompletedAt:       lo.Ternary(runeEntry.CompletedAt.IsZero(), nil, lo.ToPtr(runeEntry.CompletedAt.Unix())),
			CompletedAtHeight: runeEntry.CompletedAtHeight,
			HoldersCount:      len(holdingBalances),
			Extend: tokenInfoExtend{
				Entry: entry{
					Divisibility: runeEntry.Divisibility,
					Premine:      runeEntry.Premine,
					Rune:         runeEntry.SpacedRune.Rune,
					Spacers:      runeEntry.SpacedRune.Spacers,
					Symbol:       string(runeEntry.Symbol),
					Terms: entryTerms{
						Amount:      lo.FromPtr(terms.Amount),
						Cap:         lo.FromPtr(terms.Cap),
						HeightStart: terms.HeightStart,
						HeightEnd:   terms.HeightEnd,
						OffsetStart: terms.OffsetStart,
						OffsetEnd:   terms.OffsetEnd,
					},
					Turbo: runeEntry.Turbo,
				},
			},
		},
	}

	return errors.WithStack(ctx.JSON(resp))
}
