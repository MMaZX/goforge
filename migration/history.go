package migration

import (
	"context"
	"time"
)

// Record is a row of the goforge_migrations history table.
type Record struct {
	Version       uint64
	Name          string
	Checksum      string
	AppliedAt     time.Time
	ExecutionTime time.Duration
	Batch         int
	Dirty         bool
}

// HistoryStore persists migration history. Providers implement it on top of
// their own SQL dialect (column types, quoting) while the engine only
// depends on this interface.
type HistoryStore interface {
	// EnsureTable creates the history table if it does not exist yet.
	EnsureTable(ctx context.Context, db DB) error
	// List returns all recorded migrations ordered by version.
	List(ctx context.Context, db DB) ([]Record, error)
	// Begin records a migration as started (dirty=true) before it runs.
	// For providers without transactional DDL this is what allows a crash
	// mid-migration to be detected on the next run.
	Begin(ctx context.Context, db DB, version uint64, name, checksum string, batch int) error
	// Complete marks a previously begun migration as finished successfully.
	Complete(ctx context.Context, db DB, version uint64, executionTime time.Duration) error
	// Remove deletes a migration's record, used when rolling back.
	Remove(ctx context.Context, db DB, version uint64) error
}

// Provider encapsulates the database-specific behavior the engine needs:
// locking strategy, transactional DDL support and history storage.
type Provider interface {
	Name() string
	// SupportsTransactionalDDL reports whether a migration's Up/Down and its
	// history bookkeeping can be wrapped in a single transaction.
	SupportsTransactionalDDL() bool
	Locker(db DB) Locker
	History() HistoryStore
}
