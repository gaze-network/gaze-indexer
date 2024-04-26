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
	"bits" BIGINT NOT NULL,
	"nonce" BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS bitcoin_blocks_block_hash_idx ON "bitcoin_blocks" USING HASH ("block_hash");

CREATE TABLE IF NOT EXISTS "bitcoin_transactions" (
	"tx_hash" TEXT NOT NULL, -- can't use as primary key because block v1 has duplicate tx hashes (coinbase tx). See: https://github.com/bitcoin/bitcoin/commit/a206b0ea12eb4606b93323268fc81a4f1f952531
	"version" INT NOT NULL,
	"locktime" BIGINT NOT NULL,
	"block_height" INT NOT NULL,
	"block_hash" TEXT NOT NULL,
	"idx" INT NOT NULL,
	PRIMARY KEY ("block_height", "idx")
);

CREATE INDEX IF NOT EXISTS bitcoin_transactions_tx_hash_idx ON "bitcoin_transactions" USING HASH ("tx_hash");
CREATE INDEX IF NOT EXISTS bitcoin_transactions_block_hash_idx ON "bitcoin_transactions" USING HASH ("block_hash");

CREATE TABLE IF NOT EXISTS "bitcoin_transaction_txouts" (
	"tx_hash" TEXT NOT NULL,
	"tx_idx" INT NOT NULL,
	"pkscript" TEXT NOT NULL, -- Hex String
	"value" BIGINT NOT NULL,
	"is_spent" BOOLEAN NOT NULL DEFAULT false,
	PRIMARY KEY ("tx_hash", "tx_idx")
);

CREATE INDEX IF NOT EXISTS bitcoin_transaction_txouts_pkscript_idx ON "bitcoin_transaction_txouts" USING HASH ("pkscript");

CREATE TABLE IF NOT EXISTS "bitcoin_transaction_txins" (
	"tx_hash" TEXT NOT NULL,
	"tx_idx" INT NOT NULL,
	"prevout_tx_hash" TEXT NOT NULL,
	"prevout_tx_idx" INT NOT NULL,
	"prevout_pkscript" TEXT NULL, -- Hex String, Can be NULL if the prevout is a coinbase transaction
	"scriptsig" TEXT NOT NULL, -- Hex String
	"witness" TEXT NOT NULL DEFAULT '', -- Hex String
	"sequence" BIGINT NOT NULL,
	PRIMARY KEY ("tx_hash", "tx_idx")
);

CREATE INDEX IF NOT EXISTS bitcoin_transaction_txins_prevout_idx ON "bitcoin_transaction_txins" USING BTREE ("prevout_tx_hash", "prevout_tx_idx");

COMMIT;