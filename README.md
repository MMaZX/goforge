# GoForge

Migraciones de base de datos portables y reutilizables para PostgreSQL y MariaDB — como módulo de Go, como CLI independiente y como servidor MCP para agentes de IA.

```
                  GoForge
                     │
       ┌─────────────┼──────────────┐
       │             │              │
       ▼             ▼              ▼
   Módulo Go        CLI            MCP
       │             │              │
       ▼             ▼              ▼
 proyectos Go    PHP legacy     agentes de IA
                     │
                     ▼
             PostgreSQL/MariaDB
```

## Para qué sirve

GoForge es una herramienta de **migración de esquema y datos de bases de datos**:
te ayuda a versionar los cambios de tu base de datos y aplicarlos de forma
controlada y repetible.

El objetivo es que puedas **incorporarla a cualquier proyecto** — uno legacy que
arrastras desde hace años o uno que estás empezando hoy — **sin reinventar la
rueda** montando tu propio sistema de migraciones de propósito general. Es
especialmente útil en proyectos que no tienen un migrador fuerte (o no tienen
ninguno): scripts sueltos de PHP, un servicio pequeño de Go, un monolito sin
framework, etc.

Como se distribuye en un **único binario** y un **CLI**, puede usarse en
cualquier entorno sin depender de PHP, Node ni de un runtime de Go. Y en
proyectos de Go puedes además embeber el motor como módulo.

La sintaxis de los comandos de migración (`migrate`, `migrate:rollback`,
`migrate:fresh`, `make:migration`, …) está inspirada directamente en
`artisan migrate` de Laravel, para que resulte familiar de inmediato.

GoForge no es un framework, ni un ORM, ni un query builder. Ejecuta migraciones
versionadas en SQL o en Go contra PostgreSQL o MariaDB, registra lo que ya se ha
aplicado en una tabla `goforge_migrations`, y se niega a ejecutar cuando detecta
manipulación (checksum que no coincide) o una caída a mitad de una migración
(estado *dirty*).

Referencia completa de comandos, flags y ejemplos: [USAGE.md](USAGE.md).

## ¿Por qué GoForge y no otro?

Existen muchos migradores para Go. GoForge tiene sentido cuando valoras
alguna de estas cosas:

| | GoForge |
|---|---|
| **Sintaxis** | Comandos tipo Laravel: `migrate`, `migrate:rollback`, `migrate:fresh`, `make:migration`. Familiar desde el primer minuto si vienes de `artisan`. |
| **Distribución** | Un único binario estático, sin dependencias de runtime. Se deja caer en un proyecto PHP, un script, un contenedor, un CI, lo que sea. |
| **Proyectos legacy** | Pensado para entrar en un proyecto que *no* tiene migrador, sin imponer un framework ni un sistema de secretos propio (usa `.env` de 12 factores). |
| **Doble uso** | El mismo motor como CLI **y** como módulo de Go embebible (`migration.Engine`), con migraciones en SQL o en Go. |
| **Agentes de IA** | Servidor MCP integrado (`goforge mcp`): un agente puede consultar estado y aplicar migraciones de forma acotada, sin poder ejecutar SQL arbitrario. |
| **Seguridad de datos** | Se niega a ejecutar ante checksum alterado o estado *dirty* de una caída previa; bloqueo por `pg_advisory_lock` / `GET_LOCK`. |

Si ya usas `golang-migrate`, `goose` o `atlas` y te funcionan, no hay razón
para cambiar. GoForge no compite en features de esquema declarativo ni en
número de motores soportados (por ahora, solo PostgreSQL y MariaDB).

## Instalación

Como binario independiente — no requiere PHP, Node ni un runtime de Go para
ejecutarse:

```sh
go install github.com/MMaZX/goforge/cmd/goforge@latest
```

Como módulo de Go, para compilar el motor dentro de tu propio programa:

```sh
go get github.com/MMaZX/goforge
```

## Inicio rápido

```sh
goforge init --driver=postgres     # o mariadb / mysql (alias de mariadb)
goforge make:migration create_users
# edita migrations/000001_create_users.{up,down}.sql
# edita .env: rellena los marcadores CHANGE_* que goforge init generó por ti
goforge migrate
goforge migrate:status
```

`goforge init --driver=<nombre>` crea tres cosas:

- `goforge.yaml`, con `database.driver` fijado al driver que elegiste y
  `database.url: ${DATABASE_URL}` — nunca una credencial literal.
- `./migrations`.
- `.env` (solo si no existe ya uno junto a `goforge.yaml`), con un
  `DATABASE_URL` ya escrito en la sintaxis correcta para ese driver, por ejemplo:

  ```sh
  # --driver=postgres
  DATABASE_URL=postgres://CHANGE_USER:CHANGE_PASSWORD@CHANGE_HOST:5432/CHANGE_DATABASE?sslmode=disable

  # --driver=mariadb (o --driver=mysql)
  DATABASE_URL=CHANGE_USER:CHANGE_PASSWORD@tcp(CHANGE_HOST:3306)/CHANGE_DATABASE?parseTime=true
  ```

  PostgreSQL y MariaDB usan sintaxis de DSN no relacionadas (una URL frente al
  formato propio de `go-sql-driver/mysql`), que es exactamente el motivo por el
  que `init` te lo escribe en lugar de dejar que lo busques. Sustituye los
  marcadores `CHANGE_*` por tus credenciales reales — `.env` ya está en
  `.gitignore` y nunca se commitea.

Cada driver que GoForge soporta — nombre canónico, alias, etiqueta legible y DSN
de ejemplo — está definido en un solo lugar: `internal/providers`. Ningún otro
punto del código fija la lista a mano.

## Configuración y credenciales

`goforge.yaml`:

```yaml
database:
  driver: postgres   # postgres, mariadb, o un alias (postgresql, pg, mysql)
  url: ${DATABASE_URL}

migrations:
  path: ./migrations
```

Las referencias `${VAR}` en `goforge.yaml` se resuelven, en este orden, desde:

1. Una variable ya exportada en el entorno (por tu shell, tu CI o tu
   orquestador) — tiene prioridad sobre `.env`.
2. Un archivo `.env` junto a `goforge.yaml` (se carga automáticamente, es
   opcional y está en `.gitignore`) — o donde apunte `--env-file <ruta>` en su
   lugar, si no lo quieres junto a `goforge.yaml` (por ejemplo `secrets/db.env`,
   una ruta gestionada por tu tooling de secretos, etc.). `goforge init` también
   respeta `--env-file`: genera el `.env` inicial (con los marcadores `CHANGE_*`)
   en esa ruta en vez de en la ubicación por defecto, creando los directorios
   padre que hagan falta.
3. Un valor literal escrito directamente en `goforge.yaml` — funciona, pero no
   se recomienda: ese archivo está pensado para commitearse.

No existe ningún archivo de credenciales ni keychain aparte de esto — es el
patrón estándar de 12 factores, elegido para que el CLI independiente funcione en
proyectos legacy sin imponer su propio sistema de secretos.

## Comandos

Todos los comandos aceptan además `--config <ruta>` (por defecto `goforge.yaml`)
y `--env-file <ruta>` (por defecto: junto a `--config`, con nombre `.env`).

```
goforge init [--driver=postgres|mariadb|mysql]
goforge migrate               [--steps N]
goforge migrate:status        [--json]
goforge migrate:rollback      [--steps N] [--json]
goforge migrate:reset         --yes
goforge migrate:fresh         --yes
goforge make:migration <nombre> [--go]
goforge validate              [--json]
goforge doctor                 [--json]
goforge generate
goforge version
goforge mcp
```

Todos los comandos admiten `--json` para salida legible por máquina.

## `goforge doctor`

Diagnostica la configuración de un proyecto de principio a fin en lugar de
obligarte a interpretar el primer error que le toque encontrar a `migrate`.
Comprueba, en este orden — cada comprobación se ejecuta aunque una anterior haya
fallado, salvo donde eso sea genuinamente imposible:

1. **Config** — `goforge.yaml` existe y parsea.
2. **`.env`** — informativo: si se encontró uno junto a `goforge.yaml`.
3. **Directorio de migraciones** — cuántas migraciones se encontraron, o por qué
   falló su carga (versión duplicada, falta el par `.up.sql`/`.down.sql`, etc.).
4. **Conexión a la base de datos** — conecta de verdad (con un límite de 10s para
   que un host inalcanzable falle rápido en lugar de colgarse), e informa de la
   versión real del servidor y del nombre de la base de datos actual.
5. **Historial de migraciones** — cuántas están aplicadas/pendientes/dirty, y
   ejecuta la misma validación de checksum/estado dirty que `goforge validate`.
6. **Bloqueo** — adquiere y libera el lock de migración
   (`pg_advisory_lock` / `GET_LOCK`), para detectar un lock que se quedó
   atascado tras una ejecución anterior que se cayó.

Las comprobaciones 5 y 6 se omiten (no fallan) cuando la propia conexión a la
base de datos no tuvo éxito. Sale con código distinto de cero si alguna
comprobación falló.

## Migraciones en Go

El CLI independiente solo ejecuta migraciones SQL directamente — nunca parsea ni
interpreta archivos `.go` en tiempo de ejecución. Para usar migraciones en Go,
genera su registro e importa el motor dentro de tu propio programa:

```sh
goforge make:migration --go add_computed_column
goforge generate   # escribe migrations/goforge_registry_gen.go
```

```go
reg, _ := migrations.Registry()
entries, _ := migration.Load(os.DirFS("./migrations"), reg)
engine, _ := migration.NewEngine(db, provider, entries)
```

## Servidor MCP

```sh
goforge mcp
```

Expone herramientas de solo lectura (`goforge_status`, `goforge_migration_list`,
`goforge_migration_pending`, `goforge_migration_validate`,
`goforge_migration_plan`) sin confirmación, y herramientas que mutan estado
(`goforge_migration_up`, `goforge_migration_rollback`) que exigen un
`confirm: true` explícito. Nunca ejecuta SQL arbitrario.

## Desarrollo

```sh
go test ./...                                  # tests unitarios
go test -tags=integration ./tests/integration/... -v   # requiere Docker
```

## Licencia

MIT
