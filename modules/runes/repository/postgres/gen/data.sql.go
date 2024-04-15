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
INSERT INTO runes_entries (rune_id, rune, spacers, premine, symbol, divisibility, terms, terms_amount, terms_cap, terms_height_start, terms_height_end, terms_offset_start, terms_offset_end, turbo, etching_block)
  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
`

type CreateRuneEntryParams struct {
	RuneID           string
	Rune             string
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
}

func (q *Queries) CreateRuneEntry(ctx context.Context, arg CreateRuneEntryParams) error {
	_, err := q.db.Exec(ctx, createRuneEntry,
		arg.RuneID,
		arg.Rune,
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
INSERT INTO runes_transactions (hash, block_height, index, timestamp, inputs, outputs, mints, burns) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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

const getIndexedBlockByHeight = `-- name: GetIndexedBlockByHeight :one
SELECT hash, height, prev_hash, event_hash, cumulative_event_hash FROM runes_indexed_blocks WHERE height = $1
`

func (q *Queries) GetIndexedBlockByHeight(ctx context.Context, height int32) (RunesIndexedBlock, error) {
	row := q.db.QueryRow(ctx, getIndexedBlockByHeight, height)
	var i RunesIndexedBlock
	err := row.Scan(
		&i.Hash,
		&i.Height,
		&i.PrevHash,
		&i.EventHash,
		&i.CumulativeEventHash,
	)
	return i, err
}

const getLatestIndexedBlock = `-- name: GetLatestIndexedBlock :one
SELECT hash, height, prev_hash, event_hash, cumulative_event_hash FROM runes_indexed_blocks ORDER BY height DESC LIMIT 1
`

func (q *Queries) GetLatestIndexedBlock(ctx context.Context) (RunesIndexedBlock, error) {
	row := q.db.QueryRow(ctx, getLatestIndexedBlock)
	var i RunesIndexedBlock
	err := row.Scan(
		&i.Hash,
		&i.Height,
		&i.PrevHash,
		&i.EventHash,
		&i.CumulativeEventHash,
	)
	return i, err
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
WITH states AS (
  -- select latest state
  SELECT DISTINCT ON (rune_id) rune_id, block_height, mints, burned_amount, completed_at, completed_at_height FROM runes_entry_states WHERE rune_id = ANY($1::text[]) ORDER BY rune_id, block_height DESC
)
SELECT runes_entries.rune_id, rune, spacers, premine, symbol, divisibility, terms, terms_amount, terms_cap, terms_height_start, terms_height_end, terms_offset_start, terms_offset_end, turbo, etching_block, states.rune_id, block_height, mints, burned_amount, completed_at, completed_at_height FROM runes_entries
  LEFT JOIN states ON runes_entries.rune_id = states.rune_id
  WHERE rune_id = ANY($1::text[])
`

type GetRuneEntriesByRuneIdsRow struct {
	RuneID            string
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

const getRuneTransactionsByHeight = `-- name: GetRuneTransactionsByHeight :many
SELECT hash, runes_transactions.block_height, index, timestamp, inputs, outputs, mints, burns, tx_hash, runes_runestones.block_height, etching, etching_divisibility, etching_premine, etching_rune, etching_spacers, etching_symbol, etching_terms, etching_terms_amount, etching_terms_cap, etching_terms_height_start, etching_terms_height_end, etching_terms_offset_start, etching_terms_offset_end, etching_turbo, edicts, mint, pointer, cenotaph, flaws FROM runes_transactions 
  LEFT JOIN runes_runestones ON runes_transactions.hash = runes_runestones.tx_hash
  WHERE runes_transactions.block_height = $1
`

type GetRuneTransactionsByHeightRow struct {
	Hash                    string
	BlockHeight             int32
	Index                   int32
	Timestamp               pgtype.Timestamp
	Inputs                  []byte
	Outputs                 []byte
	Mints                   []byte
	Burns                   []byte
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

func (q *Queries) GetRuneTransactionsByHeight(ctx context.Context, blockHeight int32) ([]GetRuneTransactionsByHeightRow, error) {
	rows, err := q.db.Query(ctx, getRuneTransactionsByHeight, blockHeight)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetRuneTransactionsByHeightRow
	for rows.Next() {
		var i GetRuneTransactionsByHeightRow
		if err := rows.Scan(
			&i.Hash,
			&i.BlockHeight,
			&i.Index,
			&i.Timestamp,
			&i.Inputs,
			&i.Outputs,
			&i.Mints,
			&i.Burns,
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

const spendOutPointBalances = `-- name: SpendOutPointBalances :exec
UPDATE runes_outpoint_balances SET spent_height = $1 WHERE rune_id = $2 AND tx_hash = $3 AND tx_idx = $4
`

type SpendOutPointBalancesParams struct {
	SpentHeight pgtype.Int4
	RuneID      string
	TxHash      string
	TxIdx       int32
}

func (q *Queries) SpendOutPointBalances(ctx context.Context, arg SpendOutPointBalancesParams) error {
	_, err := q.db.Exec(ctx, spendOutPointBalances,
		arg.SpentHeight,
		arg.RuneID,
		arg.TxHash,
		arg.TxIdx,
	)
	return err
}

const unspendOutPointBalancesSinceHeight = `-- name: UnspendOutPointBalancesSinceHeight :exec
UPDATE runes_outpoint_balances SET spent_height = NULL WHERE spent_height >= $1
`

func (q *Queries) UnspendOutPointBalancesSinceHeight(ctx context.Context, spentHeight pgtype.Int4) error {
	_, err := q.db.Exec(ctx, unspendOutPointBalancesSinceHeight, spentHeight)
	return err
}
