package usecase

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
)

func (u *Usecase) GetTransferableTransfersByPkScript(ctx context.Context, pkScript []byte, blockHeight uint64) ([]*entity.EventInscribeTransfer, error) {
	result, err := u.dg.GetTransferableTransfersByPkScript(ctx, pkScript, blockHeight)
	if err != nil {
		return nil, errors.Wrap(err, "error during GetTransferableTransfersByPkScript")
	}
	return result, nil
}
