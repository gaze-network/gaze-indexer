BEGIN;

CREATE TABLE IF NOT EXISTS blocks (
    "block_height" INTEGER NOT NULL,
    "block_hash" TEXT NOT NULL,
    "module" TEXT NOT NULL,
    PRIMARY KEY("block_height", "block_hash")
);

CREATE TABLE IF NOT EXISTS events (
    "tx_hash" TEXT NOT NULL PRIMARY KEY,
    "block_height" INTEGER NOT NULL,
    "tx_index" INTEGER NOT NULL,
    "action" INTEGER NOT NULL,
    "raw_message" BYTEA NOT NULL,
    "parsed_message" JSONB NOT NULL,
    "block_timestamp" TIMESTAMP NOT NULL,
    "block_hash" TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS node_sales (
    "block_height" INTEGER NOT NULL,
    "tx_index" INTEGER NOT NULL,
    "starts_at" TIMESTAMP NOT NULL,
    "ends_at" TIMESTAMP NOT NULL,
    "tiers" JSONB[] NOT NULL,
    "seller_public_key" TEXT NOT NULL,
    "max_per_address" INTEGER NOT NULL,
    "deploy_tx_hash" TEXT NOT NULL REFERENCES events(tx_hash) ON DELETE CASCADE,
    PRIMARY KEY ("block_height", "tx_index")
);

CREATE TABLE IF NOT EXISTS nodes (
    "sale_block" INTEGER NOT NULL,
    "sale_tx_index" INTEGER NOT NULL,
    "node_id" INTEGER NOT NULL,
    "tier_index" INTEGER NOT NULL,
    "delegated_to" TEXT NOT NULL,
    "owner_public_key" TEXT NOT NULL,
    "purchase_tx_hash" TEXT NOT NULL REFERENCES events(tx_hash) ON DELETE CASCADE,
    "delegate_tx_hash" TEXT NOT NULL REFERENCES events(tx_hash) ON DELETE SET NULL,
    PRIMARY KEY("sale_block", "sale_tx_index", "node_id"),
    FOREIGN KEY("sale_block", "sale_tx_index") REFERENCES node_sales("block_height", "tx_index")
);



COMMIT;