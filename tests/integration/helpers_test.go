//go:build integration

// Package integration runs the migration engine against real PostgreSQL and
// MariaDB containers via Docker. Run with:
//
//	go test -tags=integration ./tests/integration/...
package integration

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"

	"github.com/MMaZX/goforge/migration"
)

// provider is what each database package (postgres, mariadb) exposes; the
// suite below is written against this instead of migration.Provider
// directly so it can also drop test tables between cases.
type provider interface {
	migration.Provider
	DB() migration.DB
	Close() error
}

type harness struct {
	name          string
	transactional bool
	newProvider   func(ctx context.Context, dsn string) (provider, error)
	start         func(ctx context.Context) (dsn string, terminate func(), err error)
}

func waitForPing(ctx context.Context, driverName, dsn string) error {
	deadline := time.Now().Add(60 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		db, err := sql.Open(driverName, dsn)
		if err == nil {
			lastErr = db.PingContext(ctx)
			db.Close()
			if lastErr == nil {
				return nil
			}
		} else {
			lastErr = err
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for database: %w", lastErr)
}

func startContainer(ctx context.Context, req testcontainers.ContainerRequest) (testcontainers.Container, error) {
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}
