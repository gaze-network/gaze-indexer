package usecase

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
)

func (u *Usecase) GetDeployEvents(ctx context.Context, pkScript []byte, tick string, height uint64) ([]*entity.EventDeploy, error) {
	result, err := u.dg.GetDeployEvents(ctx, pkScript, tick, height)
	if err != nil {
		return nil, errors.Wrap(err, "error during GetDeployEvents")
	}
	return result, nil
}

func (u *Usecase) GetMintEvents(ctx context.Context, pkScript []byte, tick string, height uint64) ([]*entity.EventMint, error) {
	result, err := u.dg.GetMintEvents(ctx, pkScript, tick, height)
	if err != nil {
		return nil, errors.Wrap(err, "error during GetMintEvents")
	}
	return result, nil
}

func (u *Usecase) GetInscribeTransferEvents(ctx context.Context, pkScript []byte, tick string, height uint64) ([]*entity.EventInscribeTransfer, error) {
	result, err := u.dg.GetInscribeTransferEvents(ctx, pkScript, tick, height)
	if err != nil {
		return nil, errors.Wrap(err, "error during GetInscribeTransferEvents")
	}
	return result, nil
}

func (u *Usecase) GetTransferTransferEvents(ctx context.Context, pkScript []byte, tick string, height uint64) ([]*entity.EventTransferTransfer, error) {
	result, err := u.dg.GetTransferTransferEvents(ctx, pkScript, tick, height)
	if err != nil {
		return nil, errors.Wrap(err, "error during GetTransferTransfersEvents")
	}
	return result, nil
}
