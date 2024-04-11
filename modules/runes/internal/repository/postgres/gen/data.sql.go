// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: data.sql

package gen

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createIndexedBlock = `-- name: CreateIndexedBlock :exec
INSERT INTO runes_indexed_blocks (hash, height, event_hash, cumulative_event_hash) VALUES ($1, $2, $3, $4)
`

type CreateIndexedBlockParams struct {
	Hash                string
	Height              int32
	EventHash           string
	CumulativeEventHash string
}

func (q *Queries) CreateIndexedBlock(ctx context.Context, arg CreateIndexedBlockParams) error {
	_, err := q.db.Exec(ctx, createIndexedBlock,
		arg.Hash,
		arg.Height,
		arg.EventHash,
		arg.CumulativeEventHash,
	)
	return err
}

const deleteIndexedBlockByHash = `-- name: DeleteIndexedBlockByHash :exec
DELETE FROM runes_indexed_blocks WHERE hash = $1
`

func (q *Queries) DeleteIndexedBlockByHash(ctx context.Context, hash string) error {
	_, err := q.db.Exec(ctx, deleteIndexedBlockByHash, hash)
	return err
}

const getBalancesByPkScript = `-- name: GetBalancesByPkScript :many
SELECT DISTINCT ON (rune_id) pkscript, block_height, rune_id, amount FROM runes_balances WHERE pkscript = $1 AND block_height <= $2 ORDER BY rune_id, block_height DESC
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
			&i.Amount,
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
SELECT DISTINCT ON (pkscript) pkscript, block_height, rune_id, amount FROM runes_balances WHERE rune_id = $1 AND block_height <= $2 ORDER BY pkscript, block_height DESC
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
			&i.Amount,
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
SELECT rune_id, tx_hash, tx_idx, amount, block_height, spent_height FROM runes_outpoint_balances WHERE tx_hash = $1 AND tx_idx = $2
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
			&i.Amount,
			&i.BlockHeight,
			&i.SpentHeight,
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

const getRuneEntriesByRuneIds = `-- name: GetRuneEntriesByRuneIds :many
SELECT rune_id, rune, spacers, burned_amount, mints, premine, symbol, divisibility, terms, terms_amount, terms_cap, terms_height_start, terms_height_end, terms_offset_start, terms_offset_end, completion_time FROM runes_entries WHERE rune_id = ANY($1::text[]) FOR UPDATE
`

// using FOR UPDATE to prevent other connections for updating the same rune entries if they are being accessed by Processor
func (q *Queries) GetRuneEntriesByRuneIds(ctx context.Context, runeIds []string) ([]RunesEntry, error) {
	rows, err := q.db.Query(ctx, getRuneEntriesByRuneIds, runeIds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []RunesEntry
	for rows.Next() {
		var i RunesEntry
		if err := rows.Scan(
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

const getRuneEntriesByRunes = `-- name: GetRuneEntriesByRunes :many
SELECT rune_id, rune, spacers, burned_amount, mints, premine, symbol, divisibility, terms, terms_amount, terms_cap, terms_height_start, terms_height_end, terms_offset_start, terms_offset_end, completion_time FROM runes_entries WHERE rune = ANY($1::text[]) FOR UPDATE
`

// using FOR UPDATE to prevent other connections for updating the same rune entries if they are being accessed by Processor
func (q *Queries) GetRuneEntriesByRunes(ctx context.Context, runes []string) ([]RunesEntry, error) {
	rows, err := q.db.Query(ctx, getRuneEntriesByRunes, runes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []RunesEntry
	for rows.Next() {
		var i RunesEntry
		if err := rows.Scan(
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

const getRunesProcessorState = `-- name: GetRunesProcessorState :one
SELECT latest_block_height, latest_block_hash, latest_prev_block_hash, updated_at FROM runes_processor_state
`

func (q *Queries) GetRunesProcessorState(ctx context.Context) (RunesProcessorState, error) {
	row := q.db.QueryRow(ctx, getRunesProcessorState)
	var i RunesProcessorState
	err := row.Scan(
		&i.LatestBlockHeight,
		&i.LatestBlockHash,
		&i.LatestPrevBlockHash,
		&i.UpdatedAt,
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
	TermsHeightStart pgtype.Int4
	TermsHeightEnd   pgtype.Int4
	TermsOffsetStart pgtype.Int4
	TermsOffsetEnd   pgtype.Int4
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

const updateLatestBlock = `-- name: UpdateLatestBlock :exec
UPDATE runes_processor_state SET latest_block_height = $1, latest_block_hash = $2, latest_prev_block_hash = $3
`

type UpdateLatestBlockParams struct {
	LatestBlockHeight   int32
	LatestBlockHash     string
	LatestPrevBlockHash string
}

func (q *Queries) UpdateLatestBlock(ctx context.Context, arg UpdateLatestBlockParams) error {
	_, err := q.db.Exec(ctx, updateLatestBlock, arg.LatestBlockHeight, arg.LatestBlockHash, arg.LatestPrevBlockHash)
	return err
}
