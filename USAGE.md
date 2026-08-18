# GoForge — Usage

Reference manual for the `goforge` CLI. For the project overview and
architecture, see [README.md](README.md).

- [Install](#install)
- [Global flags](#global-flags)
- [Quick start](#quick-start)
- [Commands](#commands)
- [Configuration (`goforge.yaml`)](#configuration-goforgeyaml)
- [Credentials (`.env`)](#credentials-env)
- [Language (`--lang`)](#language---lang)
- [Migration files](#migration-files)
- [Go migrations](#go-migrations)
- [`goforge doctor`](#goforge-doctor)
- [MCP server](#mcp-server)
- [Exit codes](#exit-codes)
- [`--json` output](#--json-output)

## Install

Standalone binary — no PHP, Node, Python, or a Go toolchain required to run it:

```sh
go install github.com/MMaZX/goforge/cmd/goforge@latest
```

Or download a prebuilt binary + `SHA256SUMS.txt` from
[Releases](https://github.com/MMaZX/goforge/releases) for linux/darwin/windows,
amd64/arm64.

As a Go module, to use the migration engine directly in your own program:

```sh
go get github.com/MMaZX/goforge
```

## Global flags

Every subcommand accepts these (they must come from the root command, i.e.
before or after the subcommand name, standard Cobra flag parsing applies):

| Flag | Default | Meaning |
|---|---|---|
| `--config <path>` | `goforge.yaml` | Path to the config file. |
| `--env-file <path>` | *(alongside `--config`, named `.env`)* | Path to the `.env` file. See [Credentials](#credentials-env). |
| `--json` | `false` | Machine-readable JSON instead of human text. See [`--json` output](#--json-output). |
| `--lang <en\|es>` | *(auto-detected)* | UI language for human output. See [Language](#language---lang). |

## Quick start

```sh
goforge init --driver=postgres        # or mariadb / mysql (alias of mariadb)
goforge make:migration create_users
# edit migrations/000001_create_users.{up,down}.sql
# edit .env: fill in the CHANGE_* placeholders goforge init generated
goforge migrate
goforge migrate:status
```

## Commands

### `goforge init`

Creates `goforge.yaml`, `./migrations`, and (if one doesn't already exist) a
starter `.env` with a ready-to-edit `DATABASE_URL` in the correct syntax for
the driver you chose.

```sh
goforge init --driver=postgres|mariadb|mysql   # default: postgres
```

Fails if `--config` already points to an existing file. Never overwrites an
existing `.env` — if one is already there, it's left untouched and the CLI
tells you so.

### `goforge migrate`

Applies every pending migration, in ascending version order. In interactive terminal sessions, it displays the pending migrations and asks for confirmation (`si` in Spanish, `yes` in English, strictly lowercase). Use `--yes` / `-y` to bypass confirmation in CI/CD or automation.

```sh
goforge migrate [--steps N] [--yes]   # 0 (default) = all pending
```

### `goforge migrate:status`

Lists every known migration and whether it's applied, its batch, and
whether it's flagged dirty.

```sh
goforge migrate:status [--json]
```

### `goforge migrate:rollback`

Reverts the most recently applied batch, or the last N applied migrations
overall with `--steps`. In interactive terminal sessions, it shows a destructive warning and requires **two consecutive confirmations** (`[1/2]` and `[2/2]`, typing `si` or `yes`). Use `--yes` / `-y` to bypass confirmation.

```sh
goforge migrate:rollback [--steps N] [--yes]   # 0 (default) = last batch
```

### `goforge migrate:reset`

Reverts **every** applied migration. Destructive — prompts for **two consecutive confirmations** in interactive sessions, or requires `--yes` / `-y` in non-interactive/CI environments.

```sh
goforge migrate:reset [--yes]
```

### `goforge migrate:fresh`

Reverts every applied migration and re-applies them all from scratch.
Destructive — prompts for **two consecutive confirmations** in interactive sessions, or requires `--yes` / `-y` in non-interactive/CI environments.

```sh
goforge migrate:fresh [--yes]
```

### `goforge make:migration <name>`

Creates the next-numbered migration. SQL by default (a `.up.sql`/`.down.sql`
pair); `--go` creates a Go migration stub instead (see
[Go migrations](#go-migrations)).

```sh
goforge make:migration create_users
goforge make:migration add_computed_column --go
```

### `goforge validate`

Verifies checksums of applied migrations and detects dirty (interrupted)
state, without changing anything. Same checks `goforge doctor` runs as part
of its "Migration history" check.

```sh
goforge validate [--json]
```

### `goforge generate`

Scans `./migrations` for Go migration files and (re)writes
`migrations/goforge_registry_gen.go`, registering each one with its
checksum. Static, source-level discovery only — never interprets `.go`
files at runtime. See [Go migrations](#go-migrations).

```sh
goforge generate
```

### `goforge doctor`

Diagnoses the whole setup in one shot. See [`goforge doctor`](#goforge-doctor).

```sh
goforge doctor [--json]
```

### `goforge version`

```sh
goforge version [--json]
```

### `goforge mcp`

Runs the MCP server over stdio, for AI agents. See [MCP server](#mcp-server).

```sh
goforge mcp
```

## Configuration (`goforge.yaml`)

```yaml
database:
  driver: postgres   # postgres, mariadb, or an alias (postgresql, pg, mysql)
  url: ${DATABASE_URL}

migrations:
  path: ./migrations

language: es   # optional; see Language below
```

`driver` and `url` are required; `migrations.path` defaults to `./migrations`
if omitted. `${VAR}` references are expanded against the process
environment, after `.env` has been loaded (see below) — never write a real
credential directly in this file, since it's meant to be committed.

## Credentials (`.env`)

`${VAR}` references in `goforge.yaml` resolve, in order:

1. A variable already exported in the environment (shell, CI, orchestrator)
   — wins over `.env`.
2. A `.env` file — by default next to `goforge.yaml`, named `.env`; or
   wherever `--env-file <path>` points instead, if you want it somewhere
   else (e.g. `secrets/db.env`, a path your secrets tooling manages).
   `goforge init` respects `--env-file` too: it scaffolds the starter `.env`
   there, creating parent directories as needed.
3. A literal value written directly in `goforge.yaml` — works, but not
   recommended.

No separate credentials file or keychain exists beyond this — standard
12-factor pattern, so the standalone binary works in legacy projects without
imposing its own secrets system. `.env` is never read or written outside of
this resolution; GoForge has no other place it looks for secrets.

Example `.env` for each driver (exactly what `goforge init` generates, with
placeholders to replace):

```sh
# --driver=postgres
DATABASE_URL=postgres://CHANGE_USER:CHANGE_PASSWORD@CHANGE_HOST:5432/CHANGE_DATABASE?sslmode=disable

# --driver=mariadb (or --driver=mysql)
DATABASE_URL=CHANGE_USER:CHANGE_PASSWORD@tcp(CHANGE_HOST:3306)/CHANGE_DATABASE?parseTime=true
```

## Language (`--lang`)

Human-readable output (headers, status table, confirmation prompts, error
messages GoForge itself prints) is available in English and Spanish.
Resolved once at startup, first match wins:

1. `--lang en|es`
2. `GOFORGE_LANG` environment variable
3. `language:` in `goforge.yaml`
4. `LC_ALL`, then `LANG` (system locale, e.g. `es_PE.UTF-8` → `es`)
5. English, if nothing above matched

A key missing from the Spanish catalog falls back to English rather than
breaking. **Not translated, on purpose:** `--json` output, Cobra's own
`--help`/usage text, engine errors from `migration/`, and migration
names/paths — these stay in English so they remain a stable, greppable
contract for scripts, logs, and machine consumers regardless of `--lang`.

## Confirmations and Terminal Output

- **Strict lowercase validation**: Interactive prompts require typing exactly `si` (in Spanish) or `yes` (in English). Any uppercase input (`SI`, `YES`), mixed case, or whitespace will cancel the operation.
- **Double confirmation**: Destructive commands (`migrate:rollback`, `migrate:reset`, `migrate:fresh`) ask for confirmation twice (`[1/2]` and `[2/2]`).
- **Non-interactive environments (CI/CD)**: Running modifying commands without a TTY requires passing `--yes` (or `-y`), otherwise the command fails safely with an error.
- **Color support and `NO_COLOR`**: Output uses lightweight ANSI colors to highlight danger warnings and statuses. If the `NO_COLOR` environment variable is present or stdout is not a TTY, color escape codes are automatically disabled.

## Migration files

SQL migrations, discovered directly under `migrations.path`:

```
000001_create_users.up.sql
000001_create_users.down.sql
```

Both files are required for a given version. The SQL inside is executed
exactly as written — GoForge never rewrites or translates it between
drivers; statements are split on a semicolon at the end of a line.

## Go migrations

The standalone CLI **only** executes SQL migrations directly — it never
parses or interprets `.go` files at runtime. To use Go migrations:

```sh
goforge make:migration add_computed_column --go
goforge generate   # writes migrations/goforge_registry_gen.go
```

Then, in your own program (not the standalone CLI):

```go
reg, _ := migrations.Registry()
entries, _ := migration.Load(os.DirFS("./migrations"), reg)
engine, _ := migration.NewEngine(db, provider, entries)
```

## `goforge doctor`

Runs 6 checks, each reported even if an earlier one failed (except where
that's genuinely impossible), instead of making you interpret the first
error `migrate` happens to hit:

1. **Config** — `goforge.yaml` exists and parses. If this fails, nothing
   else runs.
2. **`.env`** — informational: whether one was found, and where.
3. **Migrations directory** — how many migrations were found, or why
   loading them failed (duplicate version, missing `.up.sql`/`.down.sql`
   pair, etc.).
4. **Database connection** — actually connects, bounded to 10s so an
   unreachable host fails fast instead of hanging; reports the real server
   version and current database name on success.
5. **Migration history** — counts applied/pending/dirty, and runs the same
   checksum/dirty-state validation as `goforge validate`. Skipped if the
   connection or the migrations directory failed.
6. **Locking** — acquires and releases the migration lock
   (`pg_advisory_lock` / `GET_LOCK`), catching a lock a previous crashed run
   left stuck. Skipped if the connection failed.

Exits non-zero if any check failed (skipped checks don't count as failures).

```sh
goforge doctor
goforge doctor --json
goforge doctor --env-file=secrets/db.env
```

## MCP server

```sh
goforge mcp
```

stdio transport, using the official Go MCP SDK. Exposes the same
`migration.Engine` the CLI uses — identical locking, checksum, and
transaction guarantees. Read-only tools need no confirmation:

- `goforge_status`
- `goforge_migration_list`
- `goforge_migration_pending`
- `goforge_migration_validate`
- `goforge_migration_plan`

Schema-modifying tools require an explicit `confirm: true` argument,
enforced by both the JSON Schema (`confirm` is a required field) and the
handler:

- `goforge_migration_up`
- `goforge_migration_rollback`

No tool executes arbitrary SQL — there is no path for an agent to run
`DROP DATABASE`, `TRUNCATE`, or any equivalent; every operation goes through
the migration engine's own reviewed migrations.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success. |
| `1` | The operation failed (migration error, etc.). |
| `2` | Configuration/setup problem (missing or invalid `goforge.yaml`, bad `--driver`, connection failure before even attempting the operation). |

`goforge doctor` is the one exception: it always exits `1` when any check
failed — including a missing `goforge.yaml` — since a failed check is its
normal, expected way of reporting a problem, not a usage error.

## `--json` output

Every command accepts `--json` for a structured, single JSON object on
stdout — stable regardless of `--lang`, meant for scripts and agents. Logs
and diagnostics go to stderr, never mixed into the JSON stream. Example:

```sh
$ goforge migrate:status --json
{
  "status": "ok",
  "migrations": [
    {"version": 1, "name": "create_users", "applied": true, "batch": 1, "applied_at": "2026-08-11T07:25:11Z"}
  ]
}
```
