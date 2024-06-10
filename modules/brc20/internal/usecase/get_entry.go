package usecase

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/common/errs"
	"github.com/gaze-network/indexer-network/modules/brc20/internal/entity"
)

func (u *Usecase) GetTickEntryByTickBatch(ctx context.Context, ticks []string) (map[string]*entity.TickEntry, error) {
	entries, err := u.dg.GetTickEntriesByTicks(ctx, ticks)
	if err != nil {
		return nil, errors.Wrap(err, "error during GetTickEntriesByTicks")
	}
	return entries, nil
}

func (u *Usecase) GetTickEntryByTickAndHeight(ctx context.Context, tick string, blockHeight uint64) (*entity.TickEntry, error) {
	entries, err := u.GetTickEntryByTickAndHeightBatch(ctx, []string{tick}, blockHeight)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	entry, ok := entries[tick]
	if !ok {
		return nil, errors.Wrap(errs.NotFound, "entry not found")
	}
	return entry, nil
}

func (u *Usecase) GetTickEntryByTickAndHeightBatch(ctx context.Context, ticks []string, blockHeight uint64) (map[string]*entity.TickEntry, error) {
	return nil, nil
}
