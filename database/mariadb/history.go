package mariadb

import (
	"context"
	"fmt"
	"time"

	"github.com/MMaZX/goforge/migration"
)

const tableName = "goforge_migrations"

type historyStore struct{}

func (historyStore) EnsureTable(ctx context.Context, db migration.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS `+tableName+` (
			version            BIGINT UNSIGNED PRIMARY KEY,
			name               VARCHAR(255) NOT NULL,
			checksum           VARCHAR(64) NOT NULL,
			applied_at         DATETIME(3) NOT NULL,
			execution_time_ms  BIGINT NOT NULL DEFAULT 0,
			batch              INT NOT NULL,
			dirty              TINYINT(1) NOT NULL DEFAULT 0
		) ENGINE=InnoDB`)
	if err != nil {
		return fmt.Errorf("mariadb: ensuring history table: %w", err)
	}
	return nil
}

func (historyStore) List(ctx context.Context, db migration.DB) ([]migration.Record, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT version, name, checksum, applied_at, execution_time_ms, batch, dirty
		FROM `+tableName+`
		ORDER BY version ASC`)
	if err != nil {
		return nil, fmt.Errorf("mariadb: listing history: %w", err)
	}
	defer rows.Close()

	var out []migration.Record
	for rows.Next() {
		var r migration.Record
		var execMS int64
		var dirty int
		if err := rows.Scan(&r.Version, &r.Name, &r.Checksum, &r.AppliedAt, &execMS, &r.Batch, &dirty); err != nil {
			return nil, fmt.Errorf("mariadb: scanning history row: %w", err)
		}
		r.ExecutionTime = time.Duration(execMS) * time.Millisecond
		r.Dirty = dirty != 0
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("mariadb: reading history rows: %w", err)
	}
	return out, nil
}

func (historyStore) Begin(ctx context.Context, db migration.DB, version uint64, name, checksum string, batch int) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO `+tableName+` (version, name, checksum, applied_at, execution_time_ms, batch, dirty)
		VALUES (?, ?, ?, ?, 0, ?, 1)`,
		version, name, checksum, time.Now(), batch)
	if err != nil {
		return fmt.Errorf("mariadb: recording migration start: %w", err)
	}
	return nil
}

func (historyStore) Complete(ctx context.Context, db migration.DB, version uint64, executionTime time.Duration) error {
	_, err := db.ExecContext(ctx, `
		UPDATE `+tableName+`
		SET dirty = 0, execution_time_ms = ?
		WHERE version = ?`,
		executionTime.Milliseconds(), version)
	if err != nil {
		return fmt.Errorf("mariadb: recording migration completion: %w", err)
	}
	return nil
}

func (historyStore) Remove(ctx context.Context, db migration.DB, version uint64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM `+tableName+` WHERE version = ?`, version)
	if err != nil {
		return fmt.Errorf("mariadb: removing history row: %w", err)
	}
	return nil
}
