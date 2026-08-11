package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MMaZX/goforge/internal/cliutil"
	"github.com/MMaZX/goforge/migration"
)

func newResetCmd(flags *globalFlags) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "migrate:reset",
		Short: "Revert every applied migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				return fmt.Errorf("migrate:reset reverts every applied migration; re-run with --yes to confirm")
			}
			engine, conn, _, err := loadEngine(cmd, flags)
			if err != nil {
				return err
			}
			defer conn.Close()

			result, err := engine.Reset(cmd.Context())
			if errors.Is(err, migration.ErrNoMigrations) {
				if flags.json {
					return cliutil.PrintJSON(cmd.OutOrStdout(), map[string]any{"status": "ok", "migrations": []any{}, "message": "nothing to reset"})
				}
				fmt.Fprintln(cmd.OutOrStdout(), "Nothing to reset.")
				return nil
			}
			if err != nil {
				return err
			}

			executed := cliutil.FromRecords(result.Executed)
			if flags.json {
				return cliutil.PrintJSON(cmd.OutOrStdout(), map[string]any{"status": "ok", "migrations": executed})
			}
			cliutil.PrintExecutedHuman(cmd.OutOrStdout(), executed)
			fmt.Fprintf(cmd.OutOrStdout(), "\n%d %s reset successfully.\n", len(executed), cliutil.Plural(len(executed)))
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "confirm this destructive operation")
	return cmd
}
