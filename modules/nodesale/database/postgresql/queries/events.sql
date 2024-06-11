-- name: RemoveEventsFromBlock :execrows
DELETE FROM events
WHERE "block_height" >= @from_block;

-- name: AddEvent :exec
INSERT INTO events("tx_hash", "block_height", "tx_index", "wallet_address", "valid", "action", 
                    "raw_message", "parsed_message", "block_timestamp", "block_hash", "metadata")
VALUES ( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: GetEventsByWallet :many
SELECT *
FROM events
WHERE wallet_address = $1;