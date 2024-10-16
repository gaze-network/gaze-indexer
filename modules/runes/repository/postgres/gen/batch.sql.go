// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: batch.sql

package gen

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const batchCreateRuneEntries = `-- name: BatchCreateRuneEntries :exec
INSERT INTO runes_entries ("rune_id", "rune", "number", "spacers", "premine", "symbol", "divisibility", "terms", "terms_amount", "terms_cap", "terms_height_start", "terms_height_end", "terms_offset_start", "terms_offset_end", "turbo", "etching_block", "etching_tx_hash", "etched_at")
VALUES(
  unnest($1::TEXT[]),
  unnest($2::TEXT[]),
  unnest($3::BIGINT[]),
  unnest($4::INT[]),
  unnest($5::DECIMAL[]),
  unnest($6::INT[]),
  unnest($7::SMALLINT[]),
  unnest($8::BOOLEAN[]),
  unnest($9::DECIMAL[]),
  unnest($10::DECIMAL[]),
  unnest($11::INT[]), -- nullable (need patch)
  unnest($12::INT[]), -- nullable (need patch)
  unnest($13::INT[]), -- nullable (need patch)
  unnest($14::INT[]), -- nullable (need patch)
  unnest($15::BOOLEAN[]),
  unnest($16::INT[]),
  unnest($17::TEXT[]),
  unnest($18::TIMESTAMP[])
)
`

type BatchCreateRuneEntriesParams struct {
	RuneIDArr           []string
	RuneArr             []string
	NumberArr           []int64
	SpacersArr          []int32
	PremineArr          []pgtype.Numeric
	SymbolArr           []int32
	DivisibilityArr     []int16
	TermsArr            []bool
	TermsAmountArr      []pgtype.Numeric
	TermsCapArr         []pgtype.Numeric
	TermsHeightStartArr []int32
	TermsHeightEndArr   []int32
	TermsOffsetStartArr []int32
	TermsOffsetEndArr   []int32
	TurboArr            []bool
	EtchingBlockArr     []int32
	EtchingTxHashArr    []string
	EtchedAtArr         []pgtype.Timestamp
}

func (q *Queries) BatchCreateRuneEntries(ctx context.Context, arg BatchCreateRuneEntriesParams) error {
	_, err := q.db.Exec(ctx, batchCreateRuneEntries,
		arg.RuneIDArr,
		arg.RuneArr,
		arg.NumberArr,
		arg.SpacersArr,
		arg.PremineArr,
		arg.SymbolArr,
		arg.DivisibilityArr,
		arg.TermsArr,
		arg.TermsAmountArr,
		arg.TermsCapArr,
		arg.TermsHeightStartArr,
		arg.TermsHeightEndArr,
		arg.TermsOffsetStartArr,
		arg.TermsOffsetEndArr,
		arg.TurboArr,
		arg.EtchingBlockArr,
		arg.EtchingTxHashArr,
		arg.EtchedAtArr,
	)
	return err
}

const batchCreateRuneEntryStates = `-- name: BatchCreateRuneEntryStates :exec
INSERT INTO runes_entry_states ("rune_id", "block_height", "mints", "burned_amount", "completed_at", "completed_at_height")
VALUES(
  unnest($1::TEXT[]),
  unnest($2::INT[]),
  unnest($3::DECIMAL[]),
  unnest($4::DECIMAL[]),
  unnest($5::TIMESTAMP[]),
  unnest($6::INT[]) -- nullable (need patch)
)
`

type BatchCreateRuneEntryStatesParams struct {
	RuneIDArr            []string
	BlockHeightArr       []int32
	MintsArr             []pgtype.Numeric
	BurnedAmountArr      []pgtype.Numeric
	CompletedAtArr       []pgtype.Timestamp
	CompletedAtHeightArr []int32
}

func (q *Queries) BatchCreateRuneEntryStates(ctx context.Context, arg BatchCreateRuneEntryStatesParams) error {
	_, err := q.db.Exec(ctx, batchCreateRuneEntryStates,
		arg.RuneIDArr,
		arg.BlockHeightArr,
		arg.MintsArr,
		arg.BurnedAmountArr,
		arg.CompletedAtArr,
		arg.CompletedAtHeightArr,
	)
	return err
}

const batchCreateRuneTransactions = `-- name: BatchCreateRuneTransactions :exec
INSERT INTO runes_transactions ("hash", "block_height", "index", "timestamp", "inputs", "outputs", "mints", "burns", "rune_etched")
VALUES (
  unnest($1::TEXT[]),
  unnest($2::INT[]),
  unnest($3::INT[]),
  unnest($4::TIMESTAMP[]),
  unnest($5::JSONB[]),
  unnest($6::JSONB[]),
  unnest($7::JSONB[]),
  unnest($8::JSONB[]),
  unnest($9::BOOLEAN[])
)
`

type BatchCreateRuneTransactionsParams struct {
	HashArr        []string
	BlockHeightArr []int32
	IndexArr       []int32
	TimestampArr   []pgtype.Timestamp
	InputsArr      [][]byte
	OutputsArr     [][]byte
	MintsArr       [][]byte
	BurnsArr       [][]byte
	RuneEtchedArr  []bool
}

func (q *Queries) BatchCreateRuneTransactions(ctx context.Context, arg BatchCreateRuneTransactionsParams) error {
	_, err := q.db.Exec(ctx, batchCreateRuneTransactions,
		arg.HashArr,
		arg.BlockHeightArr,
		arg.IndexArr,
		arg.TimestampArr,
		arg.InputsArr,
		arg.OutputsArr,
		arg.MintsArr,
		arg.BurnsArr,
		arg.RuneEtchedArr,
	)
	return err
}

const batchCreateRunesBalances = `-- name: BatchCreateRunesBalances :exec
INSERT INTO runes_balances ("pkscript", "block_height", "rune_id", "amount")
VALUES(
  unnest($1::TEXT[]),
  unnest($2::INT[]),
  unnest($3::TEXT[]),
  unnest($4::DECIMAL[])
)
`

type BatchCreateRunesBalancesParams struct {
	PkscriptArr    []string
	BlockHeightArr []int32
	RuneIDArr      []string
	AmountArr      []pgtype.Numeric
}

func (q *Queries) BatchCreateRunesBalances(ctx context.Context, arg BatchCreateRunesBalancesParams) error {
	_, err := q.db.Exec(ctx, batchCreateRunesBalances,
		arg.PkscriptArr,
		arg.BlockHeightArr,
		arg.RuneIDArr,
		arg.AmountArr,
	)
	return err
}

const batchCreateRunesOutpointBalances = `-- name: BatchCreateRunesOutpointBalances :exec
INSERT INTO runes_outpoint_balances ("rune_id", "pkscript", "tx_hash", "tx_idx", "amount", "block_height", "spent_height")
VALUES(
  unnest($1::TEXT[]),
  unnest($2::TEXT[]),
  unnest($3::TEXT[]),
  unnest($4::INT[]),
  unnest($5::DECIMAL[]),
  unnest($6::INT[]),
  unnest($7::INT[]) -- nullable (need patch)
)
`

type BatchCreateRunesOutpointBalancesParams struct {
	RuneIDArr      []string
	PkscriptArr    []string
	TxHashArr      []string
	TxIdxArr       []int32
	AmountArr      []pgtype.Numeric
	BlockHeightArr []int32
	SpentHeightArr []int32
}

func (q *Queries) BatchCreateRunesOutpointBalances(ctx context.Context, arg BatchCreateRunesOutpointBalancesParams) error {
	_, err := q.db.Exec(ctx, batchCreateRunesOutpointBalances,
		arg.RuneIDArr,
		arg.PkscriptArr,
		arg.TxHashArr,
		arg.TxIdxArr,
		arg.AmountArr,
		arg.BlockHeightArr,
		arg.SpentHeightArr,
	)
	return err
}

const batchCreateRunestones = `-- name: BatchCreateRunestones :exec
INSERT INTO runes_runestones ("tx_hash", "block_height", "etching", "etching_divisibility", "etching_premine", "etching_rune", "etching_spacers", "etching_symbol", "etching_terms", "etching_terms_amount", "etching_terms_cap", "etching_terms_height_start", "etching_terms_height_end", "etching_terms_offset_start", "etching_terms_offset_end", "etching_turbo", "edicts", "mint", "pointer", "cenotaph", "flaws")
VALUES(
  unnest($1::TEXT[]),
  unnest($2::INT[]),
  unnest($3::BOOLEAN[]),
  unnest($4::SMALLINT[]), -- nullable (need patch)
  unnest($5::DECIMAL[]),
  unnest($6::TEXT[]), -- nullable (need patch)
  unnest($7::INT[]), -- nullable (need patch)
  unnest($8::INT[]), -- nullable (need patch)
  unnest($9::BOOLEAN[]), -- nullable (need patch)
  unnest($10::DECIMAL[]),
  unnest($11::DECIMAL[]),
  unnest($12::INT[]), -- nullable (need patch)
  unnest($13::INT[]), -- nullable (need patch)
  unnest($14::INT[]), -- nullable (need patch)
  unnest($15::INT[]), -- nullable (need patch)
  unnest($16::BOOLEAN[]), -- nullable (need patch)
  unnest($17::JSONB[]),
  unnest($18::TEXT[]), -- nullable (need patch)
  unnest($19::INT[]), -- nullable (need patch)
  unnest($20::BOOLEAN[]),
  unnest($21::INT[])
)
`

type BatchCreateRunestonesParams struct {
	TxHashArr                  []string
	BlockHeightArr             []int32
	EtchingArr                 []bool
	EtchingDivisibilityArr     []int16
	EtchingPremineArr          []pgtype.Numeric
	EtchingRuneArr             []string
	EtchingSpacersArr          []int32
	EtchingSymbolArr           []int32
	EtchingTermsArr            []bool
	EtchingTermsAmountArr      []pgtype.Numeric
	EtchingTermsCapArr         []pgtype.Numeric
	EtchingTermsHeightStartArr []int32
	EtchingTermsHeightEndArr   []int32
	EtchingTermsOffsetStartArr []int32
	EtchingTermsOffsetEndArr   []int32
	EtchingTurboArr            []bool
	EdictsArr                  [][]byte
	MintArr                    []string
	PointerArr                 []int32
	CenotaphArr                []bool
	FlawsArr                   []int32
}

func (q *Queries) BatchCreateRunestones(ctx context.Context, arg BatchCreateRunestonesParams) error {
	_, err := q.db.Exec(ctx, batchCreateRunestones,
		arg.TxHashArr,
		arg.BlockHeightArr,
		arg.EtchingArr,
		arg.EtchingDivisibilityArr,
		arg.EtchingPremineArr,
		arg.EtchingRuneArr,
		arg.EtchingSpacersArr,
		arg.EtchingSymbolArr,
		arg.EtchingTermsArr,
		arg.EtchingTermsAmountArr,
		arg.EtchingTermsCapArr,
		arg.EtchingTermsHeightStartArr,
		arg.EtchingTermsHeightEndArr,
		arg.EtchingTermsOffsetStartArr,
		arg.EtchingTermsOffsetEndArr,
		arg.EtchingTurboArr,
		arg.EdictsArr,
		arg.MintArr,
		arg.PointerArr,
		arg.CenotaphArr,
		arg.FlawsArr,
	)
	return err
}

const batchSpendOutpointBalances = `-- name: BatchSpendOutpointBalances :exec
UPDATE runes_outpoint_balances
	SET "spent_height" = $1::INT
	FROM (
    SELECT 
      unnest($2::TEXT[]) AS tx_hash, 
      unnest($3::INT[]) AS tx_idx
    ) AS input
	WHERE "runes_outpoint_balances"."tx_hash" = "input"."tx_hash" AND "runes_outpoint_balances"."tx_idx" = "input"."tx_idx"
`

type BatchSpendOutpointBalancesParams struct {
	SpentHeight int32
	TxHashArr   []string
	TxIdxArr    []int32
}

func (q *Queries) BatchSpendOutpointBalances(ctx context.Context, arg BatchSpendOutpointBalancesParams) error {
	_, err := q.db.Exec(ctx, batchSpendOutpointBalances, arg.SpentHeight, arg.TxHashArr, arg.TxIdxArr)
	return err
}
