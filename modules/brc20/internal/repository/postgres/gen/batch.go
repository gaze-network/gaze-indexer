// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: batch.go

package gen

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrBatchAlreadyClosed = errors.New("batch already closed")
)

const createBalances = `-- name: CreateBalances :batchexec
INSERT INTO "brc20_balances" ("pkscript", "block_height", "tick", "overall_balance", "available_balance") VALUES ($1, $2, $3, $4, $5)
`

type CreateBalancesBatchResults struct {
	br     pgx.BatchResults
	tot    int
	closed bool
}

type CreateBalancesParams struct {
	Pkscript         string
	BlockHeight      int32
	Tick             string
	OverallBalance   pgtype.Numeric
	AvailableBalance pgtype.Numeric
}

func (q *Queries) CreateBalances(ctx context.Context, arg []CreateBalancesParams) *CreateBalancesBatchResults {
	batch := &pgx.Batch{}
	for _, a := range arg {
		vals := []interface{}{
			a.Pkscript,
			a.BlockHeight,
			a.Tick,
			a.OverallBalance,
			a.AvailableBalance,
		}
		batch.Queue(createBalances, vals...)
	}
	br := q.db.SendBatch(ctx, batch)
	return &CreateBalancesBatchResults{br, len(arg), false}
}

func (b *CreateBalancesBatchResults) Exec(f func(int, error)) {
	defer b.br.Close()
	for t := 0; t < b.tot; t++ {
		if b.closed {
			if f != nil {
				f(t, ErrBatchAlreadyClosed)
			}
			continue
		}
		_, err := b.br.Exec()
		if f != nil {
			f(t, err)
		}
	}
}

func (b *CreateBalancesBatchResults) Close() error {
	b.closed = true
	return b.br.Close()
}

const createEventDeploys = `-- name: CreateEventDeploys :batchexec
INSERT INTO "brc20_event_deploys" ("id", "inscription_id", "inscription_number", "tick", "original_tick", "tx_hash", "block_height", "tx_index", "timestamp", "pkscript", "satpoint", "total_supply", "decimals", "limit_per_mint", "is_self_mint") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
`

type CreateEventDeploysBatchResults struct {
	br     pgx.BatchResults
	tot    int
	closed bool
}

type CreateEventDeploysParams struct {
	Id                int64
	InscriptionID     string
	InscriptionNumber int64
	Tick              string
	OriginalTick      string
	TxHash            string
	BlockHeight       int32
	TxIndex           int32
	Timestamp         pgtype.Timestamp
	Pkscript          string
	Satpoint          string
	TotalSupply       pgtype.Numeric
	Decimals          int16
	LimitPerMint      pgtype.Numeric
	IsSelfMint        bool
}

func (q *Queries) CreateEventDeploys(ctx context.Context, arg []CreateEventDeploysParams) *CreateEventDeploysBatchResults {
	batch := &pgx.Batch{}
	for _, a := range arg {
		vals := []interface{}{
			a.Id,
			a.InscriptionID,
			a.InscriptionNumber,
			a.Tick,
			a.OriginalTick,
			a.TxHash,
			a.BlockHeight,
			a.TxIndex,
			a.Timestamp,
			a.Pkscript,
			a.Satpoint,
			a.TotalSupply,
			a.Decimals,
			a.LimitPerMint,
			a.IsSelfMint,
		}
		batch.Queue(createEventDeploys, vals...)
	}
	br := q.db.SendBatch(ctx, batch)
	return &CreateEventDeploysBatchResults{br, len(arg), false}
}

func (b *CreateEventDeploysBatchResults) Exec(f func(int, error)) {
	defer b.br.Close()
	for t := 0; t < b.tot; t++ {
		if b.closed {
			if f != nil {
				f(t, ErrBatchAlreadyClosed)
			}
			continue
		}
		_, err := b.br.Exec()
		if f != nil {
			f(t, err)
		}
	}
}

func (b *CreateEventDeploysBatchResults) Close() error {
	b.closed = true
	return b.br.Close()
}

const createEventInscribeTransfers = `-- name: CreateEventInscribeTransfers :batchexec
INSERT INTO "brc20_event_inscribe_transfers" ("id", "inscription_id", "inscription_number", "tick", "original_tick", "tx_hash", "block_height", "tx_index", "timestamp", "pkscript", "satpoint", "output_index", "sats_amount", "amount") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
`

type CreateEventInscribeTransfersBatchResults struct {
	br     pgx.BatchResults
	tot    int
	closed bool
}

type CreateEventInscribeTransfersParams struct {
	Id                int64
	InscriptionID     string
	InscriptionNumber int64
	Tick              string
	OriginalTick      string
	TxHash            string
	BlockHeight       int32
	TxIndex           int32
	Timestamp         pgtype.Timestamp
	Pkscript          string
	Satpoint          string
	OutputIndex       int32
	SatsAmount        int64
	Amount            pgtype.Numeric
}

func (q *Queries) CreateEventInscribeTransfers(ctx context.Context, arg []CreateEventInscribeTransfersParams) *CreateEventInscribeTransfersBatchResults {
	batch := &pgx.Batch{}
	for _, a := range arg {
		vals := []interface{}{
			a.Id,
			a.InscriptionID,
			a.InscriptionNumber,
			a.Tick,
			a.OriginalTick,
			a.TxHash,
			a.BlockHeight,
			a.TxIndex,
			a.Timestamp,
			a.Pkscript,
			a.Satpoint,
			a.OutputIndex,
			a.SatsAmount,
			a.Amount,
		}
		batch.Queue(createEventInscribeTransfers, vals...)
	}
	br := q.db.SendBatch(ctx, batch)
	return &CreateEventInscribeTransfersBatchResults{br, len(arg), false}
}

func (b *CreateEventInscribeTransfersBatchResults) Exec(f func(int, error)) {
	defer b.br.Close()
	for t := 0; t < b.tot; t++ {
		if b.closed {
			if f != nil {
				f(t, ErrBatchAlreadyClosed)
			}
			continue
		}
		_, err := b.br.Exec()
		if f != nil {
			f(t, err)
		}
	}
}

func (b *CreateEventInscribeTransfersBatchResults) Close() error {
	b.closed = true
	return b.br.Close()
}

const createEventMints = `-- name: CreateEventMints :batchexec
INSERT INTO "brc20_event_mints" ("id", "inscription_id", "inscription_number", "tick", "original_tick", "tx_hash", "block_height", "tx_index", "timestamp", "pkscript", "satpoint", "amount", "parent_id") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
`

type CreateEventMintsBatchResults struct {
	br     pgx.BatchResults
	tot    int
	closed bool
}

type CreateEventMintsParams struct {
	Id                int64
	InscriptionID     string
	InscriptionNumber int64
	Tick              string
	OriginalTick      string
	TxHash            string
	BlockHeight       int32
	TxIndex           int32
	Timestamp         pgtype.Timestamp
	Pkscript          string
	Satpoint          string
	Amount            pgtype.Numeric
	ParentID          pgtype.Text
}

func (q *Queries) CreateEventMints(ctx context.Context, arg []CreateEventMintsParams) *CreateEventMintsBatchResults {
	batch := &pgx.Batch{}
	for _, a := range arg {
		vals := []interface{}{
			a.Id,
			a.InscriptionID,
			a.InscriptionNumber,
			a.Tick,
			a.OriginalTick,
			a.TxHash,
			a.BlockHeight,
			a.TxIndex,
			a.Timestamp,
			a.Pkscript,
			a.Satpoint,
			a.Amount,
			a.ParentID,
		}
		batch.Queue(createEventMints, vals...)
	}
	br := q.db.SendBatch(ctx, batch)
	return &CreateEventMintsBatchResults{br, len(arg), false}
}

func (b *CreateEventMintsBatchResults) Exec(f func(int, error)) {
	defer b.br.Close()
	for t := 0; t < b.tot; t++ {
		if b.closed {
			if f != nil {
				f(t, ErrBatchAlreadyClosed)
			}
			continue
		}
		_, err := b.br.Exec()
		if f != nil {
			f(t, err)
		}
	}
}

func (b *CreateEventMintsBatchResults) Close() error {
	b.closed = true
	return b.br.Close()
}

const createEventTransferTransfers = `-- name: CreateEventTransferTransfers :batchexec
INSERT INTO "brc20_event_transfer_transfers" ("id", "inscription_id", "inscription_number", "tick", "original_tick", "tx_hash", "block_height", "tx_index", "timestamp", "from_pkscript", "from_satpoint", "from_input_index", "to_pkscript", "to_satpoint", "to_output_index", "spent_as_fee", "amount") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
`

type CreateEventTransferTransfersBatchResults struct {
	br     pgx.BatchResults
	tot    int
	closed bool
}

type CreateEventTransferTransfersParams struct {
	Id                int64
	InscriptionID     string
	InscriptionNumber int64
	Tick              string
	OriginalTick      string
	TxHash            string
	BlockHeight       int32
	TxIndex           int32
	Timestamp         pgtype.Timestamp
	FromPkscript      string
	FromSatpoint      string
	FromInputIndex    int32
	ToPkscript        string
	ToSatpoint        string
	ToOutputIndex     int32
	SpentAsFee        bool
	Amount            pgtype.Numeric
}

func (q *Queries) CreateEventTransferTransfers(ctx context.Context, arg []CreateEventTransferTransfersParams) *CreateEventTransferTransfersBatchResults {
	batch := &pgx.Batch{}
	for _, a := range arg {
		vals := []interface{}{
			a.Id,
			a.InscriptionID,
			a.InscriptionNumber,
			a.Tick,
			a.OriginalTick,
			a.TxHash,
			a.BlockHeight,
			a.TxIndex,
			a.Timestamp,
			a.FromPkscript,
			a.FromSatpoint,
			a.FromInputIndex,
			a.ToPkscript,
			a.ToSatpoint,
			a.ToOutputIndex,
			a.SpentAsFee,
			a.Amount,
		}
		batch.Queue(createEventTransferTransfers, vals...)
	}
	br := q.db.SendBatch(ctx, batch)
	return &CreateEventTransferTransfersBatchResults{br, len(arg), false}
}

func (b *CreateEventTransferTransfersBatchResults) Exec(f func(int, error)) {
	defer b.br.Close()
	for t := 0; t < b.tot; t++ {
		if b.closed {
			if f != nil {
				f(t, ErrBatchAlreadyClosed)
			}
			continue
		}
		_, err := b.br.Exec()
		if f != nil {
			f(t, err)
		}
	}
}

func (b *CreateEventTransferTransfersBatchResults) Close() error {
	b.closed = true
	return b.br.Close()
}

const createInscriptionEntries = `-- name: CreateInscriptionEntries :batchexec
INSERT INTO "brc20_inscription_entries" ("id", "number", "sequence_number", "delegate", "metadata", "metaprotocol", "parents", "pointer", "content", "content_encoding", "content_type", "cursed", "cursed_for_brc20", "created_at", "created_at_height") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
`

type CreateInscriptionEntriesBatchResults struct {
	br     pgx.BatchResults
	tot    int
	closed bool
}

type CreateInscriptionEntriesParams struct {
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
}

func (q *Queries) CreateInscriptionEntries(ctx context.Context, arg []CreateInscriptionEntriesParams) *CreateInscriptionEntriesBatchResults {
	batch := &pgx.Batch{}
	for _, a := range arg {
		vals := []interface{}{
			a.Id,
			a.Number,
			a.SequenceNumber,
			a.Delegate,
			a.Metadata,
			a.Metaprotocol,
			a.Parents,
			a.Pointer,
			a.Content,
			a.ContentEncoding,
			a.ContentType,
			a.Cursed,
			a.CursedForBrc20,
			a.CreatedAt,
			a.CreatedAtHeight,
		}
		batch.Queue(createInscriptionEntries, vals...)
	}
	br := q.db.SendBatch(ctx, batch)
	return &CreateInscriptionEntriesBatchResults{br, len(arg), false}
}

func (b *CreateInscriptionEntriesBatchResults) Exec(f func(int, error)) {
	defer b.br.Close()
	for t := 0; t < b.tot; t++ {
		if b.closed {
			if f != nil {
				f(t, ErrBatchAlreadyClosed)
			}
			continue
		}
		_, err := b.br.Exec()
		if f != nil {
			f(t, err)
		}
	}
}

func (b *CreateInscriptionEntriesBatchResults) Close() error {
	b.closed = true
	return b.br.Close()
}

const createInscriptionEntryStates = `-- name: CreateInscriptionEntryStates :batchexec
INSERT INTO "brc20_inscription_entry_states" ("id", "block_height", "transfer_count") VALUES ($1, $2, $3)
`

type CreateInscriptionEntryStatesBatchResults struct {
	br     pgx.BatchResults
	tot    int
	closed bool
}

type CreateInscriptionEntryStatesParams struct {
	Id            string
	BlockHeight   int32
	TransferCount int32
}

func (q *Queries) CreateInscriptionEntryStates(ctx context.Context, arg []CreateInscriptionEntryStatesParams) *CreateInscriptionEntryStatesBatchResults {
	batch := &pgx.Batch{}
	for _, a := range arg {
		vals := []interface{}{
			a.Id,
			a.BlockHeight,
			a.TransferCount,
		}
		batch.Queue(createInscriptionEntryStates, vals...)
	}
	br := q.db.SendBatch(ctx, batch)
	return &CreateInscriptionEntryStatesBatchResults{br, len(arg), false}
}

func (b *CreateInscriptionEntryStatesBatchResults) Exec(f func(int, error)) {
	defer b.br.Close()
	for t := 0; t < b.tot; t++ {
		if b.closed {
			if f != nil {
				f(t, ErrBatchAlreadyClosed)
			}
			continue
		}
		_, err := b.br.Exec()
		if f != nil {
			f(t, err)
		}
	}
}

func (b *CreateInscriptionEntryStatesBatchResults) Close() error {
	b.closed = true
	return b.br.Close()
}

const createInscriptionTransfers = `-- name: CreateInscriptionTransfers :batchexec
INSERT INTO "brc20_inscription_transfers" ("inscription_id", "inscription_number", "inscription_sequence_number", "block_height", "tx_index", "tx_hash", "from_input_index", "old_satpoint_tx_hash", "old_satpoint_out_idx", "old_satpoint_offset", "new_satpoint_tx_hash", "new_satpoint_out_idx", "new_satpoint_offset", "new_pkscript", "new_output_value", "sent_as_fee", "transfer_count") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
`

type CreateInscriptionTransfersBatchResults struct {
	br     pgx.BatchResults
	tot    int
	closed bool
}

type CreateInscriptionTransfersParams struct {
	InscriptionID             string
	InscriptionNumber         int64
	InscriptionSequenceNumber int64
	BlockHeight               int32
	TxIndex                   int32
	TxHash                    string
	FromInputIndex            int32
	OldSatpointTxHash         pgtype.Text
	OldSatpointOutIdx         pgtype.Int4
	OldSatpointOffset         pgtype.Int8
	NewSatpointTxHash         pgtype.Text
	NewSatpointOutIdx         pgtype.Int4
	NewSatpointOffset         pgtype.Int8
	NewPkscript               string
	NewOutputValue            int64
	SentAsFee                 bool
	TransferCount             int32
}

func (q *Queries) CreateInscriptionTransfers(ctx context.Context, arg []CreateInscriptionTransfersParams) *CreateInscriptionTransfersBatchResults {
	batch := &pgx.Batch{}
	for _, a := range arg {
		vals := []interface{}{
			a.InscriptionID,
			a.InscriptionNumber,
			a.InscriptionSequenceNumber,
			a.BlockHeight,
			a.TxIndex,
			a.TxHash,
			a.FromInputIndex,
			a.OldSatpointTxHash,
			a.OldSatpointOutIdx,
			a.OldSatpointOffset,
			a.NewSatpointTxHash,
			a.NewSatpointOutIdx,
			a.NewSatpointOffset,
			a.NewPkscript,
			a.NewOutputValue,
			a.SentAsFee,
			a.TransferCount,
		}
		batch.Queue(createInscriptionTransfers, vals...)
	}
	br := q.db.SendBatch(ctx, batch)
	return &CreateInscriptionTransfersBatchResults{br, len(arg), false}
}

func (b *CreateInscriptionTransfersBatchResults) Exec(f func(int, error)) {
	defer b.br.Close()
	for t := 0; t < b.tot; t++ {
		if b.closed {
			if f != nil {
				f(t, ErrBatchAlreadyClosed)
			}
			continue
		}
		_, err := b.br.Exec()
		if f != nil {
			f(t, err)
		}
	}
}

func (b *CreateInscriptionTransfersBatchResults) Close() error {
	b.closed = true
	return b.br.Close()
}

const createTickEntries = `-- name: CreateTickEntries :batchexec
INSERT INTO "brc20_tick_entries" ("tick", "original_tick", "total_supply", "decimals", "limit_per_mint", "is_self_mint", "deploy_inscription_id", "deployed_at", "deployed_at_height") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

type CreateTickEntriesBatchResults struct {
	br     pgx.BatchResults
	tot    int
	closed bool
}

type CreateTickEntriesParams struct {
	Tick                string
	OriginalTick        string
	TotalSupply         pgtype.Numeric
	Decimals            int16
	LimitPerMint        pgtype.Numeric
	IsSelfMint          bool
	DeployInscriptionID string
	DeployedAt          pgtype.Timestamp
	DeployedAtHeight    int32
}

func (q *Queries) CreateTickEntries(ctx context.Context, arg []CreateTickEntriesParams) *CreateTickEntriesBatchResults {
	batch := &pgx.Batch{}
	for _, a := range arg {
		vals := []interface{}{
			a.Tick,
			a.OriginalTick,
			a.TotalSupply,
			a.Decimals,
			a.LimitPerMint,
			a.IsSelfMint,
			a.DeployInscriptionID,
			a.DeployedAt,
			a.DeployedAtHeight,
		}
		batch.Queue(createTickEntries, vals...)
	}
	br := q.db.SendBatch(ctx, batch)
	return &CreateTickEntriesBatchResults{br, len(arg), false}
}

func (b *CreateTickEntriesBatchResults) Exec(f func(int, error)) {
	defer b.br.Close()
	for t := 0; t < b.tot; t++ {
		if b.closed {
			if f != nil {
				f(t, ErrBatchAlreadyClosed)
			}
			continue
		}
		_, err := b.br.Exec()
		if f != nil {
			f(t, err)
		}
	}
}

func (b *CreateTickEntriesBatchResults) Close() error {
	b.closed = true
	return b.br.Close()
}

const createTickEntryStates = `-- name: CreateTickEntryStates :batchexec
INSERT INTO "brc20_tick_entry_states" ("tick", "block_height", "minted_amount", "burned_amount", "completed_at", "completed_at_height") VALUES ($1, $2, $3, $4, $5, $6)
`

type CreateTickEntryStatesBatchResults struct {
	br     pgx.BatchResults
	tot    int
	closed bool
}

type CreateTickEntryStatesParams struct {
	Tick              string
	BlockHeight       int32
	MintedAmount      pgtype.Numeric
	BurnedAmount      pgtype.Numeric
	CompletedAt       pgtype.Timestamp
	CompletedAtHeight pgtype.Int4
}

func (q *Queries) CreateTickEntryStates(ctx context.Context, arg []CreateTickEntryStatesParams) *CreateTickEntryStatesBatchResults {
	batch := &pgx.Batch{}
	for _, a := range arg {
		vals := []interface{}{
			a.Tick,
			a.BlockHeight,
			a.MintedAmount,
			a.BurnedAmount,
			a.CompletedAt,
			a.CompletedAtHeight,
		}
		batch.Queue(createTickEntryStates, vals...)
	}
	br := q.db.SendBatch(ctx, batch)
	return &CreateTickEntryStatesBatchResults{br, len(arg), false}
}

func (b *CreateTickEntryStatesBatchResults) Exec(f func(int, error)) {
	defer b.br.Close()
	for t := 0; t < b.tot; t++ {
		if b.closed {
			if f != nil {
				f(t, ErrBatchAlreadyClosed)
			}
			continue
		}
		_, err := b.br.Exec()
		if f != nil {
			f(t, err)
		}
	}
}

func (b *CreateTickEntryStatesBatchResults) Close() error {
	b.closed = true
	return b.br.Close()
}
