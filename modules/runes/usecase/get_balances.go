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

func (u *Usecase) GetTotalHoldersByRuneIds(ctx context.Context, runeIds []runes.RuneId, blockHeight uint64) (map[runes.RuneId]int64, error) {
	holders, err := u.runesDg.GetTotalHoldersByRuneIds(ctx, runeIds, blockHeight)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get total holders by rune ids")
	}
	return holders, nil
}

func (u *Usecase) GetTotalHoldersByRuneId(ctx context.Context, runeId runes.RuneId, blockHeight uint64) (int64, error) {
	holders, err := u.runesDg.GetTotalHoldersByRuneIds(ctx, []runes.RuneId{runeId}, blockHeight)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get total holders by rune ids")
	}

	// defaults to zero holders if not found
	return holders[runeId], nil
}
