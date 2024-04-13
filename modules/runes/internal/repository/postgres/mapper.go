package postgres

import (
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/modules/runes/internal/entity"
	"github.com/gaze-network/indexer-network/modules/runes/internal/repository/postgres/gen"
	"github.com/gaze-network/indexer-network/modules/runes/internal/runes"
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
		createdAt = src.CreatedAt.Time
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

func mapRuneEntryModelToType(src gen.GetRuneEntriesByRuneIdsRow) (runes.RuneEntry, error) {
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
		completedAt = src.CompletedAt.Time
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
	return runes.RuneEntry{
		RuneId:            runeId,
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
	}, nil
}

func mapRuneEntryTypeToParams(src runes.RuneEntry, blockHeight uint64) (gen.CreateRuneEntryParams, gen.CreateRuneEntryStateParams, error) {
	runeId := src.RuneId.String()
	rune := src.SpacedRune.Rune.String()
	spacers := int32(src.SpacedRune.Spacers)
	mints, err := numericFromUint128(&src.Mints)
	if err != nil {
		return gen.CreateRuneEntryParams{}, gen.CreateRuneEntryStateParams{}, errors.Wrap(err, "failed to parse mints")
	}
	burnedAmount, err := numericFromUint128(&src.BurnedAmount)
	if err != nil {
		return gen.CreateRuneEntryParams{}, gen.CreateRuneEntryStateParams{}, errors.Wrap(err, "failed to parse burned amount")
	}
	premine, err := numericFromUint128(&src.Premine)
	if err != nil {
		return gen.CreateRuneEntryParams{}, gen.CreateRuneEntryStateParams{}, errors.Wrap(err, "failed to parse premine")
	}
	var completedAt pgtype.Timestamp
	if !src.CompletedAt.IsZero() {
		completedAt.Time = src.CompletedAt
		completedAt.Valid = true
	}
	var completedAtHeight pgtype.Int4
	if src.CompletedAtHeight != nil {
		completedAtHeight.Int32 = int32(*src.CompletedAtHeight)
		completedAtHeight.Valid = true
	}
	var terms bool
	var termsAmount, termsCap pgtype.Numeric
	var termsHeightStart, termsHeightEnd, termsOffsetStart, termsOffsetEnd pgtype.Int4
	if src.Terms != nil {
		terms = true
		if src.Terms.Amount != nil {
			termsAmount, err = numericFromUint128(src.Terms.Amount)
			if err != nil {
				return gen.CreateRuneEntryParams{}, gen.CreateRuneEntryStateParams{}, errors.Wrap(err, "failed to parse terms amount")
			}
		}
		if src.Terms.Cap != nil {
			termsCap, err = numericFromUint128(src.Terms.Cap)
			if err != nil {
				return gen.CreateRuneEntryParams{}, gen.CreateRuneEntryStateParams{}, errors.Wrap(err, "failed to parse terms cap")
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

	return gen.CreateRuneEntryParams{
			RuneID:           runeId,
			Rune:             rune,
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
			EtchingBlock:     int32(blockHeight),
		}, gen.CreateRuneEntryStateParams{
			BlockHeight:       int32(blockHeight),
			RuneID:            runeId,
			Mints:             mints,
			BurnedAmount:      burnedAmount,
			CompletedAt:       completedAt,
			CompletedAtHeight: completedAtHeight,
		}, nil
}

type outpointBalanceModel struct {
	PkScript       string
	Id             runes.RuneId
	Value          uint128.Uint128
	Index          uint32
	PrevTxHash     string
	PrevTxOutIndex uint32
}

func mapOutPointBalanceTypeToModel(src entity.OutPointBalance) outpointBalanceModel {
	return outpointBalanceModel{
		PkScript:       hex.EncodeToString(src.PkScript),
		Id:             src.Id,
		Value:          src.Value,
		Index:          src.Index,
		PrevTxHash:     src.PrevTxHash.String(),
		PrevTxOutIndex: src.PrevTxOutIndex,
	}
}

func mapOutPointBalanceModelToType(src outpointBalanceModel) (entity.OutPointBalance, error) {
	pkScript, err := hex.DecodeString(src.PkScript)
	if err != nil {
		return entity.OutPointBalance{}, errors.Wrap(err, "failed to decode pk script")
	}
	prevTxHash, err := chainhash.NewHashFromStr(src.PrevTxHash)
	if err != nil {
		return entity.OutPointBalance{}, errors.Wrap(err, "failed to parse prev tx hash")
	}
	return entity.OutPointBalance{
		PkScript:       pkScript,
		Id:             src.Id,
		Value:          src.Value,
		Index:          src.Index,
		PrevTxHash:     *prevTxHash,
		PrevTxOutIndex: src.PrevTxOutIndex,
	}, nil
}

// mapRuneTransactionModelToType returns params for creating a new rune transaction and (optionally) a runestone.
func mapRuneTransactionTypeToParams(src entity.RuneTransaction) (gen.CreateRuneTransactionParams, *gen.CreateRunestoneParams, error) {
	var timestamp pgtype.Timestamp
	if !src.Timestamp.IsZero() {
		timestamp.Time = src.Timestamp
		timestamp.Valid = true
	}
	inputs := lo.Map(src.Inputs, func(input *entity.OutPointBalance, _ int) outpointBalanceModel {
		return mapOutPointBalanceTypeToModel(*input)
	})
	outputs := lo.Map(src.Outputs, func(output *entity.OutPointBalance, _ int) outpointBalanceModel {
		return mapOutPointBalanceTypeToModel(*output)
	})
	inputsBytes, err := json.Marshal(inputs)
	if err != nil {
		return gen.CreateRuneTransactionParams{}, nil, errors.Wrap(err, "failed to marshal inputs")
	}
	outputsBytes, err := json.Marshal(outputs)
	if err != nil {
		return gen.CreateRuneTransactionParams{}, nil, errors.Wrap(err, "failed to marshal outputs")
	}
	mints := make(map[string]uint128.Uint128)
	for key, value := range src.Mints {
		mints[key.String()] = value
	}
	mintsBytes, err := json.Marshal(mints)
	if err != nil {
		return gen.CreateRuneTransactionParams{}, nil, errors.Wrap(err, "failed to marshal mints")
	}
	burns := make(map[string]uint128.Uint128)
	for key, value := range src.Burns {
		burns[key.String()] = value
	}
	burnsBytes, err := json.Marshal(burns)
	if err != nil {
		return gen.CreateRuneTransactionParams{}, nil, errors.Wrap(err, "failed to marshal burns")
	}

	var runestoneParams *gen.CreateRunestoneParams
	if src.Runestone != nil {
		params, err := mapRunestoneTypeToParams(*src.Runestone, src.Hash, src.BlockHeight)
		if err != nil {
			return gen.CreateRuneTransactionParams{}, nil, errors.Wrap(err, "failed to map runestone to params")
		}
		runestoneParams = &params
	}

	return gen.CreateRuneTransactionParams{
		Hash:        src.Hash.String(),
		BlockHeight: int32(src.BlockHeight),
		Timestamp:   timestamp,
		Inputs:      inputsBytes,
		Outputs:     outputsBytes,
		Mints:       mintsBytes,
		Burns:       burnsBytes,
	}, runestoneParams, nil
}

func extractModelRuneTxAndRunestone(src gen.GetRuneTransactionsByHeightRow) (gen.RunesTransaction, *gen.RunesRunestone, error) {
	var runestone *gen.RunesRunestone
	if src.TxHash.Valid {
		// these fields should never be null
		if !src.Etching.Valid {
			return gen.RunesTransaction{}, nil, errors.New("runestone etching bool is null")
		}
		if !src.EtchingTerms.Valid {
			return gen.RunesTransaction{}, nil, errors.New("runestone etching terms bool is null")
		}
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
			EtchingTerms:            src.EtchingTerms.Bool,
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
		Timestamp:   src.Timestamp,
		Inputs:      src.Inputs,
		Outputs:     src.Outputs,
		Mints:       src.Mints,
		Burns:       src.Burns,
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

	inputsRaw := make([]outpointBalanceModel, 0)
	if err := json.Unmarshal(src.Inputs, &inputsRaw); err != nil {
		return entity.RuneTransaction{}, errors.Wrap(err, "failed to unmarshal inputs")
	}
	inputs := make([]*entity.OutPointBalance, len(inputsRaw))
	for i, input := range inputsRaw {
		outpoint, err := mapOutPointBalanceModelToType(input)
		if err != nil {
			return entity.RuneTransaction{}, errors.Wrap(err, "failed to parse outpoint balance")
		}
		inputs[i] = &outpoint
	}

	outputsRaw := make([]outpointBalanceModel, 0)
	if err := json.Unmarshal(src.Outputs, &outputsRaw); err != nil {
		return entity.RuneTransaction{}, errors.Wrap(err, "failed to unmarshal outputs")
	}
	outputs := make([]*entity.OutPointBalance, len(outputsRaw))
	for i, output := range outputsRaw {
		outpoint, err := mapOutPointBalanceModelToType(output)
		if err != nil {
			return entity.RuneTransaction{}, errors.Wrap(err, "failed to parse outpoint balance")
		}
		outputs[i] = &outpoint
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
		Timestamp:   timestamp,
		Inputs:      inputs,
		Outputs:     outputs,
		Mints:       mints,
		Burns:       burns,
	}, nil
}

func mapRunestoneTypeToParams(src runes.Runestone, txHash chainhash.Hash, blockHeight uint64) (gen.CreateRunestoneParams, error) {
	var runestoneParams gen.CreateRunestoneParams

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
			runestoneParams.EtchingTerms = true
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
		if src.EtchingTerms {
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

func mapIndexedBlockModelToType(src gen.RunesIndexedBlock) (*entity.IndexedBlock, error) {
	hash, err := chainhash.NewHashFromStr(src.Hash)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse block hash")
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
		EventHash:           *eventHash,
		CumulativeEventHash: *cumulativeEventHash,
	}, nil
}

func mapIndexedBlockTypeToParams(src entity.IndexedBlock) (gen.CreateIndexedBlockParams, error) {
	return gen.CreateIndexedBlockParams{
		Height:              int32(src.Height),
		Hash:                src.Hash.String(),
		EventHash:           src.EventHash.String(),
		CumulativeEventHash: src.CumulativeEventHash.String(),
	}, nil
}
