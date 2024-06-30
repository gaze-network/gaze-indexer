package httphandler

import (
	"context"
	"slices"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/pkg/btcutils"
	"github.com/gaze-network/indexer-network/pkg/decimals"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

type getBalancesByAddressBatchRequest struct {
	Queries []getBalancesByAddressRequest `json:"queries"`
}

func (r getBalancesByAddressBatchRequest) Validate() error {
	var errList []error
	for _, query := range r.Queries {
		if query.Wallet == "" {
			errList = append(errList, errors.Errorf("queries[%d]: 'wallet' is required"))
		}
	}
	return errs.WithPublicMessage(errors.Join(errList...), "validation error")
}

type getBalancesByAddressBatchResult struct {
	List []*getBalancesByAddressResult `json:"list"`
}

type getBalancesByAddressBatchResponse = common.HttpResponse[getBalancesByAddressBatchResult]

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

	processQuery := func(ctx context.Context, query getBalancesByAddressRequest) (*getBalancesByAddressResult, error) {
		pkScript, err := btcutils.ToPkScript(h.network, query.Wallet)
		if err != nil {
			return nil, errs.NewPublicError("unable to resolve pkscript from \"wallet\"")
		}

		blockHeight := query.BlockHeight
		if blockHeight == 0 {
			blockHeight = latestBlockHeight
		}

		balances, err := h.usecase.GetBalancesByPkScript(ctx, pkScript, blockHeight)
		if err != nil {
			return nil, errors.Wrap(err, "error during GetBalancesByPkScript")
		}

		balanceRuneIds := lo.Keys(balances)
		entries, err := h.usecase.GetTickEntryByTickBatch(ctx, balanceRuneIds)
		if err != nil {
			return nil, errors.Wrap(err, "error during GetTickEntryByTickBatch")
		}

		balanceList := make([]balance, 0, len(balances))
		for id, b := range balances {
			entry := entries[id]
			balanceList = append(balanceList, balance{
				Amount:   decimals.ToUint256(b.OverallBalance, entry.Decimals),
				Id:       id,
				Name:     entry.OriginalTick,
				Symbol:   entry.Tick,
				Decimals: entry.Decimals,
				Extend: balanceExtend{
					Transferable: decimals.ToUint256(b.OverallBalance.Sub(b.AvailableBalance), entry.Decimals),
					Available:    decimals.ToUint256(b.AvailableBalance, entry.Decimals),
				},
			})
		}
		slices.SortFunc(balanceList, func(i, j balance) int {
			return j.Amount.Cmp(i.Amount)
		})

		return &getBalancesByAddressResult{
			BlockHeight: blockHeight,
			List:        balanceList,
		}, nil
	}

	results := make([]*getBalancesByAddressResult, len(req.Queries))
	eg, ectx := errgroup.WithContext(ctx.UserContext())
	for i, query := range req.Queries {
		i := i
		query := query
		eg.Go(func() error {
			result, err := processQuery(ectx, query)
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
