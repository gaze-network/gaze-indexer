package postgres

import (
	"encoding/hex"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
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

func numericFromUint128(src *uint128.Uint128) (pgtype.Numeric, error) {
	if src == nil {
		return pgtype.Numeric{}, nil
	}
	bytes := []byte(src.String())
	var result pgtype.Numeric
	err := result.UnmarshalJSON(bytes)
	if err != nil {
		return pgtype.Numeric{}, errors.WithStack(err)
	}
	return result, nil
}

func mapRuneEntryModelToType(src gen.RunesEntry) (runes.RuneEntry, error) {
	runeId, err := runes.NewRuneIdFromString(src.RuneID)
	if err != nil {
		return runes.RuneEntry{}, errors.Wrap(err, "failed to parse rune id")
	}
	burnedAmount, err := uint128FromNumeric(src.BurnedAmount)
	if err != nil {
		return runes.RuneEntry{}, errors.Wrap(err, "failed to parse burned amount")
	}
	rune, err := runes.NewRuneFromString(src.Rune)
	if err != nil {
		return runes.RuneEntry{}, errors.Wrap(err, "failed to parse rune")
	}
	mints, err := uint128FromNumeric(src.Mints)
	if err != nil {
		return runes.RuneEntry{}, errors.Wrap(err, "failed to parse mints")
	}
	premine, err := uint128FromNumeric(src.Premine)
	if err != nil {
		return runes.RuneEntry{}, errors.Wrap(err, "failed to parse premine")
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
				return runes.RuneEntry{}, errors.Wrap(err, "failed to parse terms amount")
			}
			terms.Amount = &amount
		}
		if src.TermsCap.Valid {
			cap, err := uint128FromNumeric(src.TermsCap)
			if err != nil {
				return runes.RuneEntry{}, errors.Wrap(err, "failed to parse terms cap")
			}
			terms.Cap = &cap
		}
		if src.TermsHeightStart.Valid {
			heightStart, err := uint128FromNumeric(src.TermsHeightStart)
			if err != nil {
				return runes.RuneEntry{}, errors.Wrap(err, "failed to parse terms height start")
			}
			terms.HeightStart = lo.ToPtr((heightStart.Uint64()))
		}
		if src.TermsHeightEnd.Valid {
			heightEnd, err := uint128FromNumeric(src.TermsHeightEnd)
			if err != nil {
				return runes.RuneEntry{}, errors.Wrap(err, "failed to parse terms height end")
			}
			terms.HeightEnd = lo.ToPtr((heightEnd.Uint64()))
		}
		if src.TermsOffsetStart.Valid {
			offsetStart, err := uint128FromNumeric(src.TermsOffsetStart)
			if err != nil {
				return runes.RuneEntry{}, errors.Wrap(err, "failed to parse terms offset start")
			}
			terms.OffsetStart = lo.ToPtr((offsetStart.Uint64()))
		}
		if src.TermsOffsetEnd.Valid {
			offsetEnd, err := uint128FromNumeric(src.TermsOffsetEnd)
			if err != nil {
				return runes.RuneEntry{}, errors.Wrap(err, "failed to parse terms offset end")
			}
			terms.OffsetEnd = lo.ToPtr((offsetEnd.Uint64()))
		}
	}
	return runes.RuneEntry{
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

func mapRuneEntryTypeToParams(src runes.RuneEntry) (gen.SetRuneEntryParams, error) {
	runeId := src.RuneId.String()
	rune := src.SpacedRune.Rune.String()
	spacers := int32(src.SpacedRune.Spacers)
	burnedAmount, err := numericFromUint128(&src.BurnedAmount)
	if err != nil {
		return gen.SetRuneEntryParams{}, errors.Wrap(err, "failed to parse burned amount")
	}
	mints, err := numericFromUint128(&src.Mints)
	if err != nil {
		return gen.SetRuneEntryParams{}, errors.Wrap(err, "failed to parse mints")
	}
	premine, err := numericFromUint128(&src.Premine)
	if err != nil {
		return gen.SetRuneEntryParams{}, errors.Wrap(err, "failed to parse premine")
	}
	var completionTime pgtype.Timestamp
	if !src.CompletionTime.IsZero() {
		completionTime.Time = src.CompletionTime
		completionTime.Valid = true
	}
	var terms bool
	var termsAmount, termsCap, termsHeightStart, termsHeightEnd, termsOffsetStart, termsOffsetEnd pgtype.Numeric
	if src.Terms != nil {
		terms = true
		if src.Terms.Amount != nil {
			termsAmount, err = numericFromUint128(src.Terms.Amount)
			if err != nil {
				return gen.SetRuneEntryParams{}, errors.Wrap(err, "failed to parse terms amount")
			}
		}
		if src.Terms.Cap != nil {
			termsCap, err = numericFromUint128(src.Terms.Cap)
			if err != nil {
				return gen.SetRuneEntryParams{}, errors.Wrap(err, "failed to parse terms cap")
			}
		}
		if src.Terms.HeightStart != nil {
			termsHeightStart, err = numericFromUint128(lo.ToPtr(uint128.From64(*src.Terms.HeightStart)))
			if err != nil {
				return gen.SetRuneEntryParams{}, errors.Wrap(err, "failed to parse terms height start")
			}
		}
		if src.Terms.HeightEnd != nil {
			termsHeightEnd, err = numericFromUint128(lo.ToPtr(uint128.From64(*src.Terms.HeightEnd)))
			if err != nil {
				return gen.SetRuneEntryParams{}, errors.Wrap(err, "failed to parse terms height end")
			}
		}
		if src.Terms.OffsetStart != nil {
			termsOffsetStart, err = numericFromUint128(lo.ToPtr(uint128.From64(*src.Terms.OffsetStart)))
			if err != nil {
				return gen.SetRuneEntryParams{}, errors.Wrap(err, "failed to parse terms offset start")
			}
		}
		if src.Terms.OffsetEnd != nil {
			termsOffsetEnd, err = numericFromUint128(lo.ToPtr(uint128.From64(*src.Terms.OffsetEnd)))
			if err != nil {
				return gen.SetRuneEntryParams{}, errors.Wrap(err, "failed to parse terms offset end")
			}
		}
	}

	return gen.SetRuneEntryParams{
		RuneID:           runeId,
		Rune:             rune,
		Spacers:          spacers,
		BurnedAmount:     burnedAmount,
		Mints:            mints,
		Premine:          premine,
		Symbol:           src.Symbol,
		Divisibility:     int16(src.Divisibility),
		Terms:            terms,
		TermsAmount:      termsAmount,
		TermsCap:         termsCap,
		TermsHeightStart: termsHeightStart,
		TermsHeightEnd:   termsHeightEnd,
		TermsOffsetStart: termsOffsetStart,
		TermsOffsetEnd:   termsOffsetEnd,
		CompletionTime:   completionTime,
	}, nil
}

func mapBalanceModelToType(src gen.RunesBalance) (*entity.Balance, error) {
	runeId, err := runes.NewRuneIdFromString(src.RuneID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse rune id")
	}
	amount, err := uint128FromNumeric(src.Amount)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse balance")
	}
	pkScript, err := hex.DecodeString(src.Pkscript)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse pkscript")
	}
	return &entity.Balance{
		PkScript:    pkScript,
		RuneId:      runeId,
		Amount:      amount,
		BlockHeight: uint64(src.BlockHeight),
	}, nil
}
