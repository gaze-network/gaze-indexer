package usecase

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
)

func (u *Usecase) GetUnspentOutPointBalancesByPkScript(ctx context.Context, pkScript []byte, blockHeight uint64) ([]*entity.OutPointBalance, error) {
	balances, err := u.runesDg.GetUnspentOutPointBalancesByPkScript(ctx, pkScript, blockHeight)
	if err != nil {
		return nil, errors.Wrap(err, "error during GetBalancesByPkScript")
	}
	return balances, nil
}
