// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
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

const createRuneEntry = `-- name: CreateRuneEntry :exec
INSERT INTO runes_entries (rune_id, rune, number, spacers, premine, symbol, divisibility, terms, terms_amount, terms_cap, terms_height_start, terms_height_end, terms_offset_start, terms_offset_end, turbo, etching_block, etching_tx_hash, etched_at)
  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
`

type CreateRuneEntryParams struct {
	RuneID           string
	Rune             string
	Number           int64
	Spacers          int32
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
	Turbo            bool
	EtchingBlock     int32
	EtchingTxHash    string
	EtchedAt         pgtype.Timestamp
}

func (q *Queries) CreateRuneEntry(ctx context.Context, arg CreateRuneEntryParams) error {
	_, err := q.db.Exec(ctx, createRuneEntry,
		arg.RuneID,
		arg.Rune,
		arg.Number,
		arg.Spacers,
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
		arg.Turbo,
		arg.EtchingBlock,
		arg.EtchingTxHash,
		arg.EtchedAt,
	)
	return err
}

const createRuneEntryState = `-- name: CreateRuneEntryState :exec
INSERT INTO runes_entry_states (rune_id, block_height, mints, burned_amount, completed_at, completed_at_height) VALUES ($1, $2, $3, $4, $5, $6)
`

type CreateRuneEntryStateParams struct {
	RuneID            string
	BlockHeight       int32
	Mints             pgtype.Numeric
	BurnedAmount      pgtype.Numeric
	CompletedAt       pgtype.Timestamp
	CompletedAtHeight pgtype.Int4
}

func (q *Queries) CreateRuneEntryState(ctx context.Context, arg CreateRuneEntryStateParams) error {
	_, err := q.db.Exec(ctx, createRuneEntryState,
		arg.RuneID,
		arg.BlockHeight,
		arg.Mints,
		arg.BurnedAmount,
		arg.CompletedAt,
		arg.CompletedAtHeight,
	)
	return err
}

const createRuneTransaction = `-- name: CreateRuneTransaction :exec
INSERT INTO runes_transactions (hash, block_height, index, timestamp, inputs, outputs, mints, burns, rune_etched) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

type CreateRuneTransactionParams struct {
	Hash        string
	BlockHeight int32
	Index       int32
	Timestamp   pgtype.Timestamp
	Inputs      []byte
	Outputs     []byte
	Mints       []byte
	Burns       []byte
	RuneEtched  bool
}

func (q *Queries) CreateRuneTransaction(ctx context.Context, arg CreateRuneTransactionParams) error {
	_, err := q.db.Exec(ctx, createRuneTransaction,
		arg.Hash,
		arg.BlockHeight,
		arg.Index,
		arg.Timestamp,
		arg.Inputs,
		arg.Outputs,
		arg.Mints,
		arg.Burns,
		arg.RuneEtched,
	)
	return err
}

const createRunestone = `-- name: CreateRunestone :exec
INSERT INTO runes_runestones (tx_hash, block_height, etching, etching_divisibility, etching_premine, etching_rune, etching_spacers, etching_symbol, etching_terms, etching_terms_amount, etching_terms_cap, etching_terms_height_start, etching_terms_height_end, etching_terms_offset_start, etching_terms_offset_end, etching_turbo, edicts, mint, pointer, cenotaph, flaws) 
  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
`

type CreateRunestoneParams struct {
	TxHash                  string
	BlockHeight             int32
	Etching                 bool
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
	Cenotaph                bool
	Flaws                   int32
}

func (q *Queries) CreateRunestone(ctx context.Context, arg CreateRunestoneParams) error {
	_, err := q.db.Exec(ctx, createRunestone,
		arg.TxHash,
		arg.BlockHeight,
		arg.Etching,
		arg.EtchingDivisibility,
		arg.EtchingPremine,
		arg.EtchingRune,
		arg.EtchingSpacers,
		arg.EtchingSymbol,
		arg.EtchingTerms,
		arg.EtchingTermsAmount,
		arg.EtchingTermsCap,
		arg.EtchingTermsHeightStart,
		arg.EtchingTermsHeightEnd,
		arg.EtchingTermsOffsetStart,
		arg.EtchingTermsOffsetEnd,
		arg.EtchingTurbo,
		arg.Edicts,
		arg.Mint,
		arg.Pointer,
		arg.Cenotaph,
		arg.Flaws,
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
SELECT pkscript, block_height, rune_id, amount FROM balances WHERE amount > 0
`

type GetBalancesByPkScriptParams struct {
	Pkscript    string
	BlockHeight int32
}

type GetBalancesByPkScriptRow struct {
	Pkscript    string
	BlockHeight int32
	RuneID      string
	Amount      pgtype.Numeric
}

func (q *Queries) GetBalancesByPkScript(ctx context.Context, arg GetBalancesByPkScriptParams) ([]GetBalancesByPkScriptRow, error) {
	rows, err := q.db.Query(ctx, getBalancesByPkScript, arg.Pkscript, arg.BlockHeight)
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
SELECT pkscript, block_height, rune_id, amount FROM balances WHERE amount > 0
`

type GetBalancesByRuneIdParams struct {
	RuneID      string
	BlockHeight int32
}

type GetBalancesByRuneIdRow struct {
	Pkscript    string
	BlockHeight int32
	RuneID      string
	Amount      pgtype.Numeric
}

func (q *Queries) GetBalancesByRuneId(ctx context.Context, arg GetBalancesByRuneIdParams) ([]GetBalancesByRuneIdRow, error) {
	rows, err := q.db.Query(ctx, getBalancesByRuneId, arg.RuneID, arg.BlockHeight)
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

const getRuneTransactions = `-- name: GetRuneTransactions :many
SELECT hash, runes_transactions.block_height, index, timestamp, inputs, outputs, mints, burns, rune_etched, tx_hash, runes_runestones.block_height, etching, etching_divisibility, etching_premine, etching_rune, etching_spacers, etching_symbol, etching_terms, etching_terms_amount, etching_terms_cap, etching_terms_height_start, etching_terms_height_end, etching_terms_offset_start, etching_terms_offset_end, etching_turbo, edicts, mint, pointer, cenotaph, flaws FROM runes_transactions
	LEFT JOIN runes_runestones ON runes_transactions.hash = runes_runestones.tx_hash
	WHERE (
    $1::BOOLEAN = FALSE -- if @filter_pk_script is TRUE, apply pk_script filter
    OR runes_transactions.outputs @> $2::JSONB 
    OR runes_transactions.inputs @> $2::JSONB
  ) AND (
    $3::BOOLEAN = FALSE -- if @filter_rune_id is TRUE, apply rune_id filter
    OR runes_transactions.outputs @> $4::JSONB 
    OR runes_transactions.inputs @> $4::JSONB 
    OR runes_transactions.mints ? $5 
    OR runes_transactions.burns ? $5
    OR (runes_transactions.rune_etched = TRUE AND runes_transactions.block_height = $6 AND runes_transactions.index = $7)
  ) AND (
    $8::INT = 0 OR runes_transactions.block_height = $8::INT -- if @block_height > 0, apply block_height filter
  )
`

type GetRuneTransactionsParams struct {
	FilterPkScript    bool
	PkScriptParam     []byte
	FilterRuneID      bool
	RuneIDParam       []byte
	RuneID            []byte
	RuneIDBlockHeight int32
	RuneIDTxIndex     int32
	BlockHeight       int32
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
		arg.FilterPkScript,
		arg.PkScriptParam,
		arg.FilterRuneID,
		arg.RuneIDParam,
		arg.RuneID,
		arg.RuneIDBlockHeight,
		arg.RuneIDTxIndex,
		arg.BlockHeight,
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

const getUnspentOutPointBalancesByPkScript = `-- name: GetUnspentOutPointBalancesByPkScript :many
SELECT rune_id, pkscript, tx_hash, tx_idx, amount, block_height, spent_height FROM runes_outpoint_balances WHERE pkscript = $1 AND block_height <= $2 AND (spent_height IS NULL OR spent_height > $2)
`

type GetUnspentOutPointBalancesByPkScriptParams struct {
	Pkscript    string
	BlockHeight int32
}

func (q *Queries) GetUnspentOutPointBalancesByPkScript(ctx context.Context, arg GetUnspentOutPointBalancesByPkScriptParams) ([]RunesOutpointBalance, error) {
	rows, err := q.db.Query(ctx, getUnspentOutPointBalancesByPkScript, arg.Pkscript, arg.BlockHeight)
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

const spendOutPointBalances = `-- name: SpendOutPointBalances :exec
UPDATE runes_outpoint_balances SET spent_height = $1 WHERE tx_hash = $2 AND tx_idx = $3
`

type SpendOutPointBalancesParams struct {
	SpentHeight pgtype.Int4
	TxHash      string
	TxIdx       int32
}

func (q *Queries) SpendOutPointBalances(ctx context.Context, arg SpendOutPointBalancesParams) error {
	_, err := q.db.Exec(ctx, spendOutPointBalances, arg.SpentHeight, arg.TxHash, arg.TxIdx)
	return err
}

const unspendOutPointBalancesSinceHeight = `-- name: UnspendOutPointBalancesSinceHeight :exec
UPDATE runes_outpoint_balances SET spent_height = NULL WHERE spent_height >= $1
`

func (q *Queries) UnspendOutPointBalancesSinceHeight(ctx context.Context, spentHeight pgtype.Int4) error {
	_, err := q.db.Exec(ctx, unspendOutPointBalancesSinceHeight, spentHeight)
	return err
}
