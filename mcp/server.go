// Package mcp implements the GoForge MCP server: it exposes the same
// migration.Engine the CLI uses to AI agents over stdio. Read-only tools run
// without confirmation; every tool that modifies the database schema
// requires an explicit confirm:true argument. No tool executes arbitrary
// SQL, so there is no path for an agent to run DROP DATABASE, TRUNCATE, or
// any other statement outside of the project's own reviewed migrations.
package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MMaZX/goforge/internal/config"
	"github.com/MMaZX/goforge/migration"
)

// NewServer builds the GoForge MCP server bound to engine, using cfg for
// display metadata (driver name, migrations path).
func NewServer(engine *migration.Engine, cfg *config.Config, version string) *sdk.Server {
	s := sdk.NewServer(&sdk.Implementation{Name: "goforge", Version: version}, nil)

	sdk.AddTool(s, &sdk.Tool{
		Name:        "goforge_status",
		Description: "Summarize migration state: driver, path, and how many migrations are applied/pending/dirty. Read-only.",
	}, statusHandler(engine, cfg))

	sdk.AddTool(s, &sdk.Tool{
		Name:        "goforge_migration_list",
		Description: "List every known migration with its applied state. Read-only.",
	}, listHandler(engine))

	sdk.AddTool(s, &sdk.Tool{
		Name:        "goforge_migration_pending",
		Description: "List migrations that have not been applied yet. Read-only.",
	}, pendingHandler(engine))

	sdk.AddTool(s, &sdk.Tool{
		Name:        "goforge_migration_validate",
		Description: "Verify checksums of applied migrations and detect dirty (interrupted) state. Read-only.",
	}, validateHandler(engine))

	sdk.AddTool(s, &sdk.Tool{
		Name:        "goforge_migration_plan",
		Description: "Show exactly which migrations `up` or `rollback` would run, without modifying the database. Read-only.",
	}, planHandler(engine))

	sdk.AddTool(s, &sdk.Tool{
		Name:        "goforge_migration_up",
		Description: "Apply pending migrations. Modifies the database schema. Requires confirm:true.",
	}, upHandler(engine))

	sdk.AddTool(s, &sdk.Tool{
		Name:        "goforge_migration_rollback",
		Description: "Revert previously applied migrations. Modifies the database schema. Requires confirm:true.",
	}, rollbackHandler(engine))

	return s
}

func statusHandler(engine *migration.Engine, cfg *config.Config) sdk.ToolHandlerFor[Empty, StatusOutput] {
	return func(ctx context.Context, _ *sdk.CallToolRequest, _ Empty) (*sdk.CallToolResult, StatusOutput, error) {
		entries, err := engine.Status(ctx)
		if err != nil {
			return nil, StatusOutput{}, err
		}
		out := StatusOutput{Driver: cfg.Database.Driver, MigrationsPath: cfg.Migrations.Path, Total: len(entries)}
		for _, e := range entries {
			if e.Applied {
				out.Applied++
			} else {
				out.Pending++
			}
			if e.Dirty {
				out.Dirty = true
			}
		}
		return nil, out, nil
	}
}

func listHandler(engine *migration.Engine) sdk.ToolHandlerFor[Empty, ListOutput] {
	return func(ctx context.Context, _ *sdk.CallToolRequest, _ Empty) (*sdk.CallToolResult, ListOutput, error) {
		entries, err := engine.Status(ctx)
		if err != nil {
			return nil, ListOutput{}, err
		}
		return nil, ListOutput{Migrations: rowsFromStatus(entries)}, nil
	}
}

func pendingHandler(engine *migration.Engine) sdk.ToolHandlerFor[Empty, PendingOutput] {
	return func(ctx context.Context, _ *sdk.CallToolRequest, _ Empty) (*sdk.CallToolResult, PendingOutput, error) {
		entries, err := engine.Status(ctx)
		if err != nil {
			return nil, PendingOutput{}, err
		}
		var pending []migration.StatusEntry
		for _, e := range entries {
			if !e.Applied {
				pending = append(pending, e)
			}
		}
		return nil, PendingOutput{Migrations: rowsFromStatus(pending)}, nil
	}
}

func validateHandler(engine *migration.Engine) sdk.ToolHandlerFor[Empty, ValidateOutput] {
	return func(ctx context.Context, _ *sdk.CallToolRequest, _ Empty) (*sdk.CallToolResult, ValidateOutput, error) {
		if err := engine.Validate(ctx); err != nil {
			return nil, ValidateOutput{Valid: false, Error: err.Error()}, nil
		}
		return nil, ValidateOutput{Valid: true}, nil
	}
}

func planHandler(engine *migration.Engine) sdk.ToolHandlerFor[PlanInput, PlanOutput] {
	return func(ctx context.Context, _ *sdk.CallToolRequest, in PlanInput) (*sdk.CallToolResult, PlanOutput, error) {
		dir := migration.Up
		direction := "up"
		if in.Direction == "down" {
			dir = migration.Down
			direction = "down"
		} else if in.Direction != "" && in.Direction != "up" {
			return nil, PlanOutput{}, fmt.Errorf("mcp: direction must be \"up\" or \"down\", got %q", in.Direction)
		}
		entries, err := engine.Plan(ctx, dir, in.Steps)
		if err != nil {
			return nil, PlanOutput{}, err
		}
		return nil, PlanOutput{Direction: direction, Migrations: rowsFromEntries(entries)}, nil
	}
}

func upHandler(engine *migration.Engine) sdk.ToolHandlerFor[RunInput, RunOutput] {
	return func(ctx context.Context, _ *sdk.CallToolRequest, in RunInput) (*sdk.CallToolResult, RunOutput, error) {
		if !in.Confirm {
			return nil, RunOutput{}, fmt.Errorf("mcp: goforge_migration_up modifies the database schema; call again with confirm:true")
		}
		result, err := engine.Up(ctx, in.Steps)
		if err != nil {
			return nil, RunOutput{}, err
		}
		return nil, RunOutput{Batch: result.Batch, Migrations: rowsFromRecords(result.Executed)}, nil
	}
}

func rollbackHandler(engine *migration.Engine) sdk.ToolHandlerFor[RunInput, RunOutput] {
	return func(ctx context.Context, _ *sdk.CallToolRequest, in RunInput) (*sdk.CallToolResult, RunOutput, error) {
		if !in.Confirm {
			return nil, RunOutput{}, fmt.Errorf("mcp: goforge_migration_rollback modifies the database schema; call again with confirm:true")
		}
		result, err := engine.Rollback(ctx, in.Steps)
		if err != nil {
			return nil, RunOutput{}, err
		}
		return nil, RunOutput{Migrations: rowsFromRecords(result.Executed)}, nil
	}
}
