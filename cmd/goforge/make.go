package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MMaZX/goforge/internal/cliutil"
	"github.com/MMaZX/goforge/internal/config"
)

var migrationFileRE = regexp.MustCompile(`^(\d{6,})_`)

func newMakeMigrationCmd(flags *globalFlags) *cobra.Command {
	var asGo bool
	cmd := &cobra.Command{
		Use:   "make:migration <name>",
		Short: "Create a new migration file pair (or a Go migration stub)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flags.configPath)
			if err != nil {
				return &cliutil.ConfigError{Err: err}
			}
			slug := slugify(args[0])
			nextVersion, err := nextMigrationVersion(cfg.Migrations.Path)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(cfg.Migrations.Path, 0o755); err != nil {
				return fmt.Errorf("creating migrations directory: %w", err)
			}

			base := fmt.Sprintf("%06d_%s", nextVersion, slug)
			if asGo {
				path := filepath.Join(cfg.Migrations.Path, base+".go")
				if err := os.WriteFile(path, []byte(goMigrationTemplate(nextVersion, slug)), 0o644); err != nil {
					return fmt.Errorf("writing %s: %w", path, err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Created %s\nRun `goforge generate` to register it.\n", path)
				return nil
			}

			upPath := filepath.Join(cfg.Migrations.Path, base+".up.sql")
			downPath := filepath.Join(cfg.Migrations.Path, base+".down.sql")
			if err := os.WriteFile(upPath, []byte("-- write your up migration here\n"), 0o644); err != nil {
				return fmt.Errorf("writing %s: %w", upPath, err)
			}
			if err := os.WriteFile(downPath, []byte("-- write your down migration here\n"), 0o644); err != nil {
				return fmt.Errorf("writing %s: %w", downPath, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created %s\nCreated %s\n", upPath, downPath)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asGo, "go", false, "create a Go migration stub instead of SQL files")
	cmd.Flags().Bool("sql", true, "create SQL migration files (default)")
	return cmd
}

func slugify(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	lastUnderscore := false
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			b.WriteRune(r)
			lastUnderscore = false
		default:
			if !lastUnderscore {
				b.WriteByte('_')
				lastUnderscore = true
			}
		}
	}
	return strings.Trim(b.String(), "_")
}

func nextMigrationVersion(dir string) (uint64, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return 1, nil
	}
	if err != nil {
		return 0, fmt.Errorf("reading %s: %w", dir, err)
	}
	var max uint64
	for _, e := range entries {
		m := migrationFileRE.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		v, err := strconv.ParseUint(m[1], 10, 64)
		if err == nil && v > max {
			max = v
		}
	}
	return max + 1, nil
}

func goMigrationTemplate(version uint64, slug string) string {
	typeName := fmt.Sprintf("migration%06d", version)
	return fmt.Sprintf(`package migrations

import (
	"context"

	"github.com/MMaZX/goforge/migration"
)

type %[1]s struct{}

func (%[1]s) Version() uint64 { return %[2]d }
func (%[1]s) Name() string    { return %[3]q }

func (%[1]s) Up(ctx context.Context, db migration.DB) error {
	// TODO: implement
	return nil
}

func (%[1]s) Down(ctx context.Context, db migration.DB) error {
	// TODO: implement
	return nil
}
`, typeName, version, slug)
}
