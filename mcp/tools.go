package mcp

import "github.com/MMaZX/goforge/migration"

// Empty is the input type for tools that take no arguments.
type Empty struct{}

// MigrationRow is the MCP-facing view of one migration's state.
type MigrationRow struct {
	Version   uint64 `json:"version"`
	Name      string `json:"name"`
	Applied   bool   `json:"applied"`
	Batch     int    `json:"batch,omitempty"`
	AppliedAt string `json:"applied_at,omitempty"`
	Dirty     bool   `json:"dirty,omitempty"`
}

func rowsFromStatus(entries []migration.StatusEntry) []MigrationRow {
	out := make([]MigrationRow, 0, len(entries))
	for _, e := range entries {
		row := MigrationRow{Version: e.Version, Name: e.Name, Applied: e.Applied, Batch: e.Batch, Dirty: e.Dirty}
		if !e.AppliedAt.IsZero() {
			row.AppliedAt = e.AppliedAt.Format(rfc3339)
		}
		out = append(out, row)
	}
	return out
}

func rowsFromEntries(entries []migration.Entry) []MigrationRow {
	out := make([]MigrationRow, 0, len(entries))
	for _, e := range entries {
		out = append(out, MigrationRow{Version: e.Version(), Name: e.Name()})
	}
	return out
}

func rowsFromRecords(records []migration.Record) []MigrationRow {
	out := make([]MigrationRow, 0, len(records))
	for _, r := range records {
		row := MigrationRow{Version: r.Version, Name: r.Name, Applied: true, Batch: r.Batch}
		if !r.AppliedAt.IsZero() {
			row.AppliedAt = r.AppliedAt.Format(rfc3339)
		}
		out = append(out, row)
	}
	return out
}

const rfc3339 = "2006-01-02T15:04:05Z07:00"

// StatusOutput summarizes the migration state for goforge_status.
type StatusOutput struct {
	Driver         string `json:"driver"`
	MigrationsPath string `json:"migrations_path"`
	Total          int    `json:"total"`
	Applied        int    `json:"applied"`
	Pending        int    `json:"pending"`
	Dirty          bool   `json:"dirty"`
}

// ListOutput is the result of goforge_migration_list.
type ListOutput struct {
	Migrations []MigrationRow `json:"migrations"`
}

// PendingOutput is the result of goforge_migration_pending.
type PendingOutput struct {
	Migrations []MigrationRow `json:"migrations"`
}

// ValidateOutput is the result of goforge_migration_validate.
type ValidateOutput struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}

// PlanInput selects the direction and scope for goforge_migration_plan.
type PlanInput struct {
	Direction string `json:"direction,omitempty" jsonschema:"either 'up' or 'down'; defaults to 'up'"`
	Steps     int    `json:"steps,omitempty" jsonschema:"limit how many migrations to include in the plan; 0 means all"`
}

// PlanOutput is the result of goforge_migration_plan: what would run,
// without having modified the database.
type PlanOutput struct {
	Direction  string         `json:"direction"`
	Migrations []MigrationRow `json:"migrations"`
}

// RunInput is shared by goforge_migration_up and goforge_migration_rollback.
// Confirm must be explicitly set to true: these tools modify the database
// schema, and GoForge never runs a schema-modifying operation an agent
// triggered without explicit, per-call confirmation.
type RunInput struct {
	Steps   int  `json:"steps,omitempty" jsonschema:"limit how many migrations to run; 0 means all pending (up) or the last batch (rollback)"`
	Confirm bool `json:"confirm" jsonschema:"must be true; this call runs schema-modifying migrations against the database"`
}

// RunOutput is the result of goforge_migration_up / goforge_migration_rollback.
type RunOutput struct {
	Batch      int            `json:"batch,omitempty"`
	Migrations []MigrationRow `json:"migrations"`
}
