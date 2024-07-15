// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: events.sql

package gen

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const addEvent = `-- name: AddEvent :exec
INSERT INTO events("tx_hash", "block_height", "tx_index", "wallet_address", "valid", "action", 
                    "raw_message", "parsed_message", "block_timestamp", "block_hash", "metadata")
VALUES ( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
`

type AddEventParams struct {
	TxHash         string
	BlockHeight    int64
	TxIndex        int32
	WalletAddress  string
	Valid          bool
	Action         int32
	RawMessage     []byte
	ParsedMessage  []byte
	BlockTimestamp pgtype.Timestamp
	BlockHash      string
	Metadata       []byte
}

func (q *Queries) AddEvent(ctx context.Context, arg AddEventParams) error {
	_, err := q.db.Exec(ctx, addEvent,
		arg.TxHash,
		arg.BlockHeight,
		arg.TxIndex,
		arg.WalletAddress,
		arg.Valid,
		arg.Action,
		arg.RawMessage,
		arg.ParsedMessage,
		arg.BlockTimestamp,
		arg.BlockHash,
		arg.Metadata,
	)
	return err
}

const getEventsByWallet = `-- name: GetEventsByWallet :many
SELECT tx_hash, block_height, tx_index, wallet_address, valid, action, raw_message, parsed_message, block_timestamp, block_hash, metadata
FROM events
WHERE wallet_address = $1
`

func (q *Queries) GetEventsByWallet(ctx context.Context, walletAddress string) ([]Event, error) {
	rows, err := q.db.Query(ctx, getEventsByWallet, walletAddress)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Event
	for rows.Next() {
		var i Event
		if err := rows.Scan(
			&i.TxHash,
			&i.BlockHeight,
			&i.TxIndex,
			&i.WalletAddress,
			&i.Valid,
			&i.Action,
			&i.RawMessage,
			&i.ParsedMessage,
			&i.BlockTimestamp,
			&i.BlockHash,
			&i.Metadata,
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

const removeEventsFromBlock = `-- name: RemoveEventsFromBlock :execrows
DELETE FROM events
WHERE "block_height" >= $1
`

func (q *Queries) RemoveEventsFromBlock(ctx context.Context, fromBlock int64) (int64, error) {
	result, err := q.db.Exec(ctx, removeEventsFromBlock, fromBlock)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
