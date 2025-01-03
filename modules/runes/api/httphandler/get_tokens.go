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
	Search              string         `query:"search"`
	BlockHeight         uint64         `query:"blockHeight"`
	Scope               GetTokensScope `query:"scope"`
	AdditionalFieldsRaw string         `query:"additionalFields"` // comma-separated list of additional fields
	AdditionalFields    []string
}

func (r *getTokensRequest) Validate() error {
	var errList []error
	if err := r.paginationRequest.Validate(); err != nil {
		errList = append(errList, err)
	}
	if r.Limit > getTokensMaxLimit {
		errList = append(errList, errors.Errorf("limit must be less than or equal to 1000"))
	}
	if r.Scope != "" && !r.Scope.IsValid() {
		errList = append(errList, errors.Errorf("invalid scope: %s", r.Scope))
	}

	if r.AdditionalFieldsRaw == "" {
		// temporarily set default value for backward compatibility
		r.AdditionalFieldsRaw = "holdersCount" // TODO: remove this default value after all clients are updated
	}
	r.AdditionalFields = strings.Split(r.AdditionalFieldsRaw, ",")

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
	List []*getTokenInfoResult `json:"list"`
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
	holdersCounts := make(map[runes.RuneId]int64)
	if lo.Contains(req.AdditionalFields, "holdersCount") {
		holdersCounts, err = h.usecase.GetTotalHoldersByRuneIds(ctx.UserContext(), runeIds, blockHeight)
		if err != nil {
			return errors.Wrap(err, "error during GetTotalHoldersByRuneIds")
		}
	}

	results := make([]*getTokenInfoResult, 0, len(entries))
	for _, ent := range entries {
		var holdersCount *int64
		if lo.Contains(req.AdditionalFields, "holdersCount") {
			holdersCount = lo.ToPtr(holdersCounts[ent.RuneId])
		}
		result, err := createTokenInfoResult(ent, holdersCount)
		if err != nil {
			return errors.Wrap(err, "error during createTokenInfoResult")
		}

		results = append(results, result)
	}

	return errors.WithStack(ctx.JSON(getTokensResponse{
		Result: &getTokensResult{
			List: results,
		},
	}))
}
