// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: nodesales.sql

package gen

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const addNodesale = `-- name: AddNodesale :exec
INSERT INTO node_sales("block_height", "tx_index", "starts_at", "ends_at", "tiers", "seller_public_key", "max_per_address", "deploy_tx_hash")
VALUES ( $1, $2, $3, $4, $5, $6, $7, $8)
`

type AddNodesaleParams struct {
	BlockHeight     int32
	TxIndex         int32
	StartsAt        pgtype.Timestamp
	EndsAt          pgtype.Timestamp
	Tiers           [][]byte
	SellerPublicKey string
	MaxPerAddress   int32
	DeployTxHash    string
}

func (q *Queries) AddNodesale(ctx context.Context, arg AddNodesaleParams) error {
	_, err := q.db.Exec(ctx, addNodesale,
		arg.BlockHeight,
		arg.TxIndex,
		arg.StartsAt,
		arg.EndsAt,
		arg.Tiers,
		arg.SellerPublicKey,
		arg.MaxPerAddress,
		arg.DeployTxHash,
	)
	return err
}
