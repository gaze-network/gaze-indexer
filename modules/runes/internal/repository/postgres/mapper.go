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
			heightStart := uint64(src.TermsHeightStart.Int32)
			terms.HeightStart = &heightStart
		}
		if src.TermsHeightEnd.Valid {
			heightEnd := uint64(src.TermsHeightEnd.Int32)
			terms.HeightEnd = &heightEnd
		}
		if src.TermsOffsetStart.Valid {
			offsetStart := uint64(src.TermsOffsetStart.Int32)
			terms.OffsetStart = &offsetStart
		}
		if src.TermsOffsetEnd.Valid {
			offsetEnd := uint64(src.TermsOffsetEnd.Int32)
			terms.OffsetEnd = &offsetEnd
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
	var termsAmount, termsCap pgtype.Numeric
	var termsHeightStart, termsHeightEnd, termsOffsetStart, termsOffsetEnd pgtype.Int4
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
			termsHeightStart = pgtype.Int4{
				Int32: int32(*src.Terms.HeightStart),
				Valid: true,
			}
		}
		if src.Terms.HeightEnd != nil {
			termsHeightEnd = pgtype.Int4{
				Int32: int32(*src.Terms.HeightEnd),
				Valid: true,
			}
		}
		if src.Terms.OffsetStart != nil {
			termsOffsetStart = pgtype.Int4{
				Int32: int32(*src.Terms.OffsetStart),
				Valid: true,
			}
		}
		if src.Terms.OffsetEnd != nil {
			termsOffsetEnd = pgtype.Int4{
				Int32: int32(*src.Terms.OffsetEnd),
				Valid: true,
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
