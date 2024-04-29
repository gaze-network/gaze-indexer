package usecase

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
)

func (u *Usecase) GetBalancesByPkScript(ctx context.Context, pkScript []byte, blockHeight uint64) (map[runes.RuneId]*entity.Balance, error) {
	balances, err := u.runesDg.GetBalancesByPkScript(ctx, pkScript, blockHeight)
	if err != nil {
		return nil, errors.Wrap(err, "error during GetBalancesByPkScript")
	}
	return balances, nil
}

func (u *Usecase) GetBalancesByRuneId(ctx context.Context, runeId runes.RuneId, blockHeight uint64) ([]*entity.Balance, error) {
	balances, err := u.runesDg.GetBalancesByRuneId(ctx, runeId, blockHeight)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rune holders by rune id")
	}
	return balances, nil
}
