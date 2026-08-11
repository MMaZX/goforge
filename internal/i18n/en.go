package i18n

// english is the source-of-truth catalog: every key used by the CLI must
// exist here. Other catalogs are merged over it, so a missing translation
// falls back to these strings.
var english = Catalog{
	// App-wide chrome (printed by cmd/goforge/main.go and version.go).
	"app.header":  "GoForge %s\nDatabase: %s\nMigrations: %s\n\n",
	"app.version": "GoForge %s\n",
	"app.error":   "Error: %v\n",

	// goforge init.
	"init.created":        "Created %s and ./migrations\n",
	"init.created_full":   "Created %s, ./migrations and %s\n",
	"init.driver_line":    "Driver: %s\n",
	"init.edit_env":       "Edit %s and replace CHANGE_USER, CHANGE_PASSWORD, CHANGE_HOST and CHANGE_DATABASE with your real credentials:\n  DATABASE_URL=%s\n",
	"init.env_untouched":  "%s already exists, left untouched.\n",
	"init.already_exists": "%s already exists",

	// goforge make:migration / generate / validate.
	"make.created_go":  "Created %s\nRun `goforge generate` to register it.\n",
	"make.created_sql": "Created %s\nCreated %s\n",
	"generate.created": "Generated %s\n",
	"validate.ok":      "Migrations are valid.",

	// goforge migrate / fresh.
	"migrate.nothing":       "Nothing to migrate.",
	"migrate.applied_one":   "\n%d migration applied successfully.\n",
	"migrate.applied_other": "\n%d migrations applied successfully.\n",

	// goforge migrate:rollback.
	"rollback.nothing":    "Nothing to roll back.",
	"rollback.done_one":   "\n%d migration rolled back successfully.\n",
	"rollback.done_other": "\n%d migrations rolled back successfully.\n",

	// goforge migrate:reset.
	"reset.nothing":    "Nothing to reset.",
	"reset.done_one":   "\n%d migration reset successfully.\n",
	"reset.done_other": "\n%d migrations reset successfully.\n",

	// goforge migrate:fresh / migrate:reset without --yes.
	"fresh.confirm": "migrate:fresh reverts and re-applies every migration; re-run with --yes to confirm",
	"reset.confirm": "migrate:reset reverts every applied migration; re-run with --yes to confirm",

	// goforge migrate:status (human table rendered by cliutil).
	"status.empty":   "No migrations found.",
	"status.applied": "✓ applied (batch %d)",
	"status.pending": "✗ pending",
	"status.dirty":   "[DIRTY]",

	// goforge doctor. The check names below are translated only for human
	// output; the --json report keeps the stable English Name/Detail.
	"doctor.header":               "GoForge doctor",
	"doctor.all_passed":           "All checks passed.",
	"doctor.some_failed":          "Some checks failed.",
	"doctor.checks_failed_one":    "%d check failed",
	"doctor.checks_failed_other":  "%d checks failed",
	"doctor.check.config":         "Config",
	"doctor.check.env":            ".env",
	"doctor.check.migrations_dir": "Migrations directory",
	"doctor.check.db_connection":  "Database connection",
	"doctor.check.history":        "Migration history",
	"doctor.check.locking":        "Locking",
}
