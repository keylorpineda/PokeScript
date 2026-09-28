# PokeScript — Especificación de implementación

Documento técnico para el equipo de desarrollo. Recoge todas las reglas del lenguaje en la forma en que hacen falta para programar el compilador, no en la forma en que se explican en la documentación académica.

**Stack:** Wails (contenedor) · Go (compilador e intérprete) · Svelte (interfaz) · CodeMirror 6 (editor) · Howler.js (audio) · PixiJS (opcional, salida gráfica).

---

## 0. Arquitectura y orden de construcción

```text
Svelte (frontend)                    Go (backend)
├── Gestor de proyectos    ←─────→   ├── FS: proyecto, archivos .pks
├── Editor (CodeMirror 6)  ←─────→   ├── Lexer     → []Token
├── Segmento de salida     ←─────→   ├── Parser    → *AST
├── Asistente pedagógico   ←─────→   ├── Analyzer  → *AST validado | []Diagnostic
└── Menú de consulta                 └── Interp    → efectos de salida / entrada
```

**Orden recomendado de implementación.** Cada paso debe compilar y correr antes de pasar al siguiente.

| #   | Hito                                                         | Se puede demostrar           |
| --- | ------------------------------------------------------------ | ---------------------------- |
| 1   | Lexer completo + tabla de tokens                             | Colorear en CodeMirror       |
| 2   | Parser de expresiones                                        | Evaluar `2 + 3 * 4`          |
| 3   | Parser de instrucciones y bloques                            | Árbol de un archivo completo |
| 4   | Intérprete sin tipos (`gritar`, variables, `si`, `mientras`) | Primer programa corriendo    |
| 5   | Tabla de símbolos + chequeo de tipos                         | Errores semánticos           |
| 6   | Movimientos, parámetros, retorno                             | Funciones                    |
| 7   | Colecciones (`equipo`, `mochila`)                            | Recorridos                   |
| 8   | `especie`, `segun` exhaustivo                                | Coincidencia de patrones     |
| 9   | Importaciones + grafo de dependencias                        | Proyecto multiarchivo        |
| 10  | `posible` / `fantasma` + estrechamiento                      | Nulabilidad                  |
| 11  | `ficha`                                                      | Registros                    |
| 12  | Asistente (distancia de edición)                             | Innovación                   |

> El hito 4 es el que convierte el proyecto en algo demostrable. Llegar ahí es la prioridad.

---

## 1. Léxico

### 1.1 Palabras reservadas

Ningún identificador puede coincidir con una de estas. Se escriben siempre en minúscula y sin tilde (se conserva la `ñ` en `enseñar` y `tamaño`).

| Grupo                | Palabras                                                                                              |
| -------------------- | ----------------------------------------------------------------------------------------------------- |
| Declaración de tipos | `especie` `ficha` `medalla`                                                                           |
| Tipos                | `roca` `agua` `fuego` `planta` `electrico` `equipo` `mochila` `posible`                               |
| Conectores de tipo   | `de` `a`                                                                                              |
| Literales            | `verdadero` `falso` `fantasma`                                                                        |
| Estructura           | `combate` `movimiento` `entregar` `fin` `enseñar` `desde`                                             |
| Control              | `si` `sino` `sino si` `segun` `entonces` `otro` `mientras` `recorrer` `en` `hasta` `huir` `siguiente` |
| E/S y datos          | `gritar` `capturar` `convertir` `tamaño` `aleatorio` `redondear` `sumar` `quitar` `contiene`          |
| Operadores palabra   | `resto` `igual` `diferente` `y` `o` `no`                                                              |

**Total: 48 palabras reservadas.**

**`sino si` es el único token de dos términos.** El lexer lo reconoce como una unidad: al leer `sino`, mira si el siguiente token es `si` y, si lo es, emite `SINO_SI`.

**`sino` tiene dos papeles** y el parser los distingue por posición, no el lexer:

- al inicio de una instrucción → alternativa del condicional
- dentro de una expresión → valor de reemplazo (nivel `expr_respaldo`)

### 1.2 Símbolos

| Símbolo     | Uso                                        |
| ----------- | ------------------------------------------ |
| `( )`       | agrupación, argumentos, parámetros         |
| `[ ]`       | acceso a equipo/mochila, literal de equipo |
| `{ }`       | literal de mochila, literal de ficha       |
| `,`         | separador                                  |
| `:`         | dentro de literales de mochila y ficha     |
| `.`         | acceso a campo de ficha                    |
| `=`         | asignación y valor inicial                 |
| `+ - * /`   | aritméticos                                |
| `> < >= <=` | comparación de orden                       |
| `//`        | comentario hasta fin de línea              |

### 1.3 Identificadores

- Inician con letra del alfabeto español, incluidas `ñ` y vocales con tilde.
- Continúan con letras, dígitos o `_`.
- Sensible a mayúsculas: `vida` y `Vida` son distintos.
- Regex orientativa: `^[\p{L}][\p{L}\p{Nd}_]*$` (Go: `unicode.IsLetter`).

**Convención de estilo** (advertencia, no error): tipos de `especie`/`ficha` con mayúscula inicial, valores de especie en mayúsculas, medallas en mayúsculas, datos en minúsculas.

### 1.4 Literales

| Tipo        | Forma               | Notas                                             |
| ----------- | ------------------- | ------------------------------------------------- |
| `roca`      | `[0-9]+`            | **sin signo**; el `-` siempre es operador unario  |
| `agua`      | `[0-9]+\.[0-9]+`    | punto decimal, dígitos obligatorios a ambos lados |
| `fuego`     | `'c'`               | exactamente un carácter                           |
| `planta`    | `"..."`             | escapes: `\"` `\n` `\\` únicamente                |
| `electrico` | `verdadero` `falso` |                                                   |
| nulo        | `fantasma`          | solo asignable a tipos `posible`                  |

**Escape desconocido = error léxico.** Mensaje: indicar la secuencia no reconocida.

**El punto no es ambiguo.** `6.9` es dígito·punto·dígito; `chispo.vida` es identificador·punto·identificador. Un carácter de anticipación basta: tras un dígito, si viene `.` seguido de dígito → sigue el literal; en cualquier otro caso → token `PUNTO`.

### 1.5 Líneas y espacios

- **Una instrucción por línea, sin excepciones.** El salto de línea es significativo: el lexer emite `NEWLINE`.
- La indentación **no** es significativa para el parser, pero el lexer registra la columna inicial de cada línea para la advertencia de indentación inconsistente.
- Líneas en blanco y comentarios se descartan (no emiten `NEWLINE` propio).

### 1.6 Estructura del token (Go)

```go
type Token struct {
    Kind   TokenKind
    Lexeme string
    Line   int   // 1-based
    Col    int   // 1-based
    Len    int   // para el subrayado en CodeMirror
}
```

`Line`, `Col` y `Len` deben propagarse a cada nodo del AST y a cada diagnóstico. **Sin esto no funciona el subrayado del editor**, que es la mitad del valor del asistente.

---

## 2. Gramática (EBNF)

Escrita para descenso recursivo. La precedencia está codificada en la jerarquía `expr_o` → `expr_acceso`.

```ebnf
programa         = { importacion } , { declaracion } ;
importacion      = "enseñar" , identificador , { "," , identificador } ,
                   "desde" , literal_texto ;

declaracion      = decl_medalla | decl_especie | decl_ficha
                 | decl_movimiento | bloque_combate ;
decl_medalla     = "medalla" , tipo , identificador , "=" , expresion ;
decl_especie     = "especie" , identificador , valores , "fin" ;
valores          = identificador , { "," , identificador } ;
decl_ficha       = "ficha" , identificador , { campo } , "fin" ;
campo            = tipo , identificador ;
decl_movimiento  = "movimiento" , [ tipo ] , identificador ,
                   "(" , [ parametros ] , ")" , bloque , "fin" ;
bloque_combate   = "combate" , bloque , "fin" ;
parametros       = tipo , identificador , { "," , tipo , identificador } ;

tipo             = tipo_opcional | tipo_equipo | tipo_mochila ;
tipo_opcional    = [ "posible" ] , ( tipo_simple | tipo_nombrado ) ;
tipo_simple      = "roca" | "agua" | "fuego" | "planta" | "electrico" ;
tipo_nombrado    = identificador ;
tipo_equipo      = "equipo" , "de" , tipo_contenido ;
tipo_mochila     = "mochila" , "de" , tipo_clave , "a" , tipo_contenido ;
tipo_contenido   = tipo_simple | tipo_nombrado | tipo_equipo | tipo_mochila ;
tipo_clave       = "roca" | "fuego" | "planta" | tipo_nombrado ;

bloque           = { instruccion } ;
instruccion      = decl_dato | asignacion | salida | entrada | llamada
                 | mod_coleccion | condicional | condicional_multiple
                 | ciclo | corte | retorno ;
decl_dato        = [ "medalla" ] , tipo , identificador , [ "=" , expresion ] ;
asignacion       = destino , "=" , expresion ;
destino          = identificador , { "[" , expresion , "]" | "." , identificador } ;
salida           = "gritar" , expresion , { "," , expresion } ;
entrada          = "capturar" , "(" , destino , "," , expresion , ")" ;
llamada          = identificador , "(" , [ argumentos ] , ")" ;
argumentos       = expresion , { "," , expresion } ;
mod_coleccion    = "sumar" , expresion , "a" , identificador
                 | "quitar" , identificador , "[" , expresion , "]" ;
corte            = "huir" | "siguiente" ;
retorno          = "entregar" , [ expresion ] ;

condicional      = "si" , expresion , bloque ,
                   { "sino si" , expresion , bloque } ,
                   [ "sino" , bloque ] , "fin" ;
condicional_multiple = "segun" , expresion , { alternativa } ,
                   [ "otro" , "entonces" , instruccion ] , "fin" ;
alternativa      = patron , { "," , patron } , "entonces" , instruccion ;
patron           = literal | identificador ;

ciclo            = ciclo_mientras | ciclo_recorrido | ciclo_rango ;
ciclo_mientras   = "mientras" , expresion , bloque , "fin" ;
ciclo_recorrido  = "recorrer" , identificador , [ "," , identificador ] ,
                   "en" , expresion , bloque , "fin" ;
ciclo_rango      = "recorrer" , identificador , "de" , expresion ,
                   "hasta" , expresion , bloque , "fin" ;

expresion        = expr_o ;
expr_o           = expr_y , { "o" , expr_y } ;
expr_y           = expr_no , { "y" , expr_no } ;
expr_no          = [ "no" ] , expr_igualdad ;
expr_igualdad    = expr_pertenencia ,
                   [ ( "igual" | "diferente" ) , expr_pertenencia ] ;
expr_pertenencia = expr_orden , [ "contiene" , expr_orden ] ;
expr_orden       = expr_suma , [ ( ">" | "<" | ">=" | "<=" ) , expr_suma ] ;
expr_suma        = expr_producto , { ( "+" | "-" ) , expr_producto } ;
expr_producto    = expr_unaria , { ( "*" | "/" | "resto" ) , expr_unaria } ;
expr_unaria      = [ "-" ] , expr_respaldo ;
expr_respaldo    = expr_acceso , [ "sino" , expr_acceso ] ;
expr_acceso      = primaria , { "[" , expresion , "]" | "." , identificador } ;
primaria         = literal | literal_equipo | literal_mochila | literal_ficha
                 | identificador | llamada | conversion | tamano
                 | aleatorio | redondear | "(" , expresion , ")" ;

literal_equipo   = "[" , [ expresion , { "," , expresion } ] , "]" ;
literal_mochila  = "{" , [ par , { "," , par } ] , "}" ;
par              = expresion , ":" , expresion ;
literal_ficha    = "{" , identificador , ":" , expresion ,
                   { "," , identificador , ":" , expresion } , "}" ;
conversion       = "convertir" , "(" , expresion , ")" , "a" , tipo ;
tamano           = "tamaño" , "(" , expresion , ")" ;
aleatorio        = "aleatorio" , "(" , expresion , "," , expresion , ")" ;
redondear        = "redondear" , "(" , expresion , ")" ;
```

### 2.1 Puntos que el parser debe resolver con cuidado

**Anticipación de dos tokens en `instruccion`.** Al ver un identificador, puede ser:

- `identificador (` → llamada
- `identificador =` o `identificador [` o `identificador .` → asignación
- `identificador identificador` → declaración de tipo nombrado (`Estado x`, `Pokemon p`)

**`{ }` es ambiguo entre mochila y ficha para el parser.** Solución: el parser produce un nodo genérico `LiteralLlaves` con una lista de pares; el **analizador semántico** decide si es mochila o ficha según el tipo esperado. Si las claves son identificadores simples y el tipo esperado es una ficha → ficha; si el tipo esperado es mochila → mochila con esas claves evaluadas como expresiones.

**Literales de colección sin tipo esperado = error semántico**, no sintáctico.

**`tipo_nombrado` es un identificador**, así que `tipo` no se puede distinguir de una expresión por el primer token. En `decl_dato` el parser prueba `tipo identificador`; si tras el identificador no viene otro identificador, retrocede y lo trata como expresión/asignación.

**Recuperación de errores:** al fallar, descartar tokens hasta el siguiente `NEWLINE` y continuar. Acumular hasta **20 diagnósticos sintácticos** por compilación.

**Pila de bloques.** El parser mantiene una pila con `(tipoDeBloque, línea, columna)` de cada bloque abierto. Cuando el archivo termina con la pila no vacía, el diagnóstico dice **en qué línea se abrió** el bloque sin cerrar. Este es el diagnóstico estrella del lenguaje; no lo dejen para el final.

---

## 3. Sistema de tipos

### 3.1 Tipos

```go
type Kind int
const (
    KRoca Kind = iota; KAgua; KFuego; KPlanta; KElectrico
    KEspecie; KFicha; KEquipo; KMochila
)
type Type struct {
    Kind     Kind
    Opcional bool      // posible
    Nombre   string    // especie / ficha
    Clave    *Type     // mochila
    Elem     *Type     // equipo / mochila
}
```

`posible` se aplica **solo** a tipos simples y especies. `posible equipo`, `equipo de posible X`, `posible ficha` → error.

### 3.2 Tabla de efectividades (conversiones)

`MT` = mismo tipo · `EF` = automática · `RC` = requiere `convertir` · `SE` = error

| Origen \ Destino | roca | agua | fuego | planta | electrico | especie |
| ---------------- | ---- | ---- | ----- | ------ | --------- | ------- |
| **roca**         | MT   | EF   | SE    | RC     | SE        | SE      |
| **agua**         | RC   | MT   | SE    | RC     | SE        | SE      |
| **fuego**        | SE   | SE   | MT    | RC     | SE        | SE      |
| **planta**       | RC   | RC   | SE    | MT     | SE        | SE      |
| **electrico**    | SE   | SE   | SE    | RC     | MT        | SE      |
| **especie**      | SE   | SE   | SE    | RC     | SE        | MT      |

Reglas adicionales:

- `T` → `posible T`: automática.
- `posible T` → `T`: solo dentro de un bloque con estrechamiento, o mediante `sino`.
- `equipo`/`mochila`/`ficha`: compatibles solo si el tipo es **exactamente** el mismo.
- `convertir` sobre una casilla `SE` es error de compilación aunque sea explícito.
- `convertir(agua) a roca` **trunca** (no redondea). `redondear` redondea.
- `convertir(planta) a roca/agua` puede fallar **en ejecución** si el texto no es numérico.

### 3.3 Operaciones por tipo

| Operación           | Tipos válidos                                           | Resultado                    |
| ------------------- | ------------------------------------------------------- | ---------------------------- |
| `+ - *`             | roca·roca, agua·agua, roca·agua                         | roca / agua (ensanchamiento) |
| `+`                 | planta·planta                                           | planta                       |
| `/`                 | roca·roca, agua·agua, roca·agua                         | **siempre agua**             |
| `resto`             | roca·roca                                               | roca                         |
| `-` unario          | roca, agua                                              | mismo                        |
| `> < >= <=`         | roca·roca, agua·agua, roca·agua, fuego·fuego            | electrico                    |
| `igual` `diferente` | mismos tipos (o T vs posible T)                         | electrico                    |
| `y` `o` `no`        | electrico                                               | electrico                    |
| `contiene`          | equipo·elem, mochila·clave, planta·planta, planta·fuego | electrico                    |
| `sino`              | posible T · T                                           | T                            |
| `[ ]`               | equipo[roca], mochila[clave], planta[roca]              | elem / valor / fuego         |
| `.`                 | ficha.campo                                             | tipo del campo               |
| `tamaño()`          | equipo, mochila, planta                                 | roca                         |

**No encadenables:** `> < >= <=`, `igual`, `diferente`, `contiene`. Una sola aparición por expresión; `a > b > c` es **error de sintaxis**, no de tipos.

**`no` está por debajo de las comparaciones:** `no rival igual fantasma` se lee como `no (rival igual fantasma)`.

### 3.4 Valores en ejecución (Go)

```go
type Value interface{}
// roca      → int64
// agua      → float64
// fuego     → rune
// planta    → string
// electrico → bool
// especie   → EspecieVal{Tipo, Valor string}
// equipo    → *[]Value
// mochila   → *ordered map (¡preservar orden de inserción!)
// ficha     → map[string]Value
// fantasma  → nil
```

**La mochila debe preservar el orden de inserción**, porque el documento promete que `recorrer` la visita en ese orden. Un `map` de Go no lo hace: usar slice de claves + map, o una estructura ordenada.

**Paso por valor de todo, incluidas colecciones y fichas.** Copia profunda al pasar argumentos. Esto es lo que elimina los efectos colaterales; no lo optimicen con punteros.

---

## 4. Validaciones semánticas

Dos pasadas sobre el proyecto completo.

### Pasada 1 — Recolección

1. Resolver importaciones: el archivo existe, el nombre existe en él, no hay colisión con lo declarado localmente.
2. Construir el grafo de dependencias; detectar ciclos (DFS con marcado). Al detectar uno, **reportar la cadena completa**.
3. Registrar especies (valores no repetidos entre especies del proyecto), fichas (campos no repetidos), medallas y cabeceras de movimientos.
4. Verificar que exista **exactamente un** `combate`, en el archivo principal.
5. Movimientos duplicados (mismo nombre en el proyecto) → error señalando ambas declaraciones. **No hay sobrecarga.**

### Pasada 2 — Verificación

Por cada cuerpo:

| #   | Verificación                                                                 | Severidad   |
| --- | ---------------------------------------------------------------------------- | ----------- |
| 1   | Tipos de cada expresión contra la tabla                                      | error       |
| 2   | Condición de `si`/`mientras` es `electrico` (sin veracidad implícita)        | error       |
| 3   | Cantidad y tipo de argumentos en cada llamada                                | error       |
| 4   | Movimiento con tipo entrega valor **por todos los caminos**                  | error       |
| 5   | `entregar` con valor en movimiento sin tipo, o sin valor en uno con tipo     | error       |
| 6   | **Asignación definida**: ningún camino lee un dato antes de asignarle valor  | error       |
| 7   | Reasignación de `medalla` (o modificación de su contenido)                   | error       |
| 8   | Ocultamiento de nombres (incluye variables de `recorrer` anidados)           | error       |
| 9   | `segun` sobre especie o `electrico`: exhaustivo, **nombrando los faltantes** | error       |
| 10  | `segun` sobre roca/agua/fuego/planta: `otro` obligatorio                     | error       |
| 11  | Uso de `posible` sin comprobar                                               | error       |
| 12  | `huir`/`siguiente` fuera de ciclo                                            | error       |
| 13  | Modificar una colección mientras se la recorre                               | error       |
| 14  | Literal de colección sin tipo esperado                                       | error       |
| 15  | Literal de ficha: campos completos, sin repetir, sin ajenos                  | error       |
| 16  | Asignar a la variable de recorrido (es de solo lectura)                      | error       |
| 17  | División o `resto` entre literal `0`                                         | error       |
| 18  | Rama de `segun` inalcanzable (patrón repetido, `otro` redundante)            | advertencia |
| 19  | Movimiento con tipo invocado como instrucción (valor descartado)             | advertencia |
| 20  | Dato declarado y nunca utilizado                                             | advertencia |
| 21  | Ciclo cuya condición no depende de nada modificado en su cuerpo              | advertencia |
| 22  | Indentación inconsistente con la estructura de bloques                       | advertencia |
| 23  | Convención de nombres (mayúsculas/minúsculas)                                | advertencia |

**El análisis semántico no se detiene en el primer error:** recorre todo y reporta todos.

### 4.1 Estrechamiento de `posible`

Se aplica cuando la condición es una comparación **directa** contra `fantasma`:

```text
si rival diferente fantasma   → dentro de la rama afirmativa, rival es planta
si rival igual fantasma       → dentro de la rama sino, rival es planta
```

**Se propaga a través de `y`** (no de `o`): al analizar `A y B`, B se analiza con los hechos que A dejó verdaderos. Esto es coherente con el cortocircuito.

**Se pierde** si el dato se reasigna dentro del bloque.

Implementación sugerida: un mapa `narrowed map[string]bool` en el entorno de análisis, que se apila y desapila con los bloques.

### 4.2 Asignación definida

Mismo recorrido de flujo que la verificación de retorno. Para cada dato declarado sin valor, llevar un estado `Asignado | NoAsignado | Quizá`:

- Tras un `si` con `sino`: asignado solo si **ambas** ramas lo asignan.
- Tras un `si` sin `sino`: `Quizá` → se trata como no asignado.
- Tras un `mientras` o `recorrer`: el cuerpo puede no ejecutarse → no cuenta.
- `capturar(x, ...)` asigna `x`.

---

## 5. Funciones incorporadas

| Nombre      | Firma                                                                                        | Notas                                                     |
| ----------- | -------------------------------------------------------------------------------------------- | --------------------------------------------------------- |
| `tamaño`    | `tamaño(equipo de T) → roca`<br>`tamaño(mochila de C a V) → roca`<br>`tamaño(planta) → roca` |                                                           |
| `convertir` | `convertir(T) a U → U`                                                                       | solo si la tabla lo permite                               |
| `capturar`  | `capturar(destino, planta)`                                                                  | **instrucción**, no produce valor                         |
| `aleatorio` | `aleatorio(roca, roca) → roca`                                                               | extremos inclusive; si min > max → error de ejecución     |
| `redondear` | `redondear(agua) → roca`                                                                     | mitad exacta se aleja de cero                             |
| `resto`     | `roca resto roca → roca`                                                                     | operador infijo                                           |
| `gritar`    | `gritar(expr, ...)`                                                                          | instrucción; convierte todo a texto; salta línea al final |
| `sumar`     | `sumar valor a equipo`                                                                       | instrucción                                               |
| `quitar`    | `quitar coleccion[índice o clave]`                                                           | instrucción                                               |

---

## 6. Semántica de ejecución

- Arranca en `combate`, tras evaluar las medallas de alcance de archivo en orden de aparición y registrar cabeceras.
- **Índices desde 1.** Fuera de rango → error de ejecución.
- `recorrer` sobre colección: variable implícita, solo lectura, muere con el bloque. Sobre mochila con dos variables: clave y valor.
- `recorrer n de A hasta B`: ambos extremos inclusive, paso de 1, evaluados una sola vez; si A > B el cuerpo no se ejecuta.
- `mientras` evalúa antes de cada vuelta.
- `huir` sale del ciclo más interno; `siguiente` pasa a la vuelta siguiente.
- `y` / `o` en **cortocircuito**.
- Recursión limitada a **1000** llamadas anidadas.
- Desbordamiento de `roca` (int64) → error de ejecución.
- `capturar` bloquea hasta recibir entrada; si no corresponde al tipo, explica y vuelve a preguntar (bucle interno, sin límite).

### Errores de ejecución

División o `resto` entre cero · índice fuera de rango · clave inexistente · `convertir` de texto no numérico · recursión excedida · desbordamiento · `aleatorio` con min > max.

---

## 7. Diagnósticos

Estructura única que Go envía a Svelte:

```go
type Diagnostic struct {
    Severity  string // "error" | "advertencia"
    Category  string // "lexico"|"sintactico"|"semantico"|"importacion"|"ejecucion"
    Heading   string // encabezado temático
    File      string
    Line      int
    Col       int
    Len       int    // para subrayar en CodeMirror
    Desc      string
    Cause     string
    Suggest   string // lo agrega el asistente
    Fix       *Fix   // corrección aplicable, opcional
}
type Fix struct {
    Line, Col, Len int
    Replacement    string
}
```

**Encabezados temáticos sugeridos:**

| Categoría                      | Encabezado                       |
| ------------------------------ | -------------------------------- |
| Error de tipos                 | `No es muy efectivo…`            |
| Sintáctico / bloque sin cerrar | `¡Se escapó!`                    |
| Dato sin valor / no declarado  | `¡No pasó nada!`                 |
| Ejecución                      | `¡Falló el ataque!`              |
| Importación                    | `No se encontró la ruta`         |
| Advertencia                    | `¿Seguro que quieres hacer eso?` |
| Compilación exitosa            | `¡Es superefectivo!`             |

Ejemplo completo:

```text
No es muy efectivo…
Tipo de error: semántico
Archivo: principal.pks
Línea: 12
Columna: 13
Descripción: no es posible sumar un valor de tipo roca con uno de tipo electrico.
Posible causa: la combinación de estos dos tipos no tiene efecto según la tabla de efectividades.
Sugerencia: consultar la tabla de efectividades desde el menú del entorno.
```

### 7.1 Asistente pedagógico

- **Reglas por categoría**: un mapa de `Category + subcódigo` → plantilla de explicación.
- **Distancia de edición** (Levenshtein) sobre la tabla de símbolos visible cuando un identificador no está declarado. Umbral sugerido: distancia ≤ 2, o ≤ 1/3 de la longitud. Devolver el mejor candidato como `Fix`.
- Determinista, sin servicios externos.

---

## 8. Proyecto y archivos

```text
mi_proyecto/
    proyecto.json        // metadatos: nombre, archivo principal
    principal.pks        // contiene el bloque combate
    operaciones.pks
    tipos.pks
```

- Extensión `.pks`. Sin subcarpetas en esta versión.
- Rutas de `enseñar` relativas a la carpeta del proyecto.
- El archivo principal se marca en `proyecto.json`.
- Orden dentro de un archivo: importaciones → declaraciones de archivo → `combate` (solo en el principal).
- Se importan movimientos, especies, fichas y medallas. **No** datos modificables (no existen globales).

---

## 9. Integración con el frontend

### CodeMirror 6

- `StreamLanguage` o parser propio para el resaltado. Categorías de token: palabra reservada, tipo, literal, identificador, comentario, operador, símbolo.
- **Decorations** para subrayar errores usando `Line/Col/Len`. Convertir a offsets absolutos con `doc.line(n).from + (col-1)`.
- Autocompletado desde la tabla de símbolos que devuelve Go.
- Los `Fix` se aplican con `dispatch({changes: {from, to, insert}})`.

### Wails

Métodos mínimos a exponer desde Go:

```go
CompilarProyecto(ruta string) ResultadoCompilacion
EjecutarProyecto(ruta string) // streaming de salida
EnviarEntrada(texto string)   // respuesta a capturar
ObtenerTablaEfectividades() [][]string
ObtenerPalabrasReservadas() []PalabraDoc
```

La ejecución debe ser **streaming** (eventos de Wails), no una llamada que devuelve todo al final, porque `capturar` necesita ir y volver.

### Estado en Svelte

`proyecto` · `archivoActivo` · `diagnosticos[]` · `salida[]` · `esperandoEntrada` · `estadoAsistente`

---

## 10. Programa de ejemplo completo

Sirve como prueba de integración y como el "proyecto de ejemplo" del entregable.

**`constantes.pks`**

```text
medalla roca VIDA_MAXIMA = 100
medalla roca NIVEL       = 25
```

**`tipos.pks`**

```text
especie Estado
    SANO, ENVENENADO, DORMIDO, PARALIZADO
fin

ficha Pokemon
    planta nombre
    roca   vida
    Estado estado
fin
```

**`operaciones.pks`**

```text
enseñar NIVEL desde "constantes.pks"

// Daño base con variación al azar del 85% al 100%
movimiento roca calcular_dano(roca poder)
    agua base      = poder * NIVEL / 50
    roca variacion = aleatorio(85, 100)
    agua total     = base * variacion / 100
    entregar redondear(total) + 2
fin
```

**`principal.pks`**

```text
enseñar calcular_dano desde "operaciones.pks"
enseñar Estado, Pokemon desde "tipos.pks"
enseñar VIDA_MAXIMA desde "constantes.pks"

movimiento describir(Estado actual)
    segun actual
        SANO                 entonces gritar "  Puede atacar"
        ENVENENADO           entonces gritar "  Pierde vida cada turno"
        DORMIDO, PARALIZADO  entonces gritar "  Podría no atacar"
    fin
fin

combate
    planta nombre
    capturar(nombre, "¿Cómo se llama tu Pokemon? ")

    Pokemon mio   = {nombre: nombre,   vida: VIDA_MAXIMA, estado: SANO}
    Pokemon rival = {nombre: "Bulbi", vida: VIDA_MAXIMA, estado: SANO}

    gritar "¡", mio.nombre, " entra en combate!"
    describir(mio.estado)

    recorrer turno de 1 hasta 20
        gritar "--- Turno ", turno, " ---"

        roca dano = calcular_dano(40)
        rival.vida = rival.vida - dano
        gritar mio.nombre, " ataca y hace ", dano, " de daño"

        si rival.vida <= 0
            gritar rival.nombre, " se debilitó. ¡Ganaste!"
            huir
        fin

        roca contra = calcular_dano(35)
        mio.vida = mio.vida - contra
        gritar rival.nombre, " contraataca: ", contra, " de daño"
        gritar "Vida de ", mio.nombre, ": ", mio.vida

        si mio.vida <= 0
            gritar mio.nombre, " se debilitó. Perdiste."
            huir
        fin
    fin
fin
```

Cubre: múltiples archivos, importación de movimientos, especies, fichas y medallas, `capturar`, declaración sin valor, fichas con acceso por punto, `segun` exhaustivo con agrupación, ciclo de rango, condicionales, `huir`, funciones con y sin retorno, aritmética con ensanchamiento y `redondear`.

---

## 11. Casos de prueba prioritarios

Diseñarlos **antes** de programar cada validación. Cubren el mínimo del punto 18 del enunciado.

### Errores sintácticos (5)

1. Falta un `fin` → debe indicar en qué línea se abrió el bloque
2. `a > b > c` → operador no encadenable
3. Dos instrucciones en una línea
4. Cadena sin cerrar
5. `recorrer n de 1 a 10` (falta `hasta`)

### Errores semánticos (5)

1. `roca + electrico` → remite a la tabla de efectividades
2. `segun` sobre especie con un valor sin cubrir → lo nombra
3. Uso de `posible` sin comprobar
4. Reasignación de `medalla`
5. Dato leído sin valor asignado

### Importaciones (3 + 1 circular)

1. Archivo inexistente
2. Nombre inexistente en el archivo
3. Argumentos de tipo incorrecto a un movimiento importado
4. `a.pks` → `b.pks` → `a.pks` → reportar la cadena completa

**Prueba completa**: el programa de la sección 10.

---

## 12. Recordatorios de alcance

El aplicativo vale **35%** de la nota y la defensa individual **25%**. La documentación ya está prácticamente terminada; el riesgo está todo del lado del código.

- Llegar al **hito 4** cuanto antes: un programa corriendo cambia la conversación.
- Las advertencias 20 a 23 de la sección 4 son **prescindibles**. Si el tiempo aprieta, se recortan sin que nadie lo note.
- El asistente con distancia de edición es barato y es la innovación: no lo dejen de último.
- Crear el repositorio **hoy** con commits de los tres. Es 5% que no se recupera después.
