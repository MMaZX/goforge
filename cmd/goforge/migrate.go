package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MMaZX/goforge/internal/cliutil"
	"github.com/MMaZX/goforge/internal/i18n"
	"github.com/MMaZX/goforge/migration"
)

func newMigrateCmd(flags *globalFlags) *cobra.Command {
	var steps int
	var yes bool
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Apply pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes && (flags.json || !cliutil.IsTerminal(cmd.InOrStdin())) {
				err := errors.New(i18n.T("confirm.non_tty_error"))
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

			plan, err := engine.Plan(cmd.Context(), migration.Up, steps)
			if err != nil {
				if flags.json {
					cliutil.PrintJSON(cmd.OutOrStdout(), map[string]any{"status": "error", "error": err.Error()})
				}
				return err
			}
			if len(plan) == 0 {
				if flags.json {
					return cliutil.PrintJSON(cmd.OutOrStdout(), map[string]any{"status": "ok", "migrations": []any{}, "message": "nothing to migrate"})
				}
				fmt.Fprintln(cmd.OutOrStdout(), i18n.T("migrate.nothing"))
				return nil
			}

			if !yes {
				fmt.Fprint(cmd.OutOrStdout(), cliutil.Bold(i18n.Tn("confirm.migrate.header", len(plan))))
				for _, entry := range plan {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s %06d_%s\n", cliutil.Accent(">"), entry.Version(), entry.Name())
				}
				fmt.Fprintln(cmd.OutOrStdout())

				expected := i18n.T("confirm.word")
				promptText := i18n.T("confirm.migrate.prompt") + cliutil.Warning(fmt.Sprintf(i18n.T("confirm.type_to_confirm"), expected))
				ok, err := cliutil.PromptConfirmation(cmd.InOrStdin(), cmd.OutOrStdout(), []string{promptText}, expected)
				if err != nil || !ok {
					fmt.Fprint(cmd.OutOrStdout(), cliutil.Warning(fmt.Sprintf(i18n.T("confirm.cancelled"), expected)))
					return nil
				}
				fmt.Fprintln(cmd.OutOrStdout())
			}

			result, err := engine.Up(cmd.Context(), steps)
			if errors.Is(err, migration.ErrNoMigrations) {
				if flags.json {
					return cliutil.PrintJSON(cmd.OutOrStdout(), map[string]any{"status": "ok", "migrations": []any{}, "message": "nothing to migrate"})
				}
				fmt.Fprintln(cmd.OutOrStdout(), i18n.T("migrate.nothing"))
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
			fmt.Fprint(cmd.OutOrStdout(), cliutil.SuccessBadgeLine(i18n.Tn("migrate.applied", len(executed))))
			return nil
		},
	}
	cmd.Flags().IntVar(&steps, "steps", 0, "limit how many pending migrations to apply (0 = all)")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip interactive confirmation")
	return cmd
}
