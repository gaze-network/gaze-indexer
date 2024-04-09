BEGIN;

-- Indexer Client Information

CREATE TABLE IF NOT EXISTS "runes_indexer_stats" (
	"id" BIGSERIAL PRIMARY KEY,
	"client_version" TEXT NOT NULL,
	"network" TEXT NOT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "runes_indexer_db_version" (
	"id" BIGSERIAL PRIMARY KEY,
	"version" INT NOT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO "runes_indexer_db_version" ("version") VALUES (1);

-- Runes data

CREATE TABLE IF NOT EXISTS "runes_entries" (
  "rune_id" TEXT NOT NULL PRIMARY KEY,
  "rune" TEXT NOT NULL,
  "spacers" INT NOT NULL,
  "burned_amount" DECIMAL NOT NULL,
  "mints" DECIMAL NOT NULL,
  "premine" DECIMAL NOT NULL,
  "symbol" INT NOT NULL,
  "term_amount" DECIMAL,
  "term_cap" DECIMAL,
  "term_height_start" DECIMAL, -- using DECIMAL because it is uint64 but postgres only supports up to int64
  "term_height_end" DECIMAL,
  "term_offset_start" DECIMAL,
  "term_offset_end" DECIMAL,
  "completion_time" TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS runes_entries_rune_idx ON "runes_entries" USING HASH ("rune");

CREATE TABLE IF NOT EXISTS "runes_outpoint_balances" (
  "rune_id" TEXT NOT NULL,
  "tx_hash" TEXT NOT NULL,
  "tx_idx" INT NOT NULL, -- output index
  "value" DECIMAL NOT NULL,
  PRIMARY KEY ("rune_id", "tx_hash", "tx_idx")
);

CREATE TABLE IF NOT EXISTS "runes_balances" (
  "pkscript" TEXT NOT NULL,
  "block_height" INT NOT NULL,
  "rune_id" TEXT NOT NULL,
  "value" DECIMAL NOT NULL,
  PRIMARY KEY ("pkscript", "block_height", "rune_id")
);

COMMIT;
