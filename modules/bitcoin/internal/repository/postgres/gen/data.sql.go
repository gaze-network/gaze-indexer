// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: data.sql

package gen

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

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
	Bits          int32
	Nonce         int32
}

// TODO: GetBlockHeaderByRange
// TODO: GetBlockByHeight/Hash (Join block with transactions, txins, txouts)
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
	Locktime    int32
	BlockHeight int32
	BlockHash   string
	Idx         int16
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
	WHERE "tx_hash" = $3 AND "tx_idx" = $4
	RETURNING "pkscript"
)
INSERT INTO bitcoin_transaction_txins ("tx_hash","tx_idx","prevout_tx_hash","prevout_tx_idx","prevout_pkscript","scriptsig","witness","sequence") 
VALUES ($1, $2, $3, $4, (SELECT "pkscript" FROM update_txout), $5, $6, $7)
`

type InsertTransactionTxInParams struct {
	TxHash        string
	TxIdx         int16
	PrevoutTxHash string
	PrevoutTxIdx  int16
	Scriptsig     string
	Witness       pgtype.Text
	Sequence      int32
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
	TxIdx    int16
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
