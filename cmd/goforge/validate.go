package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MMaZX/goforge/internal/cliutil"
	"github.com/MMaZX/goforge/internal/i18n"
)

func newValidateCmd(flags *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Verify migration checksums and detect dirty state",
		RunE: func(cmd *cobra.Command, args []string) error {
			engine, conn, _, err := loadEngine(cmd, flags)
			if err != nil {
				return err
			}
			defer conn.Close()

			if err := engine.Validate(cmd.Context()); err != nil {
				if flags.json {
					cliutil.PrintJSON(cmd.OutOrStdout(), map[string]any{"status": "invalid", "error": err.Error()})
				}
				return err
			}
			if flags.json {
				return cliutil.PrintJSON(cmd.OutOrStdout(), map[string]any{"status": "ok"})
			}
			fmt.Fprintln(cmd.OutOrStdout(), cliutil.Success(i18n.T("validate.ok")))
			return nil
		},
	}
}
