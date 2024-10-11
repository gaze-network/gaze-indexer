// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: data.sql

package gen

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const countRuneEntries = `-- name: CountRuneEntries :one
SELECT COUNT(*) FROM runes_entries
`

func (q *Queries) CountRuneEntries(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, countRuneEntries)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createIndexedBlock = `-- name: CreateIndexedBlock :exec
INSERT INTO runes_indexed_blocks (hash, height, prev_hash, event_hash, cumulative_event_hash) VALUES ($1, $2, $3, $4, $5)
`

type CreateIndexedBlockParams struct {
	Hash                string
	Height              int32
	PrevHash            string
	EventHash           string
	CumulativeEventHash string
}

func (q *Queries) CreateIndexedBlock(ctx context.Context, arg CreateIndexedBlockParams) error {
	_, err := q.db.Exec(ctx, createIndexedBlock,
		arg.Hash,
		arg.Height,
		arg.PrevHash,
		arg.EventHash,
		arg.CumulativeEventHash,
	)
	return err
}

const deleteIndexedBlockSinceHeight = `-- name: DeleteIndexedBlockSinceHeight :exec
DELETE FROM runes_indexed_blocks WHERE height >= $1
`

func (q *Queries) DeleteIndexedBlockSinceHeight(ctx context.Context, height int32) error {
	_, err := q.db.Exec(ctx, deleteIndexedBlockSinceHeight, height)
	return err
}

const deleteOutPointBalancesSinceHeight = `-- name: DeleteOutPointBalancesSinceHeight :exec
DELETE FROM runes_outpoint_balances WHERE block_height >= $1
`

func (q *Queries) DeleteOutPointBalancesSinceHeight(ctx context.Context, blockHeight int32) error {
	_, err := q.db.Exec(ctx, deleteOutPointBalancesSinceHeight, blockHeight)
	return err
}

const deleteRuneBalancesSinceHeight = `-- name: DeleteRuneBalancesSinceHeight :exec
DELETE FROM runes_balances WHERE block_height >= $1
`

func (q *Queries) DeleteRuneBalancesSinceHeight(ctx context.Context, blockHeight int32) error {
	_, err := q.db.Exec(ctx, deleteRuneBalancesSinceHeight, blockHeight)
	return err
}

const deleteRuneEntriesSinceHeight = `-- name: DeleteRuneEntriesSinceHeight :exec
DELETE FROM runes_entries WHERE etching_block >= $1
`

func (q *Queries) DeleteRuneEntriesSinceHeight(ctx context.Context, etchingBlock int32) error {
	_, err := q.db.Exec(ctx, deleteRuneEntriesSinceHeight, etchingBlock)
	return err
}

const deleteRuneEntryStatesSinceHeight = `-- name: DeleteRuneEntryStatesSinceHeight :exec
DELETE FROM runes_entry_states WHERE block_height >= $1
`

func (q *Queries) DeleteRuneEntryStatesSinceHeight(ctx context.Context, blockHeight int32) error {
	_, err := q.db.Exec(ctx, deleteRuneEntryStatesSinceHeight, blockHeight)
	return err
}

const deleteRuneTransactionsSinceHeight = `-- name: DeleteRuneTransactionsSinceHeight :exec
DELETE FROM runes_transactions WHERE block_height >= $1
`

func (q *Queries) DeleteRuneTransactionsSinceHeight(ctx context.Context, blockHeight int32) error {
	_, err := q.db.Exec(ctx, deleteRuneTransactionsSinceHeight, blockHeight)
	return err
}

const deleteRunestonesSinceHeight = `-- name: DeleteRunestonesSinceHeight :exec
DELETE FROM runes_runestones WHERE block_height >= $1
`

func (q *Queries) DeleteRunestonesSinceHeight(ctx context.Context, blockHeight int32) error {
	_, err := q.db.Exec(ctx, deleteRunestonesSinceHeight, blockHeight)
	return err
}

const getBalanceByPkScriptAndRuneId = `-- name: GetBalanceByPkScriptAndRuneId :one
SELECT pkscript, block_height, rune_id, amount FROM runes_balances WHERE pkscript = $1 AND rune_id = $2 AND block_height <= $3 ORDER BY block_height DESC LIMIT 1
`

type GetBalanceByPkScriptAndRuneIdParams struct {
	Pkscript    string
	RuneID      string
	BlockHeight int32
}

func (q *Queries) GetBalanceByPkScriptAndRuneId(ctx context.Context, arg GetBalanceByPkScriptAndRuneIdParams) (RunesBalance, error) {
	row := q.db.QueryRow(ctx, getBalanceByPkScriptAndRuneId, arg.Pkscript, arg.RuneID, arg.BlockHeight)
	var i RunesBalance
	err := row.Scan(
		&i.Pkscript,
		&i.BlockHeight,
		&i.RuneID,
		&i.Amount,
	)
	return i, err
}

const getBalancesByPkScript = `-- name: GetBalancesByPkScript :many
WITH balances AS (
  SELECT DISTINCT ON (rune_id) pkscript, block_height, rune_id, amount FROM runes_balances WHERE pkscript = $1 AND block_height <= $2 ORDER BY rune_id, block_height DESC
)
SELECT pkscript, block_height, rune_id, amount FROM balances WHERE amount > 0 ORDER BY amount DESC, rune_id LIMIT $3 OFFSET $4
`

type GetBalancesByPkScriptParams struct {
	Pkscript    string
	BlockHeight int32
	Limit       int32
	Offset      int32
}

type GetBalancesByPkScriptRow struct {
	Pkscript    string
	BlockHeight int32
	RuneID      string
	Amount      pgtype.Numeric
}

func (q *Queries) GetBalancesByPkScript(ctx context.Context, arg GetBalancesByPkScriptParams) ([]GetBalancesByPkScriptRow, error) {
	rows, err := q.db.Query(ctx, getBalancesByPkScript,
		arg.Pkscript,
		arg.BlockHeight,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetBalancesByPkScriptRow
	for rows.Next() {
		var i GetBalancesByPkScriptRow
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
WITH balances AS (
  SELECT DISTINCT ON (pkscript) pkscript, block_height, rune_id, amount FROM runes_balances WHERE rune_id = $1 AND block_height <= $2 ORDER BY pkscript, block_height DESC
)
SELECT pkscript, block_height, rune_id, amount FROM balances WHERE amount > 0 ORDER BY amount DESC, pkscript LIMIT $3 OFFSET $4
`

type GetBalancesByRuneIdParams struct {
	RuneID      string
	BlockHeight int32
	Limit       int32
	Offset      int32
}

type GetBalancesByRuneIdRow struct {
	Pkscript    string
	BlockHeight int32
	RuneID      string
	Amount      pgtype.Numeric
}

func (q *Queries) GetBalancesByRuneId(ctx context.Context, arg GetBalancesByRuneIdParams) ([]GetBalancesByRuneIdRow, error) {
	rows, err := q.db.Query(ctx, getBalancesByRuneId,
		arg.RuneID,
		arg.BlockHeight,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetBalancesByRuneIdRow
	for rows.Next() {
		var i GetBalancesByRuneIdRow
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

const getIndexedBlockByHeight = `-- name: GetIndexedBlockByHeight :one
SELECT height, hash, prev_hash, event_hash, cumulative_event_hash FROM runes_indexed_blocks WHERE height = $1
`

func (q *Queries) GetIndexedBlockByHeight(ctx context.Context, height int32) (RunesIndexedBlock, error) {
	row := q.db.QueryRow(ctx, getIndexedBlockByHeight, height)
	var i RunesIndexedBlock
	err := row.Scan(
		&i.Height,
		&i.Hash,
		&i.PrevHash,
		&i.EventHash,
		&i.CumulativeEventHash,
	)
	return i, err
}

const getLatestIndexedBlock = `-- name: GetLatestIndexedBlock :one
SELECT height, hash, prev_hash, event_hash, cumulative_event_hash FROM runes_indexed_blocks ORDER BY height DESC LIMIT 1
`

func (q *Queries) GetLatestIndexedBlock(ctx context.Context) (RunesIndexedBlock, error) {
	row := q.db.QueryRow(ctx, getLatestIndexedBlock)
	var i RunesIndexedBlock
	err := row.Scan(
		&i.Height,
		&i.Hash,
		&i.PrevHash,
		&i.EventHash,
		&i.CumulativeEventHash,
	)
	return i, err
}

const getOngoingRuneEntries = `-- name: GetOngoingRuneEntries :many
WITH states AS (
  -- select latest state
  SELECT DISTINCT ON (rune_id) rune_id, block_height, mints, burned_amount, completed_at, completed_at_height FROM runes_entry_states WHERE block_height <= $1::integer ORDER BY rune_id, block_height DESC
)
SELECT runes_entries.rune_id, number, rune, spacers, premine, symbol, divisibility, terms, terms_amount, terms_cap, terms_height_start, terms_height_end, terms_offset_start, terms_offset_end, turbo, etching_block, etching_tx_hash, etched_at, states.rune_id, block_height, mints, burned_amount, completed_at, completed_at_height FROM runes_entries 
  LEFT JOIN states ON runes_entries.rune_id = states.rune_id
  WHERE (
    runes_entries.terms = TRUE AND
    COALESCE(runes_entries.terms_amount, 0) != 0 AND
    COALESCE(runes_entries.terms_cap, 0) != 0 AND
    states.mints < runes_entries.terms_cap AND
    (
      runes_entries.terms_height_start IS NULL OR runes_entries.terms_height_start <= $1::integer
    ) AND (
      runes_entries.terms_height_end IS NULL OR $1::integer <= runes_entries.terms_height_end
    ) AND (
      runes_entries.terms_offset_start IS NULL OR runes_entries.terms_offset_start + runes_entries.etching_block <= $1::integer
    ) AND (
      runes_entries.terms_offset_end IS NULL OR $1::integer <= runes_entries.terms_offset_start + runes_entries.etching_block
    )

  ) AND (
    $2::text = '' OR
    runes_entries.rune ILIKE '%' || $2::text || '%'
  )
  ORDER BY states.mints DESC
  LIMIT $4 OFFSET $3
`

type GetOngoingRuneEntriesParams struct {
	Height int32
	Search string
	Offset int32
	Limit  int32
}

type GetOngoingRuneEntriesRow struct {
	RuneID            string
	Number            int64
	Rune              string
	Spacers           int32
	Premine           pgtype.Numeric
	Symbol            int32
	Divisibility      int16
	Terms             bool
	TermsAmount       pgtype.Numeric
	TermsCap          pgtype.Numeric
	TermsHeightStart  pgtype.Int4
	TermsHeightEnd    pgtype.Int4
	TermsOffsetStart  pgtype.Int4
	TermsOffsetEnd    pgtype.Int4
	Turbo             bool
	EtchingBlock      int32
	EtchingTxHash     string
	EtchedAt          pgtype.Timestamp
	RuneID_2          pgtype.Text
	BlockHeight       pgtype.Int4
	Mints             pgtype.Numeric
	BurnedAmount      pgtype.Numeric
	CompletedAt       pgtype.Timestamp
	CompletedAtHeight pgtype.Int4
}

func (q *Queries) GetOngoingRuneEntries(ctx context.Context, arg GetOngoingRuneEntriesParams) ([]GetOngoingRuneEntriesRow, error) {
	rows, err := q.db.Query(ctx, getOngoingRuneEntries,
		arg.Height,
		arg.Search,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetOngoingRuneEntriesRow
	for rows.Next() {
		var i GetOngoingRuneEntriesRow
		if err := rows.Scan(
			&i.RuneID,
			&i.Number,
			&i.Rune,
			&i.Spacers,
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
			&i.Turbo,
			&i.EtchingBlock,
			&i.EtchingTxHash,
			&i.EtchedAt,
			&i.RuneID_2,
			&i.BlockHeight,
			&i.Mints,
			&i.BurnedAmount,
			&i.CompletedAt,
			&i.CompletedAtHeight,
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

const getOutPointBalancesAtOutPoint = `-- name: GetOutPointBalancesAtOutPoint :many
SELECT rune_id, pkscript, tx_hash, tx_idx, amount, block_height, spent_height FROM runes_outpoint_balances WHERE tx_hash = $1 AND tx_idx = $2
`

type GetOutPointBalancesAtOutPointParams struct {
	TxHash string
	TxIdx  int32
}

func (q *Queries) GetOutPointBalancesAtOutPoint(ctx context.Context, arg GetOutPointBalancesAtOutPointParams) ([]RunesOutpointBalance, error) {
	rows, err := q.db.Query(ctx, getOutPointBalancesAtOutPoint, arg.TxHash, arg.TxIdx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []RunesOutpointBalance
	for rows.Next() {
		var i RunesOutpointBalance
		if err := rows.Scan(
			&i.RuneID,
			&i.Pkscript,
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

const getRuneEntries = `-- name: GetRuneEntries :many
WITH states AS (
  -- select latest state
  SELECT DISTINCT ON (rune_id) rune_id, block_height, mints, burned_amount, completed_at, completed_at_height FROM runes_entry_states WHERE block_height <= $4 ORDER BY rune_id, block_height DESC
)
SELECT runes_entries.rune_id, number, rune, spacers, premine, symbol, divisibility, terms, terms_amount, terms_cap, terms_height_start, terms_height_end, terms_offset_start, terms_offset_end, turbo, etching_block, etching_tx_hash, etched_at, states.rune_id, block_height, mints, burned_amount, completed_at, completed_at_height FROM runes_entries 
  LEFT JOIN states ON runes_entries.rune_id = states.rune_id
  WHERE (
    $1 = '' OR
    runes_entries.rune ILIKE $1 || '%'
  )
  ORDER BY runes_entries.number 
  LIMIT $3 OFFSET $2
`

type GetRuneEntriesParams struct {
	Search interface{}
	Offset int32
	Limit  int32
	Height int32
}

type GetRuneEntriesRow struct {
	RuneID            string
	Number            int64
	Rune              string
	Spacers           int32
	Premine           pgtype.Numeric
	Symbol            int32
	Divisibility      int16
	Terms             bool
	TermsAmount       pgtype.Numeric
	TermsCap          pgtype.Numeric
	TermsHeightStart  pgtype.Int4
	TermsHeightEnd    pgtype.Int4
	TermsOffsetStart  pgtype.Int4
	TermsOffsetEnd    pgtype.Int4
	Turbo             bool
	EtchingBlock      int32
	EtchingTxHash     string
	EtchedAt          pgtype.Timestamp
	RuneID_2          pgtype.Text
	BlockHeight       pgtype.Int4
	Mints             pgtype.Numeric
	BurnedAmount      pgtype.Numeric
	CompletedAt       pgtype.Timestamp
	CompletedAtHeight pgtype.Int4
}

func (q *Queries) GetRuneEntries(ctx context.Context, arg GetRuneEntriesParams) ([]GetRuneEntriesRow, error) {
	rows, err := q.db.Query(ctx, getRuneEntries,
		arg.Search,
		arg.Offset,
		arg.Limit,
		arg.Height,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetRuneEntriesRow
	for rows.Next() {
		var i GetRuneEntriesRow
		if err := rows.Scan(
			&i.RuneID,
			&i.Number,
			&i.Rune,
			&i.Spacers,
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
			&i.Turbo,
			&i.EtchingBlock,
			&i.EtchingTxHash,
			&i.EtchedAt,
			&i.RuneID_2,
			&i.BlockHeight,
			&i.Mints,
			&i.BurnedAmount,
			&i.CompletedAt,
			&i.CompletedAtHeight,
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
WITH states AS (
  -- select latest state
  SELECT DISTINCT ON (rune_id) rune_id, block_height, mints, burned_amount, completed_at, completed_at_height FROM runes_entry_states WHERE rune_id = ANY($1::text[]) ORDER BY rune_id, block_height DESC
)
SELECT runes_entries.rune_id, number, rune, spacers, premine, symbol, divisibility, terms, terms_amount, terms_cap, terms_height_start, terms_height_end, terms_offset_start, terms_offset_end, turbo, etching_block, etching_tx_hash, etched_at, states.rune_id, block_height, mints, burned_amount, completed_at, completed_at_height FROM runes_entries
  LEFT JOIN states ON runes_entries.rune_id = states.rune_id
  WHERE runes_entries.rune_id = ANY($1::text[])
`

type GetRuneEntriesByRuneIdsRow struct {
	RuneID            string
	Number            int64
	Rune              string
	Spacers           int32
	Premine           pgtype.Numeric
	Symbol            int32
	Divisibility      int16
	Terms             bool
	TermsAmount       pgtype.Numeric
	TermsCap          pgtype.Numeric
	TermsHeightStart  pgtype.Int4
	TermsHeightEnd    pgtype.Int4
	TermsOffsetStart  pgtype.Int4
	TermsOffsetEnd    pgtype.Int4
	Turbo             bool
	EtchingBlock      int32
	EtchingTxHash     string
	EtchedAt          pgtype.Timestamp
	RuneID_2          pgtype.Text
	BlockHeight       pgtype.Int4
	Mints             pgtype.Numeric
	BurnedAmount      pgtype.Numeric
	CompletedAt       pgtype.Timestamp
	CompletedAtHeight pgtype.Int4
}

func (q *Queries) GetRuneEntriesByRuneIds(ctx context.Context, runeIds []string) ([]GetRuneEntriesByRuneIdsRow, error) {
	rows, err := q.db.Query(ctx, getRuneEntriesByRuneIds, runeIds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetRuneEntriesByRuneIdsRow
	for rows.Next() {
		var i GetRuneEntriesByRuneIdsRow
		if err := rows.Scan(
			&i.RuneID,
			&i.Number,
			&i.Rune,
			&i.Spacers,
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
			&i.Turbo,
			&i.EtchingBlock,
			&i.EtchingTxHash,
			&i.EtchedAt,
			&i.RuneID_2,
			&i.BlockHeight,
			&i.Mints,
			&i.BurnedAmount,
			&i.CompletedAt,
			&i.CompletedAtHeight,
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

const getRuneEntriesByRuneIdsAndHeight = `-- name: GetRuneEntriesByRuneIdsAndHeight :many
WITH states AS (
  -- select latest state
  SELECT DISTINCT ON (rune_id) rune_id, block_height, mints, burned_amount, completed_at, completed_at_height FROM runes_entry_states WHERE rune_id = ANY($1::text[]) AND block_height <= $2 ORDER BY rune_id, block_height DESC
)
SELECT runes_entries.rune_id, number, rune, spacers, premine, symbol, divisibility, terms, terms_amount, terms_cap, terms_height_start, terms_height_end, terms_offset_start, terms_offset_end, turbo, etching_block, etching_tx_hash, etched_at, states.rune_id, block_height, mints, burned_amount, completed_at, completed_at_height FROM runes_entries
  LEFT JOIN states ON runes_entries.rune_id = states.rune_id
  WHERE runes_entries.rune_id = ANY($1::text[]) AND etching_block <= $2
`

type GetRuneEntriesByRuneIdsAndHeightParams struct {
	RuneIds []string
	Height  int32
}

type GetRuneEntriesByRuneIdsAndHeightRow struct {
	RuneID            string
	Number            int64
	Rune              string
	Spacers           int32
	Premine           pgtype.Numeric
	Symbol            int32
	Divisibility      int16
	Terms             bool
	TermsAmount       pgtype.Numeric
	TermsCap          pgtype.Numeric
	TermsHeightStart  pgtype.Int4
	TermsHeightEnd    pgtype.Int4
	TermsOffsetStart  pgtype.Int4
	TermsOffsetEnd    pgtype.Int4
	Turbo             bool
	EtchingBlock      int32
	EtchingTxHash     string
	EtchedAt          pgtype.Timestamp
	RuneID_2          pgtype.Text
	BlockHeight       pgtype.Int4
	Mints             pgtype.Numeric
	BurnedAmount      pgtype.Numeric
	CompletedAt       pgtype.Timestamp
	CompletedAtHeight pgtype.Int4
}

func (q *Queries) GetRuneEntriesByRuneIdsAndHeight(ctx context.Context, arg GetRuneEntriesByRuneIdsAndHeightParams) ([]GetRuneEntriesByRuneIdsAndHeightRow, error) {
	rows, err := q.db.Query(ctx, getRuneEntriesByRuneIdsAndHeight, arg.RuneIds, arg.Height)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetRuneEntriesByRuneIdsAndHeightRow
	for rows.Next() {
		var i GetRuneEntriesByRuneIdsAndHeightRow
		if err := rows.Scan(
			&i.RuneID,
			&i.Number,
			&i.Rune,
			&i.Spacers,
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
			&i.Turbo,
			&i.EtchingBlock,
			&i.EtchingTxHash,
			&i.EtchedAt,
			&i.RuneID_2,
			&i.BlockHeight,
			&i.Mints,
			&i.BurnedAmount,
			&i.CompletedAt,
			&i.CompletedAtHeight,
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

const getRuneIdFromRune = `-- name: GetRuneIdFromRune :one
SELECT rune_id FROM runes_entries WHERE rune = $1
`

func (q *Queries) GetRuneIdFromRune(ctx context.Context, rune string) (string, error) {
	row := q.db.QueryRow(ctx, getRuneIdFromRune, rune)
	var rune_id string
	err := row.Scan(&rune_id)
	return rune_id, err
}

const getRuneTransaction = `-- name: GetRuneTransaction :one
SELECT hash, runes_transactions.block_height, index, timestamp, inputs, outputs, mints, burns, rune_etched, tx_hash, runes_runestones.block_height, etching, etching_divisibility, etching_premine, etching_rune, etching_spacers, etching_symbol, etching_terms, etching_terms_amount, etching_terms_cap, etching_terms_height_start, etching_terms_height_end, etching_terms_offset_start, etching_terms_offset_end, etching_turbo, edicts, mint, pointer, cenotaph, flaws FROM runes_transactions
  LEFT JOIN runes_runestones ON runes_transactions.hash = runes_runestones.tx_hash
  WHERE hash = $1 LIMIT 1
`

type GetRuneTransactionRow struct {
	Hash                    string
	BlockHeight             int32
	Index                   int32
	Timestamp               pgtype.Timestamp
	Inputs                  []byte
	Outputs                 []byte
	Mints                   []byte
	Burns                   []byte
	RuneEtched              bool
	TxHash                  pgtype.Text
	BlockHeight_2           pgtype.Int4
	Etching                 pgtype.Bool
	EtchingDivisibility     pgtype.Int2
	EtchingPremine          pgtype.Numeric
	EtchingRune             pgtype.Text
	EtchingSpacers          pgtype.Int4
	EtchingSymbol           pgtype.Int4
	EtchingTerms            pgtype.Bool
	EtchingTermsAmount      pgtype.Numeric
	EtchingTermsCap         pgtype.Numeric
	EtchingTermsHeightStart pgtype.Int4
	EtchingTermsHeightEnd   pgtype.Int4
	EtchingTermsOffsetStart pgtype.Int4
	EtchingTermsOffsetEnd   pgtype.Int4
	EtchingTurbo            pgtype.Bool
	Edicts                  []byte
	Mint                    pgtype.Text
	Pointer                 pgtype.Int4
	Cenotaph                pgtype.Bool
	Flaws                   pgtype.Int4
}

func (q *Queries) GetRuneTransaction(ctx context.Context, hash string) (GetRuneTransactionRow, error) {
	row := q.db.QueryRow(ctx, getRuneTransaction, hash)
	var i GetRuneTransactionRow
	err := row.Scan(
		&i.Hash,
		&i.BlockHeight,
		&i.Index,
		&i.Timestamp,
		&i.Inputs,
		&i.Outputs,
		&i.Mints,
		&i.Burns,
		&i.RuneEtched,
		&i.TxHash,
		&i.BlockHeight_2,
		&i.Etching,
		&i.EtchingDivisibility,
		&i.EtchingPremine,
		&i.EtchingRune,
		&i.EtchingSpacers,
		&i.EtchingSymbol,
		&i.EtchingTerms,
		&i.EtchingTermsAmount,
		&i.EtchingTermsCap,
		&i.EtchingTermsHeightStart,
		&i.EtchingTermsHeightEnd,
		&i.EtchingTermsOffsetStart,
		&i.EtchingTermsOffsetEnd,
		&i.EtchingTurbo,
		&i.Edicts,
		&i.Mint,
		&i.Pointer,
		&i.Cenotaph,
		&i.Flaws,
	)
	return i, err
}

const getRuneTransactions = `-- name: GetRuneTransactions :many
SELECT hash, runes_transactions.block_height, index, timestamp, inputs, outputs, mints, burns, rune_etched, tx_hash, runes_runestones.block_height, etching, etching_divisibility, etching_premine, etching_rune, etching_spacers, etching_symbol, etching_terms, etching_terms_amount, etching_terms_cap, etching_terms_height_start, etching_terms_height_end, etching_terms_offset_start, etching_terms_offset_end, etching_turbo, edicts, mint, pointer, cenotaph, flaws FROM runes_transactions
	LEFT JOIN runes_runestones ON runes_transactions.hash = runes_runestones.tx_hash
	WHERE (
    $3::BOOLEAN = FALSE -- if @filter_pk_script is TRUE, apply pk_script filter
    OR runes_transactions.outputs @> $4::JSONB 
    OR runes_transactions.inputs @> $4::JSONB
  ) AND (
    $5::BOOLEAN = FALSE -- if @filter_rune_id is TRUE, apply rune_id filter
    OR runes_transactions.outputs @> $6::JSONB 
    OR runes_transactions.inputs @> $6::JSONB 
    OR runes_transactions.mints ? $7 
    OR runes_transactions.burns ? $7
    OR (runes_transactions.rune_etched = TRUE AND runes_transactions.block_height = $8 AND runes_transactions.index = $9)
  ) AND (
    $10 <= runes_transactions.block_height AND runes_transactions.block_height <= $11
  )
ORDER BY runes_transactions.block_height DESC, runes_transactions.index DESC LIMIT $1 OFFSET $2
`

type GetRuneTransactionsParams struct {
	Limit             int32
	Offset            int32
	FilterPkScript    bool
	PkScriptParam     []byte
	FilterRuneID      bool
	RuneIDParam       []byte
	RuneID            []byte
	RuneIDBlockHeight int32
	RuneIDTxIndex     int32
	FromBlock         int32
	ToBlock           int32
}

type GetRuneTransactionsRow struct {
	Hash                    string
	BlockHeight             int32
	Index                   int32
	Timestamp               pgtype.Timestamp
	Inputs                  []byte
	Outputs                 []byte
	Mints                   []byte
	Burns                   []byte
	RuneEtched              bool
	TxHash                  pgtype.Text
	BlockHeight_2           pgtype.Int4
	Etching                 pgtype.Bool
	EtchingDivisibility     pgtype.Int2
	EtchingPremine          pgtype.Numeric
	EtchingRune             pgtype.Text
	EtchingSpacers          pgtype.Int4
	EtchingSymbol           pgtype.Int4
	EtchingTerms            pgtype.Bool
	EtchingTermsAmount      pgtype.Numeric
	EtchingTermsCap         pgtype.Numeric
	EtchingTermsHeightStart pgtype.Int4
	EtchingTermsHeightEnd   pgtype.Int4
	EtchingTermsOffsetStart pgtype.Int4
	EtchingTermsOffsetEnd   pgtype.Int4
	EtchingTurbo            pgtype.Bool
	Edicts                  []byte
	Mint                    pgtype.Text
	Pointer                 pgtype.Int4
	Cenotaph                pgtype.Bool
	Flaws                   pgtype.Int4
}

func (q *Queries) GetRuneTransactions(ctx context.Context, arg GetRuneTransactionsParams) ([]GetRuneTransactionsRow, error) {
	rows, err := q.db.Query(ctx, getRuneTransactions,
		arg.Limit,
		arg.Offset,
		arg.FilterPkScript,
		arg.PkScriptParam,
		arg.FilterRuneID,
		arg.RuneIDParam,
		arg.RuneID,
		arg.RuneIDBlockHeight,
		arg.RuneIDTxIndex,
		arg.FromBlock,
		arg.ToBlock,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetRuneTransactionsRow
	for rows.Next() {
		var i GetRuneTransactionsRow
		if err := rows.Scan(
			&i.Hash,
			&i.BlockHeight,
			&i.Index,
			&i.Timestamp,
			&i.Inputs,
			&i.Outputs,
			&i.Mints,
			&i.Burns,
			&i.RuneEtched,
			&i.TxHash,
			&i.BlockHeight_2,
			&i.Etching,
			&i.EtchingDivisibility,
			&i.EtchingPremine,
			&i.EtchingRune,
			&i.EtchingSpacers,
			&i.EtchingSymbol,
			&i.EtchingTerms,
			&i.EtchingTermsAmount,
			&i.EtchingTermsCap,
			&i.EtchingTermsHeightStart,
			&i.EtchingTermsHeightEnd,
			&i.EtchingTermsOffsetStart,
			&i.EtchingTermsOffsetEnd,
			&i.EtchingTurbo,
			&i.Edicts,
			&i.Mint,
			&i.Pointer,
			&i.Cenotaph,
			&i.Flaws,
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

const getRunesUTXOsByPkScript = `-- name: GetRunesUTXOsByPkScript :many
SELECT tx_hash, tx_idx, max("pkscript") as pkscript, array_agg("rune_id") as rune_ids, array_agg("amount") as amounts 
  FROM runes_outpoint_balances 
  WHERE
    pkscript = $3 AND
    block_height <= $4 AND
    (spent_height IS NULL OR spent_height > $4)
  GROUP BY tx_hash, tx_idx
  ORDER BY tx_hash, tx_idx 
  LIMIT $1 OFFSET $2
`

type GetRunesUTXOsByPkScriptParams struct {
	Limit       int32
	Offset      int32
	Pkscript    string
	BlockHeight int32
}

type GetRunesUTXOsByPkScriptRow struct {
	TxHash   string
	TxIdx    int32
	Pkscript interface{}
	RuneIds  interface{}
	Amounts  interface{}
}

func (q *Queries) GetRunesUTXOsByPkScript(ctx context.Context, arg GetRunesUTXOsByPkScriptParams) ([]GetRunesUTXOsByPkScriptRow, error) {
	rows, err := q.db.Query(ctx, getRunesUTXOsByPkScript,
		arg.Limit,
		arg.Offset,
		arg.Pkscript,
		arg.BlockHeight,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetRunesUTXOsByPkScriptRow
	for rows.Next() {
		var i GetRunesUTXOsByPkScriptRow
		if err := rows.Scan(
			&i.TxHash,
			&i.TxIdx,
			&i.Pkscript,
			&i.RuneIds,
			&i.Amounts,
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

const getRunesUTXOsByRuneIdAndPkScript = `-- name: GetRunesUTXOsByRuneIdAndPkScript :many
SELECT tx_hash, tx_idx, max("pkscript") as pkscript, array_agg("rune_id") as rune_ids, array_agg("amount") as amounts 
  FROM runes_outpoint_balances 
  WHERE
    pkscript = $3 AND 
    block_height <= $4 AND 
    (spent_height IS NULL OR spent_height > $4)
  GROUP BY tx_hash, tx_idx
  HAVING array_agg("rune_id") @> $5::text[] 
  ORDER BY tx_hash, tx_idx 
  LIMIT $1 OFFSET $2
`

type GetRunesUTXOsByRuneIdAndPkScriptParams struct {
	Limit       int32
	Offset      int32
	Pkscript    string
	BlockHeight int32
	RuneIds     []string
}

type GetRunesUTXOsByRuneIdAndPkScriptRow struct {
	TxHash   string
	TxIdx    int32
	Pkscript interface{}
	RuneIds  interface{}
	Amounts  interface{}
}

func (q *Queries) GetRunesUTXOsByRuneIdAndPkScript(ctx context.Context, arg GetRunesUTXOsByRuneIdAndPkScriptParams) ([]GetRunesUTXOsByRuneIdAndPkScriptRow, error) {
	rows, err := q.db.Query(ctx, getRunesUTXOsByRuneIdAndPkScript,
		arg.Limit,
		arg.Offset,
		arg.Pkscript,
		arg.BlockHeight,
		arg.RuneIds,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetRunesUTXOsByRuneIdAndPkScriptRow
	for rows.Next() {
		var i GetRunesUTXOsByRuneIdAndPkScriptRow
		if err := rows.Scan(
			&i.TxHash,
			&i.TxIdx,
			&i.Pkscript,
			&i.RuneIds,
			&i.Amounts,
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

const getTotalHoldersByRuneIds = `-- name: GetTotalHoldersByRuneIds :many
WITH balances AS (
  SELECT DISTINCT ON (rune_id, pkscript) pkscript, block_height, rune_id, amount FROM runes_balances WHERE rune_id = ANY($1::TEXT[]) AND block_height <= $2 ORDER BY rune_id, pkscript, block_height DESC
)
SELECT rune_id, COUNT(DISTINCT pkscript) FROM balances WHERE amount > 0 GROUP BY rune_id
`

type GetTotalHoldersByRuneIdsParams struct {
	RuneIds     []string
	BlockHeight int32
}

type GetTotalHoldersByRuneIdsRow struct {
	RuneID string
	Count  int64
}

func (q *Queries) GetTotalHoldersByRuneIds(ctx context.Context, arg GetTotalHoldersByRuneIdsParams) ([]GetTotalHoldersByRuneIdsRow, error) {
	rows, err := q.db.Query(ctx, getTotalHoldersByRuneIds, arg.RuneIds, arg.BlockHeight)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetTotalHoldersByRuneIdsRow
	for rows.Next() {
		var i GetTotalHoldersByRuneIdsRow
		if err := rows.Scan(&i.RuneID, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const unspendOutPointBalancesSinceHeight = `-- name: UnspendOutPointBalancesSinceHeight :exec
UPDATE runes_outpoint_balances SET spent_height = NULL WHERE spent_height >= $1
`

func (q *Queries) UnspendOutPointBalancesSinceHeight(ctx context.Context, spentHeight pgtype.Int4) error {
	_, err := q.db.Exec(ctx, unspendOutPointBalancesSinceHeight, spentHeight)
	return err
}
