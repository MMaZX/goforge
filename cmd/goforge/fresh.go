package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MMaZX/goforge/internal/cliutil"
	"github.com/MMaZX/goforge/internal/i18n"
)

func newFreshCmd(flags *globalFlags) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "migrate:fresh",
		Short: "Revert every applied migration and re-apply them all",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes && (flags.json || !cliutil.IsTerminal(cmd.InOrStdin())) {
				err := errors.New(i18n.T("fresh.confirm"))
				if flags.json {
					cliutil.PrintJSON(cmd.OutOrStdout(), map[string]any{"status": "error", "error": err.Error()})
				}
				return err
			}

			engine, conn, cfg, err := loadEngine(cmd, flags)
			if err != nil {
				return err
			}
			defer conn.Close()

			if !flags.json {
				fmt.Fprint(cmd.OutOrStdout(), i18n.T("app.header", version, driverLabel(cfg.Database.Driver), cfg.Migrations.Path))
			}

			if !yes {
				fmt.Fprint(cmd.OutOrStdout(), cliutil.DangerBold(i18n.T("confirm.fresh.warning")))
				expected := i18n.T("confirm.word")
				step1 := i18n.T("confirm.fresh.step1") + cliutil.Warning(fmt.Sprintf(i18n.T("confirm.type_to_confirm"), expected))
				step2 := "\n" + cliutil.DangerBold(fmt.Sprintf(i18n.T("confirm.fresh.step2"), expected))

				ok, err := cliutil.PromptConfirmation(cmd.InOrStdin(), cmd.OutOrStdout(), []string{step1, step2}, expected)
				if err != nil || !ok {
					fmt.Fprint(cmd.OutOrStdout(), cliutil.Warning(fmt.Sprintf(i18n.T("confirm.cancelled"), expected)))
					return nil
				}
				fmt.Fprintln(cmd.OutOrStdout())
			}

			result, err := engine.Fresh(cmd.Context())
			if err != nil {
				return err
			}

			executed := cliutil.FromRecords(result.Executed)
			if flags.json {
				return cliutil.PrintJSON(cmd.OutOrStdout(), map[string]any{"status": "ok", "batch": result.Batch, "migrations": executed})
			}
			cliutil.PrintExecutedHuman(cmd.OutOrStdout(), executed)
			fmt.Fprint(cmd.OutOrStdout(), cliutil.SuccessBadgeLine(i18n.Tn("migrate.applied", len(executed))))
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "confirm this destructive operation")
	return cmd
}
