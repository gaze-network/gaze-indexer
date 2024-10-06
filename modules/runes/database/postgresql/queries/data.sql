-- name: GetBalancesByPkScript :many
WITH balances AS (
  SELECT DISTINCT ON (rune_id) * FROM runes_balances WHERE pkscript = $1 AND block_height <= $2 ORDER BY rune_id, block_height DESC
)
SELECT * FROM balances WHERE amount > 0 ORDER BY amount DESC, rune_id LIMIT $3 OFFSET $4;

-- name: GetBalancesByRuneId :many
WITH balances AS (
  SELECT DISTINCT ON (pkscript) * FROM runes_balances WHERE rune_id = $1 AND block_height <= $2 ORDER BY pkscript, block_height DESC
)
SELECT * FROM balances WHERE amount > 0 ORDER BY amount DESC, pkscript LIMIT $3 OFFSET $4;

-- name: GetBalanceByPkScriptAndRuneId :one
SELECT * FROM runes_balances WHERE pkscript = $1 AND rune_id = $2 AND block_height <= $3 ORDER BY block_height DESC LIMIT 1;

-- name: GetTotalHoldersByRuneIds :many
SELECT rune_id, COUNT(pkscript) FROM runes_balances WHERE rune_id = ANY(@rune_ids::TEXT[]) AND amount > 0 GROUP BY rune_id;

-- name: GetOutPointBalancesAtOutPoint :many
SELECT * FROM runes_outpoint_balances WHERE tx_hash = $1 AND tx_idx = $2;

-- name: GetRunesUTXOsByPkScript :many
SELECT tx_hash, tx_idx, max("pkscript") as pkscript, array_agg("rune_id") as rune_ids, array_agg("amount") as amounts 
  FROM runes_outpoint_balances 
  WHERE
    pkscript = @pkScript AND
    block_height <= @block_height AND
    (spent_height IS NULL OR spent_height > @block_height)
  GROUP BY tx_hash, tx_idx
  ORDER BY tx_hash, tx_idx 
  LIMIT $1 OFFSET $2;

-- name: GetRunesUTXOsByRuneIdAndPkScript :many
SELECT tx_hash, tx_idx, max("pkscript") as pkscript, array_agg("rune_id") as rune_ids, array_agg("amount") as amounts 
  FROM runes_outpoint_balances 
  WHERE
    pkscript = @pkScript AND 
    block_height <= @block_height AND 
    (spent_height IS NULL OR spent_height > @block_height)
  GROUP BY tx_hash, tx_idx
  HAVING array_agg("rune_id") @> @rune_ids::text[] 
  ORDER BY tx_hash, tx_idx 
  LIMIT $1 OFFSET $2;

-- name: GetRuneEntriesByRuneIds :many
WITH states AS (
  -- select latest state
  SELECT DISTINCT ON (rune_id) * FROM runes_entry_states WHERE rune_id = ANY(@rune_ids::text[]) ORDER BY rune_id, block_height DESC
)
SELECT * FROM runes_entries
  LEFT JOIN states ON runes_entries.rune_id = states.rune_id
  WHERE runes_entries.rune_id = ANY(@rune_ids::text[]);

-- name: GetRuneEntriesByRuneIdsAndHeight :many
WITH states AS (
  -- select latest state
  SELECT DISTINCT ON (rune_id) * FROM runes_entry_states WHERE rune_id = ANY(@rune_ids::text[]) AND block_height <= @height ORDER BY rune_id, block_height DESC
)
SELECT * FROM runes_entries
  LEFT JOIN states ON runes_entries.rune_id = states.rune_id
  WHERE runes_entries.rune_id = ANY(@rune_ids::text[]) AND etching_block <= @height;

-- name: GetRuneEntryList :many
SELECT DISTINCT ON (number) * FROM runes_entries 
LEFT JOIN runes_entry_states states ON runes_entries.rune_id = states.rune_id
ORDER BY number LIMIT @_limit OFFSET @_offset;

-- name: GetRuneIdFromRune :one
SELECT rune_id FROM runes_entries WHERE rune = $1;

-- name: GetRuneTransactions :many
SELECT * FROM runes_transactions
	LEFT JOIN runes_runestones ON runes_transactions.hash = runes_runestones.tx_hash
	WHERE (
    @filter_pk_script::BOOLEAN = FALSE -- if @filter_pk_script is TRUE, apply pk_script filter
    OR runes_transactions.outputs @> @pk_script_param::JSONB 
    OR runes_transactions.inputs @> @pk_script_param::JSONB
  ) AND (
    @filter_rune_id::BOOLEAN = FALSE -- if @filter_rune_id is TRUE, apply rune_id filter
    OR runes_transactions.outputs @> @rune_id_param::JSONB 
    OR runes_transactions.inputs @> @rune_id_param::JSONB 
    OR runes_transactions.mints ? @rune_id 
    OR runes_transactions.burns ? @rune_id
    OR (runes_transactions.rune_etched = TRUE AND runes_transactions.block_height = @rune_id_block_height AND runes_transactions.index = @rune_id_tx_index)
  ) AND (
    @from_block <= runes_transactions.block_height AND runes_transactions.block_height <= @to_block
  )
ORDER BY runes_transactions.block_height DESC, runes_transactions.index DESC LIMIT $1 OFFSET $2;

-- name: GetRuneTransaction :one
SELECT * FROM runes_transactions
  LEFT JOIN runes_runestones ON runes_transactions.hash = runes_runestones.tx_hash
  WHERE hash = $1 LIMIT 1;

-- name: CountRuneEntries :one
SELECT COUNT(*) FROM runes_entries;

-- name: CreateRuneEntry :exec
INSERT INTO runes_entries (rune_id, rune, number, spacers, premine, symbol, divisibility, terms, terms_amount, terms_cap, terms_height_start, terms_height_end, terms_offset_start, terms_offset_end, turbo, etching_block, etching_tx_hash, etched_at)
  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18);

-- name: CreateRuneEntryState :exec
INSERT INTO runes_entry_states (rune_id, block_height, mints, burned_amount, completed_at, completed_at_height) VALUES ($1, $2, $3, $4, $5, $6);

-- name: CreateRuneTransaction :exec
INSERT INTO runes_transactions (hash, block_height, index, timestamp, inputs, outputs, mints, burns, rune_etched) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: CreateRunestone :exec
INSERT INTO runes_runestones (tx_hash, block_height, etching, etching_divisibility, etching_premine, etching_rune, etching_spacers, etching_symbol, etching_terms, etching_terms_amount, etching_terms_cap, etching_terms_height_start, etching_terms_height_end, etching_terms_offset_start, etching_terms_offset_end, etching_turbo, edicts, mint, pointer, cenotaph, flaws) 
  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21);

-- name: CreateOutPointBalances :batchexec
INSERT INTO runes_outpoint_balances (rune_id, pkscript, tx_hash, tx_idx, amount, block_height, spent_height) VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: SpendOutPointBalances :exec
UPDATE runes_outpoint_balances SET spent_height = $1 WHERE tx_hash = $2 AND tx_idx = $3;

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
