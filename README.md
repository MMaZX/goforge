# GoForge

Portable, reusable database migrations for PostgreSQL and MariaDB — as a Go module, a standalone CLI, and an MCP server for AI agents.

```
                  GoForge
                     │
       ┌─────────────┼──────────────┐
       │             │              │
       ▼             ▼              ▼
   Go Module       CLI            MCP
       │             │              │
       ▼             ▼              ▼
 proyectos Go    legacy PHP     AI Agents
                     │
                     ▼
             PostgreSQL/MariaDB
```

GoForge is not a framework, an ORM, or a query builder. It runs versioned
SQL or Go migrations against PostgreSQL or MariaDB, tracks what has been
applied in a `goforge_migrations` table, and refuses to run when it detects
tampering (checksum mismatch) or a crash mid-migration (dirty state).

## Install

As a standalone binary — no PHP, Node, or Go runtime required to run it:

```sh
go install github.com/MMaZX/goforge/cmd/goforge@latest
```

As a Go module, to build the engine into your own program:

```sh
go get github.com/MMaZX/goforge
```

## Quick start

```sh
goforge init                       # creates goforge.yaml + ./migrations
goforge make:migration create_users
# edit migrations/000001_create_users.{up,down}.sql
goforge migrate
goforge migrate:status
```

`goforge.yaml`:

```yaml
database:
  driver: postgres   # or mariadb
  url: ${DATABASE_URL}

migrations:
  path: ./migrations
```

`DATABASE_URL` can come from the environment or a `.env` file next to
`goforge.yaml`.

## Commands

```
goforge init
goforge migrate               [--steps N]
goforge migrate:status        [--json]
goforge migrate:rollback      [--steps N] [--json]
goforge migrate:reset         --yes
goforge migrate:fresh         --yes
goforge make:migration <name> [--go]
goforge validate              [--json]
goforge generate
goforge version
goforge mcp
```

Every command supports `--json` for machine-readable output.

## Go migrations

The standalone CLI only executes SQL migrations directly — it never parses
or interprets `.go` files at runtime. To use Go migrations, generate their
registry and import the engine into your own program:

```sh
goforge make:migration --go add_computed_column
goforge generate   # writes migrations/goforge_registry_gen.go
```

```go
reg, _ := migrations.Registry()
entries, _ := migration.Load(os.DirFS("./migrations"), reg)
engine, _ := migration.NewEngine(db, provider, entries)
```

## MCP server

```sh
goforge mcp
```

Exposes read-only tools (`goforge_status`, `goforge_migration_list`,
`goforge_migration_pending`, `goforge_migration_validate`,
`goforge_migration_plan`) without confirmation, and mutating tools
(`goforge_migration_up`, `goforge_migration_rollback`) that require an
explicit `confirm: true`. It never executes arbitrary SQL.

## Development

```sh
go test ./...                                  # unit tests
go test -tags=integration ./tests/integration/... -v   # requires Docker
```

## License

MIT
