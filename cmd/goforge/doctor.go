package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/MMaZX/goforge/internal/cliutil"
	"github.com/MMaZX/goforge/internal/config"
	"github.com/MMaZX/goforge/internal/i18n"
	"github.com/MMaZX/goforge/internal/providers"
	"github.com/MMaZX/goforge/migration"
)

const (
	doctorConnectTimeout = 10 * time.Second
	doctorLockTimeout    = 5 * time.Second
)

func newDoctorCmd(flags *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose the database connection and migrations setup",
		RunE: func(cmd *cobra.Command, args []string) error {
			report := runDoctor(cmd.Context(), flags)

			if flags.json {
				if err := cliutil.PrintJSON(cmd.OutOrStdout(), report); err != nil {
					return err
				}
			} else {
				cliutil.PrintDoctorHuman(cmd.OutOrStdout(), report)
			}

			if !report.Healthy {
				failed := 0
				for _, c := range report.Checks {
					if !c.Skipped && !c.OK {
						failed++
					}
				}
				return errors.New(i18n.Tn("doctor.checks_failed", failed))
			}
			return nil
		},
	}
}

func runDoctor(ctx context.Context, flags *globalFlags) cliutil.DoctorReport {
	var checks []cliutil.DoctorCheck
	add := func(c cliutil.DoctorCheck) { checks = append(checks, c) }
	healthy := func() bool {
		for _, c := range checks {
			if !c.Skipped && !c.OK {
				return false
			}
		}
		return true
	}

	cfg, err := config.Load(flags.configPath, flags.envPath)
	if err != nil {
		add(cliutil.DoctorCheck{
			Name:   "Config",
			OK:     false,
			Detail: fmt.Sprintf("%s\nRun `goforge init` to create one.", err),
		})
		return cliutil.DoctorReport{Healthy: false, Checks: checks}
	}
	applyConfigLanguage(flags, cfg)
	desc, _ := providers.Resolve(cfg.Database.Driver) // already canonicalized by config.Load
	add(cliutil.DoctorCheck{
		Name:   "Config",
		OK:     true,
		Detail: fmt.Sprintf("%s (driver: %s, migrations: %s)", flags.configPath, desc.Label, cfg.Migrations.Path),
	})

	envPath := config.ResolveEnvPath(flags.configPath, flags.envPath)
	if _, err := os.Stat(envPath); err == nil {
		add(cliutil.DoctorCheck{Name: ".env", OK: true, Detail: fmt.Sprintf("found at %s", envPath)})
	} else {
		add(cliutil.DoctorCheck{Name: ".env", OK: true, Detail: "not found, relying on exported environment variables"})
	}

	entries, err := cliutil.LoadMigrations(cfg)
	migrationsOK := err == nil
	if err != nil {
		add(cliutil.DoctorCheck{Name: "Migrations directory", OK: false, Detail: err.Error()})
	} else {
		add(cliutil.DoctorCheck{
			Name:   "Migrations directory",
			OK:     true,
			Detail: fmt.Sprintf("%s (%d found)", cfg.Migrations.Path, len(entries)),
		})
	}

	connectCtx, cancel := context.WithTimeout(ctx, doctorConnectTimeout)
	start := time.Now()
	conn, err := cliutil.Connect(connectCtx, cfg)
	cancel()
	if err != nil {
		add(cliutil.DoctorCheck{Name: "Database connection", OK: false, Detail: err.Error()})
		add(cliutil.DoctorCheck{Name: "Migration history", Skipped: true, Detail: "database connection failed"})
		add(cliutil.DoctorCheck{Name: "Locking", Skipped: true, Detail: "database connection failed"})
		return cliutil.DoctorReport{Healthy: healthy(), Checks: checks}
	}
	defer conn.Close()
	elapsed := time.Since(start)

	detail := fmt.Sprintf("Connected in %s", elapsed.Round(time.Millisecond))
	var version, database string
	if err := conn.DB.QueryRowContext(ctx, desc.VersionQuery).Scan(&version); err == nil {
		detail += fmt.Sprintf("\nServer version: %s", version)
	}
	if err := conn.DB.QueryRowContext(ctx, desc.CurrentDatabaseQuery).Scan(&database); err == nil {
		detail += fmt.Sprintf("\nCurrent database: %s", database)
	}
	add(cliutil.DoctorCheck{Name: "Database connection", OK: true, Detail: detail})

	if !migrationsOK {
		add(cliutil.DoctorCheck{Name: "Migration history", Skipped: true, Detail: "migrations directory has errors, see above"})
	} else {
		engine, err := migration.NewEngine(conn.DB, conn.Provider, entries)
		if err != nil {
			add(cliutil.DoctorCheck{Name: "Migration history", OK: false, Detail: err.Error()})
		} else {
			status, statusErr := engine.Status(ctx)
			validateErr := engine.Validate(ctx)
			applied, pending, dirty := 0, 0, 0
			for _, s := range status {
				if s.Applied {
					applied++
				} else {
					pending++
				}
				if s.Dirty {
					dirty++
				}
			}
			switch {
			case statusErr != nil:
				add(cliutil.DoctorCheck{Name: "Migration history", OK: false, Detail: statusErr.Error()})
			case validateErr != nil:
				add(cliutil.DoctorCheck{
					Name: "Migration history",
					OK:   false,
					Detail: fmt.Sprintf("%d applied, %d pending, %d dirty\n%s",
						applied, pending, dirty, validateErr),
				})
			default:
				add(cliutil.DoctorCheck{
					Name:   "Migration history",
					OK:     true,
					Detail: fmt.Sprintf("%d applied, %d pending, %d dirty", applied, pending, dirty),
				})
			}
		}
	}

	lockCtx, lockCancel := context.WithTimeout(ctx, doctorLockTimeout)
	locker := conn.Provider.Locker(conn.DB)
	if err := locker.Lock(lockCtx); err != nil {
		add(cliutil.DoctorCheck{Name: "Locking", OK: false, Detail: err.Error()})
	} else if err := locker.Unlock(lockCtx); err != nil {
		add(cliutil.DoctorCheck{Name: "Locking", OK: false, Detail: err.Error()})
	} else {
		add(cliutil.DoctorCheck{Name: "Locking", OK: true, Detail: "acquired and released successfully"})
	}
	lockCancel()

	return cliutil.DoctorReport{Healthy: healthy(), Checks: checks}
}
