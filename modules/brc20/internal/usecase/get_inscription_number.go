package usecase

import (
	"context"

	"github.com/cockroachdb/errors"
)

func (u *Usecase) GetFirstLastInscriptionNumberByTick(ctx context.Context, tick string) (int64, int64, error) {
	first, last, err := u.dg.GetFirstLastInscriptionNumberByTick(ctx, tick)
	if err != nil {
		return -1, -1, errors.Wrap(err, "error during GetFirstLastInscriptionNumberByTick")
	}
	return first, last, nil
}
