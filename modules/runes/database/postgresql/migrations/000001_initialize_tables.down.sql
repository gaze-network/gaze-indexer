BEGIN;

DROP TABLE IF EXISTS "runes_indexer_stats";
DROP TABLE IF EXISTS "runes_indexer_db_version";
DROP TABLE IF EXISTS "runes_entries";
DROP TABLE IF EXISTS "runes_outpoint_balances";
DROP TABLE IF EXISTS "runes_balances";

COMMIT;
