package datagateway

import "context"

type Tx interface {
	// Commit commits the DB transaction. All changes made after Begin() will be persisted. Calling Commit() will close the current transaction.
	// If Commit() is called without a prior Begin(), it must be a no-op.
	Commit(ctx context.Context) error
	// Rollback rolls back the DB transaction. All changes made after Begin() will be discarded.
	// Rollback() must be safe to call even if no transaction is active. Hence, a defer Rollback() is safe, even if Commit() was called prior with non-error conditions.
	Rollback(ctx context.Context) error
}
