package i18n

// spanish holds the Spanish translations. Keys missing here fall back to
// the English catalog at SetLanguage time.
var spanish = Catalog{
	"app.header":  "GoForge %s\nBase de datos: %s\nMigraciones: %s\n\n",
	"app.version": "GoForge %s\n",
	"app.error":   "Error: %v\n",

	"init.created":        "Se creó %s y ./migrations\n",
	"init.created_full":   "Se crearon %s, ./migrations y %s\n",
	"init.driver_line":    "Driver: %s\n",
	"init.edit_env":       "Edita %s y reemplaza CHANGE_USER, CHANGE_PASSWORD, CHANGE_HOST y CHANGE_DATABASE con tus credenciales reales:\n  DATABASE_URL=%s\n",
	"init.env_untouched":  "%s ya existe, se dejó intacto.\n",
	"init.already_exists": "%s ya existe",

	"make.created_go":  "Se creó %s\nEjecuta `goforge generate` para registrarla.\n",
	"make.created_sql": "Se creó %s\nSe creó %s\n",
	"generate.created": "Se generó %s\n",
	"validate.ok":      "Las migraciones son válidas.",

	"migrate.nothing":       "No hay nada que migrar.",
	"migrate.applied_one":   "\n%d migración aplicada correctamente.\n",
	"migrate.applied_other": "\n%d migraciones aplicadas correctamente.\n",

	"rollback.nothing":    "No hay nada que revertir.",
	"rollback.done_one":   "\n%d migración revertida correctamente.\n",
	"rollback.done_other": "\n%d migraciones revertidas correctamente.\n",

	"reset.nothing":    "No hay nada que restablecer.",
	"reset.done_one":   "\n%d migración restablecida correctamente.\n",
	"reset.done_other": "\n%d migraciones restablecidas correctamente.\n",

	"fresh.confirm": "migrate:fresh revierte y vuelve a aplicar todas las migraciones; vuelve a ejecutar con --yes para confirmar",
	"reset.confirm": "migrate:reset revierte todas las migraciones aplicadas; vuelve a ejecutar con --yes para confirmar",

	"status.empty":   "No se encontraron migraciones.",
	"status.applied": "✓ aplicada (lote %d)",
	"status.pending": "✗ pendiente",
	"status.dirty":   "[DIRTY]",

	"doctor.header":               "GoForge doctor",
	"doctor.all_passed":           "Todas las verificaciones pasaron.",
	"doctor.some_failed":          "Algunas verificaciones fallaron.",
	"doctor.checks_failed_one":    "%d verificación falló",
	"doctor.checks_failed_other":  "%d verificaciones fallaron",
	"doctor.check.config":         "Configuración",
	"doctor.check.env":            ".env",
	"doctor.check.migrations_dir": "Directorio de migraciones",
	"doctor.check.db_connection":  "Conexión a la base de datos",
	"doctor.check.history":        "Historial de migraciones",
	"doctor.check.locking":        "Bloqueo (lock)",
}
