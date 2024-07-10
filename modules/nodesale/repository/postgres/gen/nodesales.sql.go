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
INSERT INTO node_sales("block_height", "tx_index", "name", "starts_at", "ends_at", "tiers", "seller_public_key", "max_per_address", "deploy_tx_hash", "max_discount_percentage", "seller_wallet")
VALUES ( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
`

type AddNodesaleParams struct {
	BlockHeight           int32
	TxIndex               int32
	Name                  string
	StartsAt              pgtype.Timestamp
	EndsAt                pgtype.Timestamp
	Tiers                 [][]byte
	SellerPublicKey       string
	MaxPerAddress         int32
	DeployTxHash          string
	MaxDiscountPercentage int32
	SellerWallet          string
}

func (q *Queries) AddNodesale(ctx context.Context, arg AddNodesaleParams) error {
	_, err := q.db.Exec(ctx, addNodesale,
		arg.BlockHeight,
		arg.TxIndex,
		arg.Name,
		arg.StartsAt,
		arg.EndsAt,
		arg.Tiers,
		arg.SellerPublicKey,
		arg.MaxPerAddress,
		arg.DeployTxHash,
		arg.MaxDiscountPercentage,
		arg.SellerWallet,
	)
	return err
}

const getNodesale = `-- name: GetNodesale :many
SELECT block_height, tx_index, name, starts_at, ends_at, tiers, seller_public_key, max_per_address, deploy_tx_hash, max_discount_percentage, seller_wallet
FROM node_sales
WHERE block_height = $1 AND
    tx_index = $2
`

type GetNodesaleParams struct {
	BlockHeight int32
	TxIndex     int32
}

func (q *Queries) GetNodesale(ctx context.Context, arg GetNodesaleParams) ([]NodeSale, error) {
	rows, err := q.db.Query(ctx, getNodesale, arg.BlockHeight, arg.TxIndex)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []NodeSale
	for rows.Next() {
		var i NodeSale
		if err := rows.Scan(
			&i.BlockHeight,
			&i.TxIndex,
			&i.Name,
			&i.StartsAt,
			&i.EndsAt,
			&i.Tiers,
			&i.SellerPublicKey,
			&i.MaxPerAddress,
			&i.DeployTxHash,
			&i.MaxDiscountPercentage,
			&i.SellerWallet,
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
