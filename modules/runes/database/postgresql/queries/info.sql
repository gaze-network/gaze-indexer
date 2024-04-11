-- name: GetCurrentDBVersion :one
SELECT "version" FROM runes_indexer_db_version ORDER BY id DESC LIMIT 1;

-- name: GetCurrenEventHashVersion :one
SELECT "event_hash_version" FROM runes_indexer_db_version ORDER BY id DESC LIMIT 1;

-- name: GetCurrentIndexerStats :one
SELECT "client_version", "network" FROM runes_indexer_stats ORDER BY id DESC LIMIT 1;

-- name: UpdateIndexerStats :exec
INSERT INTO runes_indexer_stats (client_version, network) VALUES ($1, $2);
