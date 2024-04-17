package usecase

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
)

func (u *Usecase) GetTransactionsByHeight(ctx context.Context, height uint64) ([]*entity.RuneTransaction, error) {
	txs, err := u.runesDg.GetRuneTransactionsByHeight(ctx, height)
	if err != nil {
		return nil, errors.Wrap(err, "error during GetTransactionsByHeight")
	}
	return txs, nil
}
