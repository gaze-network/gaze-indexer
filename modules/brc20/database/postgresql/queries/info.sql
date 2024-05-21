-- name: GetLatestIndexerState :one
SELECT * FROM brc20_indexer_state ORDER BY created_at DESC LIMIT 1;

-- name: SetIndexerState :exec
INSERT INTO brc20_indexer_state (db_version, event_hash_version) VALUES ($1, $2);

-- name: GetLatestIndexerStats :one
SELECT "client_version", "network" FROM brc20_indexer_stats ORDER BY id DESC LIMIT 1;

-- name: UpdateIndexerStats :exec
INSERT INTO brc20_indexer_stats (client_version, network) VALUES ($1, $2);
