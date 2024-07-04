package usecase

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
)

// Use limit = -1 as no limit.
func (u *Usecase) GetBalancesByPkScript(ctx context.Context, pkScript []byte, blockHeight uint64, limit int32, offset int32) ([]*entity.Balance, error) {
	balances, err := u.runesDg.GetBalancesByPkScript(ctx, pkScript, blockHeight, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "error during GetBalancesByPkScript")
	}
	return balances, nil
}

// Use limit = -1 as no limit.
func (u *Usecase) GetBalancesByRuneId(ctx context.Context, runeId runes.RuneId, blockHeight uint64, limit int32, offset int32) ([]*entity.Balance, error) {
	balances, err := u.runesDg.GetBalancesByRuneId(ctx, runeId, blockHeight, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rune holders by rune id")
	}
	return balances, nil
}
