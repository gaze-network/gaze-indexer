-- name: GetLastProcessedBlock :one
SELECT * FROM blocks
WHERE "block_height" = (SELECT MAX("block_height") FROM blocks);


-- name: GetBlock :one
SELECT * FROM blocks
WHERE "block_height" = $1;

-- name: RemoveBlockFrom :execrows
DELETE FROM blocks
WHERE "block_height" >= @from_block;

-- name: AddBlock :exec
INSERT INTO blocks ("block_height", "block_hash", "module")
VALUES ($1, $2, $3);