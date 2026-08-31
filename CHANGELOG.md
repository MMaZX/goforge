# Changelog

Todos los cambios notables de este proyecto se documentan en este archivo.

El formato se basa en [Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/)
y el proyecto sigue [Versionado Semántico](https://semver.org/lang/es/).

## [Sin publicar]

## [0.1.0] - 2026-08-31

Primera versión pública de GoForge: migraciones de base de datos portables
para PostgreSQL y MariaDB, como CLI independiente, módulo de Go y servidor
MCP.

### Añadido

- **Motor de migraciones** (`migration.Engine`): aplica y revierte
  migraciones versionadas dentro de una transacción, registra lo aplicado
  en la tabla `goforge_migrations` con su batch y su checksum.
- **Protecciones de integridad**: se niega a ejecutar si detecta un
  checksum alterado o un estado *dirty* de una ejecución anterior que se
  cayó a mitad. Bloqueo entre procesos vía `pg_advisory_lock` (PostgreSQL)
  y `GET_LOCK` (MariaDB).
- **Providers**: PostgreSQL y MariaDB, con sus alias (`postgresql`, `pg`,
  `mysql`) y DSN de ejemplo definidos en un único registro central
  (`internal/providers`).
- **CLI `goforge`** con sintaxis inspirada en `artisan`:
  - `init --driver=postgres|mariadb|mysql` — genera `goforge.yaml`,
    `./migrations` y un `.env` inicial con el `DATABASE_URL` en la sintaxis
    correcta del driver.
  - `make:migration <nombre>` — crea el par `.up.sql`/`.down.sql`
    siguiente, o un stub en Go con `--go`.
  - `migrate [--steps N]` — aplica las migraciones pendientes.
  - `migrate:status [--json]` — lista el estado de cada migración.
  - `migrate:rollback [--steps N]` — revierte el último batch o las
    últimas N migraciones.
  - `migrate:reset` — revierte todas las migraciones aplicadas.
  - `migrate:fresh` — revierte todo y lo vuelve a aplicar desde cero.
  - `validate [--json]` — verifica checksums y estado *dirty* sin cambiar
    nada.
  - `doctor [--json]` — diagnostica la configuración completa en 6
    comprobaciones (config, `.env`, directorio de migraciones, conexión,
    historial, bloqueo).
  - `generate` — (re)genera `migrations/goforge_registry_gen.go` para las
    migraciones en Go, con descubrimiento estático a nivel de código
    fuente.
  - `version`, `mcp`.
- **Confirmaciones interactivas** para comandos destructivos: validación
  estricta en minúsculas (`si` / `yes`), doble confirmación (`[1/2]`,
  `[2/2]`) en `rollback`/`reset`/`fresh`, y `--yes` / `-y` para entornos no
  interactivos y CI.
- **Salida bilingüe** (inglés / español) para el texto legible por
  humanos, vía `internal/i18n`. Resolución del idioma por `--lang`,
  `GOFORGE_LANG`, `language:` en `goforge.yaml` o el locale del sistema.
  La salida `--json`, los errores del motor y los nombres de migraciones
  se quedan en inglés a propósito.
- **Salida de terminal con color** mediante roles semánticos (`Success`,
  `Danger`, `Warning`, `Accent`, `Muted`) e insignias para el veredicto de
  una ejecución. El color solo se emite cuando stdout es una terminal
  real, además de respetar `NO_COLOR` y `TERM=dumb`.
- **Salida `--json`** en todos los comandos: un único objeto JSON en
  stdout, estable sin importar `--lang`, con logs y diagnósticos siempre en
  stderr.
- **Configuración `goforge.yaml`** con expansión de `${VAR}` resuelta desde
  el entorno del proceso, un `.env` (por defecto junto al config, o vía
  `--env-file <ruta>`) o un valor literal. Patrón de 12 factores, sin
  keychain propio.
- **Servidor MCP** (`goforge mcp`) sobre stdio con el SDK oficial de MCP
  para Go: herramientas de solo lectura (`goforge_status`,
  `goforge_migration_list`, `goforge_migration_pending`,
  `goforge_migration_validate`, `goforge_migration_plan`) sin confirmación,
  y herramientas que mutan estado (`goforge_migration_up`,
  `goforge_migration_rollback`) que exigen `confirm: true` explícito.
  Ninguna herramienta ejecuta SQL arbitrario.
- **Migraciones en Go** embebibles en tu propio programa a través de
  `migrations.Registry()` + `migration.Load` + `migration.NewEngine`.
- **Documentación** en español: `README.md` (visión general, propósito y
  comparación con otros migradores) y `USAGE.md` (manual de referencia del
  CLI). Los nombres de comandos, flags, claves de configuración, salida
  `--json` y ejemplos de código se mantienen en inglés como contrato
  estable para scripts y agentes. `CONTRIBUTING.md` con el flujo de
  contribución y las comprobaciones que exige la CI.
- **CI** (GitHub Actions): `go build`, `go vet`, `gofmt`, `go test`, tests
  de integración con Docker (PostgreSQL y MariaDB), y build multiplataforma
  para linux/darwin/windows en amd64/arm64.
- **Workflow de release**: al subir un tag `vX.Y.Z` se compilan los
  binarios de las 6 plataformas, se genera `SHA256SUMS.txt` y se publica la
  GitHub Release con notas automáticas.

[Sin publicar]: https://github.com/MMaZX/goforge/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/MMaZX/goforge/releases/tag/v0.1.0
