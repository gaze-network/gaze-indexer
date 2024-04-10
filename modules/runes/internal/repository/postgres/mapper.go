package postgres

import (
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/runes/internal/repository/postgres/gen"
	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
	"github.com/gaze-network/uint128"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/samber/lo"
)

func uint128FromNumeric(src pgtype.Numeric) (uint128.Uint128, error) {
	bytes, err := src.MarshalJSON()
	if err != nil {
		return uint128.Uint128{}, errors.WithStack(err)
	}
	result, err := uint128.FromString(string(bytes))
	if err != nil {
		return uint128.Uint128{}, errors.WithStack(err)
	}
	return result, nil
}

func mapRuneEntryModelToType(src gen.RunesEntry) (*runes.RuneEntry, error) {
	runeId, err := runes.NewRuneIdFromString(src.RuneID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse rune id")
	}
	burnedAmount, err := uint128FromNumeric(src.BurnedAmount)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse burned amount")
	}
	rune, err := runes.NewRuneFromString(src.Rune)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse rune")
	}
	mints, err := uint128FromNumeric(src.Mints)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse mints")
	}
	premine, err := uint128FromNumeric(src.Premine)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse premine")
	}
	var completionTime time.Time
	if src.CompletionTime.Valid {
		completionTime = src.CompletionTime.Time
	}
	var terms *runes.Terms
	if src.Terms {
		terms = &runes.Terms{}
		if src.TermsAmount.Valid {
			amount, err := uint128FromNumeric(src.TermsAmount)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse terms amount")
			}
			terms.Amount = &amount
		}
		if src.TermsCap.Valid {
			cap, err := uint128FromNumeric(src.TermsCap)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse terms cap")
			}
			terms.Cap = &cap
		}
		if src.TermsHeightStart.Valid {
			heightStart, err := uint128FromNumeric(src.TermsHeightStart)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse terms height start")
			}
			terms.HeightStart = lo.ToPtr((heightStart.Uint64()))
		}
		if src.TermsHeightEnd.Valid {
			heightEnd, err := uint128FromNumeric(src.TermsHeightEnd)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse terms height end")
			}
			terms.HeightEnd = lo.ToPtr((heightEnd.Uint64()))
		}
		if src.TermsOffsetStart.Valid {
			offsetStart, err := uint128FromNumeric(src.TermsOffsetStart)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse terms offset start")
			}
			terms.OffsetStart = lo.ToPtr((offsetStart.Uint64()))
		}
		if src.TermsOffsetEnd.Valid {
			offsetEnd, err := uint128FromNumeric(src.TermsOffsetEnd)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse terms offset end")
			}
			terms.OffsetEnd = lo.ToPtr((offsetEnd.Uint64()))
		}
	}
	return &runes.RuneEntry{
		RuneId:         runeId,
		SpacedRune:     runes.NewSpacedRune(rune, uint32(src.Spacers)),
		Mints:          mints,
		BurnedAmount:   burnedAmount,
		Premine:        premine,
		Symbol:         src.Symbol,
		Divisibility:   uint8(src.Divisibility),
		CompletionTime: completionTime,
		Terms:          terms,
	}, nil
}
