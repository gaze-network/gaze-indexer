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
	Ids              []string `json:"ids"`
	BlockHeight      uint64   `json:"blockHeight"`
	AdditionalFields []string `json:"additionalFields"`
}

const getTokenInfoBatchMaxQueries = 100

func (r *getTokenInfoBatchRequest) Validate() error {
	var errList []error

	if len(r.Ids) == 0 {
		errList = append(errList, errors.New("ids cannot be empty"))
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
	if lo.Contains(req.AdditionalFields, "holdersCount") {
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
		var holdersCount *int64
		if lo.Contains(req.AdditionalFields, "holdersCount") {
			holdersCount = lo.ToPtr(holdersCounts[runeId])
		}

		result, err := createTokenInfoResult(runeEntry, holdersCount)
		if err != nil {
			return errors.Wrap(err, "error during createTokenInfoResult")
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
