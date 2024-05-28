-- name: ClearDelegate :execrows
UPDATE nodes
SET "delegated_to" = NULL
WHERE "delegate_tx_hash" = NULL;