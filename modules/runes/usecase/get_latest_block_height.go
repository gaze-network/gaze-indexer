package usecase

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/core/types"
)

func (u *Usecase) GetLatestBlock(ctx context.Context) (types.BlockHeader, error) {
	blockHeader, err := u.runesDg.GetLatestBlock(ctx)
	if err != nil {
		return types.BlockHeader{}, errors.Wrap(err, "failed to get latest block")
	}
	return blockHeader, nil
}
