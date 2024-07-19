package httphandler

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

type getBalanceQuery struct {
	Wallet      string `json:"wallet"`
	Id          string `json:"id"`
	BlockHeight uint64 `json:"blockHeight"`
	Limit       int32  `json:"limit"`
	Offset      int32  `json:"offset"`
}

type getBalancesBatchRequest struct {
	Queries []getBalanceQuery `json:"queries"`
}

const getBalancesBatchMaxQueries = 100

func (r getBalancesBatchRequest) Validate() error {
	var errList []error
	if len(r.Queries) == 0 {
		errList = append(errList, errors.New("at least one query is required"))
	}
	if len(r.Queries) > getBalancesBatchMaxQueries {
		errList = append(errList, errors.Errorf("cannot exceed %d queries", getBalancesBatchMaxQueries))
	}
	for i, query := range r.Queries {
		if query.Wallet == "" {
			errList = append(errList, errors.Errorf("queries[%d]: 'wallet' is required", i))
		}
		if query.Id != "" && !isRuneIdOrRuneName(query.Id) {
			errList = append(errList, errors.Errorf("queries[%d]: 'id' is not valid rune id or rune name", i))
		}
		if query.Limit < 0 {
			errList = append(errList, errors.Errorf("queries[%d]: 'limit' must be non-negative", i))
		}
		if query.Limit > getBalancesMaxLimit {
			errList = append(errList, errors.Errorf("queries[%d]: 'limit' cannot exceed %d", i, getBalancesMaxLimit))
		}
	}
	return errs.WithPublicMessage(errors.Join(errList...), "validation error")
}

type getBalancesBatchResult struct {
	List []*getBalancesResult `json:"list"`
}

type getBalancesBatchResponse = HttpResponse[getBalancesBatchResult]

func (h *HttpHandler) GetBalancesBatch(ctx *fiber.Ctx) (err error) {
	var req getBalancesBatchRequest
	if err := ctx.BodyParser(&req); err != nil {
		return errors.WithStack(err)
	}
	if err := req.Validate(); err != nil {
		return errors.WithStack(err)
	}

	var latestBlockHeight uint64
	blockHeader, err := h.usecase.GetLatestBlock(ctx.UserContext())
	if err != nil {
		if errors.Is(err, errs.NotFound) {
			return errs.NewPublicError("latest block not found")
		}
		return errors.Wrap(err, "error during GetLatestBlock")
	}
	latestBlockHeight = uint64(blockHeader.Height)

	processQuery := func(ctx context.Context, query getBalanceQuery, queryIndex int) (*getBalancesResult, error) {
		pkScript, ok := resolvePkScript(h.network, query.Wallet)
		if !ok {
			return nil, errs.NewPublicError(fmt.Sprintf("unable to resolve pkscript from \"queries[%d].wallet\"", queryIndex))
		}

		blockHeight := query.BlockHeight
		if blockHeight == 0 {
			blockHeight = latestBlockHeight
		}

		if query.Limit == 0 {
			query.Limit = getBalancesMaxLimit
		}

		balances, err := h.usecase.GetBalancesByPkScript(ctx, pkScript, blockHeight, query.Limit, query.Offset)
		if err != nil {
			if errors.Is(err, errs.NotFound) {
				return nil, errs.NewPublicError("balances not found")
			}
			return nil, errors.Wrap(err, "error during GetBalancesByPkScript")
		}

		runeId, ok := h.resolveRuneId(ctx, query.Id)
		if ok {
			// filter out balances that don't match the requested rune id
			balances = lo.Filter(balances, func(b *entity.Balance, _ int) bool {
				return b.RuneId == runeId
			})
		}

		balanceRuneIds := lo.Map(balances, func(b *entity.Balance, _ int) runes.RuneId {
			return b.RuneId
		})
		runeEntries, err := h.usecase.GetRuneEntryByRuneIdBatch(ctx, balanceRuneIds)
		if err != nil {
			if errors.Is(err, errs.NotFound) {
				return nil, errs.NewPublicError("rune not found")
			}
			return nil, errors.Wrap(err, "error during GetRuneEntryByRuneIdBatch")
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

		result := getBalancesResult{
			BlockHeight: blockHeight,
			List:        balanceList,
		}
		return &result, nil
	}

	results := make([]*getBalancesResult, len(req.Queries))
	eg, ectx := errgroup.WithContext(ctx.UserContext())
	for i, query := range req.Queries {
		i := i
		query := query
		eg.Go(func() error {
			result, err := processQuery(ectx, query, i)
			if err != nil {
				return errors.Wrapf(err, "error during processQuery for query %d", i)
			}
			results[i] = result
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return errors.WithStack(err)
	}

	resp := getBalancesBatchResponse{
		Result: &getBalancesBatchResult{
			List: results,
		},
	}

	return errors.WithStack(ctx.JSON(resp))
}
