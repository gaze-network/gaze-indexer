// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: blocks.sql

package gen

import (
	"context"
)

const addBlock = `-- name: AddBlock :exec
INSERT INTO blocks("block_height", "block_hash", "module")
VALUES ($1, $2, $3)
`

type AddBlockParams struct {
	BlockHeight int64
	BlockHash   string
	Module      string
}

func (q *Queries) AddBlock(ctx context.Context, arg AddBlockParams) error {
	_, err := q.db.Exec(ctx, addBlock, arg.BlockHeight, arg.BlockHash, arg.Module)
	return err
}

const getBlock = `-- name: GetBlock :one
SELECT block_height, block_hash, module FROM blocks
WHERE "block_height" = $1
`

func (q *Queries) GetBlock(ctx context.Context, blockHeight int64) (Block, error) {
	row := q.db.QueryRow(ctx, getBlock, blockHeight)
	var i Block
	err := row.Scan(&i.BlockHeight, &i.BlockHash, &i.Module)
	return i, err
}

const getLastProcessedBlock = `-- name: GetLastProcessedBlock :one
SELECT block_height, block_hash, module FROM blocks
WHERE "block_height" = (SELECT MAX("block_height") FROM blocks)
`

func (q *Queries) GetLastProcessedBlock(ctx context.Context) (Block, error) {
	row := q.db.QueryRow(ctx, getLastProcessedBlock)
	var i Block
	err := row.Scan(&i.BlockHeight, &i.BlockHash, &i.Module)
	return i, err
}

const removeBlockFrom = `-- name: RemoveBlockFrom :execrows
DELETE FROM blocks
WHERE "block_height" >= $1
`

func (q *Queries) RemoveBlockFrom(ctx context.Context, fromBlock int64) (int64, error) {
	result, err := q.db.Exec(ctx, removeBlockFrom, fromBlock)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
