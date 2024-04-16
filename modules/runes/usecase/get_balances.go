package usecase

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
)

func (u *Usecase) GetBalancesByPkScript(ctx context.Context, pkScript []byte, blockHeight uint64) (map[runes.RuneId]*entity.Balance, error) {
	balances, err := u.runesDg.GetBalancesByPkScript(ctx, pkScript, blockHeight)
	if err != nil {
		return nil, errors.Wrap(err, "error during GetBalancesByPkScript")
	}
	return balances, nil
}
