BEGIN;

-- Indexer Client Information

CREATE TABLE IF NOT EXISTS "runes_indexer_stats" (
	"id" BIGSERIAL PRIMARY KEY,
	"client_version" TEXT NOT NULL,
	"network" TEXT NOT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "runes_indexer_state" (
	"id" BIGSERIAL PRIMARY KEY,
	"db_version" INT NOT NULL,
	"event_hash_version" INT NOT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS runes_indexer_state_created_at_idx ON "runes_indexer_state" USING BTREE ("created_at" DESC);

-- Runes data

CREATE TABLE IF NOT EXISTS "runes_indexed_blocks" (
	"height" INT NOT NULL PRIMARY KEY,
	"hash" TEXT NOT NULL,
	"prev_hash" TEXT NOT NULL,
	"event_hash" TEXT NOT NULL,
	"cumulative_event_hash" TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS "runes_entries" (
	"rune_id" TEXT NOT NULL PRIMARY KEY,
	"number" BIGINT NOT NULL, -- sequential number of the rune starting from 0
	"rune" TEXT NOT NULL,
	"spacers" INT NOT NULL,
	"premine" DECIMAL NOT NULL,
	"symbol" INT NOT NULL,
	"divisibility" SMALLINT NOT NULL,
	"terms" BOOLEAN NOT NULL, -- if true, then minting term exists for this entry
	"terms_amount" DECIMAL,
	"terms_cap" DECIMAL,
	"terms_height_start" INT,
	"terms_height_end" INT,
	"terms_offset_start" INT,
	"terms_offset_end" INT,
	"turbo" BOOLEAN NOT NULL,
	"etching_block" INT NOT NULL,
	"etching_tx_hash" TEXT NOT NULL,
	"etched_at" TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS runes_entries_rune_idx ON "runes_entries" USING BTREE ("rune");
CREATE UNIQUE INDEX IF NOT EXISTS runes_entries_number_idx ON "runes_entries" USING BTREE ("number");

CREATE TABLE IF NOT EXISTS "runes_entry_states" (
	"rune_id" TEXT NOT NULL,
	"block_height" INT NOT NULL,
	"mints" DECIMAL NOT NULL,
	"burned_amount" DECIMAL NOT NULL,
	"completed_at" TIMESTAMP,
	"completed_at_height" INT,
	PRIMARY KEY ("rune_id", "block_height")
);

CREATE TABLE IF NOT EXISTS "runes_transactions" (
	"hash" TEXT NOT NULL PRIMARY KEY,
	"block_height" INT NOT NULL,
	"index" INT NOT NULL,
	"timestamp" TIMESTAMP NOT NULL,
	"inputs" JSONB NOT NULL,
	"outputs" JSONB NOT NULL,
	"mints" JSONB NOT NULL,
	"burns" JSONB NOT NULL,
	"rune_etched" BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS "runes_runestones" (
	"tx_hash" TEXT NOT NULL PRIMARY KEY,
	"block_height" INT NOT NULL,
	"etching" BOOLEAN NOT NULL,
	"etching_divisibility" SMALLINT,
	"etching_premine" DECIMAL,
	"etching_rune" TEXT,
	"etching_spacers" INT,
	"etching_symbol" INT,
	"etching_terms" BOOLEAN,
	"etching_terms_amount" DECIMAL,
	"etching_terms_cap" DECIMAL,
	"etching_terms_height_start" INT,
	"etching_terms_height_end" INT,
	"etching_terms_offset_start" INT,
	"etching_terms_offset_end" INT,
	"etching_turbo" BOOLEAN,
	"edicts" JSONB NOT NULL DEFAULT '[]',
	"mint" TEXT,
	"pointer" INT,
	"cenotaph" BOOLEAN NOT NULL,
	"flaws" INT NOT NULL
);

CREATE TABLE IF NOT EXISTS "runes_outpoint_balances" (
	"rune_id" TEXT NOT NULL,
	"pkscript" TEXT NOT NULL,
	"tx_hash" TEXT NOT NULL,
	"tx_idx" INT NOT NULL, -- output index
	"amount" DECIMAL NOT NULL,
	"block_height" INT NOT NULL, -- block height when this output was created
	"spent_height" INT, -- block height when this output was spent
	PRIMARY KEY ("rune_id", "tx_hash", "tx_idx")
);

CREATE TABLE IF NOT EXISTS "runes_balances" (
	"pkscript" TEXT NOT NULL,
	"block_height" INT NOT NULL,
	"rune_id" TEXT NOT NULL,
	"amount" DECIMAL NOT NULL,
	PRIMARY KEY ("pkscript", "rune_id", "block_height")
);

COMMIT;
