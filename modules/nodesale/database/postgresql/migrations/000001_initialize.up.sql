BEGIN;

CREATE TABLE IF NOT EXISTS blocks (
    "block_height" BIGINT NOT NULL,
    "block_hash" TEXT NOT NULL,
    "module" TEXT NOT NULL,
    PRIMARY KEY("block_height", "block_hash")
);

CREATE TABLE IF NOT EXISTS events (
    "tx_hash" TEXT NOT NULL PRIMARY KEY,
    "block_height" BIGINT NOT NULL,
    "tx_index" INTEGER NOT NULL,
    "wallet_address" TEXT NOT NULL,
    "valid" BOOLEAN NOT NULL,
    "action" INTEGER NOT NULL,
    "raw_message" BYTEA NOT NULL,
    "parsed_message" JSONB NOT NULL DEFAULT '{}',
    "block_timestamp" TIMESTAMP NOT NULL,
    "block_hash" TEXT NOT NULL,
    "metadata" JSONB NOT NULL DEFAULT '{}'
);

INSERT INTO events("tx_hash", "block_height", "tx_index",
                 "wallet_address", "valid", "action", 
                "raw_message", "parsed_message", "block_timestamp",
                "block_hash", "metadata")
VALUES ('', -1, -1,
        '', false, -1,
        '', '{}', NOW(),
        '', '{}');

CREATE TABLE IF NOT EXISTS node_sales (
    "block_height" BIGINT NOT NULL,
    "tx_index" INTEGER NOT NULL,
    "name" TEXT NOT NULL,
    "starts_at" TIMESTAMP NOT NULL,
    "ends_at" TIMESTAMP NOT NULL,
    "tiers" JSONB[] NOT NULL,
    "seller_public_key" TEXT NOT NULL,
    "max_per_address" INTEGER NOT NULL,
    "deploy_tx_hash" TEXT NOT NULL REFERENCES events(tx_hash) ON DELETE CASCADE,
    "max_discount_percentage" INTEGER NOT NULL,
    "seller_wallet" TEXT NOT NULL,
    PRIMARY KEY ("block_height", "tx_index")
);

CREATE TABLE IF NOT EXISTS nodes (
    "sale_block" BIGINT NOT NULL,
    "sale_tx_index" INTEGER NOT NULL,
    "node_id" INTEGER NOT NULL,
    "tier_index" INTEGER NOT NULL,
    "delegated_to" TEXT NOT NULL DEFAULT '',
    "owner_public_key" TEXT NOT NULL,
    "purchase_tx_hash" TEXT NOT NULL REFERENCES events(tx_hash) ON DELETE CASCADE,
    "delegate_tx_hash" TEXT NOT NULL DEFAULT '' REFERENCES events(tx_hash) ON DELETE SET DEFAULT,
    PRIMARY KEY("sale_block", "sale_tx_index", "node_id"),
    FOREIGN KEY("sale_block", "sale_tx_index") REFERENCES node_sales("block_height", "tx_index")
);



COMMIT;