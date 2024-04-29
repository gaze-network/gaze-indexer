BEGIN;

-- DROP INDEX
DROP INDEX IF EXISTS bitcoin_blocks_block_hash_idx;
DROP INDEX IF EXISTS bitcoin_transactions_tx_hash_idx;
DROP INDEX IF EXISTS bitcoin_transactions_block_hash_idx;
DROP INDEX IF EXISTS bitcoin_transaction_txouts_pkscript_idx;
DROP INDEX IF EXISTS bitcoin_transaction_txins_prevout_idx;

-- DROP TABLE
DROP TABLE IF EXISTS "bitcoin_indexer_stats";
DROP TABLE IF EXISTS "bitcoin_indexer_db_version";
DROP TABLE IF EXISTS "bitcoin_transaction_txins";
DROP TABLE IF EXISTS "bitcoin_transaction_txouts";
DROP TABLE IF EXISTS "bitcoin_transactions";
DROP TABLE IF EXISTS "bitcoin_blocks";

COMMIT;