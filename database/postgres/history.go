package postgres

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
			version            BIGINT PRIMARY KEY,
			name               TEXT NOT NULL,
			checksum           TEXT NOT NULL,
			applied_at         TIMESTAMPTZ NOT NULL,
			execution_time_ms  BIGINT NOT NULL DEFAULT 0,
			batch              INT NOT NULL,
			dirty              BOOLEAN NOT NULL DEFAULT FALSE
		)`)
	if err != nil {
		return fmt.Errorf("postgres: ensuring history table: %w", err)
	}
	return nil
}

func (historyStore) List(ctx context.Context, db migration.DB) ([]migration.Record, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT version, name, checksum, applied_at, execution_time_ms, batch, dirty
		FROM `+tableName+`
		ORDER BY version ASC`)
	if err != nil {
		return nil, fmt.Errorf("postgres: listing history: %w", err)
	}
	defer rows.Close()

	var out []migration.Record
	for rows.Next() {
		var r migration.Record
		var execMS int64
		if err := rows.Scan(&r.Version, &r.Name, &r.Checksum, &r.AppliedAt, &execMS, &r.Batch, &r.Dirty); err != nil {
			return nil, fmt.Errorf("postgres: scanning history row: %w", err)
		}
		r.ExecutionTime = time.Duration(execMS) * time.Millisecond
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: reading history rows: %w", err)
	}
	return out, nil
}

func (historyStore) Begin(ctx context.Context, db migration.DB, version uint64, name, checksum string, batch int) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO `+tableName+` (version, name, checksum, applied_at, execution_time_ms, batch, dirty)
		VALUES ($1, $2, $3, now(), 0, $4, TRUE)`,
		version, name, checksum, batch)
	if err != nil {
		return fmt.Errorf("postgres: recording migration start: %w", err)
	}
	return nil
}

func (historyStore) Complete(ctx context.Context, db migration.DB, version uint64, executionTime time.Duration) error {
	_, err := db.ExecContext(ctx, `
		UPDATE `+tableName+`
		SET dirty = FALSE, execution_time_ms = $2
		WHERE version = $1`,
		version, executionTime.Milliseconds())
	if err != nil {
		return fmt.Errorf("postgres: recording migration completion: %w", err)
	}
	return nil
}

func (historyStore) Remove(ctx context.Context, db migration.DB, version uint64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM `+tableName+` WHERE version = $1`, version)
	if err != nil {
		return fmt.Errorf("postgres: removing history row: %w", err)
	}
	return nil
}
