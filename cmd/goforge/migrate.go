package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MMaZX/goforge/internal/cliutil"
	"github.com/MMaZX/goforge/migration"
)

func newMigrateCmd(flags *globalFlags) *cobra.Command {
	var steps int
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Apply pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			engine, conn, cfg, err := loadEngine(cmd, flags)
			if err != nil {
				return err
			}
			defer conn.Close()

			if !flags.json {
				fmt.Fprintf(cmd.OutOrStdout(), "GoForge %s\nDatabase: %s\nMigrations: %s\n\n", version, driverLabel(cfg.Database.Driver), cfg.Migrations.Path)
			}

			result, err := engine.Up(cmd.Context(), steps)
			if errors.Is(err, migration.ErrNoMigrations) {
				if flags.json {
					return cliutil.PrintJSON(cmd.OutOrStdout(), map[string]any{"status": "ok", "migrations": []any{}, "message": "nothing to migrate"})
				}
				fmt.Fprintln(cmd.OutOrStdout(), "Nothing to migrate.")
				return nil
			}
			if err != nil {
				if flags.json {
					cliutil.PrintJSON(cmd.OutOrStdout(), map[string]any{"status": "error", "error": err.Error()})
				}
				return err
			}

			executed := cliutil.FromRecords(result.Executed)
			if flags.json {
				return cliutil.PrintJSON(cmd.OutOrStdout(), map[string]any{"status": "ok", "batch": result.Batch, "migrations": executed})
			}
			cliutil.PrintExecutedHuman(cmd.OutOrStdout(), executed)
			fmt.Fprintf(cmd.OutOrStdout(), "\n%d %s applied successfully.\n", len(executed), cliutil.Plural(len(executed)))
			return nil
		},
	}
	cmd.Flags().IntVar(&steps, "steps", 0, "limit how many pending migrations to apply (0 = all)")
	return cmd
}
