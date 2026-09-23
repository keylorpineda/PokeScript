# PokeScript

Lenguaje de programación pedagógico con temática Pokémon, más su IDE de escritorio. Es el proyecto del curso de Paradigmas de la UNA, II ciclo 2026. El equipo tiene tres integrantes.

## Fuente de verdad

`docs/PokeScript_Especificacion_Implementacion.md` es la especificación completa: léxico, gramática EBNF, sistema de tipos, validaciones, semántica de ejecución, diagnósticos y la integración con el frontend. Ante cualquier duda sobre el lenguaje, manda ese documento.

La especificación completa se carga en el contexto con esta importación:

@docs/PokeScript_Especificacion_Implementacion.md

## Stack

- **Wails**: contenedor de escritorio.
- **Go**: backend con el compilador y el intérprete (`Lexer → Parser → Analyzer → Interp`).
- **Svelte**: interfaz, con CodeMirror 6 como editor, Howler.js para audio y PixiJS como opción para salida gráfica.

## Reglas clave para el código

- Cada token, nodo del AST y diagnóstico lleva `Line`, `Col` y `Len` (base 1). El subrayado del editor depende de eso.
- Los índices de `equipo` empiezan en 1.
- La mochila preserva el orden de inserción.
- Todo se pasa por valor, con copia profunda incluso de colecciones y fichas.
- El análisis semántico reporta todos los errores; no se detiene en el primero.
- El parser se recupera hasta el siguiente `NEWLINE` y acumula como máximo 20 diagnósticos sintácticos.
- Los mensajes al usuario van en español y usan los encabezados temáticos de la sección 7 de la especificación.
- La ejecución se transmite por eventos de Wails, porque `capturar` necesita ir y volver.

## Orden de trabajo

El plan del equipo (roles, sprints, tareas `T<id>`, contratos entre módulos y decisiones) está en `docs/PLAN.md`. Las tareas nuevas y las decisiones se registran ahí.

Seguir los hitos de la sección 0 de la especificación. La prioridad es llegar al hito 4: el intérprete corriendo el primer programa.

## Commits

Los mensajes van **en inglés**, en imperativo y empezando en minúscula. El formato es obligatorio y lo validan Husky y commitlint:

```text
PKS type(scope): description
```

Los tipos, los alcances y los ejemplos están en `CONTRIBUTING.md`.

## Estructura y pruebas

- El compilador y el intérprete viven en `internal/` (por ejemplo `internal/lexer`, `internal/parser`, `internal/analizador`, `internal/interprete`, `internal/asistente`). Esos paquetes **no** importan Wails, así que se prueban sin interfaz.
- En la raíz quedan solo `main.go` y `app.go` de Wails, que hacen de puente.
- Las pruebas usan `testing` de la librería estándar, en archivos `*_test.go` al lado del código, con estilo de tabla.
- Los programas `.pks` de prueba van en `testdata/` dentro de cada paquete.
- Para correrlas: `go test ./internal/...`

## CI

`.github/workflows/ci.yml` corre en cada push a `main`/`dev` y en cada PR:

1. commitlint sobre todos los commits nuevos.
2. `gofmt`, `go vet` y `go test -race` sobre `internal/` (solo si existe `go.mod`).
3. `pnpm run build` del frontend (solo si existe `frontend/package.json`).
