// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: data.sql

package gen

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const getBlockByHeight = `-- name: GetBlockByHeight :one
SELECT block_height, block_hash, version, merkle_root, prev_block_hash, timestamp, bits, nonce FROM bitcoin_blocks WHERE block_height = $1
`

func (q *Queries) GetBlockByHeight(ctx context.Context, blockHeight int32) (BitcoinBlock, error) {
	row := q.db.QueryRow(ctx, getBlockByHeight, blockHeight)
	var i BitcoinBlock
	err := row.Scan(
		&i.BlockHeight,
		&i.BlockHash,
		&i.Version,
		&i.MerkleRoot,
		&i.PrevBlockHash,
		&i.Timestamp,
		&i.Bits,
		&i.Nonce,
	)
	return i, err
}

const getBlocksByHeightRange = `-- name: GetBlocksByHeightRange :many
SELECT block_height, block_hash, version, merkle_root, prev_block_hash, timestamp, bits, nonce FROM bitcoin_blocks WHERE block_height >= $1 AND block_height <= $2 ORDER BY block_height ASC
`

type GetBlocksByHeightRangeParams struct {
	FromHeight int32
	ToHeight   int32
}

func (q *Queries) GetBlocksByHeightRange(ctx context.Context, arg GetBlocksByHeightRangeParams) ([]BitcoinBlock, error) {
	rows, err := q.db.Query(ctx, getBlocksByHeightRange, arg.FromHeight, arg.ToHeight)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []BitcoinBlock
	for rows.Next() {
		var i BitcoinBlock
		if err := rows.Scan(
			&i.BlockHeight,
			&i.BlockHash,
			&i.Version,
			&i.MerkleRoot,
			&i.PrevBlockHash,
			&i.Timestamp,
			&i.Bits,
			&i.Nonce,
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

const getLatestBlockHeader = `-- name: GetLatestBlockHeader :one
SELECT block_height, block_hash, version, merkle_root, prev_block_hash, timestamp, bits, nonce FROM bitcoin_blocks ORDER BY block_height DESC LIMIT 1
`

func (q *Queries) GetLatestBlockHeader(ctx context.Context) (BitcoinBlock, error) {
	row := q.db.QueryRow(ctx, getLatestBlockHeader)
	var i BitcoinBlock
	err := row.Scan(
		&i.BlockHeight,
		&i.BlockHash,
		&i.Version,
		&i.MerkleRoot,
		&i.PrevBlockHash,
		&i.Timestamp,
		&i.Bits,
		&i.Nonce,
	)
	return i, err
}

const getTransactionByHash = `-- name: GetTransactionByHash :one
SELECT tx_hash, version, locktime, block_height, block_hash, idx FROM bitcoin_transactions WHERE tx_hash = $1
`

func (q *Queries) GetTransactionByHash(ctx context.Context, txHash string) (BitcoinTransaction, error) {
	row := q.db.QueryRow(ctx, getTransactionByHash, txHash)
	var i BitcoinTransaction
	err := row.Scan(
		&i.TxHash,
		&i.Version,
		&i.Locktime,
		&i.BlockHeight,
		&i.BlockHash,
		&i.Idx,
	)
	return i, err
}

const getTransactionTxInsByTxHashes = `-- name: GetTransactionTxInsByTxHashes :many
SELECT tx_hash, tx_idx, prevout_tx_hash, prevout_tx_idx, prevout_pkscript, scriptsig, witness, sequence FROM bitcoin_transaction_txins WHERE tx_hash = ANY($1::TEXT[])
`

func (q *Queries) GetTransactionTxInsByTxHashes(ctx context.Context, txHashes []string) ([]BitcoinTransactionTxin, error) {
	rows, err := q.db.Query(ctx, getTransactionTxInsByTxHashes, txHashes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []BitcoinTransactionTxin
	for rows.Next() {
		var i BitcoinTransactionTxin
		if err := rows.Scan(
			&i.TxHash,
			&i.TxIdx,
			&i.PrevoutTxHash,
			&i.PrevoutTxIdx,
			&i.PrevoutPkscript,
			&i.Scriptsig,
			&i.Witness,
			&i.Sequence,
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

const getTransactionTxOutsByTxHashes = `-- name: GetTransactionTxOutsByTxHashes :many
SELECT tx_hash, tx_idx, pkscript, value, is_spent FROM bitcoin_transaction_txouts WHERE tx_hash = ANY($1::TEXT[])
`

func (q *Queries) GetTransactionTxOutsByTxHashes(ctx context.Context, txHashes []string) ([]BitcoinTransactionTxout, error) {
	rows, err := q.db.Query(ctx, getTransactionTxOutsByTxHashes, txHashes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []BitcoinTransactionTxout
	for rows.Next() {
		var i BitcoinTransactionTxout
		if err := rows.Scan(
			&i.TxHash,
			&i.TxIdx,
			&i.Pkscript,
			&i.Value,
			&i.IsSpent,
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

const getTransactionsByHeightRange = `-- name: GetTransactionsByHeightRange :many
SELECT tx_hash, version, locktime, block_height, block_hash, idx FROM bitcoin_transactions WHERE block_height >= $1 AND block_height <= $2
`

type GetTransactionsByHeightRangeParams struct {
	FromHeight int32
	ToHeight   int32
}

func (q *Queries) GetTransactionsByHeightRange(ctx context.Context, arg GetTransactionsByHeightRangeParams) ([]BitcoinTransaction, error) {
	rows, err := q.db.Query(ctx, getTransactionsByHeightRange, arg.FromHeight, arg.ToHeight)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []BitcoinTransaction
	for rows.Next() {
		var i BitcoinTransaction
		if err := rows.Scan(
			&i.TxHash,
			&i.Version,
			&i.Locktime,
			&i.BlockHeight,
			&i.BlockHash,
			&i.Idx,
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

const insertBlock = `-- name: InsertBlock :exec
INSERT INTO bitcoin_blocks ("block_height","block_hash","version","merkle_root","prev_block_hash","timestamp","bits","nonce") VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

type InsertBlockParams struct {
	BlockHeight   int32
	BlockHash     string
	Version       int32
	MerkleRoot    string
	PrevBlockHash string
	Timestamp     pgtype.Timestamptz
	Bits          int64
	Nonce         int64
}

func (q *Queries) InsertBlock(ctx context.Context, arg InsertBlockParams) error {
	_, err := q.db.Exec(ctx, insertBlock,
		arg.BlockHeight,
		arg.BlockHash,
		arg.Version,
		arg.MerkleRoot,
		arg.PrevBlockHash,
		arg.Timestamp,
		arg.Bits,
		arg.Nonce,
	)
	return err
}

const insertTransaction = `-- name: InsertTransaction :exec
INSERT INTO bitcoin_transactions ("tx_hash","version","locktime","block_height","block_hash","idx") VALUES ($1, $2, $3, $4, $5, $6)
`

type InsertTransactionParams struct {
	TxHash      string
	Version     int32
	Locktime    int64
	BlockHeight int32
	BlockHash   string
	Idx         int32
}

func (q *Queries) InsertTransaction(ctx context.Context, arg InsertTransactionParams) error {
	_, err := q.db.Exec(ctx, insertTransaction,
		arg.TxHash,
		arg.Version,
		arg.Locktime,
		arg.BlockHeight,
		arg.BlockHash,
		arg.Idx,
	)
	return err
}

const insertTransactionTxIn = `-- name: InsertTransactionTxIn :exec
WITH update_txout AS (
	UPDATE "bitcoin_transaction_txouts"
	SET "is_spent" = true
	WHERE "tx_hash" = $3 AND "tx_idx" = $4 AND "is_spent" = false -- TODO: should throw an error if already spent
	RETURNING "pkscript"
)
INSERT INTO bitcoin_transaction_txins ("tx_hash","tx_idx","prevout_tx_hash","prevout_tx_idx","prevout_pkscript","scriptsig","witness","sequence") 
VALUES ($1, $2, $3, $4, (SELECT "pkscript" FROM update_txout), $5, $6, $7)
`

type InsertTransactionTxInParams struct {
	TxHash        string
	TxIdx         int32
	PrevoutTxHash string
	PrevoutTxIdx  int32
	Scriptsig     string
	Witness       pgtype.Text
	Sequence      int64
}

func (q *Queries) InsertTransactionTxIn(ctx context.Context, arg InsertTransactionTxInParams) error {
	_, err := q.db.Exec(ctx, insertTransactionTxIn,
		arg.TxHash,
		arg.TxIdx,
		arg.PrevoutTxHash,
		arg.PrevoutTxIdx,
		arg.Scriptsig,
		arg.Witness,
		arg.Sequence,
	)
	return err
}

const insertTransactionTxOut = `-- name: InsertTransactionTxOut :exec
INSERT INTO bitcoin_transaction_txouts ("tx_hash","tx_idx","pkscript","value") VALUES ($1, $2, $3, $4)
`

type InsertTransactionTxOutParams struct {
	TxHash   string
	TxIdx    int32
	Pkscript string
	Value    int64
}

func (q *Queries) InsertTransactionTxOut(ctx context.Context, arg InsertTransactionTxOutParams) error {
	_, err := q.db.Exec(ctx, insertTransactionTxOut,
		arg.TxHash,
		arg.TxIdx,
		arg.Pkscript,
		arg.Value,
	)
	return err
}

const revertData = `-- name: RevertData :exec
WITH delete_tx AS (
	DELETE FROM "bitcoin_transactions" WHERE "block_height" >= $1
	RETURNING "tx_hash"
), delete_txin AS (
	DELETE FROM  "bitcoin_transaction_txins" WHERE "tx_hash" = ANY(SELECT "tx_hash" FROM delete_tx)
	RETURNING "prevout_tx_hash", "prevout_tx_idx"
), delete_txout AS (
	DELETE FROM  "bitcoin_transaction_txouts" WHERE "tx_hash" = ANY(SELECT "tx_hash" FROM delete_tx)
	RETURNING NULL
), revert_txout_spent AS (
	UPDATE "bitcoin_transaction_txouts"
	SET "is_spent" = false
	WHERE ("tx_hash", "tx_idx") IN (SELECT "prevout_tx_hash", "prevout_tx_idx" FROM delete_txin)
	RETURNING NULL
)
DELETE FROM "bitcoin_blocks" WHERE "bitcoin_blocks"."block_height" >= $1
`

func (q *Queries) RevertData(ctx context.Context, fromHeight int32) error {
	_, err := q.db.Exec(ctx, revertData, fromHeight)
	return err
}
