-- name: AddNodesale :exec
INSERT INTO node_sales ("block_height", "tx_index", "name", "starts_at", "ends_at", "tiers", "seller_public_key", "max_per_address", "deploy_tx_hash", "max_discount_percentage", "seller_wallet")
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: GetNodesale :many
SELECT *
FROM node_sales
WHERE block_height = $1 AND
    tx_index = $2;