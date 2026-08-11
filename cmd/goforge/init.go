package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MMaZX/goforge/internal/cliutil"
	"github.com/MMaZX/goforge/internal/config"
	"github.com/MMaZX/goforge/internal/i18n"
	"github.com/MMaZX/goforge/internal/providers"
)

func newInitCmd(flags *globalFlags) *cobra.Command {
	var driverFlag string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create goforge.yaml, ./migrations and a starter .env",
		RunE: func(cmd *cobra.Command, args []string) error {
			desc, err := providers.Resolve(driverFlag)
			if err != nil {
				return &cliutil.ConfigError{Err: fmt.Errorf("--driver: %w", err)}
			}

			if _, err := os.Stat(flags.configPath); err == nil {
				return errors.New(i18n.T("init.already_exists", flags.configPath))
			}
			if err := os.WriteFile(flags.configPath, []byte(configTemplate(desc.Driver)), 0o644); err != nil {
				return fmt.Errorf("writing %s: %w", flags.configPath, err)
			}
			if err := os.MkdirAll("migrations", 0o755); err != nil {
				return fmt.Errorf("creating migrations directory: %w", err)
			}

			envPath := config.ResolveEnvPath(flags.configPath, flags.envPath)
			envCreated := false
			if _, err := os.Stat(envPath); os.IsNotExist(err) {
				if dir := filepath.Dir(envPath); dir != "." {
					if err := os.MkdirAll(dir, 0o755); err != nil {
						return fmt.Errorf("creating %s: %w", dir, err)
					}
				}
				if err := os.WriteFile(envPath, []byte(envTemplate(desc)), 0o600); err != nil {
					return fmt.Errorf("writing %s: %w", envPath, err)
				}
				envCreated = true
			} else if err != nil {
				return fmt.Errorf("checking %s: %w", envPath, err)
			}

			if flags.json {
				return cliutil.PrintJSON(cmd.OutOrStdout(), map[string]any{
					"status":      "ok",
					"driver":      desc.Driver,
					"config":      flags.configPath,
					"migrations":  "migrations",
					"env":         envPath,
					"env_created": envCreated,
					"example_dsn": desc.ExampleDSN,
				})
			}

			out := cmd.OutOrStdout()
			if envCreated {
				fmt.Fprint(out, i18n.T("init.created_full", flags.configPath, envPath))
				fmt.Fprint(out, i18n.T("init.driver_line", desc.Label))
				fmt.Fprintln(out)
				fmt.Fprint(out, i18n.T("init.edit_env", envPath, desc.ExampleDSN))
			} else {
				fmt.Fprint(out, i18n.T("init.created", flags.configPath))
				fmt.Fprint(out, i18n.T("init.driver_line", desc.Label))
				fmt.Fprint(out, i18n.T("init.env_untouched", envPath))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&driverFlag, "driver", "postgres", "database driver: "+strings.Join(providers.Names(), ", "))
	return cmd
}

func configTemplate(driver string) string {
	return fmt.Sprintf(`database:
  driver: %s
  url: ${DATABASE_URL}

migrations:
  path: ./migrations
`, driver)
}

func envTemplate(desc providers.Descriptor) string {
	return fmt.Sprintf("DATABASE_URL=%s\n", desc.ExampleDSN)
}
