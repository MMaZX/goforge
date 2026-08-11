package main

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/MMaZX/goforge/internal/cliutil"
	"github.com/MMaZX/goforge/internal/config"
	"github.com/MMaZX/goforge/migration"
)

func loadEngine(cmd *cobra.Command, flags *globalFlags) (*migration.Engine, *cliutil.Connection, *config.Config, error) {
	cfg, err := config.Load(flags.configPath)
	if err != nil {
		return nil, nil, nil, &cliutil.ConfigError{Err: err}
	}
	engine, conn, err := cliutil.BuildEngine(cmd.Context(), cfg)
	if err != nil {
		return nil, nil, nil, &cliutil.ConfigError{Err: err}
	}
	return engine, conn, cfg, nil
}

func driverLabel(driver string) string {
	switch driver {
	case "postgres":
		return "PostgreSQL"
	case "mariadb":
		return "MariaDB"
	default:
		return strings.ToUpper(driver[:1]) + driver[1:]
	}
}
