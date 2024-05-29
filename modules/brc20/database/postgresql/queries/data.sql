-- name: GetLatestIndexedBlock :one
SELECT * FROM "brc20_indexed_blocks" ORDER BY "height" DESC LIMIT 1;

-- name: GetIndexedBlockByHeight :one
SELECT * FROM "brc20_indexed_blocks" WHERE "height" = $1;

-- name: GetLatestProcessorStats :one
SELECT * FROM "brc20_processor_stats" ORDER BY "block_height" DESC LIMIT 1;

-- name: GetInscriptionsInOutPoints :many
SELECT "brc20_inscription_transfers".* FROM (
    SELECT
      unnest(@tx_hash_arr::text[]) AS "tx_hash",
      unnest(@tx_out_idx_arr::int[]) AS "tx_out_idx"
  ) "inputs"
  INNER JOIN "brc20_inscription_transfers" ON "inputs"."tx_hash" = "brc20_inscription_transfers"."new_satpoint_tx_hash" AND "inputs"."tx_out_idx" = "brc20_inscription_transfers"."new_satpoint_out_idx";

-- name: GetInscriptionEntriesByIds :many
WITH "states" AS (
  -- select latest state
  SELECT DISTINCT ON ("id") * FROM "brc20_inscription_entry_states" WHERE "id" = ANY(@inscription_ids::text[]) ORDER BY "id", "block_height" DESC
)
SELECT * FROM "brc20_inscription_entries"
  LEFT JOIN "states" ON "brc20_inscription_entries"."id" = "states"."id"
  WHERE "brc20_inscription_entries"."id" = ANY(@inscription_ids::text[]);

-- name: CreateIndexedBlock :exec
INSERT INTO "brc20_indexed_blocks" ("height", "hash", "event_hash", "cumulative_event_hash") VALUES ($1, $2, $3, $4);

-- name: CreateProcessorStats :exec
INSERT INTO "brc20_processor_stats" ("block_height", "cursed_inscription_count", "blessed_inscription_count", "lost_sats") VALUES ($1, $2, $3, $4);

-- name: CreateInscriptionEntries :batchexec
INSERT INTO "brc20_inscription_entries" ("id", "number", "sequence_number", "delegate", "metadata", "metaprotocol", "parents", "pointer", "content", "content_encoding", "content_type", "cursed", "cursed_for_brc20", "created_at", "created_at_height") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15);

-- name: CreateInscriptionEntryStates :batchexec
INSERT INTO "brc20_inscription_entry_states" ("id", "block_height", "transfer_count") VALUES ($1, $2, $3);

-- name: CreateInscriptionTransfers :batchexec
INSERT INTO "brc20_inscription_transfers" ("inscription_id", "block_height", "old_satpoint_tx_hash", "old_satpoint_out_idx", "old_satpoint_offset", "new_satpoint_tx_hash", "new_satpoint_out_idx", "new_satpoint_offset", "new_pkscript", "new_output_value", "sent_as_fee") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
