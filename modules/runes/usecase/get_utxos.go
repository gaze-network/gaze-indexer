package usecase

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
)

func (u *Usecase) GetRunesUTXOsByPkScript(ctx context.Context, pkScript []byte, blockHeight uint64, limit int32, offset int32) ([]*entity.RunesUTXO, error) {
	balances, err := u.runesDg.GetRunesUTXOsByPkScript(ctx, pkScript, blockHeight, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "error during GetBalancesByPkScript")
	}
	return balances, nil
}

func (u *Usecase) GetRunesUTXOsByRuneIdAndPkScript(ctx context.Context, runeId runes.RuneId, pkScript []byte, blockHeight uint64, limit int32, offset int32) ([]*entity.RunesUTXO, error) {
	balances, err := u.runesDg.GetRunesUTXOsByRuneIdAndPkScript(ctx, runeId, pkScript, blockHeight, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "error during GetBalancesByPkScript")
	}
	return balances, nil
}
