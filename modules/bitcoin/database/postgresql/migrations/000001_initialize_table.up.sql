BEGIN;

-- Indexer Client Information

CREATE TABLE IF NOT EXISTS "bitcoin_indexer_stats" (
	"id" BIGSERIAL PRIMARY KEY,
	"client_version" TEXT NOT NULL,
	"network" TEXT NOT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "bitcoin_indexer_db_version" (
	"id" BIGSERIAL PRIMARY KEY,
	"version" INT NOT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO "bitcoin_indexer_db_version" ("version") VALUES (1);

-- Bitcoin Data

CREATE TABLE IF NOT EXISTS "bitcoin_blocks" (
	"block_height" INT NOT NULL PRIMARY KEY,
	"block_hash" TEXT NOT NULL,
	"version" INT NOT NULL,
	"merkle_root" TEXT NOT NULL,
	"prev_block_hash" TEXT NOT NULL,
	"timestamp" TIMESTAMP WITH TIME ZONE NOT NULL,
	"bits" INT NOT NULL,
	"nonce" INT NOT NULL
);

CREATE INDEX IF NOT EXISTS bitcoin_blocks_block_hash_idx ON "bitcoin_blocks" USING HASH ("block_hash");

CREATE TABLE IF NOT EXISTS "bitcoin_transactions" (
	"tx_hash" TEXT NOT NULL PRIMARY KEY,
	"version" INT NOT NULL,
	"locktime" INT NOT NULL,
	"block_height" INT NOT NULL,
	"block_hash" TEXT NOT NULL,
	"idx" SMALLINT NOT NULL
);

CREATE INDEX IF NOT EXISTS bitcoin_transactions_block_height_idx ON "bitcoin_transactions" USING BTREE ("block_height");
CREATE INDEX IF NOT EXISTS bitcoin_transactions_block_hash_idx ON "bitcoin_transactions" USING BTREE ("block_hash");

CREATE TABLE IF NOT EXISTS "bitcoin_transaction_txouts" (
	"tx_hash" TEXT NOT NULL,
	"tx_idx" SMALLINT NOT NULL,
	"pkscript" TEXT NOT NULL, -- Hex String
	"value" BIGINT NOT NULL,
	"is_spent" BOOLEAN NOT NULL DEFAULT false,
	PRIMARY KEY ("tx_hash", "tx_idx")
);

CREATE INDEX IF NOT EXISTS bitcoin_transaction_txouts_pkscript_idx ON "bitcoin_transaction_txouts" USING HASH ("pkscript");

CREATE TABLE IF NOT EXISTS "bitcoin_transaction_txins" (
	"tx_hash" TEXT NOT NULL,
	"tx_idx" SMALLINT NOT NULL,
	"prevout_tx_hash" TEXT NOT NULL,
	"prevout_tx_idx" SMALLINT NOT NULL,
	"prevout_pkscript" TEXT NOT NULL, -- Hex String
	"scriptsig" TEXT NOT NULL, -- Hex String
	"witness" TEXT, -- Hex String
	"sequence" INT NOT NULL,
	PRIMARY KEY ("tx_hash", "tx_idx")
);

CREATE UNIQUE INDEX IF NOT EXISTS bitcoin_transaction_txins_prevout_idx ON "bitcoin_transaction_txins" USING BTREE ("prevout_tx_hash", "prevout_tx_idx");

COMMIT;