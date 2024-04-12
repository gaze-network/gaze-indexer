package usecase

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
)

func (u *Usecase) GetRuneIdFromRune(ctx context.Context, rune runes.Rune) (runes.RuneId, error) {
	runeId, err := u.runesDg.GetRuneIdFromRune(ctx, rune)
	if err != nil {
		return runes.RuneId{}, errors.Wrap(err, "failed to get rune entry by rune")
	}
	return runeId, nil
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
