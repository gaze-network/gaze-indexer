-- name: GetLatestIndexedBlock :one
SELECT * FROM "brc20_indexed_blocks" ORDER BY "height" DESC LIMIT 1;

-- name: GetIndexedBlockByHeight :one
SELECT * FROM "brc20_indexed_blocks" WHERE "height" = $1;

-- name: GetLatestProcessorStats :one
SELECT * FROM "brc20_processor_stats" ORDER BY "block_height" DESC LIMIT 1;

-- name: GetInscriptionTransfersInOutPoints :many
SELECT "it".*, "ie"."content"   FROM (
    SELECT
      unnest(@tx_hash_arr::text[]) AS "tx_hash",
      unnest(@tx_out_idx_arr::int[]) AS "tx_out_idx"
  ) "inputs"
  INNER JOIN "brc20_inscription_transfers" it ON "inputs"."tx_hash" = "it"."new_satpoint_tx_hash" AND "inputs"."tx_out_idx" = "it"."new_satpoint_out_idx"
  LEFT JOIN "brc20_inscription_entries" ie ON "it"."inscription_id" = "ie"."id";
  ;

-- name: GetInscriptionEntriesByIds :many
WITH "states" AS (
  -- select latest state
  SELECT DISTINCT ON ("id") * FROM "brc20_inscription_entry_states" WHERE "id" = ANY(@inscription_ids::text[]) ORDER BY "id", "block_height" DESC
)
SELECT * FROM "brc20_inscription_entries"
  LEFT JOIN "states" ON "brc20_inscription_entries"."id" = "states"."id"
  WHERE "brc20_inscription_entries"."id" = ANY(@inscription_ids::text[]);

-- name: GetTickEntriesByTicks :many
WITH "states" AS (
  -- select latest state
  SELECT DISTINCT ON ("tick") * FROM "brc20_tick_entry_states" WHERE "tick" = ANY(@ticks::text[]) ORDER BY "tick", "block_height" DESC
)
SELECT * FROM "brc20_tick_entries"
  LEFT JOIN "states" ON "brc20_tick_entries"."tick" = "states"."tick"
  WHERE "brc20_tick_entries"."tick" = ANY(@ticks::text[]);

-- name: CreateIndexedBlock :exec
INSERT INTO "brc20_indexed_blocks" ("height", "hash", "event_hash", "cumulative_event_hash") VALUES ($1, $2, $3, $4);

-- name: CreateProcessorStats :exec
INSERT INTO "brc20_processor_stats" ("block_height", "cursed_inscription_count", "blessed_inscription_count", "lost_sats") VALUES ($1, $2, $3, $4);

-- name: CreateTickEntries :batchexec
INSERT INTO "brc20_tick_entries" ("tick", "original_tick", "total_supply", "decimals", "limit_per_mint", "is_self_mint", "deploy_inscription_id", "deployed_at", "deployed_at_height") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: CreateTickEntryStates :batchexec
INSERT INTO "brc20_tick_entry_states" ("tick", "block_height", "minted_amount", "burned_amount", "completed_at", "completed_at_height") VALUES ($1, $2, $3, $4, $5, $6);

-- name: CreateInscriptionEntries :batchexec
INSERT INTO "brc20_inscription_entries" ("id", "number", "sequence_number", "delegate", "metadata", "metaprotocol", "parents", "pointer", "content", "content_encoding", "content_type", "cursed", "cursed_for_brc20", "created_at", "created_at_height") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15);

-- name: CreateInscriptionEntryStates :batchexec
INSERT INTO "brc20_inscription_entry_states" ("id", "block_height", "transfer_count") VALUES ($1, $2, $3);

-- name: CreateInscriptionTransfers :batchexec
INSERT INTO "brc20_inscription_transfers" ("inscription_id", "block_height", "tx_index", "old_satpoint_tx_hash", "old_satpoint_out_idx", "old_satpoint_offset", "new_satpoint_tx_hash", "new_satpoint_out_idx", "new_satpoint_offset", "new_pkscript", "new_output_value", "sent_as_fee") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: CreateEventDeploys :batchexec
INSERT INTO "brc20_event_deploys" ("inscription_id", "inscription_number", "tick", "original_tick", "tx_hash", "block_height", "tx_index", "timestamp", "pkscript", "satpoint", "total_supply", "decimals", "limit_per_mint", "is_self_mint") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);

-- name: CreateEventMints :batchexec
INSERT INTO "brc20_event_mints" ("inscription_id", "inscription_number", "tick", "original_tick", "tx_hash", "block_height", "tx_index", "timestamp", "pkscript", "satpoint", "amount", "parent_id") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: CreateEventInscribeTransfers :batchexec
INSERT INTO "brc20_event_inscribe_transfers" ("inscription_id", "inscription_number", "tick", "original_tick", "tx_hash", "block_height", "tx_index", "timestamp", "pkscript", "satpoint". "output_index", "sats_amount", "amount") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: CreateEventTransferTransfers :batchexec
INSERT INTO "brc20_event_transfer_transfers" ("inscription_id", "inscription_number", "tick", "original_tick", "tx_hash", "block_height", "tx_index", "timestamp", "from_pkscript", "from_satpoint", "from_input_index", "to_pkscript", "to_satpoint", "to_output_index", "amount") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15);

-- name: DeleteIndexedBlocksSinceHeight :exec
DELETE FROM "brc20_indexed_blocks" WHERE "height" >= $1;

-- name: DeleteProcessorStatsSinceHeight :exec
DELETE FROM "brc20_processor_stats" WHERE "block_height" >= $1;

-- name: DeleteTickEntriesSinceHeight :exec
DELETE FROM "brc20_tick_entries" WHERE "deployed_at_height" >= $1;

-- name: DeleteTickEntryStatesSinceHeight :exec
DELETE FROM "brc20_tick_entry_states" WHERE "block_height" >= $1;

-- name: DeleteEventDeploysSinceHeight :exec
DELETE FROM "brc20_event_deploys" WHERE "block_height" >= $1;

-- name: DeleteEventMintsSinceHeight :exec
DELETE FROM "brc20_event_mints" WHERE "block_height" >= $1;

-- name: DeleteEventInscribeTransfersSinceHeight :exec
DELETE FROM "brc20_event_inscribe_transfers" WHERE "block_height" >= $1;

-- name: DeleteEventTransferTransfersSinceHeight :exec
DELETE FROM "brc20_event_transfer_transfers" WHERE "block_height" >= $1;

-- name: DeleteBalancesSinceHeight :exec
DELETE FROM "brc20_balances" WHERE "block_height" >= $1;

-- name: DeleteInscriptionEntriesSinceHeight :exec
DELETE FROM "brc20_inscription_entries" WHERE "created_at_height" >= $1;

-- name: DeleteInscriptionEntryStatesSinceHeight :exec
DELETE FROM "brc20_inscription_entry_states" WHERE "block_height" >= $1;

-- name: DeleteInscriptionTransfersSinceHeight :exec
DELETE FROM "brc20_inscription_transfers" WHERE "block_height" >= $1;
