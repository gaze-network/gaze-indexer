-- name: RemoveEventsFromBlock :execrows
DELETE FROM events
WHERE "block_height" >= @from_block;
