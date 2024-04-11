package usecase

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
)

func (u *Usecase) GetRuneEntryByRune(ctx context.Context, rune runes.Rune) (*runes.RuneEntry, error) {
	runeEntry, err := u.runesDg.GetRuneEntryByRune(ctx, rune)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rune entry by rune")
	}
	return runeEntry, nil
}

func (u *Usecase) GetRuneEntryByRuneId(ctx context.Context, runeId runes.RuneId) (*runes.RuneEntry, error) {
	runeEntry, err := u.runesDg.GetRuneEntryByRuneId(ctx, runeId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rune entry by rune id")
	}
	return runeEntry, nil
}

func (u *Usecase) GetRuneEntryByRuneIdBatch(ctx context.Context, runeIds []runes.RuneId) (map[runes.RuneId]*runes.RuneEntry, error) {
	runeEntry, err := u.runesDg.GetRuneEntryByRuneIdBatch(ctx, runeIds)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rune entry by rune id")
	}
	return runeEntry, nil
}
