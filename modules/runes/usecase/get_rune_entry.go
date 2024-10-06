package usecase

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
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
		return nil, errors.Wrap(err, "failed to get rune entries by rune ids")
	}
	return runeEntry, nil
}

func (u *Usecase) GetRuneEntryByRuneIdAndHeight(ctx context.Context, runeId runes.RuneId, blockHeight uint64) (*runes.RuneEntry, error) {
	runeEntry, err := u.runesDg.GetRuneEntryByRuneIdAndHeight(ctx, runeId, blockHeight)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rune entry by rune id and height")
	}
	return runeEntry, nil
}

func (u *Usecase) GetRuneEntryByRuneIdAndHeightBatch(ctx context.Context, runeIds []runes.RuneId, blockHeight uint64) (map[runes.RuneId]*runes.RuneEntry, error) {
	runeEntry, err := u.runesDg.GetRuneEntryByRuneIdAndHeightBatch(ctx, runeIds, blockHeight)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rune entries by rune ids and height")
	}
	return runeEntry, nil
}

func (u *Usecase) GetRuneEntries(ctx context.Context, search string, blockHeight uint64, limit, offset int32) ([]*runes.RuneEntry, error) {
	entries, err := u.runesDg.GetRuneEntries(ctx, search, blockHeight, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to listing rune entries")
	}
	return entries, nil
}

func (u *Usecase) GetMintingRuneEntries(ctx context.Context, search string, blockHeight uint64, limit, offset int32) ([]*runes.RuneEntry, error) {
	entries, err := u.runesDg.GetMintingRuneEntries(ctx, search, blockHeight, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to listing rune entries")
	}
	return entries, nil
}
