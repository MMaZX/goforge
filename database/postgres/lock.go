package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"time"
)

// lockKey is the pg_advisory_lock key GoForge uses. It is a fixed value so
// that every goforge process against the same database contends for the
// same lock, regardless of working directory or migrations path.
var lockKey = func() int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte("goforge_migrations_lock"))
	return int64(h.Sum64())
}()

// advisoryLocker serializes migration runs using PostgreSQL's session-level
// advisory locks. A dedicated connection is held for the lock's lifetime,
// since pg_advisory_lock/unlock are tied to the session that acquired them
// and would be meaningless through a pooled, rotating connection.
type advisoryLocker struct {
	pool *sql.DB
	conn *sql.Conn
}

func (l *advisoryLocker) Lock(ctx context.Context) error {
	conn, err := l.pool.Conn(ctx)
	if err != nil {
		return fmt.Errorf("postgres: acquiring connection for lock: %w", err)
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		var acquired bool
		if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", lockKey).Scan(&acquired); err != nil {
			conn.Close()
			return fmt.Errorf("postgres: acquiring advisory lock: %w", err)
		}
		if acquired {
			l.conn = conn
			return nil
		}
		select {
		case <-ctx.Done():
			conn.Close()
			return fmt.Errorf("postgres: waiting for advisory lock: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

func (l *advisoryLocker) Unlock(ctx context.Context) error {
	if l.conn == nil {
		return nil
	}
	defer l.conn.Close()
	if _, err := l.conn.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", lockKey); err != nil {
		return fmt.Errorf("postgres: releasing advisory lock: %w", err)
	}
	return nil
}
