package usecase

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
)

func (u *Usecase) GetDeployEventByTick(ctx context.Context, tick string) (*entity.EventDeploy, error) {
	result, err := u.dg.GetDeployEventByTick(ctx, tick)
	if err != nil {
		return nil, errors.Wrap(err, "error during GetDeployEventByTick")
	}
	return result, nil
}
