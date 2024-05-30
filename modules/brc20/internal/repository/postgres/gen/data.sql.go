// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: data.sql

package gen

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createIndexedBlock = `-- name: CreateIndexedBlock :exec
INSERT INTO "brc20_indexed_blocks" ("height", "hash", "event_hash", "cumulative_event_hash") VALUES ($1, $2, $3, $4)
`

type CreateIndexedBlockParams struct {
	Height              int32
	Hash                string
	EventHash           string
	CumulativeEventHash string
}

func (q *Queries) CreateIndexedBlock(ctx context.Context, arg CreateIndexedBlockParams) error {
	_, err := q.db.Exec(ctx, createIndexedBlock,
		arg.Height,
		arg.Hash,
		arg.EventHash,
		arg.CumulativeEventHash,
	)
	return err
}

const createProcessorStats = `-- name: CreateProcessorStats :exec
INSERT INTO "brc20_processor_stats" ("block_height", "cursed_inscription_count", "blessed_inscription_count", "lost_sats") VALUES ($1, $2, $3, $4)
`

type CreateProcessorStatsParams struct {
	BlockHeight             int32
	CursedInscriptionCount  int32
	BlessedInscriptionCount int32
	LostSats                int64
}

func (q *Queries) CreateProcessorStats(ctx context.Context, arg CreateProcessorStatsParams) error {
	_, err := q.db.Exec(ctx, createProcessorStats,
		arg.BlockHeight,
		arg.CursedInscriptionCount,
		arg.BlessedInscriptionCount,
		arg.LostSats,
	)
	return err
}

const getIndexedBlockByHeight = `-- name: GetIndexedBlockByHeight :one
SELECT height, hash, event_hash, cumulative_event_hash FROM "brc20_indexed_blocks" WHERE "height" = $1
`

func (q *Queries) GetIndexedBlockByHeight(ctx context.Context, height int32) (Brc20IndexedBlock, error) {
	row := q.db.QueryRow(ctx, getIndexedBlockByHeight, height)
	var i Brc20IndexedBlock
	err := row.Scan(
		&i.Height,
		&i.Hash,
		&i.EventHash,
		&i.CumulativeEventHash,
	)
	return i, err
}

const getInscriptionEntriesByIds = `-- name: GetInscriptionEntriesByIds :many
WITH "states" AS (
  -- select latest state
  SELECT DISTINCT ON ("id") id, block_height, transfer_count FROM "brc20_inscription_entry_states" WHERE "id" = ANY($1::text[]) ORDER BY "id", "block_height" DESC
)
SELECT brc20_inscription_entries.id, number, sequence_number, delegate, metadata, metaprotocol, parents, pointer, content, content_encoding, content_type, cursed, cursed_for_brc20, created_at, created_at_height, states.id, block_height, transfer_count FROM "brc20_inscription_entries"
  LEFT JOIN "states" ON "brc20_inscription_entries"."id" = "states"."id"
  WHERE "brc20_inscription_entries"."id" = ANY($1::text[])
`

type GetInscriptionEntriesByIdsRow struct {
	Id              string
	Number          int64
	SequenceNumber  int64
	Delegate        pgtype.Text
	Metadata        []byte
	Metaprotocol    pgtype.Text
	Parents         []string
	Pointer         pgtype.Int8
	Content         []byte
	ContentEncoding pgtype.Text
	ContentType     pgtype.Text
	Cursed          bool
	CursedForBrc20  bool
	CreatedAt       pgtype.Timestamp
	CreatedAtHeight int32
	Id_2            pgtype.Text
	BlockHeight     pgtype.Int4
	TransferCount   pgtype.Int4
}

func (q *Queries) GetInscriptionEntriesByIds(ctx context.Context, inscriptionIds []string) ([]GetInscriptionEntriesByIdsRow, error) {
	rows, err := q.db.Query(ctx, getInscriptionEntriesByIds, inscriptionIds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetInscriptionEntriesByIdsRow
	for rows.Next() {
		var i GetInscriptionEntriesByIdsRow
		if err := rows.Scan(
			&i.Id,
			&i.Number,
			&i.SequenceNumber,
			&i.Delegate,
			&i.Metadata,
			&i.Metaprotocol,
			&i.Parents,
			&i.Pointer,
			&i.Content,
			&i.ContentEncoding,
			&i.ContentType,
			&i.Cursed,
			&i.CursedForBrc20,
			&i.CreatedAt,
			&i.CreatedAtHeight,
			&i.Id_2,
			&i.BlockHeight,
			&i.TransferCount,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getInscriptionsInOutPoints = `-- name: GetInscriptionsInOutPoints :many
SELECT brc20_inscription_transfers.inscription_id, brc20_inscription_transfers.block_height, brc20_inscription_transfers.tx_index, brc20_inscription_transfers.old_satpoint_tx_hash, brc20_inscription_transfers.old_satpoint_out_idx, brc20_inscription_transfers.old_satpoint_offset, brc20_inscription_transfers.new_satpoint_tx_hash, brc20_inscription_transfers.new_satpoint_out_idx, brc20_inscription_transfers.new_satpoint_offset, brc20_inscription_transfers.new_pkscript, brc20_inscription_transfers.new_output_value, brc20_inscription_transfers.sent_as_fee FROM (
    SELECT
      unnest($1::text[]) AS "tx_hash",
      unnest($2::int[]) AS "tx_out_idx"
  ) "inputs"
  INNER JOIN "brc20_inscription_transfers" ON "inputs"."tx_hash" = "brc20_inscription_transfers"."new_satpoint_tx_hash" AND "inputs"."tx_out_idx" = "brc20_inscription_transfers"."new_satpoint_out_idx"
`

type GetInscriptionsInOutPointsParams struct {
	TxHashArr   []string
	TxOutIdxArr []int32
}

func (q *Queries) GetInscriptionsInOutPoints(ctx context.Context, arg GetInscriptionsInOutPointsParams) ([]Brc20InscriptionTransfer, error) {
	rows, err := q.db.Query(ctx, getInscriptionsInOutPoints, arg.TxHashArr, arg.TxOutIdxArr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Brc20InscriptionTransfer
	for rows.Next() {
		var i Brc20InscriptionTransfer
		if err := rows.Scan(
			&i.InscriptionID,
			&i.BlockHeight,
			&i.TxIndex,
			&i.OldSatpointTxHash,
			&i.OldSatpointOutIdx,
			&i.OldSatpointOffset,
			&i.NewSatpointTxHash,
			&i.NewSatpointOutIdx,
			&i.NewSatpointOffset,
			&i.NewPkscript,
			&i.NewOutputValue,
			&i.SentAsFee,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getLatestIndexedBlock = `-- name: GetLatestIndexedBlock :one
SELECT height, hash, event_hash, cumulative_event_hash FROM "brc20_indexed_blocks" ORDER BY "height" DESC LIMIT 1
`

func (q *Queries) GetLatestIndexedBlock(ctx context.Context) (Brc20IndexedBlock, error) {
	row := q.db.QueryRow(ctx, getLatestIndexedBlock)
	var i Brc20IndexedBlock
	err := row.Scan(
		&i.Height,
		&i.Hash,
		&i.EventHash,
		&i.CumulativeEventHash,
	)
	return i, err
}

const getLatestProcessorStats = `-- name: GetLatestProcessorStats :one
SELECT block_height, cursed_inscription_count, blessed_inscription_count, lost_sats FROM "brc20_processor_stats" ORDER BY "block_height" DESC LIMIT 1
`

func (q *Queries) GetLatestProcessorStats(ctx context.Context) (Brc20ProcessorStat, error) {
	row := q.db.QueryRow(ctx, getLatestProcessorStats)
	var i Brc20ProcessorStat
	err := row.Scan(
		&i.BlockHeight,
		&i.CursedInscriptionCount,
		&i.BlessedInscriptionCount,
		&i.LostSats,
	)
	return i, err
}
