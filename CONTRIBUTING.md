# Cómo contribuir a PokeScript

## Preparación

```bash
npm install
```

Esto instala Husky y activa los hooks de Git (`.husky/`). Cada integrante debe correrlo una vez después de clonar.

Para que `git commit` abra la plantilla con la guía:

```bash
git config commit.template .gitmessage
```

## Estándar de commits

Los mensajes de commit van **en inglés**, con este formato:

```text
PKS type(scope): description
```

- `PKS` es obligatorio y va como primera palabra, sin corchetes.
- `type` es obligatorio. `(scope)` es opcional.
- La descripción va en inglés y en imperativo: `add`, `fix`, `remove` (no `added` ni `adds`).
- La descripción empieza en minúscula y no lleva punto final. Se permiten siglas como `AST` o `SINO_SI`.
- El encabezado tiene como máximo 100 caracteres.
- El cuerpo es opcional, va en inglés y separado por una línea en blanco. Explica qué cambió y por qué.

### Tipos

| Tipo       | Uso                                            |
| ---------- | ---------------------------------------------- |
| `feat`     | Funcionalidad nueva                            |
| `fix`      | Corrección de un error                         |
| `docs`     | Solo documentación                             |
| `refactor` | Cambio interno que no altera el comportamiento |
| `test`     | Casos de prueba                                |
| `style`    | Formato, sin cambios de lógica                 |
| `perf`     | Rendimiento                                    |
| `build`    | Dependencias, empaquetado, Wails               |
| `ci`       | Integración continua                           |
| `chore`    | Mantenimiento y configuración                  |
| `revert`   | Revierte un commit anterior                    |

### Alcances (scopes)

| Scope         | Parte del proyecto                                     |
| ------------- | ------------------------------------------------------ |
| `lexer`       | Análisis léxico y tokens                               |
| `parser`      | Gramática y AST                                        |
| `analyzer`    | Tabla de símbolos, tipos y validaciones semánticas     |
| `interpreter` | Ejecución                                              |
| `assistant`   | Asistente pedagógico y distancia de edición            |
| `diagnostics` | Estructura y mensajes de diagnóstico                   |
| `project`     | Gestor de proyectos, archivos `.pks` y `proyecto.json` |
| `editor`      | CodeMirror: resaltado, subrayado, autocompletado       |
| `ui`          | Interfaz Svelte en general                             |
| `wails`       | Puente Go ↔ Svelte                                     |
| `examples`    | Programas de ejemplo                                   |
| `tests`       | Infraestructura de pruebas                             |
| `docs`        | Documentación                                          |
| `config`      | Configuración del repositorio                          |

### Ejemplos

```text
PKS feat(lexer): recognize sino si as a single SINO_SI token
PKS fix(parser): report the line where an unclosed block was opened
PKS test(analyzer): add cases for non-exhaustive segun
PKS docs: update the type effectiveness table
PKS chore(config): set up husky and commitlint
```

## Ramas

Las ramas salen de `dev` y se nombran **en inglés**, en minúsculas y con guiones: `<tipo>/<área>-<qué>`.

| Tipo       | Uso                           | Ejemplo                      |
| ---------- | ----------------------------- | ---------------------------- |
| `feature/` | Funcionalidad nueva           | `feature/parser-expressions` |
| `fix/`     | Corrección de un error        | `fix/lexer-string-escapes`   |
| `docs/`    | Solo documentación            | `docs/user-manual`           |
| `test/`    | Solo pruebas                  | `test/analyzer-cases`        |
| `chore/`   | Configuración y mantenimiento | `chore/ci-cache`             |

El ID de la tarea del plan (`T2.1`) va en el título del PR, no en el nombre de la rama. El flujo completo está en [docs/PLAN.md](docs/PLAN.md#-flujo-de-trabajo-en-git).

## Hooks

| Hook         | Qué hace                                                                                          |
| ------------ | ------------------------------------------------------------------------------------------------- |
| `commit-msg` | Valida el mensaje con commitlint (`commitlint.config.js`)                                         |
| `pre-commit` | Corre lint-staged (Prettier, markdownlint, CSpell) y `gofmt` sobre los `.go` si Go está instalado |

Si un commit se rechaza, corrige el mensaje y vuelve a intentarlo. No uses `--no-verify`.

## Estándares de código

Al abrir el proyecto, VS Code ofrece instalar las extensiones recomendadas de `.vscode/extensions.json`. Acéptalas.

| Herramienta       | Qué hace                                                                  | Configuración                           |
| ----------------- | ------------------------------------------------------------------------- | --------------------------------------- |
| EditorConfig      | Codificación UTF-8, finales de línea LF y sangría por tipo de archivo     | `.editorconfig`                         |
| `.gitattributes`  | Obliga LF en el repo, para que no haya diffs falsos entre Windows y Linux | `.gitattributes`                        |
| Prettier          | Formato de JS, Svelte, JSON, YAML, CSS y Markdown                         | `.prettierrc.json`                      |
| gofmt / goimports | Formato de Go                                                             | Estándar de Go                          |
| golangci-lint     | Análisis estático de Go                                                   | `.golangci.yml`                         |
| markdownlint      | Estilo de los `.md`                                                       | `.markdownlint-cli2.jsonc`              |
| CSpell            | Ortografía en español e inglés                                            | `cspell.json`, `.cspell/pokescript.txt` |

### Comandos

```bash
npm run format   # formatea todo con Prettier
npm run lint     # revisa formato, markdown y ortografía (lo mismo que el CI)
npm run spell    # solo ortografía
```

### Cuando CSpell marca una palabra válida

Agrégala a `.cspell/pokescript.txt`, en la sección que corresponda. Los identificadores sin tilde (`expresion`, `coleccion`, …) van ahí, porque en el código son intencionales.

### Qué pasa en cada commit

1. `pre-commit` corre `lint-staged` sobre los archivos en stage. Prettier los formatea, markdownlint revisa los `.md` y CSpell revisa la ortografía. Además, `gofmt` revisa los `.go`.
2. `commit-msg` valida el formato `PKS type(scope): description`.
3. En GitHub, el CI repite todo lo anterior y corre las pruebas.
