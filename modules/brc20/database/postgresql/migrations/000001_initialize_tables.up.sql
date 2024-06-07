BEGIN;

-- Indexer Client Information

CREATE TABLE IF NOT EXISTS "brc20_indexer_states" (
	"id" BIGSERIAL PRIMARY KEY,
	"client_version" TEXT NOT NULL,
	"network" TEXT NOT NULL,
	"db_version" INT NOT NULL,
	"event_hash_version" INT NOT NULL,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS brc20_indexer_state_created_at_idx ON "brc20_indexer_states" USING BTREE ("created_at" DESC);

-- BRC20 data

CREATE TABLE IF NOT EXISTS "brc20_indexed_blocks" (
	"height" INT NOT NULL PRIMARY KEY,
	"hash" TEXT NOT NULL,
	"event_hash" TEXT NOT NULL,
	"cumulative_event_hash" TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS "brc20_processor_stats" (
	"block_height" INT NOT NULL PRIMARY KEY,
	"cursed_inscription_count" INT NOT NULL,
	"blessed_inscription_count" INT NOT NULL,
	"lost_sats" BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS "brc20_tick_entries" (
	"tick" TEXT NOT NULL PRIMARY KEY, -- lowercase of original_tick
	"original_tick" TEXT NOT NULL,
	"total_supply" DECIMAL NOT NULL,
	"decimals" SMALLINT NOT NULL,
	"limit_per_mint" DECIMAL NOT NULL,
	"is_self_mint" BOOLEAN NOT NULL,
	"deploy_inscription_id" TEXT NOT NULL,
	"deployed_at" TIMESTAMP NOT NULL,
	"deployed_at_height" INT NOT NULL
);

CREATE TABLE IF NOT EXISTS "brc20_tick_entry_states" (
	"tick" TEXT NOT NULL,
	"block_height" INT NOT NULL,
	"minted_amount" DECIMAL NOT NULL,
	"burned_amount" DECIMAL NOT NULL,
	"completed_at" TIMESTAMP,
	"completed_at_height" INT,
	PRIMARY KEY ("tick", "block_height")
);

CREATE SEQUENCE IF NOT EXISTS brc20_event_id_seq;

CREATE TABLE IF NOT EXISTS "brc20_event_deploys" (
	"id" BIGINT PRIMARY KEY DEFAULT nextval('brc20_event_id_seq'),
	"inscription_id" TEXT NOT NULL,
	"inscription_number" BIGINT NOT NULL,
	"tick" TEXT NOT NULL, -- lowercase of original_tick
	"original_tick" TEXT NOT NULL,
	"tx_hash" TEXT NOT NULL,
	"block_height" INT NOT NULL,
	"tx_index" INT NOT NULL,
	"timestamp" TIMESTAMP NOT NULL,
	
	"pkscript" TEXT NOT NULL,
	"satpoint" TEXT NOT NULL,
	"total_supply" DECIMAL NOT NULL,
	"decimals" SMALLINT NOT NULL,
	"limit_per_mint" DECIMAL NOT NULL,
	"is_self_mint" BOOLEAN NOT NULL
);
CREATE INDEX IF NOT EXISTS brc20_event_deploys_block_height_idx ON "brc20_event_deploys" USING BTREE ("block_height");

CREATE TABLE IF NOT EXISTS "brc20_event_mints" (
	"id" BIGINT PRIMARY KEY DEFAULT nextval('brc20_event_id_seq'),
	"inscription_id" TEXT NOT NULL,
	"inscription_number" BIGINT NOT NULL,
	"tick" TEXT NOT NULL, -- lowercase of original_tick
	"original_tick" TEXT NOT NULL,
	"tx_hash" TEXT NOT NULL,
	"block_height" INT NOT NULL,
	"tx_index" INT NOT NULL,
	"timestamp" TIMESTAMP NOT NULL,

	"pkscript" TEXT NOT NULL,
	"satpoint" TEXT NOT NULL,
	"amount" DECIMAL NOT NULL,
	"parent_id" TEXT -- requires parent deploy inscription id if minting a self-mint ticker
);
CREATE INDEX IF NOT EXISTS brc20_event_mints_block_height_idx ON "brc20_event_mints" USING BTREE ("block_height");

CREATE TABLE IF NOT EXISTS "brc20_event_inscribe_transfers" (
	"id" BIGINT PRIMARY KEY DEFAULT nextval('brc20_event_id_seq'),
	"inscription_id" TEXT NOT NULL,
	"inscription_number" BIGINT NOT NULL,
	"tick" TEXT NOT NULL, -- lowercase of original_tick
	"original_tick" TEXT NOT NULL,
	"tx_hash" TEXT NOT NULL,
	"block_height" INT NOT NULL,
	"tx_index" INT NOT NULL,
	"timestamp" TIMESTAMP NOT NULL,

	"pkscript" TEXT NOT NULL,
	"satpoint" TEXT NOT NULL,
	"output_index" INT NOT NULL,
	"sats_amount" BIGINT NOT NULL,
	"amount" DECIMAL NOT NULL
);
CREATE INDEX IF NOT EXISTS brc20_event_inscribe_transfers_block_height_idx ON "brc20_event_inscribe_transfers" USING BTREE ("block_height");

CREATE TABLE IF NOT EXISTS "brc20_event_transfer_transfers" (
	"id" BIGINT PRIMARY KEY DEFAULT nextval('brc20_event_id_seq'),
	"inscription_id" TEXT NOT NULL,
	"inscription_number" BIGINT NOT NULL,
	"tick" TEXT NOT NULL, -- lowercase of original_tick
	"original_tick" TEXT NOT NULL,
	"tx_hash" TEXT NOT NULL,
	"block_height" INT NOT NULL,
	"tx_index" INT NOT NULL,
	"timestamp" TIMESTAMP NOT NULL,

	"from_pkscript" TEXT NOT NULL,
	"from_satpoint" TEXT NOT NULL,
	"from_input_index" INT NOT NULL,
	"to_pkscript" TEXT NOT NULL,
	"to_satpoint" TEXT NOT NULL,
	"to_output_index" INT NOT NULL,
	"amount" DECIMAL NOT NULL
);
CREATE INDEX IF NOT EXISTS brc20_event_transfer_transfers_block_height_idx ON "brc20_event_transfer_transfers" USING BTREE ("block_height");

CREATE TABLE IF NOT EXISTS "brc20_balances" (
	"pkscript" TEXT NOT NULL,
	"block_height" INT NOT NULL,
	"tick" TEXT NOT NULL,
	"overall_balance" DECIMAL NOT NULL,
	"available_balance" DECIMAL NOT NULL,
	PRIMARY KEY ("pkscript", "tick", "block_height")
);

CREATE TABLE IF NOT EXISTS "brc20_inscription_entries" (
	"id" TEXT NOT NULL PRIMARY KEY,
	"number" BIGINT NOT NULL,
	"sequence_number" BIGINT NOT NULL,
	"delegate" TEXT, -- delegate inscription id
	"metadata" BYTEA,
	"metaprotocol" TEXT,
	"parents" TEXT[], -- parent inscription id, 0.14 only supports 1 parent per inscription
	"pointer" BIGINT,
	"content" JSONB, -- can use jsonb because we only track brc20 inscriptions
	"content_encoding" TEXT,
	"content_type" TEXT,
	"cursed" BOOLEAN NOT NULL, -- inscriptions after jubilee are no longer cursed in 0.14, which affects inscription number
	"cursed_for_brc20" BOOLEAN NOT NULL, -- however, inscriptions that would normally be cursed are still considered cursed for brc20
	"created_at" TIMESTAMP NOT NULL,
	"created_at_height" INT NOT NULL
);

CREATE TABLE IF NOT EXISTS "brc20_inscription_entry_states" (
	"id" TEXT NOT NULL,
	"block_height" INT NOT NULL,
	"transfer_count" INT NOT NULL,
	PRIMARY KEY ("id", "block_height")
);

CREATE TABLE IF NOT EXISTS "brc20_inscription_transfers" (
	"inscription_id" TEXT NOT NULL,
	"block_height" INT NOT NULL,
	"tx_index" INT NOT NULL,
	"old_satpoint_tx_hash" TEXT,
	"old_satpoint_out_idx" INT,
	"old_satpoint_offset" BIGINT,
	"new_satpoint_tx_hash" TEXT,
	"new_satpoint_out_idx" INT,
	"new_satpoint_offset" BIGINT,
	"new_pkscript" TEXT NOT NULL,
	"new_output_value" BIGINT NOT NULL,
	"sent_as_fee" BOOLEAN NOT NULL,
	PRIMARY KEY ("inscription_id", "block_height", "tx_index")
);
CREATE INDEX IF NOT EXISTS brc20_inscription_transfers_block_height_tx_index_idx ON "brc20_inscription_transfers" USING BTREE ("block_height", "tx_index");
CREATE INDEX IF NOT EXISTS brc20_inscription_transfers_new_satpoint_idx ON "brc20_inscription_transfers" USING BTREE ("new_satpoint_tx_hash", "new_satpoint_out_idx", "new_satpoint_offset");

COMMIT;
