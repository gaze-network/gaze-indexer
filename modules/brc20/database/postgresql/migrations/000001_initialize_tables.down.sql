BEGIN;

DROP TABLE IF EXISTS "brc20_indexer_states";
DROP TABLE IF EXISTS "brc20_indexed_blocks";
DROP TABLE IF EXISTS "brc20_processor_stats";
DROP TABLE IF EXISTS "brc20_tickers";
DROP TABLE IF EXISTS "brc20_ticker_states";
DROP TABLE IF EXISTS "brc20_deploy_events";
DROP TABLE IF EXISTS "brc20_mint_events";
DROP TABLE IF EXISTS "brc20_transfer_events";
DROP TABLE IF EXISTS "brc20_balances";
DROP TABLE IF EXISTS "brc20_inscription_entries";
DROP TABLE IF EXISTS "brc20_inscription_entry_states";
DROP TABLE IF EXISTS "brc20_inscription_transfers";

COMMIT;