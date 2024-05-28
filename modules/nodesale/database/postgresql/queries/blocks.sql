-- name: GetLastProcessedBlock :one
SELECT * FROM blocks
WHERE "block_height" = (SELECT MAX("block_height") FROM blocks);


-- name: GetBlock :one
SELECT * FROM blocks
WHERE "block_height" = $1;

-- name: RemoveBlockFrom :execrows
DELETE FROM blocks
WHERE "block_height" >= @from_block;