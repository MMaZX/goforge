package cliutil

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/MMaZX/goforge/internal/i18n"
	"github.com/MMaZX/goforge/migration"
)

// PrintJSON writes v to w as indented JSON, followed by a newline.
func PrintJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// ExecutedMigration is the JSON/human view of one migration.Record produced
// by a migrate/rollback run.
type ExecutedMigration struct {
	Version     uint64 `json:"version"`
	Name        string `json:"name"`
	ExecutionMS int64  `json:"execution_time_ms"`
	AppliedAt   string `json:"applied_at,omitempty"`
}

func fromRecord(r migration.Record) ExecutedMigration {
	em := ExecutedMigration{Version: r.Version, Name: r.Name, ExecutionMS: r.ExecutionTime.Milliseconds()}
	if !r.AppliedAt.IsZero() {
		em.AppliedAt = r.AppliedAt.Format(time.RFC3339)
	}
	return em
}

// FromRecords converts a slice of migration.Record for JSON/human output.
func FromRecords(records []migration.Record) []ExecutedMigration {
	out := make([]ExecutedMigration, 0, len(records))
	for _, r := range records {
		out = append(out, fromRecord(r))
	}
	return out
}

// PrintExecutedHuman prints the "✓ 000001_create_users" lines shown after a
// successful migrate/rollback run. The line carries no words, so it is
// identical in every language.
func PrintExecutedHuman(w io.Writer, executed []ExecutedMigration) {
	for _, m := range executed {
		fmt.Fprintf(w, "✓ %06d_%s\n", m.Version, m.Name)
	}
}

// StatusRow is the JSON/human view of one migration.StatusEntry.
type StatusRow struct {
	Version   uint64 `json:"version"`
	Name      string `json:"name"`
	Applied   bool   `json:"applied"`
	Batch     int    `json:"batch,omitempty"`
	AppliedAt string `json:"applied_at,omitempty"`
	Dirty     bool   `json:"dirty,omitempty"`
}

func FromStatusEntries(entries []migration.StatusEntry) []StatusRow {
	out := make([]StatusRow, 0, len(entries))
	for _, e := range entries {
		row := StatusRow{Version: e.Version, Name: e.Name, Applied: e.Applied, Batch: e.Batch, Dirty: e.Dirty}
		if !e.AppliedAt.IsZero() {
			row.AppliedAt = e.AppliedAt.Format(time.RFC3339)
		}
		out = append(out, row)
	}
	return out
}

// PrintStatusHuman renders a status table for humans, translated into the
// active language. The --json path (FromStatusEntries) is never translated.
func PrintStatusHuman(w io.Writer, rows []StatusRow) {
	for _, r := range rows {
		mark := i18n.T("status.pending")
		if r.Applied {
			mark = i18n.T("status.applied", r.Batch)
			if r.Dirty {
				mark += " " + i18n.T("status.dirty")
			}
		}
		fmt.Fprintf(w, "%06d_%-40s %s\n", r.Version, r.Name, mark)
	}
}
