// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: info.sql

package gen

import (
	"context"
)

const getCurrenEventHashVersion = `-- name: GetCurrenEventHashVersion :one
SELECT "event_hash_version" FROM runes_indexer_db_version ORDER BY id DESC LIMIT 1
`

func (q *Queries) GetCurrenEventHashVersion(ctx context.Context) (int32, error) {
	row := q.db.QueryRow(ctx, getCurrenEventHashVersion)
	var event_hash_version int32
	err := row.Scan(&event_hash_version)
	return event_hash_version, err
}

const getCurrentDBVersion = `-- name: GetCurrentDBVersion :one
SELECT "version" FROM runes_indexer_db_version ORDER BY id DESC LIMIT 1
`

func (q *Queries) GetCurrentDBVersion(ctx context.Context) (int32, error) {
	row := q.db.QueryRow(ctx, getCurrentDBVersion)
	var version int32
	err := row.Scan(&version)
	return version, err
}

const getCurrentIndexerStats = `-- name: GetCurrentIndexerStats :one
SELECT "client_version", "network" FROM runes_indexer_stats ORDER BY id DESC LIMIT 1
`

type GetCurrentIndexerStatsRow struct {
	ClientVersion string
	Network       string
}

func (q *Queries) GetCurrentIndexerStats(ctx context.Context) (GetCurrentIndexerStatsRow, error) {
	row := q.db.QueryRow(ctx, getCurrentIndexerStats)
	var i GetCurrentIndexerStatsRow
	err := row.Scan(&i.ClientVersion, &i.Network)
	return i, err
}

const updateIndexerStats = `-- name: UpdateIndexerStats :exec
INSERT INTO runes_indexer_stats (client_version, network) VALUES ($1, $2)
`

type UpdateIndexerStatsParams struct {
	ClientVersion string
	Network       string
}

func (q *Queries) UpdateIndexerStats(ctx context.Context, arg UpdateIndexerStatsParams) error {
	_, err := q.db.Exec(ctx, updateIndexerStats, arg.ClientVersion, arg.Network)
	return err
}
