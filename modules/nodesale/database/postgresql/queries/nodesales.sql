-- name: AddNodesale :exec
INSERT INTO node_sales("block_height", "tx_index", "name", "starts_at", "ends_at", "tiers", "seller_public_key", "max_per_address", "deploy_tx_hash")
VALUES ( $1, $2, $3, $4, $5, $6, $7, $8, $9);