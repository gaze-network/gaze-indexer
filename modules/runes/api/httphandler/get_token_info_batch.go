package httphandler

import (
	"fmt"
	"net/url"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
)

type getTokenInfoBatchRequest struct {
	Ids                 []string `json:"ids"`
	BlockHeight         uint64   `json:"blockHeight"`
	IncludeHoldersCount *bool    `json:"includeHoldersCount"`
}

func (r *getTokenInfoBatchRequest) ParseDefault() error {
	if r.IncludeHoldersCount == nil {
		r.IncludeHoldersCount = lo.ToPtr(false)
	}
	return nil
}

const getTokenInfoBatchMaxQueries = 100

func (r *getTokenInfoBatchRequest) Validate() error {
	var errList []error

	if len(r.Ids) == 0 {
		errList = append(errList, errors.New("at least one query is required"))
	}
	if len(r.Ids) > getTokenInfoBatchMaxQueries {
		errList = append(errList, errors.Errorf("cannot query more than %d ids", getTokenInfoBatchMaxQueries))
	}
	for i := range r.Ids {
		id, err := url.QueryUnescape(r.Ids[i])
		if err != nil {
			return errors.WithStack(err)
		}
		r.Ids[i] = id
		if !isRuneIdOrRuneName(r.Ids[i]) {
			errList = append(errList, errors.Errorf("ids[%d]: id '%s' is not valid rune id or rune name", i, r.Ids[i]))
		}
	}

	return errs.WithPublicMessage(errors.Join(errList...), "validation error")
}

type getTokenInfoBatchResult struct {
	List []*getTokenInfoResult `json:"list"`
}
type getTokenInfoBatchResponse = HttpResponse[getTokenInfoBatchResult]

func (h *HttpHandler) GetTokenInfoBatch(ctx *fiber.Ctx) (err error) {
	var req getTokenInfoBatchRequest
	if err := ctx.BodyParser(&req); err != nil {
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

	runeIds := make([]runes.RuneId, 0)
	for i, id := range req.Ids {
		runeId, ok := h.resolveRuneId(ctx.UserContext(), id)
		if !ok {
			return errs.NewPublicError(fmt.Sprintf("unable to resolve rune id \"%s\" from \"ids[%d]\"", id, i))
		}
		runeIds = append(runeIds, runeId)
	}

	runeEntries, err := h.usecase.GetRuneEntryByRuneIdAndHeightBatch(ctx.UserContext(), runeIds, blockHeight)
	if err != nil {
		return errors.Wrap(err, "error during GetRuneEntryByRuneIdAndHeightBatch")
	}
	holdersCounts := make(map[runes.RuneId]int64)
	if *req.IncludeHoldersCount {
		holdersCounts, err = h.usecase.GetTotalHoldersByRuneIds(ctx.UserContext(), runeIds, blockHeight)
		if err != nil {
			if errors.Is(err, errs.NotFound) {
				return errs.NewPublicError("rune not found")
			}
			return errors.Wrap(err, "error during GetBalancesByRuneId")
		}
	}

	results := make([]*getTokenInfoResult, 0, len(runeIds))

	for _, runeId := range runeIds {
		runeEntry, ok := runeEntries[runeId]
		if !ok {
			return errs.NewPublicError(fmt.Sprintf("rune not found: %s", runeId))
		}
		holdersCount := holdersCounts[runeId]

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

		result := &getTokenInfoResult{
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
		}
		results = append(results, result)
	}

	resp := getTokenInfoBatchResponse{
		Result: &getTokenInfoBatchResult{
			List: results,
		},
	}

	return errors.WithStack(ctx.JSON(resp))
}
