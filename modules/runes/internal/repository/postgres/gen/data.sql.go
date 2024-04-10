// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: data.sql

package gen

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const getBalancesByPkScript = `-- name: GetBalancesByPkScript :many
SELECT DISTINCT ON (rune_id) pkscript, block_height, rune_id, value FROM runes_balances WHERE pkscript = $1 AND block_height <= $2 ORDER BY rune_id, block_height DESC
`

type GetBalancesByPkScriptParams struct {
	Pkscript    string
	BlockHeight int32
}

func (q *Queries) GetBalancesByPkScript(ctx context.Context, arg GetBalancesByPkScriptParams) ([]RunesBalance, error) {
	rows, err := q.db.Query(ctx, getBalancesByPkScript, arg.Pkscript, arg.BlockHeight)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []RunesBalance
	for rows.Next() {
		var i RunesBalance
		if err := rows.Scan(
			&i.Pkscript,
			&i.BlockHeight,
			&i.RuneID,
			&i.Value,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getBalancesByRuneId = `-- name: GetBalancesByRuneId :many
SELECT DISTINCT ON (pkscript) pkscript, block_height, rune_id, value FROM runes_balances WHERE rune_id = $1 AND block_height <= $2 ORDER BY pkscript, block_height DESC
`

type GetBalancesByRuneIdParams struct {
	RuneID      string
	BlockHeight int32
}

func (q *Queries) GetBalancesByRuneId(ctx context.Context, arg GetBalancesByRuneIdParams) ([]RunesBalance, error) {
	rows, err := q.db.Query(ctx, getBalancesByRuneId, arg.RuneID, arg.BlockHeight)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []RunesBalance
	for rows.Next() {
		var i RunesBalance
		if err := rows.Scan(
			&i.Pkscript,
			&i.BlockHeight,
			&i.RuneID,
			&i.Value,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getOutPointBalances = `-- name: GetOutPointBalances :many
SELECT rune_id, tx_hash, tx_idx, value FROM runes_outpoint_balances WHERE tx_hash = $1 AND tx_idx = $2
`

type GetOutPointBalancesParams struct {
	TxHash string
	TxIdx  int32
}

func (q *Queries) GetOutPointBalances(ctx context.Context, arg GetOutPointBalancesParams) ([]RunesOutpointBalance, error) {
	rows, err := q.db.Query(ctx, getOutPointBalances, arg.TxHash, arg.TxIdx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []RunesOutpointBalance
	for rows.Next() {
		var i RunesOutpointBalance
		if err := rows.Scan(
			&i.RuneID,
			&i.TxHash,
			&i.TxIdx,
			&i.Value,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRuneEntryByRune = `-- name: GetRuneEntryByRune :one
SELECT rune_id, rune, spacers, burned_amount, mints, premine, symbol, divisibility, terms, terms_amount, terms_cap, terms_height_start, terms_height_end, terms_offset_start, terms_offset_end, completion_time FROM runes_entries WHERE rune = $1 FOR UPDATE
`

// using FOR UPDATE to prevent other connections for updating the same rune entries if they are being accessed by Processor
func (q *Queries) GetRuneEntryByRune(ctx context.Context, rune string) (RunesEntry, error) {
	row := q.db.QueryRow(ctx, getRuneEntryByRune, rune)
	var i RunesEntry
	err := row.Scan(
		&i.RuneID,
		&i.Rune,
		&i.Spacers,
		&i.BurnedAmount,
		&i.Mints,
		&i.Premine,
		&i.Symbol,
		&i.Divisibility,
		&i.Terms,
		&i.TermsAmount,
		&i.TermsCap,
		&i.TermsHeightStart,
		&i.TermsHeightEnd,
		&i.TermsOffsetStart,
		&i.TermsOffsetEnd,
		&i.CompletionTime,
	)
	return i, err
}

const getRuneEntryByRuneId = `-- name: GetRuneEntryByRuneId :one
SELECT rune_id, rune, spacers, burned_amount, mints, premine, symbol, divisibility, terms, terms_amount, terms_cap, terms_height_start, terms_height_end, terms_offset_start, terms_offset_end, completion_time FROM runes_entries WHERE rune_id = $1 FOR UPDATE
`

// using FOR UPDATE to prevent other connections for updating the same rune entries if they are being accessed by Processor
func (q *Queries) GetRuneEntryByRuneId(ctx context.Context, runeID string) (RunesEntry, error) {
	row := q.db.QueryRow(ctx, getRuneEntryByRuneId, runeID)
	var i RunesEntry
	err := row.Scan(
		&i.RuneID,
		&i.Rune,
		&i.Spacers,
		&i.BurnedAmount,
		&i.Mints,
		&i.Premine,
		&i.Symbol,
		&i.Divisibility,
		&i.Terms,
		&i.TermsAmount,
		&i.TermsCap,
		&i.TermsHeightStart,
		&i.TermsHeightEnd,
		&i.TermsOffsetStart,
		&i.TermsOffsetEnd,
		&i.CompletionTime,
	)
	return i, err
}

const setRuneEntry = `-- name: SetRuneEntry :exec
INSERT INTO runes_entries (rune_id, rune, spacers, burned_amount, mints, premine, symbol, divisibility, terms, terms_amount, terms_cap, terms_height_start, terms_height_end, terms_offset_start, terms_offset_end, completion_time) 
  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) 
  ON CONFLICT (rune_id) DO UPDATE SET (burned_amount, mints, completion_time) = (excluded.burned_amount, excluded.mints, excluded.completion_time)
`

type SetRuneEntryParams struct {
	RuneID           string
	Rune             string
	Spacers          int32
	BurnedAmount     pgtype.Numeric
	Mints            pgtype.Numeric
	Premine          pgtype.Numeric
	Symbol           int32
	Divisibility     int16
	Terms            bool
	TermsAmount      pgtype.Numeric
	TermsCap         pgtype.Numeric
	TermsHeightStart pgtype.Numeric
	TermsHeightEnd   pgtype.Numeric
	TermsOffsetStart pgtype.Numeric
	TermsOffsetEnd   pgtype.Numeric
	CompletionTime   pgtype.Timestamp
}

func (q *Queries) SetRuneEntry(ctx context.Context, arg SetRuneEntryParams) error {
	_, err := q.db.Exec(ctx, setRuneEntry,
		arg.RuneID,
		arg.Rune,
		arg.Spacers,
		arg.BurnedAmount,
		arg.Mints,
		arg.Premine,
		arg.Symbol,
		arg.Divisibility,
		arg.Terms,
		arg.TermsAmount,
		arg.TermsCap,
		arg.TermsHeightStart,
		arg.TermsHeightEnd,
		arg.TermsOffsetStart,
		arg.TermsOffsetEnd,
		arg.CompletionTime,
	)
	return err
}

const updateLatestBlockHeight = `-- name: UpdateLatestBlockHeight :exec
UPDATE runes_processor_state SET latest_block_height = $1
`

func (q *Queries) UpdateLatestBlockHeight(ctx context.Context, latestBlockHeight int32) error {
	_, err := q.db.Exec(ctx, updateLatestBlockHeight, latestBlockHeight)
	return err
}
