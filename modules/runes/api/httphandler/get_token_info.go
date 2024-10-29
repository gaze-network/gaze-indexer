package httphandler

import (
	"fmt"
	"net/url"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/uint128"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
)

type getTokenInfoRequest struct {
	Id                  string `params:"id"`
	BlockHeight         uint64 `query:"blockHeight"`
	IncludeHoldersCount *bool  `query:"includeHoldersCount"`
}

func (r *getTokenInfoRequest) ParseDefault() error {
	if r.IncludeHoldersCount == nil {
		r.IncludeHoldersCount = lo.ToPtr(false)
	}
	return nil
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
	Divisibility  uint8           `json:"divisibility"`
	Premine       uint128.Uint128 `json:"premine"`
	Rune          runes.Rune      `json:"rune"`
	Spacers       uint32          `json:"spacers"`
	Symbol        string          `json:"symbol"`
	Terms         entryTerms      `json:"terms"`
	Turbo         bool            `json:"turbo"`
	EtchingTxHash string          `json:"etchingTxHash"`
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
	HoldersCount      int64            `json:"holdersCount"`
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
	var holdersCount int64
	if *req.IncludeHoldersCount {
		holdersCount, err = h.usecase.GetTotalHoldersByRuneId(ctx.UserContext(), runeId, blockHeight)
		if err != nil {
			if errors.Is(err, errs.NotFound) {
				return errs.NewPublicError("rune not found")
			}
			return errors.Wrap(err, "error during GetBalancesByRuneId")
		}
	}

	result, err := createTokenInfoResult(runeEntry, holdersCount)
	if err != nil {
		return errors.Wrap(err, "error during createTokenInfoResult")
	}

	resp := getTokenInfoResponse{
		Result: result,
	}

	return errors.WithStack(ctx.JSON(resp))
}

func createTokenInfoResult(runeEntry *runes.RuneEntry, holdersCount int64) (*getTokenInfoResult, error) {
	totalSupply, err := runeEntry.Supply()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get total supply of rune")
	}
	mintedAmount, err := runeEntry.MintedAmount()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get minted amount of rune")
	}
	circulatingSupply := mintedAmount.Sub(runeEntry.BurnedAmount)

	terms := lo.FromPtr(runeEntry.Terms)

	return &getTokenInfoResult{
		Id:                runeEntry.RuneId,
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
		HoldersCount:      holdersCount,
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
				Turbo:         runeEntry.Turbo,
				EtchingTxHash: runeEntry.EtchingTxHash.String(),
			},
		},
	}, nil
}
