// Package postgres implements the migration.Provider for PostgreSQL.
package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	"github.com/MMaZX/goforge/migration"
)

// Provider is the PostgreSQL migration.Provider: it owns the connection
// pool and knows how to lock, store history and run DDL transactionally.
type Provider struct {
	pool *sql.DB
}

// Open connects to PostgreSQL and verifies the connection.
//
// It builds the pool from a pq.Connector (via sql.OpenDB) rather than
// sql.Open("postgres", dsn): lib/pq's Driver does not implement
// driver.DriverContext, so sql.Open falls back to a path that ignores the
// context entirely during the TCP dial — a ctx deadline would never
// actually cut off connecting to an unreachable host. Connector.Connect
// does honor it.
func Open(ctx context.Context, dsn string) (*Provider, error) {
	connector, err := pq.NewConnector(dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: opening connection: %w", err)
	}
	pool := sql.OpenDB(connector)
	if err := pool.PingContext(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: connecting: %w", err)
	}
	return &Provider{pool: pool}, nil
}

// DB returns the migration.DB the engine should use. The concrete value
// also implements migration.TxCapableDB, which NewEngine detects since this
// provider reports SupportsTransactionalDDL() == true.
func (p *Provider) DB() migration.DB { return &db{p.pool} }

// Close releases the underlying connection pool.
func (p *Provider) Close() error { return p.pool.Close() }

func (p *Provider) Name() string                   { return "postgres" }
func (p *Provider) SupportsTransactionalDDL() bool { return true }

func (p *Provider) Locker(migration.DB) migration.Locker {
	return &advisoryLocker{pool: p.pool}
}

func (p *Provider) History() migration.HistoryStore {
	return historyStore{}
}

// db adapts *sql.DB to migration.TxCapableDB: *sql.DB already satisfies
// migration.DB through its embedded methods, it only needs BeginTx to
// return migration.Tx instead of the concrete *sql.Tx.
type db struct {
	*sql.DB
}

func (d *db) BeginTx(ctx context.Context, opts *sql.TxOptions) (migration.Tx, error) {
	tx, err := d.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
