BEGIN;

DROP TABLE IF EXISTS "brc20_indexer_states";
DROP TABLE IF EXISTS "brc20_indexed_blocks";
DROP TABLE IF EXISTS "brc20_processor_stats";
DROP TABLE IF EXISTS "brc20_tick_entries";
DROP TABLE IF EXISTS "brc20_tick_entry_states";
DROP TABLE IF EXISTS "brc20_event_deploys";
DROP TABLE IF EXISTS "brc20_event_mints";
DROP TABLE IF EXISTS "brc20_event_inscribe_transfers";
DROP TABLE IF EXISTS "brc20_event_transfer_transfers";
DROP TABLE IF EXISTS "brc20_balances";
DROP TABLE IF EXISTS "brc20_inscription_entries";
DROP TABLE IF EXISTS "brc20_inscription_entry_states";
DROP TABLE IF EXISTS "brc20_inscription_transfers";

COMMIT;
