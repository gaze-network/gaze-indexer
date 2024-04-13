-- name: GetLatestIndexerState :one
SELECT * FROM runes_indexer_state ORDER BY created_at DESC LIMIT 1;

-- name: SetIndexerState :exec
INSERT INTO runes_indexer_state (db_version, event_hash_version) VALUES ($1, $2);

-- name: GetCurrentIndexerStats :one
SELECT "client_version", "network" FROM runes_indexer_stats ORDER BY id DESC LIMIT 1;

-- name: UpdateIndexerStats :exec
INSERT INTO runes_indexer_stats (client_version, network) VALUES ($1, $2);
