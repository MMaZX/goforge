package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MMaZX/goforge/internal/cliutil"
	"github.com/MMaZX/goforge/migration"
)

func newRollbackCmd(flags *globalFlags) *cobra.Command {
	var steps int
	cmd := &cobra.Command{
		Use:   "migrate:rollback",
		Short: "Revert the last applied batch of migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			engine, conn, _, err := loadEngine(cmd, flags)
			if err != nil {
				return err
			}
			defer conn.Close()

			result, err := engine.Rollback(cmd.Context(), steps)
			if errors.Is(err, migration.ErrNoMigrations) {
				if flags.json {
					return cliutil.PrintJSON(cmd.OutOrStdout(), map[string]any{"status": "ok", "migrations": []any{}, "message": "nothing to roll back"})
				}
				fmt.Fprintln(cmd.OutOrStdout(), "Nothing to roll back.")
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
				return cliutil.PrintJSON(cmd.OutOrStdout(), map[string]any{"status": "ok", "migrations": executed})
			}
			cliutil.PrintExecutedHuman(cmd.OutOrStdout(), executed)
			fmt.Fprintf(cmd.OutOrStdout(), "\n%d %s rolled back successfully.\n", len(executed), cliutil.Plural(len(executed)))
			return nil
		},
	}
	cmd.Flags().IntVar(&steps, "steps", 0, "how many applied migrations to roll back (0 = last batch)")
	return cmd
}
