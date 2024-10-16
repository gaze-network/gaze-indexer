package gen

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v5/pgtype"
)

type BatchCreateRuneEntriesPatchedParams struct {
	BatchCreateRuneEntriesParams
	TermsHeightStartArr []pgtype.Int4
	TermsHeightEndArr   []pgtype.Int4
	TermsOffsetStartArr []pgtype.Int4
	TermsOffsetEndArr   []pgtype.Int4
}

func (q *Queries) BatchCreateRuneEntriesPatched(ctx context.Context, arg BatchCreateRuneEntriesPatchedParams) error {
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
	return errors.WithStack(err)
}

type BatchCreateRuneEntryStatesPatchedParams struct {
	BatchCreateRuneEntryStatesParams
	CompletedAtHeightArr []pgtype.Int4
}

func (q *Queries) BatchCreateRuneEntryStatesPatched(ctx context.Context, arg BatchCreateRuneEntryStatesPatchedParams) error {
	_, err := q.db.Exec(ctx, batchCreateRuneEntryStates,
		arg.RuneIDArr,
		arg.BlockHeightArr,
		arg.MintsArr,
		arg.BurnedAmountArr,
		arg.CompletedAtArr,
		arg.CompletedAtHeightArr,
	)
	return errors.WithStack(err)
}

type BatchCreateRunesOutpointBalancesPatchedParams struct {
	BatchCreateRunesOutpointBalancesParams
	SpentHeightArr []pgtype.Int4
}

func (q *Queries) BatchCreateRunesOutpointBalancesPatched(ctx context.Context, arg BatchCreateRunesOutpointBalancesPatchedParams) error {
	_, err := q.db.Exec(ctx, batchCreateRunesOutpointBalances,
		arg.RuneIDArr,
		arg.PkscriptArr,
		arg.TxHashArr,
		arg.TxIdxArr,
		arg.AmountArr,
		arg.BlockHeightArr,
		arg.SpentHeightArr,
	)
	return errors.WithStack(err)
}

// const batchCreateRunestones = `-- name: BatchCreateRunestones :exec
// INSERT INTO runes_runestones ("tx_hash", "block_height", "etching", "etching_divisibility", "etching_premine", "etching_rune", "etching_spacers", "etching_symbol", "etching_terms", "etching_terms_amount", "etching_terms_cap", "etching_terms_height_start", "etching_terms_height_end", "etching_terms_offset_start", "etching_terms_offset_end", "etching_turbo", "edicts", "mint", "pointer", "cenotaph", "flaws")
// VALUES(
//   unnest($1::TEXT[]),
//   unnest($2::INT[]),
//   unnest($3::BOOLEAN[]),
//   unnest($4::SMALLINT[]), -- nullable (need patch)
//   unnest($5::DECIMAL[]),
//   unnest($6::TEXT[]), -- nullable (need patch)
//   unnest($7::INT[]), -- nullable (need patch)
//   unnest($8::INT[]), -- nullable (need patch)
//   unnest($9::BOOLEAN[]), -- nullable (need patch)
//   unnest($10::DECIMAL[]),
//   unnest($11::DECIMAL[]),
//   unnest($12::INT[]), -- nullable (need patch)
//   unnest($13::INT[]), -- nullable (need patch)
//   unnest($14::INT[]), -- nullable (need patch)
//   unnest($15::INT[]), -- nullable (need patch)
//   unnest($16::BOOLEAN[]), -- nullable (need patch)
//   unnest($17::JSONB[]),
//   unnest($18::TEXT[]), -- nullable (need patch)
//   unnest($19::INT[]), -- nullable (need patch)
//   unnest($20::BOOLEAN[]),
//   unnest($21::INT[])
// )
// `

type BatchCreateRunestonesPatchedParams struct {
	BatchCreateRunestonesParams
	EtchingDivisibilityArr     []pgtype.Int2
	EtchingRuneArr             []pgtype.Text
	EtchingSpacersArr          []pgtype.Int4
	EtchingSymbolArr           []pgtype.Int4
	EtchingTermsArr            []pgtype.Bool
	EtchingTermsHeightStartArr []pgtype.Int4
	EtchingTermsHeightEndArr   []pgtype.Int4
	EtchingTermsOffsetStartArr []pgtype.Int4
	EtchingTermsOffsetEndArr   []pgtype.Int4
	EtchingTurboArr            []pgtype.Bool
	MintArr                    []pgtype.Text
	PointerArr                 []pgtype.Int4
}

func (q *Queries) BatchCreateRunestonesPatched(ctx context.Context, arg BatchCreateRunestonesPatchedParams) error {
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
	return errors.WithStack(err)
}
