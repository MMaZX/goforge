// Package cliutil holds the plumbing shared by every goforge subcommand:
// connecting to the configured database, building the migration.Engine and
// rendering output in both human and --json form.
package cliutil

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/MMaZX/goforge/database/mariadb"
	"github.com/MMaZX/goforge/database/postgres"
	"github.com/MMaZX/goforge/internal/config"
	"github.com/MMaZX/goforge/internal/providers"
	"github.com/MMaZX/goforge/migration"
)

// Connection bundles a Provider with the DB the engine should use against
// it, and a way to close both.
type Connection struct {
	Provider migration.Provider
	DB       migration.DB
	close    func() error
}

func (c *Connection) Close() error { return c.close() }

// Connect opens a database connection for cfg.Database.Driver. See
// internal/providers for the supported drivers; config.Load already
// canonicalizes cfg.Database.Driver, this Resolve call is defense in depth
// for callers that build a Config by hand.
func Connect(ctx context.Context, cfg *config.Config) (*Connection, error) {
	desc, err := providers.Resolve(cfg.Database.Driver)
	if err != nil {
		return nil, fmt.Errorf("cliutil: %w", err)
	}

	switch desc.Driver {
	case "postgres":
		p, err := postgres.Open(ctx, cfg.Database.URL)
		if err != nil {
			return nil, err
		}
		return &Connection{Provider: p, DB: p.DB(), close: p.Close}, nil
	case "mariadb":
		p, err := mariadb.Open(ctx, cfg.Database.URL)
		if err != nil {
			return nil, err
		}
		return &Connection{Provider: p, DB: p.DB(), close: p.Close}, nil
	default:
		return nil, fmt.Errorf("cliutil: unsupported database driver %q (supported: %s)", desc.Driver, strings.Join(providers.Names(), ", "))
	}
}

// LoadMigrations discovers SQL migrations under cfg.Migrations.Path. Go
// migrations are not loaded here: the standalone CLI never interprets .go
// files, they are used by importing the migration engine as a library
// together with the registry `goforge generate` produces.
func LoadMigrations(cfg *config.Config) ([]migration.Entry, error) {
	entries, err := migration.Load(os.DirFS(cfg.Migrations.Path), nil)
	if err != nil {
		return nil, fmt.Errorf("cliutil: loading migrations from %s: %w", cfg.Migrations.Path, err)
	}
	return entries, nil
}

// BuildEngine connects to the database and constructs the engine in one
// step. The caller is responsible for closing the returned Connection.
func BuildEngine(ctx context.Context, cfg *config.Config) (*migration.Engine, *Connection, error) {
	conn, err := Connect(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}
	entries, err := LoadMigrations(cfg)
	if err != nil {
		conn.Close()
		return nil, nil, err
	}
	engine, err := migration.NewEngine(conn.DB, conn.Provider, entries)
	if err != nil {
		conn.Close()
		return nil, nil, err
	}
	return engine, conn, nil
}
