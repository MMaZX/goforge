package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MMaZX/goforge/internal/cliutil"
)

func newFreshCmd(flags *globalFlags) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "migrate:fresh",
		Short: "Revert every applied migration and re-apply them all",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				return fmt.Errorf("migrate:fresh reverts and re-applies every migration; re-run with --yes to confirm")
			}
			engine, conn, _, err := loadEngine(cmd, flags)
			if err != nil {
				return err
			}
			defer conn.Close()

			result, err := engine.Fresh(cmd.Context())
			if err != nil {
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
	cmd.Flags().BoolVar(&yes, "yes", false, "confirm this destructive operation")
	return cmd
}
