//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/MMaZX/goforge/database/mariadb"
)

func TestMariaDBIntegration(t *testing.T) {
	runSuite(t, harness{
		name:          "mariadb",
		transactional: false,
		start: func(ctx context.Context) (string, func(), error) {
			req := testcontainers.ContainerRequest{
				Image:        "mariadb:11",
				ExposedPorts: []string{"3306/tcp"},
				Env: map[string]string{
					"MARIADB_ROOT_PASSWORD": "goforge",
					"MARIADB_DATABASE":      "goforge",
				},
				WaitingFor: wait.ForLog("mariadbd: ready for connections").WithOccurrence(1),
			}
			c, err := startContainer(ctx, req)
			if err != nil {
				return "", nil, err
			}
			host, err := c.Host(ctx)
			if err != nil {
				return "", nil, err
			}
			port, err := c.MappedPort(ctx, "3306/tcp")
			if err != nil {
				return "", nil, err
			}
			dsn := fmt.Sprintf("root:goforge@tcp(%s:%s)/goforge?parseTime=true", host, port.Port())
			if err := waitForPing(ctx, "mysql", dsn); err != nil {
				return "", nil, err
			}
			return dsn, func() { c.Terminate(ctx) }, nil
		},
		newProvider: func(ctx context.Context, dsn string) (provider, error) {
			return mariadb.Open(ctx, dsn)
		},
	})
}
