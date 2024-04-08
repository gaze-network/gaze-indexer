-- name: GetCurrentDBVersion :one
SELECT "version" FROM bitcoin_indexer_db_version ORDER BY id DESC LIMIT 1;

-- name: GetCurrentIndexerStats :one
SELECT "client_version", "network" FROM bitcoin_indexer_stats ORDER BY id DESC LIMIT 1;

-- name: UpdateIndexerStats :exec
INSERT INTO bitcoin_indexer_stats (client_version, network) VALUES ($1, $2);