package main

import (
	"github.com/spf13/cobra"

	"github.com/MMaZX/goforge/internal/cliutil"
	"github.com/MMaZX/goforge/internal/config"
	"github.com/MMaZX/goforge/internal/i18n"
	"github.com/MMaZX/goforge/internal/providers"
	"github.com/MMaZX/goforge/migration"
)

// applyConfigLanguage re-resolves the UI language once a config is loaded,
// so language: in goforge.yaml participates (still below --lang and
// GOFORGE_LANG in priority).
func applyConfigLanguage(flags *globalFlags, cfg *config.Config) {
	i18n.SetLanguage(i18n.Resolve(flags.lang, cfg.Language))
}

func loadEngine(cmd *cobra.Command, flags *globalFlags) (*migration.Engine, *cliutil.Connection, *config.Config, error) {
	cfg, err := config.Load(flags.configPath, flags.envPath)
	if err != nil {
		return nil, nil, nil, &cliutil.ConfigError{Err: err}
	}
	applyConfigLanguage(flags, cfg)
	engine, conn, err := cliutil.BuildEngine(cmd.Context(), cfg)
	if err != nil {
		return nil, nil, nil, &cliutil.ConfigError{Err: err}
	}
	return engine, conn, cfg, nil
}

func driverLabel(driver string) string {
	if d, err := providers.Resolve(driver); err == nil {
		return d.Label
	}
	return driver
}
