-- name: BatchCreateRunesBalances :exec
INSERT INTO runes_balances ("pkscript", "block_height", "rune_id", "amount")
VALUES(
  unnest(@pkscript_arr::TEXT[]),
  unnest(@block_height_arr::INT[]),
  unnest(@rune_id_arr::TEXT[]),
  unnest(@amount_arr::DECIMAL[])
);

-- name: BatchCreateRuneEntries :exec
INSERT INTO runes_entries ("rune_id", "rune", "number", "spacers", "premine", "symbol", "divisibility", "terms", "terms_amount", "terms_cap", "terms_height_start", "terms_height_end", "terms_offset_start", "terms_offset_end", "turbo", "etching_block", "etching_tx_hash", "etched_at")
VALUES(
  unnest(@rune_id_arr::TEXT[]),
  unnest(@rune_arr::TEXT[]),
  unnest(@number_arr::BIGINT[]),
  unnest(@spacers_arr::INT[]),
  unnest(@premine_arr::DECIMAL[]),
  unnest(@symbol_arr::INT[]),
  unnest(@divisibility_arr::SMALLINT[]),
  unnest(@terms_arr::BOOLEAN[]),
  unnest(@terms_amount_arr), -- (DECIMAL[]) don't cast to allow null values
  unnest(@terms_cap_arr), -- (DECIMAL[]) don't cast to allow null values
  unnest(@terms_height_start_arr), -- (INT[]) don't cast to allow null values
  unnest(@terms_height_end_arr), -- (INT[]) don't cast to allow null values
  unnest(@terms_offset_start_arr), -- (INT[]) don't cast to allow null values
  unnest(@terms_offset_end_arr), -- (INT[]) don't cast to allow null values
  unnest(@turbo_arr::BOOLEAN[]),
  unnest(@etching_block_arr::INT[]),
  unnest(@etching_tx_hash_arr::TEXT[]),
  unnest(@etched_at_arr::TIMESTAMP[])
);

-- name: BatchCreateRuneEntryStates :exec
INSERT INTO runes_entry_states ("rune_id", "block_height", "mints", "burned_amount", "completed_at", "completed_at_height")
VALUES(
  unnest(@rune_id_arr::TEXT[]),
  unnest(@block_height_arr::INT[]),
  unnest(@mints_arr::DECIMAL[]),
  unnest(@burned_amount_arr::DECIMAL[]),
  unnest(@completed_at_arr::TIMESTAMP[]),
  unnest(@completed_at_height_arr) -- (INT[]) don't cast to allow null values
);

-- name: BatchCreateRunesOutpointBalances :exec
INSERT INTO runes_outpoint_balances ("rune_id", "pkscript", "tx_hash", "tx_idx", "amount", "block_height", "spent_height")
VALUES(
  unnest(@rune_id_arr::TEXT[]),
  unnest(@pkscript_arr::TEXT[]),
  unnest(@tx_hash_arr::TEXT[]),
  unnest(@tx_idx_arr::INT[]),
  unnest(@amount_arr::DECIMAL[]),
  unnest(@block_height_arr::INT[]),
  unnest(@spent_height_arr) -- (INT[]) don't cast to allow null values
);

-- name: BatchSpendOutpointBalances :exec
UPDATE runes_outpoint_balances
	SET "spent_height" = @spent_height::INT
	FROM (
    SELECT 
      unnest(@tx_hash_arr::TEXT[]) AS tx_hash, 
      unnest(@tx_idx_arr::INT[]) AS tx_idx
    ) AS input
	WHERE "runes_outpoint_balances"."tx_hash" = "input"."tx_hash" AND "runes_outpoint_balances"."tx_idx" = "input"."tx_idx";

-- name: BatchCreateRunestones :exec
INSERT INTO runes_runestones ("tx_hash", "block_height", "etching", "etching_divisibility", "etching_premine", "etching_rune", "etching_spacers", "etching_symbol", "etching_terms", "etching_terms_amount", "etching_terms_cap", "etching_terms_height_start", "etching_terms_height_end", "etching_terms_offset_start", "etching_terms_offset_end", "etching_turbo", "edicts", "mint", "pointer", "cenotaph", "flaws")
VALUES(
  unnest(@tx_hash_arr::TEXT[]),
  unnest(@block_height_arr::INT[]),
  unnest(@etching_arr::BOOLEAN[]),
  unnest(@etching_divisibility_arr), -- (SMALLINT[]) don't cast to allow null values
  unnest(@etching_premine_arr), -- (DECIMAL[]) don't cast to allow null values
  unnest(@etching_rune_arr), -- (TEXT[]) don't cast to allow null values
  unnest(@etching_spacers_arr), -- (INT[]) don't cast to allow null values
  unnest(@etching_symbol_arr), -- (INT[]) don't cast to allow null values
  unnest(@etching_terms_arr), -- (BOOLEAN[]) don't cast to allow null values
  unnest(@etching_terms_amount_arr), -- (DECIMAL[]) don't cast to allow null values
  unnest(@etching_terms_cap_arr), -- (DECIMAL[]) don't cast to allow null values
  unnest(@etching_terms_height_start_arr), -- (INT[]) don't cast to allow null values
  unnest(@etching_terms_height_end_arr), -- (INT[]) don't cast to allow null values
  unnest(@etching_terms_offset_start_arr), -- (INT[]) don't cast to allow null values
  unnest(@etching_terms_offset_end_arr), -- (INT[]) don't cast to allow null values
  unnest(@etching_turbo_arr), -- (BOOLEAN[]) don't cast to allow null values
  unnest(@edicts_arr::JSONB[]),
  unnest(@mint_arr), -- (TEXT[]) don't cast to allow null values
  unnest(@pointer_arr), -- (INT[]) don't cast to allow null values
  unnest(@cenotaph_arr::BOOLEAN[]),
  unnest(@flaws_arr::INT[])
);

-- name: BatchCreateRuneTransactions :exec
INSERT INTO runes_transactions ("hash", "block_height", "index", "timestamp", "inputs", "outputs", "mints", "burns", "rune_etched")
VALUES (
  unnest(@hash_arr::TEXT[]),
  unnest(@block_height_arr::INT[]),
  unnest(@index_arr::INT[]),
  unnest(@timestamp_arr::TIMESTAMP[]),
  unnest(@inputs_arr::JSONB[]),
  unnest(@outputs_arr::JSONB[]),
  unnest(@mints_arr::JSONB[]),
  unnest(@burns_arr::JSONB[]),
  unnest(@rune_etched_arr::BOOLEAN[])
);
