package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const defaultConfigTemplate = `database:
  driver: postgres
  url: ${DATABASE_URL}

migrations:
  path: ./migrations
`

func newInitCmd(flags *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create goforge.yaml and the migrations directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat(flags.configPath); err == nil {
				return fmt.Errorf("%s already exists", flags.configPath)
			}
			if err := os.WriteFile(flags.configPath, []byte(defaultConfigTemplate), 0o644); err != nil {
				return fmt.Errorf("writing %s: %w", flags.configPath, err)
			}
			if err := os.MkdirAll("migrations", 0o755); err != nil {
				return fmt.Errorf("creating migrations directory: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created %s and ./migrations\n", flags.configPath)
			return nil
		},
	}
}
