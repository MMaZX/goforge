package migration

import (
	"context"
	"database/sql"
)

// DB is the minimal database surface the migration engine depends on.
// Concrete providers (postgres, mariadb) wrap a *sql.DB or *sql.Tx to satisfy it.
type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Tx is an in-flight transaction. *sql.Tx satisfies this interface, but the
// engine depends only on this instead of the concrete type so that
// transactional execution can be unit tested without a real database.
type Tx interface {
	DB
	Commit() error
	Rollback() error
}

// TxCapableDB is implemented by connections that can open transactions.
// Providers that support transactional DDL must pass a DB satisfying this
// interface to the engine.
type TxCapableDB interface {
	DB
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
}
