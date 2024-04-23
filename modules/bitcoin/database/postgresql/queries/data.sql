-- name: GetLatestBlockHeader :one
SELECT * FROM bitcoin_blocks ORDER BY block_height DESC LIMIT 1;

-- name: InsertBlock :exec
INSERT INTO bitcoin_blocks ("block_height","block_hash","version","merkle_root","prev_block_hash","timestamp","bits","nonce") VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: BatchInsertBlocks :exec
INSERT INTO bitcoin_blocks ("block_height","block_hash","version","merkle_root","prev_block_hash","timestamp","bits","nonce")
VALUES (
	unnest(@block_height_arr::INT[]),
	unnest(@block_hash_arr::TEXT[]),
	unnest(@version_arr::INT[]),
	unnest(@merkle_root_arr::TEXT[]),
	unnest(@prev_block_hash_arr::TEXT[]),
	unnest(@timestamp_arr::TIMESTAMP WITH TIME ZONE[]), -- or use TIMESTAMPTZ
	unnest(@bits_arr::BIGINT[]),
	unnest(@nonce_arr::BIGINT[])
);

-- name: BatchInsertTransactions :exec
INSERT INTO bitcoin_transactions ("tx_hash","version","locktime","block_height","block_hash","idx")
VALUES (
	unnest(@tx_hash_arr::TEXT[]),
	unnest(@version_arr::INT[]),
	unnest(@locktime_arr::BIGINT[]),
	unnest(@block_height_arr::INT[]),
	unnest(@block_hash_arr::TEXT[]),
	unnest(@idx_arr::INT[])
);

-- name: BatchInsertTransactionTxIns :exec
WITH update_txout AS (
	UPDATE "bitcoin_transaction_txouts"
	SET "is_spent" = true
	FROM (SELECT unnest(@prevout_tx_hash_arr::TEXT[]) as tx_hash, unnest(@prevout_tx_idx_arr::INT[]) as tx_idx) as txin
	WHERE "bitcoin_transaction_txouts"."tx_hash" = txin.tx_hash AND "bitcoin_transaction_txouts"."tx_idx" = txin.tx_idx AND "is_spent" = false
	RETURNING "bitcoin_transaction_txouts"."tx_hash", "bitcoin_transaction_txouts"."tx_idx", "pkscript"
), prepare_insert AS (
	SELECT input.tx_hash, input.tx_idx, prevout_tx_hash, prevout_tx_idx, update_txout.pkscript as prevout_pkscript, scriptsig, witness, sequence
		FROM (
			SELECT 
				unnest(@tx_hash_arr::TEXT[]) as tx_hash,
				unnest(@tx_idx_arr::INT[]) as tx_idx,
				unnest(@prevout_tx_hash_arr::TEXT[]) as prevout_tx_hash,
				unnest(@prevout_tx_idx_arr::INT[]) as prevout_tx_idx,
				unnest(@scriptsig_arr::TEXT[]) as scriptsig,
				unnest(@witness_arr::TEXT[]) as witness,
				unnest(@sequence_arr::INT[]) as sequence
		) input LEFT JOIN update_txout ON "update_txout"."tx_hash" = "input"."prevout_tx_hash" AND "update_txout"."tx_idx" = "input"."prevout_tx_idx"
)
INSERT INTO bitcoin_transaction_txins ("tx_hash","tx_idx","prevout_tx_hash","prevout_tx_idx", "prevout_pkscript","scriptsig","witness","sequence") 
SELECT "tx_hash", "tx_idx", "prevout_tx_hash", "prevout_tx_idx", "prevout_pkscript", "scriptsig", "witness", "sequence" FROM prepare_insert;

-- name: BatchInsertTransactionTxOuts :exec
INSERT INTO bitcoin_transaction_txouts ("tx_hash","tx_idx","pkscript","value")
VALUES (
	unnest(@tx_hash_arr::TEXT[]),
	unnest(@tx_idx_arr::INT[]),
	unnest(@pkscript_arr::TEXT[]),
	unnest(@value_arr::BIGINT[])
);

-- name: RevertData :exec
WITH delete_tx AS (
	DELETE FROM "bitcoin_transactions" WHERE "block_height" >= @from_height
	RETURNING "tx_hash"
), delete_txin AS (
	DELETE FROM  "bitcoin_transaction_txins" WHERE "tx_hash" = ANY(SELECT "tx_hash" FROM delete_tx)
	RETURNING "prevout_tx_hash", "prevout_tx_idx"
), revert_txout_spent AS (
	UPDATE "bitcoin_transaction_txouts"
	SET "is_spent" = false
	WHERE ("tx_hash", "tx_idx") IN (SELECT "prevout_tx_hash", "prevout_tx_idx" FROM delete_txin)
	RETURNING NULL
), delete_txout AS (
	DELETE FROM  "bitcoin_transaction_txouts" WHERE "tx_hash" = ANY(SELECT "tx_hash" FROM delete_tx)
	RETURNING NULL
)
DELETE FROM "bitcoin_blocks" WHERE "bitcoin_blocks"."block_height" >= @from_height;

-- name: GetBlockByHeight :one
SELECT * FROM bitcoin_blocks WHERE block_height = $1;

-- name: GetBlocksByHeightRange :many
SELECT * FROM bitcoin_blocks WHERE block_height >= @from_height AND block_height <= @to_height ORDER BY block_height ASC;

-- name: GetTransactionsByHeightRange :many
SELECT * FROM bitcoin_transactions WHERE block_height >= @from_height AND block_height <= @to_height;

-- name: GetTransactionByHash :one
SELECT * FROM bitcoin_transactions WHERE tx_hash = $1;

-- name: GetTransactionTxOutsByTxHashes :many
SELECT * FROM bitcoin_transaction_txouts WHERE tx_hash = ANY(@tx_hashes::TEXT[]);

-- name: GetTransactionTxInsByTxHashes :many
SELECT * FROM bitcoin_transaction_txins WHERE tx_hash = ANY(@tx_hashes::TEXT[]);
