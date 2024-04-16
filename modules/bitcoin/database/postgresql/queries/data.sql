-- name: GetLatestBlockHeader :one
SELECT * FROM bitcoin_blocks ORDER BY block_height DESC LIMIT 1;

-- TODO: GetBlockHeaderByRange

-- TODO: GetBlockByHeight/Hash (Join block with transactions, txins, txouts)

-- name: InsertBlock :exec
INSERT INTO bitcoin_blocks ("block_height","block_hash","version","merkle_root","prev_block_hash","timestamp","bits","nonce") VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: InsertTransaction :exec
INSERT INTO bitcoin_transactions ("tx_hash","version","locktime","block_height","block_hash","idx") VALUES ($1, $2, $3, $4, $5, $6);

-- name: InsertTransactionTxOut :exec
INSERT INTO bitcoin_transaction_txouts ("tx_hash","tx_idx","pkscript","value") VALUES ($1, $2, $3, $4);

-- name: InsertTransactionTxIn :exec
WITH update_txout AS (
	UPDATE "bitcoin_transaction_txouts"
	SET "is_spent" = true
	WHERE "tx_hash" = $3 AND "tx_idx" = $4 AND "is_spent" = false -- TODO: should throw an error if already spent
	RETURNING "pkscript"
)
INSERT INTO bitcoin_transaction_txins ("tx_hash","tx_idx","prevout_tx_hash","prevout_tx_idx","prevout_pkscript","scriptsig","witness","sequence") 
VALUES ($1, $2, $3, $4, (SELECT "pkscript" FROM update_txout), $5, $6, $7);

-- name: RevertData :exec
WITH delete_tx AS (
	DELETE FROM "bitcoin_transactions" WHERE "block_height" >= @from_height
	RETURNING "tx_hash"
), delete_txin AS (
	DELETE FROM  "bitcoin_transaction_txins" WHERE "tx_hash" = ANY(SELECT "tx_hash" FROM delete_tx)
	RETURNING "prevout_tx_hash", "prevout_tx_idx"
), delete_txout AS (
	DELETE FROM  "bitcoin_transaction_txouts" WHERE "tx_hash" = ANY(SELECT "tx_hash" FROM delete_tx)
	RETURNING NULL
), revert_txout_spent AS (
	UPDATE "bitcoin_transaction_txouts"
	SET "is_spent" = false
	WHERE ("tx_hash", "tx_idx") IN (SELECT "prevout_tx_hash", "prevout_tx_idx" FROM delete_txin)
	RETURNING NULL
)
DELETE FROM "bitcoin_blocks" WHERE "bitcoin_blocks"."block_height" >= @from_height;

-- name: GetBlocksByHeightRange :many
SELECT * FROM bitcoin_blocks WHERE block_height >= @from_height AND block_height <= @to_height ORDER BY block_height ASC;

-- name: GetTransactionsByHeightRange :many
SELECT * FROM bitcoin_transactions WHERE block_height >= @from_height AND block_height <= @to_height;

-- name: GetTransactionTxOutsByTxHashes :many
SELECT * FROM bitcoin_transaction_txouts WHERE tx_hash = ANY(@tx_hashes::TEXT[]);

-- name: GetTransactionTxInsByTxHashes :many
SELECT * FROM bitcoin_transaction_txins WHERE tx_hash = ANY(@tx_hashes::TEXT[]);
