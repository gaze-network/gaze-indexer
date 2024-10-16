package postgres

import (
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/repository/postgres/gen"
	"github.com/gaze-network/indexer-network/modules/runes/runes"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"github.com/gaze-network/uint128"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/samber/lo"
)

func uint128FromNumeric(src pgtype.Numeric) (*uint128.Uint128, error) {
	if !src.Valid {
		return nil, nil
	}
	bytes, err := src.MarshalJSON()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	result, err := uint128.FromString(string(bytes))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &result, nil
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

func mapIndexerStateModelToType(src gen.RunesIndexerState) entity.IndexerState {
	var createdAt time.Time
	if src.CreatedAt.Valid {
		createdAt = src.CreatedAt.Time.UTC()
	}
	return entity.IndexerState{
		DBVersion:        src.DbVersion,
		EventHashVersion: src.EventHashVersion,
		CreatedAt:        createdAt,
	}
}

func mapIndexerStateTypeToParams(src entity.IndexerState) gen.SetIndexerStateParams {
	return gen.SetIndexerStateParams{
		DbVersion:        src.DBVersion,
		EventHashVersion: src.EventHashVersion,
	}
}

func mapRuneEntryModelToType(src gen.GetRuneEntriesRow) (runes.RuneEntry, error) {
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
	var completedAt time.Time
	if src.CompletedAt.Valid {
		completedAt = src.CompletedAt.Time.UTC()
	}
	var completedAtHeight *uint64
	if src.CompletedAtHeight.Valid {
		completedAtHeight = lo.ToPtr(uint64(src.CompletedAtHeight.Int32))
	}
	var terms *runes.Terms
	if src.Terms {
		terms = &runes.Terms{}
		if src.TermsAmount.Valid {
			amount, err := uint128FromNumeric(src.TermsAmount)
			if err != nil {
				return runes.RuneEntry{}, errors.Wrap(err, "failed to parse terms amount")
			}
			terms.Amount = amount
		}
		if src.TermsCap.Valid {
			cap, err := uint128FromNumeric(src.TermsCap)
			if err != nil {
				return runes.RuneEntry{}, errors.Wrap(err, "failed to parse terms cap")
			}
			terms.Cap = cap
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
	etchingTxHash, err := chainhash.NewHashFromStr(src.EtchingTxHash)
	if err != nil {
		return runes.RuneEntry{}, errors.Wrap(err, "failed to parse etching tx hash")
	}
	var etchedAt time.Time
	if src.EtchedAt.Valid {
		etchedAt = src.EtchedAt.Time.UTC()
	}
	return runes.RuneEntry{
		RuneId:            runeId,
		Number:            uint64(src.Number),
		Divisibility:      uint8(src.Divisibility),
		Premine:           lo.FromPtr(premine),
		SpacedRune:        runes.NewSpacedRune(rune, uint32(src.Spacers)),
		Symbol:            src.Symbol,
		Terms:             terms,
		Turbo:             src.Turbo,
		Mints:             lo.FromPtr(mints),
		BurnedAmount:      lo.FromPtr(burnedAmount),
		CompletedAt:       completedAt,
		CompletedAtHeight: completedAtHeight,
		EtchingBlock:      uint64(src.EtchingBlock),
		EtchingTxHash:     *etchingTxHash,
		EtchedAt:          etchedAt,
	}, nil
}

func mapRuneEntryTypeToParams(src runes.RuneEntry) (gen.CreateRuneEntryParams, error) {
	runeId := src.RuneId.String()
	rune := src.SpacedRune.Rune.String()
	spacers := int32(src.SpacedRune.Spacers)
	premine, err := numericFromUint128(&src.Premine)
	if err != nil {
		return gen.CreateRuneEntryParams{}, errors.Wrap(err, "failed to parse premine")
	}
	var terms bool
	var termsAmount, termsCap pgtype.Numeric
	var termsHeightStart, termsHeightEnd, termsOffsetStart, termsOffsetEnd pgtype.Int4
	if src.Terms != nil {
		terms = true
		if src.Terms.Amount != nil {
			termsAmount, err = numericFromUint128(src.Terms.Amount)
			if err != nil {
				return gen.CreateRuneEntryParams{}, errors.Wrap(err, "failed to parse terms amount")
			}
		}
		if src.Terms.Cap != nil {
			termsCap, err = numericFromUint128(src.Terms.Cap)
			if err != nil {
				return gen.CreateRuneEntryParams{}, errors.Wrap(err, "failed to parse terms cap")
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
	etchedAt := pgtype.Timestamp{Time: src.EtchedAt.UTC(), Valid: true}

	return gen.CreateRuneEntryParams{
		RuneID:           runeId,
		Rune:             rune,
		Number:           int64(src.Number),
		Spacers:          spacers,
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
		Turbo:            src.Turbo,
		EtchingBlock:     int32(src.EtchingBlock),
		EtchingTxHash:    src.EtchingTxHash.String(),
		EtchedAt:         etchedAt,
	}, nil
}

func mapRuneEntryStatesTypeToParams(src runes.RuneEntry, blockHeight uint64) (gen.CreateRuneEntryStateParams, error) {
	runeId := src.RuneId.String()
	mints, err := numericFromUint128(&src.Mints)
	if err != nil {
		return gen.CreateRuneEntryStateParams{}, errors.Wrap(err, "failed to parse mints")
	}
	burnedAmount, err := numericFromUint128(&src.BurnedAmount)
	if err != nil {
		return gen.CreateRuneEntryStateParams{}, errors.Wrap(err, "failed to parse burned amount")
	}
	var completedAt pgtype.Timestamp
	if !src.CompletedAt.IsZero() {
		completedAt.Time = src.CompletedAt.UTC()
		completedAt.Valid = true
	}
	var completedAtHeight pgtype.Int4
	if src.CompletedAtHeight != nil {
		completedAtHeight.Int32 = int32(*src.CompletedAtHeight)
		completedAtHeight.Valid = true
	}

	return gen.CreateRuneEntryStateParams{
		BlockHeight:       int32(blockHeight),
		RuneID:            runeId,
		Mints:             mints,
		BurnedAmount:      burnedAmount,
		CompletedAt:       completedAt,
		CompletedAtHeight: completedAtHeight,
	}, nil
}

func mapRuneEntryTypeToParamsBatch(srcs []*runes.RuneEntry) (gen.BatchCreateRuneEntriesPatchedParams, error) {
	var batchParams gen.BatchCreateRuneEntriesPatchedParams
	batchParams.RuneIDArr = make([]string, 0, len(srcs))
	batchParams.RuneArr = make([]string, 0, len(srcs))
	batchParams.NumberArr = make([]int64, 0, len(srcs))
	batchParams.SpacersArr = make([]int32, 0, len(srcs))
	batchParams.PremineArr = make([]pgtype.Numeric, 0, len(srcs))
	batchParams.SymbolArr = make([]int32, 0, len(srcs))
	batchParams.DivisibilityArr = make([]int16, 0, len(srcs))
	batchParams.TermsArr = make([]bool, 0, len(srcs))
	batchParams.TermsAmountArr = make([]pgtype.Numeric, 0, len(srcs))
	batchParams.TermsCapArr = make([]pgtype.Numeric, 0, len(srcs))
	batchParams.TermsHeightStartArr = make([]pgtype.Int4, 0, len(srcs))
	batchParams.TermsHeightEndArr = make([]pgtype.Int4, 0, len(srcs))
	batchParams.TermsOffsetStartArr = make([]pgtype.Int4, 0, len(srcs))
	batchParams.TermsOffsetEndArr = make([]pgtype.Int4, 0, len(srcs))
	batchParams.TurboArr = make([]bool, 0, len(srcs))
	batchParams.EtchingBlockArr = make([]int32, 0, len(srcs))
	batchParams.EtchingTxHashArr = make([]string, 0, len(srcs))
	batchParams.EtchedAtArr = make([]pgtype.Timestamp, 0, len(srcs))

	for i, src := range srcs {
		param, err := mapRuneEntryTypeToParams(*src)
		if err != nil {
			return gen.BatchCreateRuneEntriesPatchedParams{}, errors.Wrapf(err, "failed to map rune entry to params batch at index %d", i)
		}

		batchParams.RuneIDArr = append(batchParams.RuneIDArr, param.RuneID)
		batchParams.RuneArr = append(batchParams.RuneArr, param.Rune)
		batchParams.NumberArr = append(batchParams.NumberArr, param.Number)
		batchParams.SpacersArr = append(batchParams.SpacersArr, param.Spacers)
		batchParams.PremineArr = append(batchParams.PremineArr, param.Premine)
		batchParams.SymbolArr = append(batchParams.SymbolArr, param.Symbol)
		batchParams.DivisibilityArr = append(batchParams.DivisibilityArr, param.Divisibility)
		batchParams.TermsArr = append(batchParams.TermsArr, param.Terms)
		batchParams.TermsAmountArr = append(batchParams.TermsAmountArr, param.TermsAmount)
		batchParams.TermsCapArr = append(batchParams.TermsCapArr, param.TermsCap)
		batchParams.TermsHeightStartArr = append(batchParams.TermsHeightStartArr, param.TermsHeightStart)
		batchParams.TermsHeightEndArr = append(batchParams.TermsHeightEndArr, param.TermsHeightEnd)
		batchParams.TermsOffsetStartArr = append(batchParams.TermsOffsetStartArr, param.TermsOffsetStart)
		batchParams.TermsOffsetEndArr = append(batchParams.TermsOffsetEndArr, param.TermsOffsetEnd)
		batchParams.TurboArr = append(batchParams.TurboArr, param.Turbo)
		batchParams.EtchingBlockArr = append(batchParams.EtchingBlockArr, param.EtchingBlock)
		batchParams.EtchingTxHashArr = append(batchParams.EtchingTxHashArr, param.EtchingTxHash)
		batchParams.EtchedAtArr = append(batchParams.EtchedAtArr, param.EtchedAt)
	}

	return batchParams, nil
}

func mapRuneEntryStatesTypeToParamsBatch(srcs []*runes.RuneEntry, blockHeight uint64) (gen.BatchCreateRuneEntryStatesPatchedParams, error) {
	var batchParams gen.BatchCreateRuneEntryStatesPatchedParams
	batchParams.RuneIDArr = make([]string, 0, len(srcs))
	batchParams.BlockHeightArr = make([]int32, 0, len(srcs))
	batchParams.MintsArr = make([]pgtype.Numeric, 0, len(srcs))
	batchParams.BurnedAmountArr = make([]pgtype.Numeric, 0, len(srcs))
	batchParams.CompletedAtArr = make([]pgtype.Timestamp, 0, len(srcs))
	batchParams.CompletedAtHeightArr = make([]pgtype.Int4, 0, len(srcs))

	for i, src := range srcs {
		param, err := mapRuneEntryStatesTypeToParams(*src, blockHeight)
		if err != nil {
			return gen.BatchCreateRuneEntryStatesPatchedParams{}, errors.Wrapf(err, "failed to map rune entry states to params batch at index %d", i)
		}

		batchParams.RuneIDArr = append(batchParams.RuneIDArr, param.RuneID)
		batchParams.BlockHeightArr = append(batchParams.BlockHeightArr, param.BlockHeight)
		batchParams.MintsArr = append(batchParams.MintsArr, param.Mints)
		batchParams.BurnedAmountArr = append(batchParams.BurnedAmountArr, param.BurnedAmount)
		batchParams.CompletedAtArr = append(batchParams.CompletedAtArr, param.CompletedAt)
		batchParams.CompletedAtHeightArr = append(batchParams.CompletedAtHeightArr, param.CompletedAtHeight)
	}

	return batchParams, nil
}

func mapRuneTransactionTypeToParams(src entity.RuneTransaction) (gen.CreateRuneTransactionParams, error) {
	var timestamp pgtype.Timestamp
	if !src.Timestamp.IsZero() {
		timestamp.Time = src.Timestamp.UTC()
		timestamp.Valid = true
	}
	inputsBytes, err := json.Marshal(src.Inputs)
	if err != nil {
		return gen.CreateRuneTransactionParams{}, errors.Wrap(err, "failed to marshal inputs")
	}
	outputsBytes, err := json.Marshal(src.Outputs)
	if err != nil {
		return gen.CreateRuneTransactionParams{}, errors.Wrap(err, "failed to marshal outputs")
	}
	mints := make(map[string]uint128.Uint128)
	for key, value := range src.Mints {
		mints[key.String()] = value
	}
	mintsBytes, err := json.Marshal(mints)
	if err != nil {
		return gen.CreateRuneTransactionParams{}, errors.Wrap(err, "failed to marshal mints")
	}
	burns := make(map[string]uint128.Uint128)
	for key, value := range src.Burns {
		burns[key.String()] = value
	}
	burnsBytes, err := json.Marshal(burns)
	if err != nil {
		return gen.CreateRuneTransactionParams{}, errors.Wrap(err, "failed to marshal burns")
	}

	return gen.CreateRuneTransactionParams{
		Hash:        src.Hash.String(),
		BlockHeight: int32(src.BlockHeight),
		Index:       int32(src.Index),
		Timestamp:   timestamp,
		Inputs:      inputsBytes,
		Outputs:     outputsBytes,
		Mints:       mintsBytes,
		Burns:       burnsBytes,
		RuneEtched:  src.RuneEtched,
	}, nil
}

func mapRuneTransactionTypeToParamsBatch(srcs []*entity.RuneTransaction) (gen.BatchCreateRuneTransactionsParams, error) {
	batchParams := gen.BatchCreateRuneTransactionsParams{
		HashArr:        make([]string, 0, len(srcs)),
		BlockHeightArr: make([]int32, 0, len(srcs)),
		IndexArr:       make([]int32, 0, len(srcs)),
		TimestampArr:   make([]pgtype.Timestamp, 0, len(srcs)),
		RuneEtchedArr:  make([]bool, 0, len(srcs)),
	}
	inputsArr := make([][]byte, 0, len(srcs))
	outputsArr := make([][]byte, 0, len(srcs))
	mintsArr := make([][]byte, 0, len(srcs))
	burnsArr := make([][]byte, 0, len(srcs))

	for i, src := range srcs {
		param, err := mapRuneTransactionTypeToParams(*src)
		if err != nil {
			return gen.BatchCreateRuneTransactionsParams{}, errors.Wrapf(err, "failed to map rune transaction to params batch at index %d", i)
		}

		batchParams.HashArr = append(batchParams.HashArr, param.Hash)
		batchParams.BlockHeightArr = append(batchParams.BlockHeightArr, param.BlockHeight)
		batchParams.IndexArr = append(batchParams.IndexArr, param.Index)
		batchParams.TimestampArr = append(batchParams.TimestampArr, param.Timestamp)
		batchParams.RuneEtchedArr = append(batchParams.RuneEtchedArr, param.RuneEtched)

		inputsArr = append(inputsArr, param.Inputs)
		outputsArr = append(outputsArr, param.Outputs)
		mintsArr = append(mintsArr, param.Mints)
		burnsArr = append(burnsArr, param.Burns)
	}

	batchParams.InputsArr = inputsArr
	batchParams.OutputsArr = outputsArr
	batchParams.MintsArr = mintsArr
	batchParams.BurnsArr = burnsArr

	return batchParams, nil
}

func extractModelRuneTxAndRunestone(src gen.GetRuneTransactionsRow) (gen.RunesTransaction, *gen.RunesRunestone, error) {
	var runestone *gen.RunesRunestone
	if src.TxHash.Valid {
		// these fields should never be null
		if !src.Cenotaph.Valid {
			return gen.RunesTransaction{}, nil, errors.New("runestone cenotaph is null")
		}
		if !src.Flaws.Valid {
			return gen.RunesTransaction{}, nil, errors.New("runestone flaws is null")
		}
		runestone = &gen.RunesRunestone{
			TxHash:                  src.TxHash.String,
			BlockHeight:             src.BlockHeight,
			Etching:                 src.Etching.Bool,
			EtchingDivisibility:     src.EtchingDivisibility,
			EtchingPremine:          src.EtchingPremine,
			EtchingRune:             src.EtchingRune,
			EtchingSpacers:          src.EtchingSpacers,
			EtchingSymbol:           src.EtchingSymbol,
			EtchingTerms:            src.EtchingTerms,
			EtchingTermsAmount:      src.EtchingTermsAmount,
			EtchingTermsCap:         src.EtchingTermsCap,
			EtchingTermsHeightStart: src.EtchingTermsHeightStart,
			EtchingTermsHeightEnd:   src.EtchingTermsHeightEnd,
			EtchingTermsOffsetStart: src.EtchingTermsOffsetStart,
			EtchingTermsOffsetEnd:   src.EtchingTermsOffsetEnd,
			Edicts:                  src.Edicts,
			Mint:                    src.Mint,
			Pointer:                 src.Pointer,
			Cenotaph:                src.Cenotaph.Bool,
			Flaws:                   src.Flaws.Int32,
		}
	}
	return gen.RunesTransaction{
		Hash:        src.Hash,
		BlockHeight: src.BlockHeight,
		Index:       src.Index,
		Timestamp:   src.Timestamp,
		Inputs:      src.Inputs,
		Outputs:     src.Outputs,
		Mints:       src.Mints,
		Burns:       src.Burns,
		RuneEtched:  src.RuneEtched,
	}, runestone, nil
}

func mapRuneTransactionModelToType(src gen.RunesTransaction) (entity.RuneTransaction, error) {
	hash, err := chainhash.NewHashFromStr(src.Hash)
	if err != nil {
		return entity.RuneTransaction{}, errors.Wrap(err, "failed to parse transaction hash")
	}
	var timestamp time.Time
	if src.Timestamp.Valid {
		timestamp = src.Timestamp.Time
	}

	inputs := make([]*entity.TxInputOutput, 0)
	if err := json.Unmarshal(src.Inputs, &inputs); err != nil {
		return entity.RuneTransaction{}, errors.Wrap(err, "failed to unmarshal inputs")
	}
	outputs := make([]*entity.TxInputOutput, 0)
	if err := json.Unmarshal(src.Outputs, &outputs); err != nil {
		return entity.RuneTransaction{}, errors.Wrap(err, "failed to unmarshal outputs")
	}
	mintsRaw := make(map[string]uint128.Uint128)
	if err := json.Unmarshal(src.Mints, &mintsRaw); err != nil {
		return entity.RuneTransaction{}, errors.Wrap(err, "failed to unmarshal mints")
	}
	mints := make(map[runes.RuneId]uint128.Uint128)
	for key, value := range mintsRaw {
		runeId, err := runes.NewRuneIdFromString(key)
		if err != nil {
			return entity.RuneTransaction{}, errors.Wrap(err, "failed to parse rune id")
		}
		mints[runeId] = value
	}

	burnsRaw := make(map[string]uint128.Uint128)
	if err := json.Unmarshal(src.Burns, &burnsRaw); err != nil {
		return entity.RuneTransaction{}, errors.Wrap(err, "failed to unmarshal burns")
	}
	burns := make(map[runes.RuneId]uint128.Uint128)
	for key, value := range burnsRaw {
		runeId, err := runes.NewRuneIdFromString(key)
		if err != nil {
			return entity.RuneTransaction{}, errors.Wrap(err, "failed to parse rune id")
		}
		burns[runeId] = value
	}

	return entity.RuneTransaction{
		Hash:        *hash,
		BlockHeight: uint64(src.BlockHeight),
		Index:       uint32(src.Index),
		Timestamp:   timestamp,
		Inputs:      inputs,
		Outputs:     outputs,
		Mints:       mints,
		Burns:       burns,
		RuneEtched:  src.RuneEtched,
	}, nil
}

func mapRunestoneTypeToParams(src runes.Runestone, txHash chainhash.Hash, blockHeight uint64) (gen.CreateRunestoneParams, error) {
	var runestoneParams gen.CreateRunestoneParams

	// TODO: optimize serialized edicts
	edictsBytes, err := json.Marshal(src.Edicts)
	if err != nil {
		return gen.CreateRunestoneParams{}, errors.Wrap(err, "failed to marshal runestone edicts")
	}
	runestoneParams = gen.CreateRunestoneParams{
		TxHash:      txHash.String(),
		BlockHeight: int32(blockHeight),
		Edicts:      edictsBytes,
		Cenotaph:    src.Cenotaph,
		Flaws:       int32(src.Flaws),
	}
	if src.Etching != nil {
		runestoneParams.Etching = true
		etching := *src.Etching
		if etching.Divisibility != nil {
			runestoneParams.EtchingDivisibility = pgtype.Int2{Int16: int16(*etching.Divisibility), Valid: true}
		}
		if etching.Premine != nil {
			premine, err := numericFromUint128(etching.Premine)
			if err != nil {
				return gen.CreateRunestoneParams{}, errors.Wrap(err, "failed to parse etching premine")
			}
			runestoneParams.EtchingPremine = premine
		}
		if etching.Rune != nil {
			runestoneParams.EtchingRune = pgtype.Text{String: etching.Rune.String(), Valid: true}
		}
		if etching.Spacers != nil {
			runestoneParams.EtchingSpacers = pgtype.Int4{Int32: int32(*etching.Spacers), Valid: true}
		}
		if etching.Symbol != nil {
			runestoneParams.EtchingSymbol = pgtype.Int4{Int32: *etching.Symbol, Valid: true}
		}
		if etching.Terms != nil {
			runestoneParams.EtchingTerms = pgtype.Bool{Bool: true, Valid: true}
			terms := *etching.Terms
			if terms.Amount != nil {
				amount, err := numericFromUint128(terms.Amount)
				if err != nil {
					return gen.CreateRunestoneParams{}, errors.Wrap(err, "failed to parse etching terms amount")
				}
				runestoneParams.EtchingTermsAmount = amount
			}
			if terms.Cap != nil {
				cap, err := numericFromUint128(terms.Cap)
				if err != nil {
					return gen.CreateRunestoneParams{}, errors.Wrap(err, "failed to parse etching terms cap")
				}
				runestoneParams.EtchingTermsCap = cap
			}
			if terms.HeightStart != nil {
				runestoneParams.EtchingTermsHeightStart = pgtype.Int4{Int32: int32(*terms.HeightStart), Valid: true}
			}
			if terms.HeightEnd != nil {
				runestoneParams.EtchingTermsHeightEnd = pgtype.Int4{Int32: int32(*terms.HeightEnd), Valid: true}
			}
			if terms.OffsetStart != nil {
				runestoneParams.EtchingTermsOffsetStart = pgtype.Int4{Int32: int32(*terms.OffsetStart), Valid: true}
			}
			if terms.OffsetEnd != nil {
				runestoneParams.EtchingTermsOffsetEnd = pgtype.Int4{Int32: int32(*terms.OffsetEnd), Valid: true}
			}
		}
		runestoneParams.EtchingTurbo = pgtype.Bool{Bool: etching.Turbo, Valid: true}
	}
	if src.Mint != nil {
		runestoneParams.Mint = pgtype.Text{String: src.Mint.String(), Valid: true}
	}
	if src.Pointer != nil {
		runestoneParams.Pointer = pgtype.Int4{Int32: int32(*src.Pointer), Valid: true}
	}

	return runestoneParams, nil
}

func mapRunestoneTypeToParamsBatch(srcs []*entity.RuneTransaction) (gen.BatchCreateRunestonesPatchedParams, error) {
	var batchParams gen.BatchCreateRunestonesPatchedParams
	batchParams.TxHashArr = make([]string, 0, len(srcs))
	batchParams.BlockHeightArr = make([]int32, 0, len(srcs))
	batchParams.EtchingArr = make([]bool, 0, len(srcs))
	batchParams.EtchingDivisibilityArr = make([]pgtype.Int2, 0, len(srcs))
	batchParams.EtchingPremineArr = make([]pgtype.Numeric, 0, len(srcs))
	batchParams.EtchingRuneArr = make([]pgtype.Text, 0, len(srcs))
	batchParams.EtchingSpacersArr = make([]pgtype.Int4, 0, len(srcs))
	batchParams.EtchingSymbolArr = make([]pgtype.Int4, 0, len(srcs))
	batchParams.EtchingTermsArr = make([]pgtype.Bool, 0, len(srcs))
	batchParams.EtchingTermsAmountArr = make([]pgtype.Numeric, 0, len(srcs))
	batchParams.EtchingTermsCapArr = make([]pgtype.Numeric, 0, len(srcs))
	batchParams.EtchingTermsHeightStartArr = make([]pgtype.Int4, 0, len(srcs))
	batchParams.EtchingTermsHeightEndArr = make([]pgtype.Int4, 0, len(srcs))
	batchParams.EtchingTermsOffsetStartArr = make([]pgtype.Int4, 0, len(srcs))
	batchParams.EtchingTermsOffsetEndArr = make([]pgtype.Int4, 0, len(srcs))
	batchParams.EtchingTurboArr = make([]pgtype.Bool, 0, len(srcs))
	batchParams.EdictsArr = make([][]byte, 0, len(srcs))
	batchParams.MintArr = make([]pgtype.Text, 0, len(srcs))
	batchParams.PointerArr = make([]pgtype.Int4, 0, len(srcs))
	batchParams.CenotaphArr = make([]bool, 0, len(srcs))
	batchParams.FlawsArr = make([]int32, 0, len(srcs))

	for i, src := range srcs {
		if src.Runestone == nil {
			continue
		}
		param, err := mapRunestoneTypeToParams(*src.Runestone, src.Hash, src.BlockHeight)
		if err != nil {
			return gen.BatchCreateRunestonesPatchedParams{}, errors.Wrapf(err, "failed to map runestone to params batch at index %d", i)
		}

		batchParams.TxHashArr = append(batchParams.TxHashArr, param.TxHash)
		batchParams.BlockHeightArr = append(batchParams.BlockHeightArr, param.BlockHeight)
		batchParams.EtchingArr = append(batchParams.EtchingArr, param.Etching)
		batchParams.EtchingDivisibilityArr = append(batchParams.EtchingDivisibilityArr, param.EtchingDivisibility)
		batchParams.EtchingPremineArr = append(batchParams.EtchingPremineArr, param.EtchingPremine)
		batchParams.EtchingRuneArr = append(batchParams.EtchingRuneArr, param.EtchingRune)
		batchParams.EtchingSpacersArr = append(batchParams.EtchingSpacersArr, param.EtchingSpacers)
		batchParams.EtchingSymbolArr = append(batchParams.EtchingSymbolArr, param.EtchingSymbol)
		batchParams.EtchingTermsArr = append(batchParams.EtchingTermsArr, param.EtchingTerms)
		batchParams.EtchingTermsAmountArr = append(batchParams.EtchingTermsAmountArr, param.EtchingTermsAmount)
		batchParams.EtchingTermsCapArr = append(batchParams.EtchingTermsCapArr, param.EtchingTermsCap)
		batchParams.EtchingTermsHeightStartArr = append(batchParams.EtchingTermsHeightStartArr, param.EtchingTermsHeightStart)
		batchParams.EtchingTermsHeightEndArr = append(batchParams.EtchingTermsHeightEndArr, param.EtchingTermsHeightEnd)
		batchParams.EtchingTermsOffsetStartArr = append(batchParams.EtchingTermsOffsetStartArr, param.EtchingTermsOffsetStart)
		batchParams.EtchingTermsOffsetEndArr = append(batchParams.EtchingTermsOffsetEndArr, param.EtchingTermsOffsetEnd)
		batchParams.EtchingTurboArr = append(batchParams.EtchingTurboArr, param.EtchingTurbo)
		batchParams.EdictsArr = append(batchParams.EdictsArr, param.Edicts)
		batchParams.MintArr = append(batchParams.MintArr, param.Mint)
		batchParams.PointerArr = append(batchParams.PointerArr, param.Pointer)
		batchParams.CenotaphArr = append(batchParams.CenotaphArr, param.Cenotaph)
		batchParams.FlawsArr = append(batchParams.FlawsArr, param.Flaws)
	}

	return batchParams, nil
}

func mapRunestoneModelToType(src gen.RunesRunestone) (runes.Runestone, error) {
	runestone := runes.Runestone{
		Cenotaph: src.Cenotaph,
		Flaws:    runes.Flaws(src.Flaws),
	}
	if src.Etching {
		etching := runes.Etching{}
		if src.EtchingDivisibility.Valid {
			divisibility := uint8(src.EtchingDivisibility.Int16)
			etching.Divisibility = &divisibility
		}
		if src.EtchingPremine.Valid {
			premine, err := uint128FromNumeric(src.EtchingPremine)
			if err != nil {
				return runes.Runestone{}, errors.Wrap(err, "failed to parse etching premine")
			}
			etching.Premine = premine
		}
		if src.EtchingRune.Valid {
			rune, err := runes.NewRuneFromString(src.EtchingRune.String)
			if err != nil {
				return runes.Runestone{}, errors.Wrap(err, "failed to parse etching rune")
			}
			etching.Rune = &rune
		}
		if src.EtchingSpacers.Valid {
			spacers := uint32(src.EtchingSpacers.Int32)
			etching.Spacers = &spacers
		}
		if src.EtchingSymbol.Valid {
			var symbol rune = src.EtchingSymbol.Int32
			etching.Symbol = &symbol
		}
		if src.EtchingTerms.Valid && src.EtchingTerms.Bool {
			terms := runes.Terms{}
			if src.EtchingTermsAmount.Valid {
				amount, err := uint128FromNumeric(src.EtchingTermsAmount)
				if err != nil {
					return runes.Runestone{}, errors.Wrap(err, "failed to parse etching terms amount")
				}
				terms.Amount = amount
			}
			if src.EtchingTermsCap.Valid {
				cap, err := uint128FromNumeric(src.EtchingTermsCap)
				if err != nil {
					return runes.Runestone{}, errors.Wrap(err, "failed to parse etching terms cap")
				}
				terms.Cap = cap
			}
			if src.EtchingTermsHeightStart.Valid {
				heightStart := uint64(src.EtchingTermsHeightStart.Int32)
				terms.HeightStart = &heightStart
			}
			if src.EtchingTermsHeightEnd.Valid {
				heightEnd := uint64(src.EtchingTermsHeightEnd.Int32)
				terms.HeightEnd = &heightEnd
			}
			if src.EtchingTermsOffsetStart.Valid {
				offsetStart := uint64(src.EtchingTermsOffsetStart.Int32)
				terms.OffsetStart = &offsetStart
			}
			if src.EtchingTermsOffsetEnd.Valid {
				offsetEnd := uint64(src.EtchingTermsOffsetEnd.Int32)
				terms.OffsetEnd = &offsetEnd
			}
			etching.Terms = &terms
		}
		etching.Turbo = src.EtchingTurbo.Valid && src.EtchingTurbo.Bool
		runestone.Etching = &etching
	}
	if src.Mint.Valid {
		mint, err := runes.NewRuneIdFromString(src.Mint.String)
		if err != nil {
			return runes.Runestone{}, errors.Wrap(err, "failed to parse mint")
		}
		runestone.Mint = &mint
	}
	if src.Pointer.Valid {
		pointer := uint64(src.Pointer.Int32)
		runestone.Pointer = &pointer
	}
	// Edicts
	{
		if err := json.Unmarshal(src.Edicts, &runestone.Edicts); err != nil {
			return runes.Runestone{}, errors.Wrap(err, "failed to unmarshal edicts")
		}
		if len(runestone.Edicts) == 0 {
			runestone.Edicts = nil
		}
	}
	return runestone, nil
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
		Amount:      lo.FromPtr(amount),
		BlockHeight: uint64(src.BlockHeight),
	}, nil
}

func mapBalanceTypeToParams(src entity.Balance) (gen.CreateRuneBalanceParams, error) {
	amount, err := numericFromUint128(&src.Amount)
	if err != nil {
		return gen.CreateRuneBalanceParams{}, errors.Wrap(err, "failed to parse amount")
	}
	return gen.CreateRuneBalanceParams{
		RuneID:      src.RuneId.String(),
		Amount:      amount,
		Pkscript:    hex.EncodeToString(src.PkScript),
		BlockHeight: int32(src.BlockHeight),
	}, nil
}

func mapBalanceTypeToParamsBatch(srcs []*entity.Balance) (gen.BatchCreateRunesBalancesParams, error) {
	batchParams := gen.BatchCreateRunesBalancesParams{
		RuneIDArr:      make([]string, 0, len(srcs)),
		AmountArr:      make([]pgtype.Numeric, 0, len(srcs)),
		PkscriptArr:    make([]string, 0, len(srcs)),
		BlockHeightArr: make([]int32, 0, len(srcs)),
	}

	for i, src := range srcs {
		param, err := mapBalanceTypeToParams(*src)
		if err != nil {
			return gen.BatchCreateRunesBalancesParams{}, errors.Wrapf(err, "failed to map balance to params batch at index %d", i)
		}

		batchParams.RuneIDArr = append(batchParams.RuneIDArr, param.RuneID)
		batchParams.AmountArr = append(batchParams.AmountArr, param.Amount)
		batchParams.PkscriptArr = append(batchParams.PkscriptArr, param.Pkscript)
		batchParams.BlockHeightArr = append(batchParams.BlockHeightArr, param.BlockHeight)
	}

	return batchParams, nil
}

func mapIndexedBlockModelToType(src gen.RunesIndexedBlock) (*entity.IndexedBlock, error) {
	hash, err := chainhash.NewHashFromStr(src.Hash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse block hash")
	}
	prevBlockHash, err := chainhash.NewHashFromStr(src.PrevHash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse prev block hash")
	}
	eventHash, err := chainhash.NewHashFromStr(src.EventHash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse event hash")
	}
	cumulativeEventHash, err := chainhash.NewHashFromStr(src.CumulativeEventHash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse cumulative event hash")
	}
	return &entity.IndexedBlock{
		Height:              int64(src.Height),
		Hash:                *hash,
		PrevHash:            *prevBlockHash,
		EventHash:           *eventHash,
		CumulativeEventHash: *cumulativeEventHash,
	}, nil
}

func mapIndexedBlockTypeToParams(src entity.IndexedBlock) (gen.CreateIndexedBlockParams, error) {
	return gen.CreateIndexedBlockParams{
		Height:              int32(src.Height),
		Hash:                src.Hash.String(),
		PrevHash:            src.PrevHash.String(),
		EventHash:           src.EventHash.String(),
		CumulativeEventHash: src.CumulativeEventHash.String(),
	}, nil
}

func mapRunesUTXOModelToType(src gen.GetRunesUTXOsByPkScriptRow) (entity.RunesUTXO, error) {
	pkScriptRaw, ok := src.Pkscript.(string)
	if !ok {
		return entity.RunesUTXO{}, errors.New("pkscript from database is not string")
	}
	pkScript, err := hex.DecodeString(pkScriptRaw)
	if err != nil {
		return entity.RunesUTXO{}, errors.Wrap(err, "failed to parse pkscript")
	}
	txHash, err := chainhash.NewHashFromStr(src.TxHash)
	if err != nil {
		return entity.RunesUTXO{}, errors.Wrap(err, "failed to parse tx hash")
	}
	runeIdsRaw, ok := src.RuneIds.([]interface{})
	if !ok {
		return entity.RunesUTXO{}, errors.New("src.RuneIds is not a slice")
	}
	runeIds := make([]string, 0, len(runeIdsRaw))
	for i, raw := range runeIdsRaw {
		runeId, ok := raw.(string)
		if !ok {
			return entity.RunesUTXO{}, errors.Errorf("src.RuneIds[%d] is not a string", i)
		}
		runeIds = append(runeIds, runeId)
	}
	amountsRaw, ok := src.Amounts.([]interface{})
	if !ok {
		return entity.RunesUTXO{}, errors.New("amounts from database is not a slice")
	}
	amounts := make([]pgtype.Numeric, 0, len(amountsRaw))
	for i, raw := range amountsRaw {
		amount, ok := raw.(pgtype.Numeric)
		if !ok {
			return entity.RunesUTXO{}, errors.Errorf("src.Amounts[%d] is not pgtype.Numeric", i)
		}
		amounts = append(amounts, amount)
	}
	if len(runeIds) != len(amounts) {
		return entity.RunesUTXO{}, errors.New("rune ids and amounts have different lengths")
	}

	runesBalances := make([]entity.RunesUTXOBalance, 0, len(runeIds))
	for i := range runeIds {
		runeId, err := runes.NewRuneIdFromString(runeIds[i])
		if err != nil {
			return entity.RunesUTXO{}, errors.Wrap(err, "failed to parse rune id")
		}
		amount, err := uint128FromNumeric(amounts[i])
		if err != nil {
			return entity.RunesUTXO{}, errors.Wrap(err, "failed to parse amount")
		}
		runesBalances = append(runesBalances, entity.RunesUTXOBalance{
			RuneId: runeId,
			Amount: lo.FromPtr(amount),
		})
	}
	return entity.RunesUTXO{
		PkScript: pkScript,
		OutPoint: wire.OutPoint{
			Hash:  *txHash,
			Index: uint32(src.TxIdx),
		},
		RuneBalances: runesBalances,
	}, nil
}

func mapOutPointBalanceModelToType(src gen.RunesOutpointBalance) (entity.OutPointBalance, error) {
	runeId, err := runes.NewRuneIdFromString(src.RuneID)
	if err != nil {
		return entity.OutPointBalance{}, errors.Wrap(err, "failed to parse rune id")
	}
	amount, err := uint128FromNumeric(src.Amount)
	if err != nil {
		return entity.OutPointBalance{}, errors.Wrap(err, "failed to parse balance")
	}
	pkScript, err := hex.DecodeString(src.Pkscript)
	if err != nil {
		return entity.OutPointBalance{}, errors.Wrap(err, "failed to parse pkscript")
	}
	txHash, err := chainhash.NewHashFromStr(src.TxHash)
	if err != nil {
		return entity.OutPointBalance{}, errors.Wrap(err, "failed to parse tx hash")
	}
	var spentHeight *uint64
	if src.SpentHeight.Valid {
		spentHeight = lo.ToPtr(uint64(src.SpentHeight.Int32))
	}
	return entity.OutPointBalance{
		PkScript: pkScript,
		RuneId:   runeId,
		Amount:   lo.FromPtr(amount),
		OutPoint: wire.OutPoint{
			Hash:  *txHash,
			Index: uint32(src.TxIdx),
		},
		BlockHeight: uint64(src.BlockHeight),
		SpentHeight: spentHeight,
	}, nil
}

func mapOutPointBalanceTypeToParams(src entity.OutPointBalance) (gen.CreateOutPointBalanceParams, error) {
	amount, err := numericFromUint128(&src.Amount)
	if err != nil {
		return gen.CreateOutPointBalanceParams{}, errors.Wrap(err, "failed to parse amount")
	}
	var spentHeight pgtype.Int4
	if src.SpentHeight != nil {
		spentHeight = pgtype.Int4{Int32: int32(*src.SpentHeight), Valid: true}
	}
	return gen.CreateOutPointBalanceParams{
		TxHash:      src.OutPoint.Hash.String(),
		TxIdx:       int32(src.OutPoint.Index),
		Pkscript:    hex.EncodeToString(src.PkScript),
		RuneID:      src.RuneId.String(),
		Amount:      amount,
		BlockHeight: int32(src.BlockHeight),
		SpentHeight: spentHeight,
	}, nil
}

func mapOutPointBalanceTypeToParamsBatch(srcs []*entity.OutPointBalance) (gen.BatchCreateRunesOutpointBalancesPatchedParams, error) {
	var batchParams gen.BatchCreateRunesOutpointBalancesPatchedParams
	batchParams.TxHashArr = make([]string, 0, len(srcs))
	batchParams.TxIdxArr = make([]int32, 0, len(srcs))
	batchParams.PkscriptArr = make([]string, 0, len(srcs))
	batchParams.RuneIDArr = make([]string, 0, len(srcs))
	batchParams.AmountArr = make([]pgtype.Numeric, 0, len(srcs))
	batchParams.BlockHeightArr = make([]int32, 0, len(srcs))
	batchParams.SpentHeightArr = make([]pgtype.Int4, 0, len(srcs))

	for i, src := range srcs {
		param, err := mapOutPointBalanceTypeToParams(*src)
		if err != nil {
			return gen.BatchCreateRunesOutpointBalancesPatchedParams{}, errors.Wrapf(err, "failed to map outpoint balance to params batch at index %d", i)
		}

		batchParams.TxHashArr = append(batchParams.TxHashArr, param.TxHash)
		batchParams.TxIdxArr = append(batchParams.TxIdxArr, param.TxIdx)
		batchParams.PkscriptArr = append(batchParams.PkscriptArr, param.Pkscript)
		batchParams.RuneIDArr = append(batchParams.RuneIDArr, param.RuneID)
		batchParams.AmountArr = append(batchParams.AmountArr, param.Amount)
		batchParams.BlockHeightArr = append(batchParams.BlockHeightArr, param.BlockHeight)
		batchParams.SpentHeightArr = append(batchParams.SpentHeightArr, param.SpentHeight)
	}

	return batchParams, nil
}
