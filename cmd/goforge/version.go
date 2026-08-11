package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MMaZX/goforge/internal/cliutil"
	"github.com/MMaZX/goforge/internal/i18n"
)

func newVersionCmd(flags *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the GoForge version",
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags.json {
				return cliutil.PrintJSON(cmd.OutOrStdout(), map[string]any{"version": version})
			}
			fmt.Fprint(cmd.OutOrStdout(), i18n.T("app.version", version))
			return nil
		},
	}
}
