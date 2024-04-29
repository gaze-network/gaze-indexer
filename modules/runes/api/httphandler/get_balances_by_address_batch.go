package httphandler

import (
	"context"
	"fmt"
	"slices"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

type getBalanceQuery struct {
	Wallet      string `json:"wallet"`
	Id          string `json:"id"`
	BlockHeight uint64 `json:"blockHeight"`
}

type getBalancesByAddressBatchRequest struct {
	Queries []getBalanceQuery `json:"queries"`
}

func (r getBalancesByAddressBatchRequest) Validate() error {
	var errList []error
	for _, query := range r.Queries {
		if query.Wallet == "" {
			errList = append(errList, errors.Errorf("queries[%d]: 'wallet' is required"))
		}
		if query.Id != "" && !isRuneIdOrRuneName(query.Id) {
			errList = append(errList, errors.Errorf("queries[%d]: 'id' is not valid rune id or rune name"))
		}
	}
	return errs.WithPublicMessage(errors.Join(errList...), "validation error")
}

type getBalancesByAddressBatchResult struct {
	List []*getBalancesByAddressResult `json:"list"`
}

type getBalancesByAddressBatchResponse = HttpResponse[getBalancesByAddressBatchResult]

func (h *HttpHandler) GetBalancesByAddressBatch(ctx *fiber.Ctx) (err error) {
	var req getBalancesByAddressBatchRequest
	if err := ctx.BodyParser(&req); err != nil {
		return errors.WithStack(err)
	}
	if err := req.Validate(); err != nil {
		return errors.WithStack(err)
	}

	var latestBlockHeight uint64
	blockHeader, err := h.usecase.GetLatestBlock(ctx.UserContext())
	if err != nil {
		return errors.Wrap(err, "error during GetLatestBlock")
	}
	latestBlockHeight = uint64(blockHeader.Height)

	processQuery := func(ctx context.Context, query getBalanceQuery, queryIndex int) (*getBalancesByAddressResult, error) {
		pkScript, ok := resolvePkScript(h.network, query.Wallet)
		if !ok {
			return nil, errs.NewPublicError(fmt.Sprintf("unable to resolve pkscript from \"queries[%d].wallet\"", queryIndex))
		}

		blockHeight := query.BlockHeight
		if blockHeight == 0 {
			blockHeight = latestBlockHeight
		}

		balances, err := h.usecase.GetBalancesByPkScript(ctx, pkScript, blockHeight)
		if err != nil {
			return nil, errors.Wrap(err, "error during GetBalancesByPkScript")
		}

		runeId, ok := h.resolveRuneId(ctx, query.Id)
		if ok {
			// filter out balances that don't match the requested rune id
			for key := range balances {
				if key != runeId {
					delete(balances, key)
				}
			}
		}

		balanceRuneIds := lo.Keys(balances)
		runeEntries, err := h.usecase.GetRuneEntryByRuneIdBatch(ctx, balanceRuneIds)
		if err != nil {
			return nil, errors.Wrap(err, "error during GetRuneEntryByRuneIdBatch")
		}

		balanceList := make([]balance, 0, len(balances))
		for id, b := range balances {
			runeEntry := runeEntries[id]
			balanceList = append(balanceList, balance{
				Amount:   b.Amount,
				Id:       id,
				Name:     runeEntry.SpacedRune,
				Symbol:   string(runeEntry.Symbol),
				Decimals: runeEntry.Divisibility,
			})
		}
		slices.SortFunc(balanceList, func(i, j balance) int {
			return j.Amount.Cmp(i.Amount)
		})

		result := getBalancesByAddressResult{
			BlockHeight: blockHeight,
			List:        balanceList,
		}
		return &result, nil
	}

	results := make([]*getBalancesByAddressResult, len(req.Queries))
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

	resp := getBalancesByAddressBatchResponse{
		Result: &getBalancesByAddressBatchResult{
			List: results,
		},
	}

	return errors.WithStack(ctx.JSON(resp))
}
