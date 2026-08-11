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
goforge init --driver=postgres     # or mariadb / mysql (an alias of mariadb)
goforge make:migration create_users
# edit migrations/000001_create_users.{up,down}.sql
# edit .env: fill in the CHANGE_* placeholders goforge init generated for you
goforge migrate
goforge migrate:status
```

`goforge init --driver=<name>` creates three things:

- `goforge.yaml`, with `database.driver` set to the driver you chose and
  `database.url: ${DATABASE_URL}` — never a literal credential.
- `./migrations`.
- `.env` (only if one doesn't already exist next to `goforge.yaml`), with a
  `DATABASE_URL` already in the correct syntax for that driver, e.g.:

  ```sh
  # --driver=postgres
  DATABASE_URL=postgres://CHANGE_USER:CHANGE_PASSWORD@CHANGE_HOST:5432/CHANGE_DATABASE?sslmode=disable

  # --driver=mariadb (or --driver=mysql)
  DATABASE_URL=CHANGE_USER:CHANGE_PASSWORD@tcp(CHANGE_HOST:3306)/CHANGE_DATABASE?parseTime=true
  ```

  PostgreSQL and MariaDB use unrelated DSN syntaxes (a URL vs.
  `go-sql-driver/mysql`'s own format), which is exactly why `init` writes it
  out for you instead of leaving you to look it up. Replace the `CHANGE_*`
  placeholders with your real credentials — `.env` is already in
  `.gitignore` and is never committed.

Every driver GoForge supports — canonical name, aliases, human label, and
example DSN — is defined in one place: `internal/providers`. Nothing else in
the codebase hardcodes the list.

## Configuration and credentials

`goforge.yaml`:

```yaml
database:
  driver: postgres   # postgres, mariadb, or an alias (postgresql, pg, mysql)
  url: ${DATABASE_URL}

migrations:
  path: ./migrations
```

`${VAR}` references in `goforge.yaml` are resolved, in order, from:

1. A variable already exported in the environment (e.g. by your shell, CI,
   or orchestrator) — takes precedence over `.env`.
2. A `.env` file next to `goforge.yaml` (auto-loaded, optional, git-ignored)
   — or wherever `--env-file <path>` points instead, if you don't want it
   next to `goforge.yaml` (e.g. `secrets/db.env`, a path managed by your
   secrets tooling, etc.). `goforge init` also respects `--env-file`: it
   scaffolds the starter `.env` (with the `CHANGE_*` placeholders) at that
   path instead of the default location, creating parent directories as
   needed.
3. A literal value written directly in `goforge.yaml` — works, but not
   recommended: that file is meant to be committed.

No separate credentials file or keychain exists beyond this — it's the
standard 12-factor pattern, chosen so the standalone CLI works in legacy
projects without imposing its own secrets system.

## Commands

Every command also accepts `--config <path>` (default `goforge.yaml`) and
`--env-file <path>` (default: alongside `--config`, named `.env`).

```
goforge init [--driver=postgres|mariadb|mysql]
goforge migrate               [--steps N]
goforge migrate:status        [--json]
goforge migrate:rollback      [--steps N] [--json]
goforge migrate:reset         --yes
goforge migrate:fresh         --yes
goforge make:migration <name> [--go]
goforge validate              [--json]
goforge doctor                 [--json]
goforge generate
goforge version
goforge mcp
```

Every command supports `--json` for machine-readable output.

## `goforge doctor`

Diagnoses a project's setup end-to-end instead of making you interpret the
first error `migrate` happens to hit. It checks, in order — each one runs
even if an earlier one failed, except where that's genuinely impossible:

1. **Config** — `goforge.yaml` exists and parses.
2. **`.env`** — informational: whether one was found next to `goforge.yaml`.
3. **Migrations directory** — how many migrations were found, or why loading
   them failed (duplicate version, missing `.up.sql`/`.down.sql` pair, etc.).
4. **Database connection** — actually connects (bounded to 10s so an
   unreachable host fails fast instead of hanging), and reports the real
   server version and current database name.
5. **Migration history** — how many are applied/pending/dirty, and runs the
   same checksum/dirty-state validation as `goforge validate`.
6. **Locking** — acquires and releases the migration lock
   (`pg_advisory_lock` / `GET_LOCK`), to catch a lock left stuck by a
   previous crashed run.

Checks 5 and 6 are skipped (not failed) when the database connection itself
didn't succeed. Exits non-zero if any check failed.

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
