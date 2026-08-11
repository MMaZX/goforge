package mariadb

import (
	"context"
	"database/sql"
	"fmt"
)

const lockName = "goforge_migrations_lock"

// namedLocker serializes migration runs using MariaDB's GET_LOCK/RELEASE_LOCK,
// which are session-scoped like PostgreSQL's advisory locks, so a dedicated
// connection is held for the lock's lifetime.
type namedLocker struct {
	pool *sql.DB
	conn *sql.Conn
}

func (l *namedLocker) Lock(ctx context.Context) error {
	conn, err := l.pool.Conn(ctx)
	if err != nil {
		return fmt.Errorf("mariadb: acquiring connection for lock: %w", err)
	}

	var acquired int
	// timeout in seconds; a long-lived one lets the caller's ctx be the
	// real deadline while still not blocking forever if ctx has none.
	if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, 3600)", lockName).Scan(&acquired); err != nil {
		conn.Close()
		return fmt.Errorf("mariadb: acquiring named lock: %w", err)
	}
	if acquired != 1 {
		conn.Close()
		return fmt.Errorf("mariadb: named lock %q held by another connection", lockName)
	}
	l.conn = conn
	return nil
}

func (l *namedLocker) Unlock(ctx context.Context) error {
	if l.conn == nil {
		return nil
	}
	defer l.conn.Close()
	if _, err := l.conn.ExecContext(ctx, "SELECT RELEASE_LOCK(?)", lockName); err != nil {
		return fmt.Errorf("mariadb: releasing named lock: %w", err)
	}
	return nil
}
