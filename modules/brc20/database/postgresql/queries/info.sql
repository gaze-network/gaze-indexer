-- name: GetLatestIndexerState :one
SELECT * FROM brc20_indexer_states ORDER BY created_at DESC LIMIT 1;

-- name: CreateIndexerState :exec
INSERT INTO brc20_indexer_states (client_version, network, db_version, event_hash_version) VALUES ($1, $2, $3, $4);
