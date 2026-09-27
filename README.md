<div align="center">

<img src="docs/assets/banner.svg" alt="PokeScript: aprende a programar, ¡atrápalos a todos!" width="100%">

<img src="https://img.shields.io/badge/Go-compilador-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
<img src="https://img.shields.io/badge/Wails-escritorio-DF0000?style=for-the-badge&logo=wails&logoColor=white" alt="Wails">
<img src="https://img.shields.io/badge/Svelte-interfaz-FF3E00?style=for-the-badge&logo=svelte&logoColor=white" alt="Svelte">
<img src="https://img.shields.io/badge/CodeMirror-6-D30707?style=for-the-badge&logo=codemirror&logoColor=white" alt="CodeMirror 6">

<img src="https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/versions/generation-v/black-white/animated/137.gif" alt="Porygon" height="56">
<img src="https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/versions/generation-v/black-white/animated/25.gif" alt="Pikachu" height="56">
<img src="https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/versions/generation-v/black-white/animated/4.gif" alt="Charmander" height="56">
<img src="https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/versions/generation-v/black-white/animated/7.gif" alt="Squirtle" height="56">
<img src="https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/versions/generation-v/black-white/animated/1.gif" alt="Bulbasaur" height="56">
<img src="https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/versions/generation-v/black-white/animated/74.gif" alt="Geodude" height="56">

</div>

<img src="docs/assets/pokedex.svg" alt="Entrada de Pokédex: PokeScript, un lenguaje en español para aprender a programar" width="100%">

<img src="docs/assets/guias/charizard.svg" alt="Charizard: ¡Hola, entrenador! PokeScript es un lenguaje en español para aprender a programar. Los datos tienen tipo, las funciones son movimientos y los errores se explican en tu idioma." width="100%">

Trae su propio editor de escritorio. Mientras lo terminamos, los programas ya corren desde la terminal.

<sub>Proyecto del curso Paradigmas de Programación · Universidad Nacional · II ciclo 2026</sub>

<img src="docs/assets/divisor.svg" alt="" width="100%">

## Contenido

- [El editor](#el-editor)
- [Tipos](#tipos)
- [Lo básico](#lo-básico)
- [Programas completos](#programas-completos)
- [Por dentro](#por-dentro)
- [Medallas](#medallas)
- [Probarlo](#probarlo)
- [Contribuir](#contribuir)

<img src="docs/assets/divisor.svg" alt="" width="100%">

## El editor

<img src="docs/assets/guias/rotom.svg" alt="Rotom: Así va a verse el IDE: el código, el asistente al lado y la salida abajo. Si escribes curra en vez de curar, el asistente te ofrece el arreglo." width="100%">

<img src="docs/assets/editor.svg" alt="Editor de PokeScript con un programa, el asistente y la salida" width="100%">

<img src="docs/assets/divisor.svg" alt="" width="100%">

## Tipos

<img src="docs/assets/guias/pikachu.svg" alt="Pikachu: Cada tipo de dato es un tipo de Pokémon. El mío es electrico: solo sé decir verdadero o falso, pero sin mí no hay si ni mientras." width="100%">

<div align="center">

|                                                                               Pokémon                                                                               | Tipo        | Qué guarda                 | Ejemplo                          |
| :-----------------------------------------------------------------------------------------------------------------------------------------------------------------: | ----------- | -------------------------- | -------------------------------- |
|  <img src="https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/versions/generation-v/black-white/animated/74.gif" alt="Geodude" height="48">   | `roca`      | Números enteros            | `roca vida = 100`                |
|  <img src="https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/versions/generation-v/black-white/animated/7.gif" alt="Squirtle" height="48">   | `agua`      | Números decimales          | `agua precision = 0.95`          |
| <img src="https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/versions/generation-v/black-white/animated/4.gif" alt="Charmander" height="48">  | `fuego`     | Un solo carácter           | `fuego inicial = 'P'`            |
|  <img src="https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/versions/generation-v/black-white/animated/1.gif" alt="Bulbasaur" height="48">  | `planta`    | Texto                      | `planta nombre = "Bulbasaur"`    |
|  <img src="https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/versions/generation-v/black-white/animated/25.gif" alt="Pikachu" height="48">   | `electrico` | `verdadero` o `falso`      | `electrico listo = verdadero`    |
|   <img src="https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/versions/generation-v/black-white/animated/201.gif" alt="Unown" height="48">   | `especie`   | Un valor de una lista fija | `Estado e = DORMIDO`             |
|  <img src="https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/versions/generation-v/black-white/animated/143.gif" alt="Snorlax" height="48">  | `equipo`    | Lista (índices desde 1)    | `equipo de roca niveles`         |
| <img src="https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/versions/generation-v/black-white/animated/474.gif" alt="Porygon-Z" height="48"> | `mochila`   | Diccionario ordenado       | `mochila de planta a roca bolsa` |

</div>

Y como en los combates, no todos se llevan bien. La tabla de efectividades dice qué se convierte solo, qué necesita `convertir` y qué no se puede hacer.

<img src="docs/assets/efectividades.svg" alt="Tabla de efectividades entre los tipos" width="100%">

Si un programa mezcla dos tipos que no se llevan, no llega a ejecutarse:

<img src="docs/assets/combate.svg" alt="ROCA usa + contra ELECTRICO: no es muy efectivo" width="100%">

Por ejemplo, este programa:

```text
combate
    roca x = 1 + verdadero
fin
```

recibe esta respuesta:

```text
No es muy efectivo…
Tipo de error: semántico
Archivo: principal.pks
Línea: 2
Columna: 14
Descripción: no es posible sumar un valor de tipo roca con uno de tipo electrico.
Posible causa: la combinación de estos dos tipos no tiene efecto según la tabla de efectividades.
Sugerencia: consultar la tabla de efectividades desde el menú del entorno.
```

Cada clase de problema tiene su encabezado:

| Encabezado                     | Cuándo aparece                                                         |
| ------------------------------ | ---------------------------------------------------------------------- |
| ¡Es superefectivo!             | El programa compiló sin errores                                        |
| No es muy efectivo…            | Tipos que no se llevan o una regla del lenguaje que no se cumple       |
| ¡Se escapó!                    | Algo mal escrito, como un bloque sin su `fin`                          |
| ¡No pasó nada!                 | Un nombre que no existe o un dato que se usa antes de tener valor      |
| No se encontró la ruta         | Un problema al importar otro archivo                                   |
| ¡Falló el ataque!              | El programa se cayó mientras corría, por ejemplo al dividir entre cero |
| ¿Seguro que quieres hacer eso? | Un aviso: el programa corre, pero algo se ve raro                      |

<img src="docs/assets/divisor.svg" alt="" width="100%">

## Lo básico

Lo necesario para leer y escribir PokeScript. Cada bloque se puede copiar tal cual.

### Hola, mundo

<img src="docs/assets/guias/charmander.svg" alt="Charmander: Todo programa empieza en combate y termina en su fin. Lo de adentro corre de arriba abajo, una instrucción por línea." width="100%">

```text
combate
    gritar "¡Hola, mundo Pokémon!"
fin
```

### Datos y medallas

<img src="docs/assets/guias/onix.svg" alt="Onix: Cada dato se declara con su tipo. Una medalla es un valor que no cambia nunca, duro como la roca. Por eso va en mayúsculas." width="100%">

```text
medalla roca NIVEL_MAXIMO = 100

combate
    roca nivel = 5
    agua precision = 0.85
    fuego rango = 'S'
    planta nombre = "Pikachu"
    electrico brillante = falso

    nivel = nivel + 1
    gritar nombre, " está en el nivel ", nivel, " de ", NIVEL_MAXIMO
fin
```

### Decisiones

<img src="docs/assets/guias/psyduck.svg" alt="Psyduck: Aquí no hay «más o menos»: la condición de un si da verdadero o falso. si vida no compila; si vida &gt; 0 sí." width="100%">

```text
si vida > 50
    gritar "¡En plena forma!"
sino si vida > 0
    gritar "Necesita una poción"
sino
    gritar "Se debilitó…"
fin
```

Para muchos casos está `segun`. Si elige entre números o textos necesita la rama `otro`, porque esos valores no se acaban nunca:

```text
segun ataque
    'F'      entonces gritar "¡Lanzallamas!"
    'A', 'H' entonces gritar "¡Hidrobomba!"
    otro     entonces gritar "Placaje"
fin
```

### Ciclos

<img src="docs/assets/guias/tauros.svg" alt="Tauros: mientras embiste una y otra vez hasta que la condición deja de cumplirse. recorrer pasa por un rango o por un equipo. huir sale del ciclo y siguiente salta a la otra vuelta." width="100%">

```text
// De un número a otro (ambos incluidos)
recorrer n de 1 hasta 5
    gritar "Pokébola número ", n
fin

// Por cada elemento de una colección
recorrer pokemon en mi_equipo
    si pokemon igual "Magikarp"
        siguiente
    fin
    gritar pokemon, " está listo"
fin

// Mientras se cumpla la condición
mientras vida > 0
    vida = vida - 10
    si vida < 20
        huir
    fin
fin
```

### Movimientos

<img src="docs/assets/guias/machamp.svg" alt="Machamp: Los movimientos son las funciones. Unos devuelven un valor con entregar; otros solo hacen algo, como gritar un mensaje. ¡Cuatro brazos, cero efectos colaterales!" width="100%">

```text
movimiento roca calcular_dano(roca ataque, roca defensa)
    entregar ataque * 2 - defensa
fin

movimiento saludar(planta entrenador)
    gritar "¡Hola, ", entrenador, "! Bienvenido al Centro Pokémon"
fin

combate
    saludar("Ash")
    roca dano = calcular_dano(55, 40)
    gritar "El ataque hizo ", dano, " de daño"
fin
```

### Equipo y mochila

<img src="docs/assets/guias/kangaskhan.svg" alt="Kangaskhan: Un equipo es una lista. Una mochila guarda cada cosa con su clave, como mi bolsa, y recuerda el orden. Se cuenta desde 1, como el primer Pokémon de tu equipo." width="100%">

```text
equipo de planta equipo_ash = ["Pikachu", "Charizard"]
sumar "Bulbasaur" a equipo_ash
gritar equipo_ash[1]
quitar equipo_ash[2]
gritar "Quedan ", tamaño(equipo_ash), " Pokémon"

mochila de planta a roca objetos = {"Poción": 3, "Pokébola": 10}
objetos["Poción"] = objetos["Poción"] - 1
si objetos contiene "Pokébola"
    gritar "Tienes ", objetos["Pokébola"], " Pokébolas"
fin
```

### Especies y fichas

<img src="docs/assets/guias/eevee.svg" alt="Eevee: Una especie es una lista cerrada de valores, como mis evoluciones: no hay más que esas. Una ficha junta varios datos bajo un mismo nombre." width="100%">

```text
especie Clima
    SOLEADO, LLUVIA, GRANIZO
fin

ficha Entrenador
    planta nombre
    roca   medallas
    Clima  clima_favorito
fin

combate
    Entrenador misty = {nombre: "Misty", medallas: 2, clima_favorito: LLUVIA}
    misty.medallas = misty.medallas + 1
    gritar misty.nombre, " ya tiene ", misty.medallas, " medallas"
fin
```

### Conversiones

<img src="docs/assets/guias/ditto.svg" alt="Ditto: Cambiar de forma es lo mío, pero aquí se pide con convertir. Ojo: convertir(agua) a roca corta los decimales; para redondear está redondear." width="100%">

```text
planta texto = "25"
roca numero = convertir(texto) a roca
agua decimal = numero
planta mensaje = convertir(numero) a planta
roca redondo = redondear(7.5)
roca dado = aleatorio(1, 6)
```

### Varios archivos

<img src="docs/assets/guias/abra.svg" alt="Abra: Con enseñar … desde un archivo se teletransporta lo que declara otro: movimientos, especies, fichas y medallas." width="100%">

```text
enseñar calcular_dano desde "operaciones.pks"
enseñar Estado, Pokemon desde "tipos.pks"
```

## Programas completos

Cada animación es un programa real con su salida real. Están en [`ejemplos/`](ejemplos) y las pruebas del proyecto los ejecutan.

### Entrenamiento

<img src="docs/assets/guias/machop.svg" alt="Machop: Un equipo de niveles, una mochila de objetos y varios ciclos para entrenar. ¡A sudar!" width="100%">

<div align="center">
<img src="docs/assets/ejemplo-colecciones.svg" alt="Programa con equipo, mochila, recorrer y mientras, y su salida" width="100%">
</div>

<details>
<summary>Ver el código</summary>

```text
// entrenamiento.pks · equipo, mochila y ciclos
combate
    equipo de planta mi_equipo = ["Pikachu", "Charmander", "Squirtle"]
    sumar "Bulbasaur" a mi_equipo
    gritar "Tu equipo tiene ", tamaño(mi_equipo), " Pokémon"

    mochila de planta a roca bayas = {"Aranja": 3, "Zreza": 0, "Meloc": 5}
    recorrer baya, cantidad en bayas
        si cantidad igual 0
            siguiente
        fin
        gritar "  ", baya, " x", cantidad
    fin

    roca nivel = 5
    mientras nivel < 10
        nivel = nivel + 2
    fin
    gritar "Nivel final: ", nivel
fin
```

</details>

### Estados alterados

<img src="docs/assets/guias/gengar.svg" alt="Gengar: Una especie con los estados alterados, un segun que los cubre todos y movimientos que dicen qué le pasa a cada Pokémon. Je, je." width="100%">

<div align="center">
<img src="docs/assets/ejemplo-especies.svg" alt="Programa con especie, ficha, segun exhaustivo y movimientos, y su salida" width="100%">
</div>

<details>
<summary>Ver el código</summary>

```text
// estados.pks · especie, ficha, segun y movimientos
especie Estado
    SANO, ENVENENADO, DORMIDO
fin

ficha Pokemon
    planta nombre
    roca   vida
    Estado estado
fin

movimiento roca pasar_turno(Pokemon p)
    roca dano = 0
    segun p.estado
        SANO        entonces gritar p.nombre, " está en plena forma"
        ENVENENADO  entonces dano = 10
        DORMIDO     entonces gritar p.nombre, " sigue dormido…"
    fin
    entregar p.vida - dano
fin

combate
    Pokemon bulbi = {nombre: "Bulbasaur", vida: 45, estado: ENVENENADO}
    recorrer t de 1 hasta 3
        bulbi.vida = pasar_turno(bulbi)
        gritar "Turno ", t, ": ", bulbi.nombre, " tiene ", bulbi.vida, " PS"
    fin
fin
```

</details>

### Centro Pokémon

<img src="docs/assets/guias/chansey.svg" alt="Chansey: capturar se queda esperando lo que escribas. Si la respuesta no sirve, lo explica y vuelve a preguntar, con la paciencia de un Centro Pokémon." width="100%">

<div align="center">
<img src="docs/assets/ejemplo-captura.svg" alt="Programa que pide datos con capturar y valida la entrada" width="100%">
</div>

<details>
<summary>Ver el código</summary>

```text
// centro.pks · capturar, contiene y sino si
movimiento electrico es_legendario(planta nombre)
    equipo de planta legendarios = ["Mewtwo", "Lugia", "Rayquaza"]
    entregar legendarios contiene nombre
fin

combate
    planta nombre
    roca nivel
    capturar(nombre, "¿Qué Pokémon atrapaste? ")
    capturar(nivel, "¿De qué nivel? ")

    si es_legendario(nombre)
        gritar "¡Increíble! ", nombre, " es legendario"
    sino si nivel >= 50
        gritar nombre, " ya es todo un veterano"
    sino
        agua progreso = nivel / 100
        gritar nombre, " va al ", redondear(progreso * 100), "% del camino"
    fin
fin
```

</details>

### Cuando falta un `fin`

<img src="docs/assets/guias/magikarp.svg" alt="Magikarp: Olvidar un fin es el error más común al empezar. El mensaje dice qué bloque quedó abierto, en qué línea empezó y cuál es el que falta cerrar." width="100%">

<div align="center">
<img src="docs/assets/ejemplo-error.svg" alt="PokeScript detecta un bloque sin cerrar y explica dónde se abrió" width="100%">
</div>

### Combate por turnos

<img src="docs/assets/guias/blastoise.svg" alt="Blastoise: El programa de la especificación: cuatro archivos que se importan entre sí. Tu Pokémon contra Bulbi, veinte turnos como máximo." width="100%">

<details>
<summary>constantes.pks</summary>

```text
medalla roca VIDA_MAXIMA = 100
medalla roca NIVEL       = 25
```

</details>

<details>
<summary>tipos.pks</summary>

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

</details>

<details>
<summary>operaciones.pks</summary>

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

</details>

<details>
<summary>principal.pks</summary>

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

</details>

<img src="docs/assets/divisor.svg" alt="" width="100%">

## Por dentro

<img src="docs/assets/guias/porygon.svg" alt="Porygon: Un programa pasa por cuatro etapas antes de mostrar algo. Si el analizador no encuentra un nombre, el asistente busca qué quisiste escribir." width="100%">

<img src="docs/assets/recorrido.svg" alt="El recorrido de un programa por el compilador" width="100%">

El compilador y el intérprete están escritos en Go. El editor usa Svelte y CodeMirror 6, y Wails los junta en una aplicación de escritorio. Las reglas completas del lenguaje están en la [especificación](docs/PokeScript_Especificacion_Implementacion.md).

<img src="docs/assets/divisor.svg" alt="" width="100%">

## Medallas

<img src="docs/assets/guias/dragonite.svg" alt="Dragonite: Cada hito es una medalla: primero los ocho gimnasios de Kanto y después el Alto Mando. Falta la Medalla Roca, que llega con los colores del editor." width="100%">

<img src="docs/assets/medallas.svg" alt="Estuche de medallas: los ocho gimnasios de Kanto y el Alto Mando, 11 de 12" width="100%">

<img src="docs/assets/divisor.svg" alt="" width="100%">

## Probarlo

<img src="docs/assets/guias/snorlax.svg" alt="Snorlax: Para despertarme hace falta Go 1.23 o más nuevo. Si vas a trabajar en el proyecto, también Node.js 22 y pnpm 10. Con npm no me muevo. Zzz…" width="100%">

Descargas: [Go](https://go.dev/dl/) · [Node.js](https://nodejs.org) · [pnpm](https://pnpm.io/installation)

```bash
git clone https://github.com/keylorpineda/PokeScript.git
cd PokeScript
pnpm install
git config commit.template .gitmessage
```

Para correr un programa:

```bash
go run ./cmd/pks ejemplos/hola.pks
```

Y el combate por turnos, que tiene cuatro archivos:

```bash
go run ./cmd/pks ejemplos/combate
```

```text
PokeScript/
├── internal/   compilador, intérprete y asistente
├── cmd/pks/    el comando para la terminal
├── ejemplos/   los programas de este README
└── docs/       especificación, plan del equipo y animaciones
```

<img src="docs/assets/divisor.svg" alt="" width="100%">

## Contribuir

Los commits van en inglés con el formato `PKS type(scope): description`, y cada cambio entra a `dev` por un pull request. Los detalles están en [CONTRIBUTING.md](CONTRIBUTING.md) y el reparto de tareas en [docs/PLAN.md](docs/PLAN.md).

<img src="docs/assets/divisor.svg" alt="" width="100%">

<div align="center">

<img src="https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/versions/generation-v/black-white/animated/25.gif" alt="Pikachu" height="40">

<sub>Proyecto académico sin fines de lucro. Pokémon y sus nombres son marcas de Nintendo, Game Freak y The Pokémon Company; este proyecto no está afiliado a ellas. Sprites de <a href="https://github.com/PokeAPI/sprites">PokeAPI</a>.</sub>

</div>
