package migration

import "context"

// Locker prevents two processes from running migrations concurrently
// against the same database. Implementations are provider-specific
// (PostgreSQL advisory locks, MariaDB GET_LOCK).
type Locker interface {
	Lock(ctx context.Context) error
	Unlock(ctx context.Context) error
}
