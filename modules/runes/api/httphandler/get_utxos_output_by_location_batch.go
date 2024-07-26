package httphandler

import (
	"context"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/indexer-network/modules/runes/usecase"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

type getUTXOsOutputByLocationQuery struct {
	TxHash      string `json:"txHash"`
	OutputIndex int32  `json:"outputIndex"`
}

type getUTXOsOutputByLocationBatchRequest struct {
	Queries []getUTXOsOutputByLocationQuery `json:"queries"`
}

const getUTXOsOutputByLocationBatchMaxQueries = 100

func (r getUTXOsOutputByLocationBatchRequest) Validate() error {
	var errList []error
	if len(r.Queries) == 0 {
		errList = append(errList, errors.New("at least one query is required"))
	}
	if len(r.Queries) > getUTXOsOutputByLocationBatchMaxQueries {
		errList = append(errList, errors.Errorf("cannot exceed %d queries", getUTXOsOutputByLocationBatchMaxQueries))
	}
	for i, query := range r.Queries {
		if query.TxHash == "" {
			errList = append(errList, errors.Errorf("queries[%d]: 'txHash' is required", i))
		}
		if query.OutputIndex < 0 {
			errList = append(errList, errors.Errorf("queries[%d]: 'outputIndex' must be non-negative", i))
		}
	}
	return errs.WithPublicMessage(errors.Join(errList...), "validation error")
}

type getUTXOsOutputByLocationBatchResult struct {
	List []*utxoItem `json:"list"`
}

type getUTXOsOutputByLocationBatchResponse = HttpResponse[getUTXOsOutputByLocationBatchResult]

func (h *HttpHandler) GetUTXOsOutputByLocationBatch(ctx *fiber.Ctx) (err error) {
	var req getUTXOsOutputByLocationBatchRequest
	if err := ctx.BodyParser(&req); err != nil {
		return errors.WithStack(err)
	}
	if err := req.Validate(); err != nil {
		return errors.WithStack(err)
	}

	processQuery := func(ctx context.Context, query getUTXOsOutputByLocationQuery, queryIndex int) (*utxoItem, error) {
		txHash, err := chainhash.NewHashFromStr(query.TxHash)
		if err != nil {
			return nil, errs.NewPublicError(fmt.Sprintf("unable to parse txHash from \"queries[%d].txHash\"", queryIndex))
		}

		utxo, err := h.usecase.GetUTXOsOutputByLocation(ctx, *txHash, uint32(query.OutputIndex))
		if err != nil {
			if errors.Is(err, usecase.ErrUTXONotFound) {
				return nil, errs.NewPublicError(fmt.Sprintf("utxo not found for queries[%d]", queryIndex))
			}
			return nil, errors.WithStack(err)
		}

		runeIds := make(map[runes.RuneId]struct{}, 0)
		for _, balance := range utxo.RuneBalances {
			runeIds[balance.RuneId] = struct{}{}
		}
		runeIdsList := lo.Keys(runeIds)
		runeEntries, err := h.usecase.GetRuneEntryByRuneIdBatch(ctx, runeIdsList)
		if err != nil {
			if errors.Is(err, errs.NotFound) {
				return nil, errs.NewPublicError(fmt.Sprintf("rune entries not found for queries[%d]", queryIndex))
			}
			return nil, errors.Wrap(err, "error during GetRuneEntryByRuneIdBatch")
		}

		runeBalances := make([]runeBalance, 0, len(utxo.RuneBalances))
		for _, balance := range utxo.RuneBalances {
			runeEntry := runeEntries[balance.RuneId]
			runeBalances = append(runeBalances, runeBalance{
				RuneId:       balance.RuneId,
				Rune:         runeEntry.SpacedRune,
				Symbol:       string(runeEntry.Symbol),
				Amount:       balance.Amount,
				Divisibility: runeEntry.Divisibility,
			})
		}

		return &utxoItem{
			TxHash:      utxo.OutPoint.Hash,
			OutputIndex: utxo.OutPoint.Index,
			Sats:        utxo.Sats,
			Extend: utxoExtend{
				Runes: runeBalances,
			},
		}, nil
	}

	results := make([]*utxoItem, len(req.Queries))
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

	resp := getUTXOsOutputByLocationBatchResponse{
		Result: &getUTXOsOutputByLocationBatchResult{
			List: results,
		},
	}

	return errors.WithStack(ctx.JSON(resp))
}
