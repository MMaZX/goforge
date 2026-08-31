# GoForge — Uso

Manual de referencia del CLI `goforge`. Para la visión general del proyecto y
la arquitectura, consulta [README.md](README.md).

- [Instalación](#instalación)
- [Flags globales](#flags-globales)
- [Inicio rápido](#inicio-rápido)
- [Comandos](#comandos)
- [Configuración (`goforge.yaml`)](#configuración-goforgeyaml)
- [Credenciales (`.env`)](#credenciales-env)
- [Idioma (`--lang`)](#idioma---lang)
- [Archivos de migración](#archivos-de-migración)
- [Migraciones en Go](#migraciones-en-go)
- [`goforge doctor`](#goforge-doctor)
- [Servidor MCP](#servidor-mcp)
- [Códigos de salida](#códigos-de-salida)
- [Salida `--json`](#salida---json)

## Instalación

Binario independiente — no requiere PHP, Node, Python ni un toolchain de Go para
ejecutarse:

```sh
go install github.com/MMaZX/goforge/cmd/goforge@latest
```

O descarga un binario precompilado + `SHA256SUMS.txt` desde
[Releases](https://github.com/MMaZX/goforge/releases) para linux/darwin/windows,
amd64/arm64.

Como módulo de Go, para usar el motor de migraciones directamente en tu propio
programa:

```sh
go get github.com/MMaZX/goforge
```

## Flags globales

Todos los subcomandos aceptan estas flags (deben ir en el comando raíz, es
decir, antes o después del nombre del subcomando; se aplica el parseo de flags
estándar de Cobra):

| Flag | Por defecto | Significado |
|---|---|---|
| `--config <ruta>` | `goforge.yaml` | Ruta al archivo de configuración. |
| `--env-file <ruta>` | *(junto a `--config`, con nombre `.env`)* | Ruta al archivo `.env`. Ver [Credenciales](#credenciales-env). |
| `--json` | `false` | JSON legible por máquina en lugar de texto para humanos. Ver [Salida `--json`](#salida---json). |
| `--lang <en\|es>` | *(autodetectado)* | Idioma de la interfaz para la salida humana. Ver [Idioma](#idioma---lang). |

## Inicio rápido

```sh
goforge init --driver=postgres        # o mariadb / mysql (alias de mariadb)
goforge make:migration create_users
# edita migrations/000001_create_users.{up,down}.sql
# edita .env: rellena los marcadores CHANGE_* que goforge init generó
goforge migrate
goforge migrate:status
```

## Comandos

### `goforge init`

Crea `goforge.yaml`, `./migrations` y (si no existe ya) un `.env` inicial con un
`DATABASE_URL` listo para editar en la sintaxis correcta para el driver que
elegiste.

```sh
goforge init --driver=postgres|mariadb|mysql   # por defecto: postgres
```

Falla si `--config` ya apunta a un archivo existente. Nunca sobrescribe un
`.env` existente — si ya hay uno, lo deja intacto y el CLI te lo indica.

### `goforge migrate`

Aplica todas las migraciones pendientes, en orden ascendente de versión. En
sesiones de terminal interactivas, muestra las migraciones pendientes y pide
confirmación (`si` en español, `yes` en inglés, estrictamente en minúsculas).
Usa `--yes` / `-y` para saltarte la confirmación en CI/CD o automatización.

```sh
goforge migrate [--steps N] [--yes]   # 0 (por defecto) = todas las pendientes
```

### `goforge migrate:status`

Lista todas las migraciones conocidas y si están aplicadas, su batch, y si
están marcadas como dirty.

```sh
goforge migrate:status [--json]
```

### `goforge migrate:rollback`

Revierte el último batch aplicado, o las últimas N migraciones aplicadas en
total con `--steps`. En sesiones de terminal interactivas, muestra una
advertencia destructiva y exige **dos confirmaciones consecutivas** (`[1/2]` y
`[2/2]`, escribiendo `si` o `yes`). Usa `--yes` / `-y` para saltarte la
confirmación.

```sh
goforge migrate:rollback [--steps N] [--yes]   # 0 (por defecto) = último batch
```

### `goforge migrate:reset`

Revierte **todas** las migraciones aplicadas. Destructivo — pide **dos
confirmaciones consecutivas** en sesiones interactivas, o requiere `--yes` /
`-y` en entornos no interactivos / CI.

```sh
goforge migrate:reset [--yes]
```

### `goforge migrate:fresh`

Revierte todas las migraciones aplicadas y las vuelve a aplicar desde cero.
Destructivo — pide **dos confirmaciones consecutivas** en sesiones
interactivas, o requiere `--yes` / `-y` en entornos no interactivos / CI.

```sh
goforge migrate:fresh [--yes]
```

### `goforge make:migration <nombre>`

Crea la siguiente migración numerada. SQL por defecto (un par
`.up.sql`/`.down.sql`); `--go` crea en su lugar un stub de migración en Go (ver
[Migraciones en Go](#migraciones-en-go)).

```sh
goforge make:migration create_users
goforge make:migration add_computed_column --go
```

### `goforge validate`

Verifica los checksums de las migraciones aplicadas y detecta estado dirty
(interrumpido), sin cambiar nada. Son las mismas comprobaciones que
`goforge doctor` ejecuta como parte de su comprobación de "Historial de
migraciones".

```sh
goforge validate [--json]
```

### `goforge generate`

Escanea `./migrations` en busca de archivos de migración en Go y (re)escribe
`migrations/goforge_registry_gen.go`, registrando cada uno con su checksum.
Descubrimiento estático, a nivel de código fuente únicamente — nunca interpreta
archivos `.go` en tiempo de ejecución. Ver
[Migraciones en Go](#migraciones-en-go).

```sh
goforge generate
```

### `goforge doctor`

Diagnostica toda la configuración de una sola vez. Ver
[`goforge doctor`](#goforge-doctor).

```sh
goforge doctor [--json]
```

### `goforge version`

```sh
goforge version [--json]
```

### `goforge mcp`

Ejecuta el servidor MCP sobre stdio, para agentes de IA. Ver
[Servidor MCP](#servidor-mcp).

```sh
goforge mcp
```

## Configuración (`goforge.yaml`)

```yaml
database:
  driver: postgres   # postgres, mariadb, o un alias (postgresql, pg, mysql)
  url: ${DATABASE_URL}

migrations:
  path: ./migrations

language: es   # opcional; ver Idioma más abajo
```

`driver` y `url` son obligatorios; `migrations.path` toma el valor por defecto
`./migrations` si se omite. Las referencias `${VAR}` se expanden contra el
entorno del proceso, después de haber cargado `.env` (ver más abajo) — nunca
escribas una credencial real directamente en este archivo, ya que está pensado
para commitearse.

## Credenciales (`.env`)

Las referencias `${VAR}` en `goforge.yaml` se resuelven, en este orden:

1. Una variable ya exportada en el entorno (shell, CI, orquestador) — gana
   sobre `.env`.
2. Un archivo `.env` — por defecto junto a `goforge.yaml`, con nombre `.env`; o
   donde apunte `--env-file <ruta>` en su lugar, si lo quieres en otro sitio
   (por ejemplo `secrets/db.env`, una ruta que gestione tu tooling de secretos).
   `goforge init` también respeta `--env-file`: genera ahí el `.env` inicial,
   creando los directorios padre que hagan falta.
3. Un valor literal escrito directamente en `goforge.yaml` — funciona, pero no
   se recomienda.

No existe ningún archivo de credenciales ni keychain aparte de esto — patrón
estándar de 12 factores, para que el binario independiente funcione en proyectos
legacy sin imponer su propio sistema de secretos. `.env` nunca se lee ni se
escribe fuera de esta resolución; GoForge no tiene ningún otro lugar donde
busque secretos.

Ejemplo de `.env` para cada driver (exactamente lo que genera `goforge init`,
con marcadores para reemplazar):

```sh
# --driver=postgres
DATABASE_URL=postgres://CHANGE_USER:CHANGE_PASSWORD@CHANGE_HOST:5432/CHANGE_DATABASE?sslmode=disable

# --driver=mariadb (o --driver=mysql)
DATABASE_URL=CHANGE_USER:CHANGE_PASSWORD@tcp(CHANGE_HOST:3306)/CHANGE_DATABASE?parseTime=true
```

## Idioma (`--lang`)

La salida legible por humanos (cabeceras, tabla de estado, prompts de
confirmación, mensajes de error que imprime el propio GoForge) está disponible
en inglés y español. Se resuelve una sola vez al arrancar; gana la primera
coincidencia:

1. `--lang en|es`
2. Variable de entorno `GOFORGE_LANG`
3. `language:` en `goforge.yaml`
4. `LC_ALL`, luego `LANG` (locale del sistema, p. ej. `es_PE.UTF-8` → `es`)
5. Inglés, si nada de lo anterior coincidió

Una clave que falte en el catálogo español recae en inglés en lugar de romper.
**No se traduce, a propósito:** la salida `--json`, el texto de
`--help`/uso propio de Cobra, los errores del motor en `migration/`, y los
nombres/rutas de migraciones — se mantienen en inglés para que sigan siendo un
contrato estable y grepeable para scripts, logs y consumidores máquina, sin
importar `--lang`.

## Confirmaciones y salida de terminal

- **Validación estricta en minúsculas**: los prompts interactivos exigen
  escribir exactamente `si` (en español) o `yes` (en inglés). Cualquier entrada
  en mayúsculas (`SI`, `YES`), mayúsculas y minúsculas mezcladas, o espacios en
  blanco cancelará la operación.
- **Doble confirmación**: los comandos destructivos (`migrate:rollback`,
  `migrate:reset`, `migrate:fresh`) piden confirmación dos veces (`[1/2]` y
  `[2/2]`).
- **Entornos no interactivos (CI/CD)**: ejecutar comandos que modifican estado
  sin un TTY requiere pasar `--yes` (o `-y`); de lo contrario el comando falla
  de forma segura con un error.
- **Soporte de color y `NO_COLOR`**: la salida usa colores ANSI ligeros para
  resaltar advertencias de peligro y estados. Si la variable de entorno
  `NO_COLOR` está presente o stdout no es un TTY, los códigos de escape de color
  se desactivan automáticamente.

## Archivos de migración

Migraciones SQL, descubiertas directamente bajo `migrations.path`:

```
000001_create_users.up.sql
000001_create_users.down.sql
```

Ambos archivos son obligatorios para una versión dada. El SQL de dentro se
ejecuta exactamente como está escrito — GoForge nunca lo reescribe ni lo traduce
entre drivers; las sentencias se separan por un punto y coma al final de una
línea.

## Migraciones en Go

El CLI independiente **solo** ejecuta migraciones SQL directamente — nunca
parsea ni interpreta archivos `.go` en tiempo de ejecución. Para usar
migraciones en Go:

```sh
goforge make:migration add_computed_column --go
goforge generate   # escribe migrations/goforge_registry_gen.go
```

Después, en tu propio programa (no en el CLI independiente):

```go
reg, _ := migrations.Registry()
entries, _ := migration.Load(os.DirFS("./migrations"), reg)
engine, _ := migration.NewEngine(db, provider, entries)
```

## `goforge doctor`

Ejecuta 6 comprobaciones, cada una reportada aunque una anterior haya fallado
(salvo donde eso sea genuinamente imposible), en lugar de obligarte a
interpretar el primer error que le toque encontrar a `migrate`:

1. **Config** — `goforge.yaml` existe y parsea. Si esto falla, no se ejecuta
   nada más.
2. **`.env`** — informativo: si se encontró uno, y dónde.
3. **Directorio de migraciones** — cuántas migraciones se encontraron, o por
   qué falló su carga (versión duplicada, falta el par `.up.sql`/`.down.sql`,
   etc.).
4. **Conexión a la base de datos** — conecta de verdad, con un límite de 10s
   para que un host inalcanzable falle rápido en lugar de colgarse; en caso de
   éxito informa de la versión real del servidor y del nombre de la base de
   datos actual.
5. **Historial de migraciones** — cuenta aplicadas/pendientes/dirty, y ejecuta
   la misma validación de checksum/estado dirty que `goforge validate`. Se omite
   si la conexión o el directorio de migraciones fallaron.
6. **Bloqueo** — adquiere y libera el lock de migración
   (`pg_advisory_lock` / `GET_LOCK`), detectando un lock que una ejecución
   anterior que se cayó dejó atascado. Se omite si la conexión falló.

Sale con código distinto de cero si alguna comprobación falló (las
comprobaciones omitidas no cuentan como fallos).

```sh
goforge doctor
goforge doctor --json
goforge doctor --env-file=secrets/db.env
```

## Servidor MCP

```sh
goforge mcp
```

Transporte stdio, usando el SDK oficial de MCP para Go. Expone el mismo
`migration.Engine` que usa el CLI — idénticas garantías de locking, checksum y
transacción. Las herramientas de solo lectura no necesitan confirmación:

- `goforge_status`
- `goforge_migration_list`
- `goforge_migration_pending`
- `goforge_migration_validate`
- `goforge_migration_plan`

Las herramientas que modifican el esquema requieren un argumento explícito
`confirm: true`, exigido tanto por el JSON Schema (`confirm` es un campo
obligatorio) como por el handler:

- `goforge_migration_up`
- `goforge_migration_rollback`

Ninguna herramienta ejecuta SQL arbitrario — no hay ningún camino para que un
agente ejecute `DROP DATABASE`, `TRUNCATE` ni nada equivalente; cada operación
pasa por las migraciones revisadas del propio motor de migraciones.

## Códigos de salida

| Código | Significado |
|---|---|
| `0` | Éxito. |
| `1` | La operación falló (error de migración, etc.). |
| `2` | Problema de configuración/setup (`goforge.yaml` ausente o inválido, `--driver` incorrecto, fallo de conexión antes siquiera de intentar la operación). |

`goforge doctor` es la única excepción: siempre sale con `1` cuando alguna
comprobación falló — incluido un `goforge.yaml` ausente — ya que una
comprobación fallida es su forma normal y esperada de reportar un problema, no
un error de uso.

## Salida `--json`

Todos los comandos aceptan `--json` para obtener un único objeto JSON
estructurado en stdout — estable sin importar `--lang`, pensado para scripts y
agentes. Los logs y diagnósticos van a stderr, nunca mezclados en el flujo
JSON. Ejemplo:

```sh
$ goforge migrate:status --json
{
  "status": "ok",
  "migrations": [
    {"version": 1, "name": "create_users", "applied": true, "batch": 1, "applied_at": "2026-08-11T07:25:11Z"}
  ]
}
```
