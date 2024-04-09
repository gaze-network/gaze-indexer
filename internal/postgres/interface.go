package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Make sure that interfaces are compatible with the pgx package
var (
	_ DB = (*pgx.Conn)(nil)
	_ DB = (*pgxpool.Conn)(nil)
)

// Queryable is an interface that can be used to execute queries and commands
type Queryable interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

// TxQueryable is an interface that can be used to execute queries and commands within a transaction
type TxQueryable interface {
	Queryable
	Begin(context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

// DB is an interface that can be used to execute queries and commands, and also to send batches
type DB interface {
	Queryable
	TxQueryable
	SendBatch(ctx context.Context, b *pgx.Batch) (br pgx.BatchResults)
	Ping(ctx context.Context) error
}
