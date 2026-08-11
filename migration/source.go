package migration

import (
	"context"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var sqlFileRE = regexp.MustCompile(`^(\d{6,})_(.+)\.(up|down)\.sql$`)

// LoadSQLMigrations discovers *.up.sql / *.down.sql pairs directly under the
// root of fsys, named "000001_create_users.up.sql". Both files must exist
// for a given version. The developer's SQL is executed as written; GoForge
// never rewrites or translates it between providers.
func LoadSQLMigrations(fsys fs.FS) ([]Entry, error) {
	type pair struct {
		name     string
		up, down []byte
		haveUp   bool
		haveDown bool
	}
	pairs := make(map[uint64]*pair)

	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("migration: reading migrations directory: %w", err)
	}

	for _, de := range entries {
		if de.IsDir() {
			continue
		}
		m := sqlFileRE.FindStringSubmatch(de.Name())
		if m == nil {
			continue
		}
		version, err := strconv.ParseUint(m[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("migration: invalid version in %q: %w", de.Name(), err)
		}
		name, direction := m[2], m[3]

		content, err := fs.ReadFile(fsys, de.Name())
		if err != nil {
			return nil, fmt.Errorf("migration: reading %q: %w", de.Name(), err)
		}

		p, ok := pairs[version]
		if !ok {
			p = &pair{name: name}
			pairs[version] = p
		}
		if p.name != name {
			return nil, fmt.Errorf("migration: version %d has mismatched names %q and %q", version, p.name, name)
		}
		switch direction {
		case "up":
			p.up, p.haveUp = content, true
		case "down":
			p.down, p.haveDown = content, true
		}
	}

	out := make([]Entry, 0, len(pairs))
	for version, p := range pairs {
		if !p.haveUp {
			return nil, fmt.Errorf("migration: version %d (%s) is missing its .up.sql file", version, p.name)
		}
		if !p.haveDown {
			return nil, fmt.Errorf("migration: version %d (%s) is missing its .down.sql file", version, p.name)
		}
		checksum := Checksum(append(append([]byte{}, p.up...), p.down...))
		out = append(out, Entry{
			Migration: &sqlMigration{version: version, name: p.name, upSQL: string(p.up), downSQL: string(p.down)},
			Checksum:  checksum,
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Version() < out[j].Version() })
	return out, nil
}

type sqlMigration struct {
	version        uint64
	name           string
	upSQL, downSQL string
}

func (m *sqlMigration) Version() uint64 { return m.version }
func (m *sqlMigration) Name() string    { return m.name }

func (m *sqlMigration) Up(ctx context.Context, db DB) error {
	return execSQLScript(ctx, db, m.upSQL)
}

func (m *sqlMigration) Down(ctx context.Context, db DB) error {
	return execSQLScript(ctx, db, m.downSQL)
}

// execSQLScript runs a migration file's statements in order. Statements
// must be separated by a semicolon at the end of a line; this simple rule
// keeps GoForge from having to parse dialect-specific SQL, but means a
// semicolon inside a string or a multi-line function body must not sit at
// end-of-line.
func execSQLScript(ctx context.Context, db DB, script string) error {
	for _, stmt := range splitStatements(script) {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migration: executing statement: %w", err)
		}
	}
	return nil
}

func splitStatements(script string) []string {
	raw := strings.Split(script, ";\n")
	stmts := make([]string, 0, len(raw))
	for _, s := range raw {
		s = strings.TrimSpace(s)
		s = strings.TrimSuffix(s, ";")
		if s == "" {
			continue
		}
		stmts = append(stmts, s)
	}
	return stmts
}
