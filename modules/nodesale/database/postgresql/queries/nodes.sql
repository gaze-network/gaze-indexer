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