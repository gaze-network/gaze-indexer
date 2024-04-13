-- name: GetBalancesByPkScript :many
SELECT DISTINCT ON (rune_id) * FROM runes_balances WHERE pkscript = $1 AND block_height <= $2 ORDER BY rune_id, block_height DESC;

-- name: GetBalancesByRuneId :many
SELECT DISTINCT ON (pkscript) * FROM runes_balances WHERE rune_id = $1 AND block_height <= $2 ORDER BY pkscript, block_height DESC;

-- name: GetBalanceByPkScriptAndRuneId :one
SELECT * FROM runes_balances WHERE pkscript = $1 AND rune_id = $2 AND block_height <= $3 ORDER BY block_height DESC LIMIT 1;

-- name: GetOutPointBalances :many
SELECT * FROM runes_outpoint_balances WHERE tx_hash = $1 AND tx_idx = $2;

-- name: GetRuneEntriesByRuneIds :many
WITH states AS (
  -- select latest state
  SELECT DISTINCT ON (rune_id) * FROM runes_entry_states WHERE rune_id = ANY(@rune_ids::text[]) ORDER BY rune_id, block_height DESC
)
SELECT * FROM runes_entries
  LEFT JOIN states ON runes_entries.rune_id = states.rune_id
  WHERE rune_id = ANY(@rune_ids::text[]);

-- name: GetRuneIdFromRune :one
SELECT rune_id FROM runes_entries WHERE rune = $1;

-- name: GetRuneTransactionsByHeight :many
SELECT * FROM runes_transactions 
  LEFT JOIN runes_runestones ON runes_transactions.hash = runes_runestones.tx_hash
  WHERE runes_transactions.block_height = $1;

-- name: CreateRuneEntry :exec
INSERT INTO runes_entries (rune_id, rune, spacers, premine, symbol, divisibility, terms, terms_amount, terms_cap, terms_height_start, terms_height_end, terms_offset_start, terms_offset_end, turbo, etching_block)
  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15);

-- name: CreateRuneEntryState :exec
INSERT INTO runes_entry_states (rune_id, block_height, mints, burned_amount, completed_at, completed_at_height) VALUES ($1, $2, $3, $4, $5, $6);

-- name: CreateRuneTransaction :exec
INSERT INTO runes_transactions (hash, block_height, timestamp, inputs, outputs, mints, burns) VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: CreateRunestone :exec
INSERT INTO runes_runestones (tx_hash, block_height, etching, etching_divisibility, etching_premine, etching_rune, etching_spacers, etching_symbol, etching_terms, etching_terms_amount, etching_terms_cap, etching_terms_height_start, etching_terms_height_end, etching_terms_offset_start, etching_terms_offset_end, etching_turbo, edicts, mint, pointer, cenotaph, flaws) 
  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21);

-- name: CreateOutPointBalances :batchexec
INSERT INTO runes_outpoint_balances (rune_id, tx_hash, tx_idx, amount, block_height, spent_height) VALUES ($1, $2, $3, $4, $5, $6);

-- name: SpendOutPointBalances :exec
UPDATE runes_outpoint_balances SET spent_height = $1 WHERE rune_id = $2 AND tx_hash = $3 AND tx_idx = $4;

-- name: CreateRuneBalanceAtBlock :batchexec
INSERT INTO runes_balances (pkscript, block_height, rune_id, amount) VALUES ($1, $2, $3, $4);

-- name: GetLatestIndexedBlock :one
SELECT * FROM runes_indexed_blocks ORDER BY height DESC LIMIT 1;

-- name: GetIndexedBlockByHeight :one
SELECT * FROM runes_indexed_blocks WHERE height = $1;

-- name: CreateIndexedBlock :exec
INSERT INTO runes_indexed_blocks (hash, height, prev_hash, event_hash, cumulative_event_hash) VALUES ($1, $2, $3, $4, $5);

-- name: DeleteIndexedBlockSinceHeight :exec
DELETE FROM runes_indexed_blocks WHERE height >= $1;

-- name: DeleteRuneEntriesSinceHeight :exec
DELETE FROM runes_entries WHERE etching_block >= $1;

-- name: DeleteRuneEntryStatesSinceHeight :exec
DELETE FROM runes_entry_states WHERE block_height >= $1;

-- name: DeleteRuneTransactionsSinceHeight :exec
DELETE FROM runes_transactions WHERE block_height >= $1;

-- name: DeleteRunestonesSinceHeight :exec
DELETE FROM runes_runestones WHERE block_height >= $1;

-- name: DeleteOutPointBalancesSinceHeight :exec
DELETE FROM runes_outpoint_balances WHERE block_height >= $1;

-- name: UnspendOutPointBalancesSinceHeight :exec
UPDATE runes_outpoint_balances SET spent_height = NULL WHERE spent_height >= $1;

-- name: DeleteRuneBalancesSinceHeight :exec
DELETE FROM runes_balances WHERE block_height >= $1;
