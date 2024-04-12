BEGIN;

DROP TABLE IF EXISTS "runes_indexer_stats";
DROP TABLE IF EXISTS "runes_indexer_db_version";
DROP TABLE IF EXISTS "runes_processor_state";
DROP TABLE IF EXISTS "runes_indexed_blocks";
DROP TABLE IF EXISTS "runes_entries";
DROP TABLE IF EXISTS "runes_entry_states";
DROP TABLE IF EXISTS "runes_transactions";
DROP TABLE IF EXISTS "runes_runestones";
DROP TABLE IF EXISTS "runes_outpoint_balances";
DROP TABLE IF EXISTS "runes_balances"; 

COMMIT;
