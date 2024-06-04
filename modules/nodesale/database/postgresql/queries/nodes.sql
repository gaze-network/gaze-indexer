-- name: ClearDelegate :execrows
UPDATE nodes
SET "delegated_to" = NULL
WHERE "delegate_tx_hash" = NULL;

-- name: SetDelegates :execrows
UPDATE nodes
SET delegated_to = @delegatee
WHERE sale_block = $1 AND
    sale_tx_index = $2 AND
    node_id = ANY (@node_ids::int[]);

-- name: GetNodes :many
SELECT *
FROM nodes
WHERE sale_block = $1 AND
    sale_tx_index = $2 AND
    node_id = ANY (@node_ids::int[]);


-- name: GetNodesByOwner :many
SELECT * 
FROM nodes
WHERE sale_block = $1 AND
    sale_tx_index = $2 AND
    owner_public_key = $3
ORDER BY tier_index;

-- name: AddNode :exec
INSERT INTO nodes(sale_block, sale_tx_index, node_id, tier_index, delegated_to, owner_public_key, purchase_tx_hash, delegate_tx_hash)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);