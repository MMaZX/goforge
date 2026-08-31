# Contribuir a GoForge

Gracias por tu interés en mejorar GoForge. Este documento explica cómo
proponer cambios y qué se espera de una contribución para que se pueda
integrar rápido.

## Antes de empezar

- **Bugs y dudas**: abre un [issue](https://github.com/MMaZX/goforge/issues)
  describiendo qué esperabas, qué pasó, y cómo reproducirlo (versión de
  GoForge —`goforge version`—, driver, sistema operativo).
- **Cambios grandes o nuevas features**: abre primero un issue para
  discutir el enfoque antes de escribir código. Así evitas trabajo que
  luego no encaje con la dirección del proyecto.
- **Cambios pequeños** (typos, un bug acotado, un test que falta): manda el
  Pull Request directamente.

## Requisitos

- Go (la versión está en [`go.mod`](go.mod), campo `go`).
- Docker, solo si vas a ejecutar los tests de integración.
- No hace falta nada más: GoForge no depende de PHP, Node ni Python.

## Flujo de trabajo

1. Haz un fork y clona tu fork.
2. Crea una rama a partir de `main`:
   `git checkout -b fix/descripcion-corta`.
3. Haz tus cambios en commits pequeños y con sentido propio.
4. Asegúrate de que pasa todo lo que corre la CI (ver más abajo).
5. Sube la rama a tu fork y abre el Pull Request contra `main`.
6. Responde a los comentarios de la revisión con nuevos commits (no
   reescribas la historia mientras la revisión está en curso).

## Qué comprueba la CI

Tu PR no se fusiona hasta que esto pasa en verde. Puedes ejecutarlo todo en
local:

```sh
go build ./...
go vet ./...
gofmt -l .          # no debe listar ningún archivo
go test ./...
```

Tests de integración (necesitan Docker; levantan PostgreSQL y MariaDB en
contenedores):

```sh
go test -tags=integration ./tests/integration/... -v -timeout 10m
```

## Estilo de código

- Código formateado con `gofmt` (sin excepciones; la CI lo exige).
- Sigue las convenciones del código que ya hay alrededor: nombres,
  densidad de comentarios, manejo de errores con `%w`, etc.
- Los mensajes visibles para el usuario van por `internal/i18n` con su
  clave en inglés **y** en español. La salida `--json`, los errores del
  motor (`migration/`) y los nombres de migraciones se quedan en inglés a
  propósito — son un contrato estable para scripts y agentes.
- Todo comportamiento nuevo necesita tests. Los tests son table-driven
  donde tenga sentido.

## Mensajes de commit

Se usa [Conventional Commits](https://www.conventionalcommits.org/):

```
<tipo>(<ámbito opcional>): <descripción en imperativo>
```

Tipos habituales: `feat`, `fix`, `docs`, `refactor`, `test`, `perf`,
`build`, `ci`, `chore`. Ejemplos:

```
feat(cli): add --dry-run flag to migrate
fix(engine): release advisory lock when a migration panics
docs: translate USAGE.md to Spanish
```

Un cambio incompatible lleva `!` tras el tipo/ámbito y una nota
`BREAKING CHANGE:` en el cuerpo.

## Alcance de un Pull Request

- Un PR = un cambio lógico. Si te encuentras arreglando tres cosas
  distintas, son tres PRs.
- No mezcles reformateos masivos con cambios de lógica: se hacen
  imposibles de revisar.
- Actualiza `README.md` / `USAGE.md` si tu cambio afecta a comandos,
  flags o comportamiento observable.

## Buenas primeras contribuciones

Busca issues con la etiqueta
[`good first issue`](https://github.com/MMaZX/goforge/labels/good%20first%20issue).
Si no hay ninguno abierto y quieres empezar, comenta en un issue existente
y te ayudamos a acotar algo.

## Licencia

Al contribuir aceptas que tu código se publique bajo la licencia
[MIT](LICENSE) del proyecto.
