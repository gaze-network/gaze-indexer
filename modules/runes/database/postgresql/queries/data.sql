-- name: GetBalancesByPkScript :many
SELECT DISTINCT ON (rune_id) * FROM runes_balances WHERE pkscript = $1 AND block_height <= $2 ORDER BY rune_id, block_height DESC;

-- name: GetBalancesByRuneId :many
SELECT DISTINCT ON (pkscript) * FROM runes_balances WHERE rune_id = $1 AND block_height <= $2 ORDER BY pkscript, block_height DESC;

-- name: GetOutPointBalances :many
SELECT * FROM runes_outpoint_balances WHERE rune_id = $1 AND tx_hash = $2 AND tx_idx = $3;

-- name: GetRuneEntriesByRuneIds :many
SELECT * FROM runes_entries WHERE rune_id = ANY(@rune_ids::text[]);

-- name: GetRuneEntriesByRunes :many
SELECT * FROM runes_entries WHERE rune = ANY(@rune_ids::text[]);
