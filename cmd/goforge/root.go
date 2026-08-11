package main

import (
	"github.com/spf13/cobra"

	"github.com/MMaZX/goforge/internal/i18n"
)

// version is set at build time via -ldflags "-X main.version=v0.1.0".
var version = "dev"

type globalFlags struct {
	configPath string
	envPath    string
	json       bool
	lang       string
}

func newRootCmd() *cobra.Command {
	flags := &globalFlags{}

	root := &cobra.Command{
		Use:           "goforge",
		Short:         "GoForge: portable database migrations for Go, legacy projects, and AI agents.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Resolve without the config file here (flag > env > system
			// locale); commands that load goforge.yaml re-resolve with
			// cfg.Language (see loadEngine / runDoctor).
			i18n.SetLanguage(i18n.Resolve(flags.lang, ""))
			return nil
		},
	}
	root.PersistentFlags().StringVar(&flags.configPath, "config", "goforge.yaml", "path to goforge.yaml")
	root.PersistentFlags().StringVar(&flags.envPath, "env-file", "", "path to the .env file (default: alongside --config, named .env)")
	root.PersistentFlags().BoolVar(&flags.json, "json", false, "output machine-readable JSON instead of human-readable text")
	root.PersistentFlags().StringVar(&flags.lang, "lang", "", "UI language for human output: en or es (default: GOFORGE_LANG, language: in goforge.yaml, or system locale)")

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
		newDoctorCmd(flags),
	)
	return root
}
