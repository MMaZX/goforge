package main

import (
	"github.com/spf13/cobra"
)

// version is set at build time via -ldflags "-X main.version=v0.1.0".
var version = "dev"

type globalFlags struct {
	configPath string
	json       bool
}

func newRootCmd() *cobra.Command {
	flags := &globalFlags{}

	root := &cobra.Command{
		Use:           "goforge",
		Short:         "GoForge: portable database migrations for Go, legacy projects, and AI agents.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVar(&flags.configPath, "config", "goforge.yaml", "path to goforge.yaml")
	root.PersistentFlags().BoolVar(&flags.json, "json", false, "output machine-readable JSON instead of human-readable text")

	root.AddCommand(
		newInitCmd(flags),
		newMigrateCmd(flags),
		newStatusCmd(flags),
		newRollbackCmd(flags),
		newResetCmd(flags),
		newFreshCmd(flags),
		newMakeMigrationCmd(flags),
		newValidateCmd(flags),
		newVersionCmd(flags),
		newGenerateCmd(flags),
		newMCPCmd(flags),
	)
	return root
}
