package usecase

import (
	"context"

	"github.com/cockroachdb/errors"
)

func (u *Usecase) GetLatestBlockHeight(ctx context.Context) (uint64, error) {
	height, err := u.runesDg.GetLatestBlockHeight(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get latest block height")
	}
	return height, nil
}
