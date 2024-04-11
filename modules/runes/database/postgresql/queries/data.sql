-- name: GetBalancesByPkScript :many
SELECT DISTINCT ON (rune_id) * FROM runes_balances WHERE pkscript = $1 AND block_height <= $2 ORDER BY rune_id, block_height DESC;

-- name: GetBalancesByRuneId :many
SELECT DISTINCT ON (pkscript) * FROM runes_balances WHERE rune_id = $1 AND block_height <= $2 ORDER BY pkscript, block_height DESC;

-- name: GetOutPointBalances :many
SELECT * FROM runes_outpoint_balances WHERE tx_hash = $1 AND tx_idx = $2;

-- using FOR UPDATE to prevent other connections for updating the same rune entries if they are being accessed by Processor
-- name: GetRuneEntriesByRuneIds :many
SELECT * FROM runes_entries WHERE rune_id = ANY(@rune_ids::text[]) FOR UPDATE;

-- using FOR UPDATE to prevent other connections for updating the same rune entries if they are being accessed by Processor
-- name: GetRuneEntriesByRunes :many
SELECT * FROM runes_entries WHERE rune = ANY(@runes::text[]) FOR UPDATE;

-- name: SetRuneEntry :exec
INSERT INTO runes_entries (rune_id, rune, spacers, burned_amount, mints, premine, symbol, divisibility, terms, terms_amount, terms_cap, terms_height_start, terms_height_end, terms_offset_start, terms_offset_end, completion_time) 
  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) 
  ON CONFLICT (rune_id) DO UPDATE SET (burned_amount, mints, completion_time) = (excluded.burned_amount, excluded.mints, excluded.completion_time);

-- name: CreateRuneBalancesAtOutPoint :batchexec
INSERT INTO runes_outpoint_balances (rune_id, tx_hash, tx_idx, amount) VALUES ($1, $2, $3, $4);

-- name: CreateRuneBalanceAtBlock :batchexec
INSERT INTO runes_balances (pkscript, block_height, rune_id, amount) VALUES ($1, $2, $3, $4);

-- name: GetRunesProcessorState :one
SELECT * FROM runes_processor_state;

-- name: UpdateLatestBlock :exec
UPDATE runes_processor_state SET latest_block_height = $1, latest_block_hash = $2, latest_prev_block_hash = $3;

-- name: CreateIndexedBlock :exec
INSERT INTO runes_indexed_blocks (hash, height, event_hash, cumulative_event_hash) VALUES ($1, $2, $3, $4);

-- name: DeleteIndexedBlockByHash :exec
DELETE FROM runes_indexed_blocks WHERE hash = $1;
