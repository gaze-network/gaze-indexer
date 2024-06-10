package usecase

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
)

func (u *Usecase) GetBalancesByPkScript(ctx context.Context, pkScript []byte, blockHeight uint64) (map[string]*entity.Balance, error) {
	balances, err := u.dg.GetBalancesByPkScript(ctx, pkScript, blockHeight)
	if err != nil {
		return nil, errors.Wrap(err, "error during GetBalancesByPkScript")
	}
	return balances, nil
}

func (u *Usecase) GetBalancesByTick(ctx context.Context, tick string, blockHeight uint64) ([]*entity.Balance, error) {
	balances, err := u.dg.GetBalancesByTick(ctx, tick, blockHeight)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get balance by tick")
	}
	return balances, nil
}
