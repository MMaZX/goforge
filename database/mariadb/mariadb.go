// Package mariadb implements the migration.Provider for MariaDB.
package mariadb

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"github.com/MMaZX/goforge/migration"
)

// Provider is the MariaDB migration.Provider. MariaDB DDL statements cause
// an implicit commit, so unlike PostgreSQL, migrations here cannot be
// wrapped in a rollback-able transaction; the history table's dirty flag is
// what makes a crash mid-migration detectable instead.
type Provider struct {
	pool *sql.DB
}

// Open connects to MariaDB and verifies the connection.
func Open(ctx context.Context, dsn string) (*Provider, error) {
	pool, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mariadb: opening connection: %w", err)
	}
	if err := pool.PingContext(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("mariadb: connecting: %w", err)
	}
	return &Provider{pool: pool}, nil
}

// DB returns the migration.DB the engine should use.
func (p *Provider) DB() migration.DB { return p.pool }

// Close releases the underlying connection pool.
func (p *Provider) Close() error { return p.pool.Close() }

func (p *Provider) Name() string                   { return "mariadb" }
func (p *Provider) SupportsTransactionalDDL() bool { return false }

func (p *Provider) Locker(migration.DB) migration.Locker {
	return &namedLocker{pool: p.pool}
}

func (p *Provider) History() migration.HistoryStore {
	return historyStore{}
}
