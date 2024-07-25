-- name: GetLastProcessedBlock :one
SELECT * FROM blocks ORDER BY block_height DESC LIMIT 1;  


-- name: GetBlock :one
SELECT * FROM blocks
WHERE "block_height" = $1;

-- name: RemoveBlockFrom :execrows
DELETE FROM blocks
WHERE "block_height" >= @from_block;

-- name: CreateBlock :exec
INSERT INTO blocks ("block_height", "block_hash", "module")
VALUES ($1, $2, $3);