package postgres

import (
	"context"

	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/runes/internal/datagateway"
	"github.com/gaze-network/indexer-network/modules/runes/internal/repository/postgres/gen"
	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
	"github.com/gaze-network/uint128"
	"github.com/jackc/pgx/v5"
)

var _ datagateway.RunesDataGateway = (*Repository)(nil)

func (r *Repository) GetRunesBalancesAtOutPoint(ctx context.Context, outPoint wire.OutPoint) (map[runes.RuneId]uint128.Uint128, error) {
	balances, err := r.queries.GetOutPointBalances(ctx, gen.GetOutPointBalancesParams{
		TxHash: outPoint.Hash.String(),
		TxIdx:  int32(outPoint.Index),
	})
	if err != nil {
		return nil, errors.Wrap(err, "error during query")
	}

	result := make(map[runes.RuneId]uint128.Uint128, len(balances))
	for _, balance := range balances {
		runeId, err := runes.NewRuneIdFromString(balance.RuneID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse RuneId")
		}
		value, err := uint128FromNumeric(balance.Value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse balance")
		}
		result[runeId] = value
	}
	return result, nil
}

func (r *Repository) GetRuneIdByRune(ctx context.Context, rune runes.Rune) (runes.RuneId, error) {
	runeEntryModel, err := r.queries.GetRuneEntryByRune(ctx, rune.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return runes.RuneId{}, errors.WithStack(errs.NotFound)
		}
		return runes.RuneId{}, errors.Wrap(err, "error during query")
	}

	runeId, err := runes.NewRuneIdFromString(runeEntryModel.RuneID)
	if err != nil {
		return runes.RuneId{}, errors.Wrap(err, "failed to parse rune id")
	}
	return runeId, nil
}

func (r *Repository) GetRuneEntryByRuneId(ctx context.Context, runeId runes.RuneId) (*runes.RuneEntry, error) {
	runeEntryModel, err := r.queries.GetRuneEntryByRuneId(ctx, runeId.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.WithStack(errs.NotFound)
		}
		return nil, errors.Wrap(err, "error during query")
	}

	runeEntry, err := mapRuneEntryModelToType(runeEntryModel)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse rune entry model")
	}
	return runeEntry, nil
}

func (r *Repository) SetRuneEntry(ctx context.Context, entry *runes.RuneEntry) error {
	panic("implement me")
}

func (r *Repository) CreateRunesBalanceAtOutPoint(ctx context.Context, outPoint wire.OutPoint, balances map[runes.RuneId]uint128.Uint128) error {
	panic("implement me")
}

func (r *Repository) CreateRuneBalance(ctx context.Context, pkScript string, runeId runes.RuneId, blockHeight uint64, balance uint128.Uint128) error {
	panic("implement me")
}

func (r *Repository) UpdateLatestBlockHeight(ctx context.Context, blockHeight uint64) error {
	panic("implement me")
}
