//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/MMaZX/goforge/database/postgres"
)

func TestPostgresIntegration(t *testing.T) {
	runSuite(t, harness{
		name:          "postgres",
		transactional: true,
		start: func(ctx context.Context) (string, func(), error) {
			req := testcontainers.ContainerRequest{
				Image:        "postgres:16-alpine",
				ExposedPorts: []string{"5432/tcp"},
				Env: map[string]string{
					"POSTGRES_PASSWORD": "goforge",
					"POSTGRES_DB":       "goforge",
				},
				WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
			}
			c, err := startContainer(ctx, req)
			if err != nil {
				return "", nil, err
			}
			host, err := c.Host(ctx)
			if err != nil {
				return "", nil, err
			}
			port, err := c.MappedPort(ctx, "5432/tcp")
			if err != nil {
				return "", nil, err
			}
			dsn := fmt.Sprintf("postgres://postgres:goforge@%s:%s/goforge?sslmode=disable", host, port.Port())
			if err := waitForPing(ctx, "postgres", dsn); err != nil {
				return "", nil, err
			}
			return dsn, func() { c.Terminate(ctx) }, nil
		},
		newProvider: func(ctx context.Context, dsn string) (provider, error) {
			return postgres.Open(ctx, dsn)
		},
	})
}
