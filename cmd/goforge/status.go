package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MMaZX/goforge/internal/cliutil"
)

func newStatusCmd(flags *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate:status",
		Short: "Show which migrations are applied and which are pending",
		RunE: func(cmd *cobra.Command, args []string) error {
			engine, conn, _, err := loadEngine(cmd, flags)
			if err != nil {
				return err
			}
			defer conn.Close()

			entries, err := engine.Status(cmd.Context())
			if err != nil {
				return err
			}
			rows := cliutil.FromStatusEntries(entries)
			if flags.json {
				return cliutil.PrintJSON(cmd.OutOrStdout(), map[string]any{"status": "ok", "migrations": rows})
			}
			if len(rows) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No migrations found.")
				return nil
			}
			cliutil.PrintStatusHuman(cmd.OutOrStdout(), rows)
			return nil
		},
	}
}
