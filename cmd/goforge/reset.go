package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MMaZX/goforge/internal/cliutil"
	"github.com/MMaZX/goforge/internal/i18n"
	"github.com/MMaZX/goforge/migration"
)

func newResetCmd(flags *globalFlags) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "migrate:reset",
		Short: "Revert every applied migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes && (flags.json || !cliutil.IsTerminal(cmd.InOrStdin())) {
				err := errors.New(i18n.T("reset.confirm"))
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
				fmt.Fprint(cmd.OutOrStdout(), cliutil.BoldRed(i18n.T("confirm.reset.warning")))
				expected := i18n.T("confirm.word")
				step1 := i18n.T("confirm.reset.step1") + cliutil.Yellow(fmt.Sprintf(i18n.T("confirm.type_to_confirm"), expected))
				step2 := "\n" + cliutil.BoldRed(fmt.Sprintf(i18n.T("confirm.reset.step2"), expected))

				ok, err := cliutil.PromptConfirmation(cmd.InOrStdin(), cmd.OutOrStdout(), []string{step1, step2}, expected)
				if err != nil || !ok {
					fmt.Fprint(cmd.OutOrStdout(), cliutil.Yellow(fmt.Sprintf(i18n.T("confirm.cancelled"), expected)))
					return nil
				}
				fmt.Fprintln(cmd.OutOrStdout())
			}

			result, err := engine.Reset(cmd.Context())
			if errors.Is(err, migration.ErrNoMigrations) {
				if flags.json {
					return cliutil.PrintJSON(cmd.OutOrStdout(), map[string]any{"status": "ok", "migrations": []any{}, "message": "nothing to reset"})
				}
				fmt.Fprintln(cmd.OutOrStdout(), i18n.T("reset.nothing"))
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
			fmt.Fprint(cmd.OutOrStdout(), i18n.Tn("reset.done", len(executed)))
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "confirm this destructive operation")
	return cmd
}
