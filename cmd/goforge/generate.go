package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/MMaZX/goforge/internal/cliutil"
	"github.com/MMaZX/goforge/internal/config"
	"github.com/MMaZX/goforge/internal/generator"
)

func newGenerateCmd(flags *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "generate",
		Short: "Generate the registry file for Go migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(flags.configPath, flags.envPath)
			if err != nil {
				return &cliutil.ConfigError{Err: err}
			}
			if err := generator.Generate(cfg.Migrations.Path); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Generated %s\n", filepath.Join(cfg.Migrations.Path, generator.OutputFileName))
			return nil
		},
	}
}
