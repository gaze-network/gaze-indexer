-- name: GetBalancesByPkScript :many
SELECT DISTINCT ON (rune_id) * FROM runes_balances WHERE pkscript = $1 AND block_height <= $2 ORDER BY rune_id, block_height DESC;

-- name: GetBalancesByRuneId :many
SELECT DISTINCT ON (pkscript) * FROM runes_balances WHERE rune_id = $1 AND block_height <= $2 ORDER BY pkscript, block_height DESC;

-- name: GetOutPointBalances :many
SELECT * FROM runes_outpoint_balances WHERE tx_hash = $1 AND tx_idx = $2;

-- name: GetRuneEntryByRuneId :one
SELECT * FROM runes_entries WHERE rune_id = $1;

-- name: GetRuneEntryByRune :one
SELECT * FROM runes_entries WHERE rune = $1;
