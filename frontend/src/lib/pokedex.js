// Contenido de la Pokédex de consulta: todo el lenguaje, por secciones. Cada
// entrada tiene un Pokémon que la explica, un texto corto, un ejemplo y
// notas. Las palabras reservadas y la tabla de efectividades vienen del
// backend (internal/consulta) y se agregan en Pokedex.svelte.

export const SECCIONES = [
  { id: 'inicio', nombre: 'Bienvenida', objeto: 'poke-ball' },
  { id: 'tipos', nombre: 'Tipos de datos', objeto: 'great-ball' },
  { id: 'efectividades', nombre: 'Efectividades', objeto: 'ultra-ball' },
  { id: 'operadores', nombre: 'Operadores', objeto: 'rare-candy' },
  { id: 'control', nombre: 'Decisiones y ciclos', objeto: 'town-map' },
  { id: 'movimientos', nombre: 'Movimientos', objeto: 'potion' },
  { id: 'colecciones', nombre: 'Equipo y mochila', objeto: 'exp-share' },
  { id: 'incorporadas', nombre: 'Funciones incorporadas', objeto: 'master-ball' },
  { id: 'proyectos', nombre: 'Proyectos', objeto: 'town-map' },
  { id: 'errores', nombre: 'Mensajes de error', objeto: 'poke-ball' },
  { id: 'palabras', nombre: 'Palabras reservadas', objeto: 'great-ball' },
];

// Pokémon que presenta cada grupo de palabras reservadas.
export const POKEMON_GRUPO = {
  'Declaración de tipos': 'porygon',
  Tipos: 'eevee',
  'Conectores de tipo': 'ditto',
  Literales: 'pikachu',
  Estructura: 'machamp',
  Control: 'tauros',
  'E/S y datos': 'chansey',
  'Operadores palabra': 'alakazam',
};

export const ENTRADAS = [
  // ─── Bienvenida ───────────────────────────────────────────────────────────
  {
    seccion: 'inicio',
    titulo: 'Cómo usar la Pokédex',
    pokemon: 'porygon',
    texto:
      'Aquí está todo PokeScript. A la izquierda eliges la sección y en la lista, la entrada. Con las flechas del teclado también te mueves, y con el buscador encuentras cualquier palabra.',
    ejemplo: 'combate\n    gritar "¡Hola, mundo Pokémon!"\nfin',
    notas: [
      'Todo programa empieza en el bloque `combate` del archivo principal.',
      'Una instrucción por línea: el salto de línea importa.',
      'Los comentarios empiezan con `//` y llegan hasta el final de la línea.',
    ],
  },
  {
    seccion: 'inicio',
    titulo: 'Reglas de oro',
    pokemon: 'pikachu',
    texto:
      'Cinco reglas que conviene saber desde el primer día. El compilador las revisa todas antes de correr tu programa.',
    notas: [
      'Cada dato se declara con su tipo: `roca vida = 100`.',
      'Los índices de un equipo empiezan en 1, no en 0.',
      'Todo se pasa por valor: un movimiento recibe una copia y nunca cambia tus datos.',
      'Un dato no se puede leer antes de darle valor.',
      'Los nombres distinguen mayúsculas: `vida` y `Vida` son distintos.',
    ],
  },

  // ─── Tipos de datos ───────────────────────────────────────────────────────
  {
    seccion: 'tipos',
    titulo: 'roca',
    tipo: 'roca',
    pokemon: 'geodude',
    texto: 'Números enteros, sin decimales. Firmes como una roca.',
    ejemplo: 'roca vida = 100\nroca deuda = -5',
    notas: [
      'El `-` siempre es un operador: `-5` es «menos cinco».',
      'Si una roca se pasa del límite, el programa falla con un desbordamiento.',
      'Una roca pasa sola a `agua` cuando hace falta: `agua x = 3` vale 3.0.',
    ],
  },
  {
    seccion: 'tipos',
    titulo: 'agua',
    tipo: 'agua',
    pokemon: 'squirtle',
    texto: 'Números con decimales. Fluyen entre los enteros.',
    ejemplo: 'agua precision = 0.95\nagua mitad = 7 / 2',
    notas: [
      'Hacen falta dígitos a los dos lados del punto: `0.5`, nunca `.5`.',
      'La división `/` siempre da agua, aunque los dos números sean roca.',
      'Para pasar a roca: `convertir(x) a roca` corta los decimales; `redondear(x)` redondea.',
    ],
  },
  {
    seccion: 'tipos',
    titulo: 'fuego',
    tipo: 'fuego',
    pokemon: 'charmander',
    texto: 'Un solo carácter, entre comillas simples. Una chispa, no una fogata.',
    ejemplo: "fuego inicial = 'P'\nfuego separador = '-'",
    notas: [
      "Exactamente un carácter: `'ab'` es un error.",
      'Los fuegos se pueden comparar con `<` y `>` (orden alfabético).',
      'Al pedir una letra de un texto obtienes un fuego: `nombre[1]`.',
    ],
  },
  {
    seccion: 'tipos',
    titulo: 'planta',
    tipo: 'planta',
    pokemon: 'bulbasaur',
    texto: 'Texto, entre comillas dobles. Crece tanto como quieras.',
    ejemplo: 'planta nombre = "Bulbasaur"\nplanta saludo = "¡Hola, " + nombre + "!"',
    notas: [
      'Escapes permitidos: `\\"` comillas, `\\n` salto de línea y `\\\\` barra.',
      'Dos plantas se juntan con `+`.',
      '`tamaño(nombre)` dice cuántas letras tiene.',
    ],
  },
  {
    seccion: 'tipos',
    titulo: 'electrico',
    tipo: 'electrico',
    pokemon: 'pikachu',
    texto: 'Verdadero o falso. La energía de cada decisión.',
    ejemplo: 'electrico listo = verdadero\nsi listo\n    gritar "¡Al ataque!"\nfin',
    notas: [
      'La condición de un `si` o un `mientras` tiene que ser eléctrica.',
      'No hay «verdad implícita»: `si vida` no compila; `si vida > 0` sí.',
      'Se combinan con `y`, `o` y `no`.',
    ],
  },
  {
    seccion: 'tipos',
    titulo: 'especie',
    tipo: 'especie',
    pokemon: 'unown',
    texto:
      'Una lista cerrada de valores con nombre. Como las formas de Unown: solo existen las que existen.',
    ejemplo: 'especie Estado\n    SANO, DORMIDO, ENVENENADO\nfin\n\nEstado actual = DORMIDO',
    notas: [
      'El nombre de la especie va con mayúscula inicial; sus valores, en mayúsculas.',
      'Dos especies del proyecto no pueden repetir un valor.',
      'Un `segun` sobre una especie tiene que cubrir todos sus valores.',
    ],
  },
  {
    seccion: 'tipos',
    titulo: 'ficha',
    tipo: 'ficha',
    pokemon: 'porygon',
    texto: 'Un registro: varios datos con nombre juntos, como la ficha de un Pokémon.',
    ejemplo:
      'ficha Pokemon\n    planta nombre\n    roca   vida\nfin\n\nPokemon mio = {nombre: "Charmy", vida: 35}\ngritar mio.nombre',
    notas: [
      'Al crear una ficha hay que dar todos sus campos, sin repetir ni inventar otros.',
      'Cada campo se lee y se cambia con punto: `mio.vida = 40`.',
    ],
  },
  {
    seccion: 'tipos',
    titulo: 'equipo',
    tipo: 'equipo',
    pokemon: 'snorlax',
    texto:
      'Una lista de valores del mismo tipo. El primero es el 1, como el primer Pokémon de tu equipo.',
    ejemplo: 'equipo de roca niveles = [5, 12, 30]\ngritar niveles[1]\nsumar 45 a niveles',
    notas: [
      'Pedir una posición que no existe es un error de ejecución.',
      '`tamaño(niveles)` dice cuántos hay.',
      'Puede guardar otros equipos: `equipo de equipo de roca`.',
    ],
  },
  {
    seccion: 'tipos',
    titulo: 'mochila',
    tipo: 'mochila',
    pokemon: 'kangaskhan',
    texto:
      'Guarda valores por clave, como una bolsa con compartimentos. Recuerda el orden en que metiste cada cosa.',
    ejemplo:
      'mochila de planta a roca bolsa = {"pocion": 3, "superball": 1}\ngritar bolsa["pocion"]\nquitar bolsa["superball"]',
    notas: [
      'Las claves pueden ser roca, fuego, planta o una especie.',
      '`recorrer clave, valor en bolsa` las visita en el orden en que entraron.',
      'Pedir una clave que no existe es un error de ejecución.',
    ],
  },
  {
    seccion: 'tipos',
    titulo: 'posible y fantasma',
    tipo: 'posible',
    pokemon: 'gengar',
    texto:
      'Un dato `posible` puede no tener nada adentro: `fantasma`. Hay que revisarlo antes de usarlo.',
    ejemplo:
      'posible planta rival = fantasma\nsi rival diferente fantasma\n    gritar "Rival: ", rival\nfin\ngritar rival sino "nadie"',
    notas: [
      'Solo los tipos simples y las especies pueden ser `posible`.',
      'Dentro de `si rival diferente fantasma`, `rival` ya se usa como planta.',
      '`rival sino "nadie"` da el valor, o «nadie» si es fantasma.',
    ],
  },

  // ─── Efectividades ────────────────────────────────────────────────────────
  {
    seccion: 'efectividades',
    titulo: 'La tabla de efectividades',
    pokemon: 'mewtwo',
    especial: 'tabla',
    texto:
      'Dice qué pasa cuando un valor de un tipo se usa donde se espera otro. Pasa el mouse sobre una casilla para leerla.',
    notas: [
      'MT: mismo tipo. EF: pasa solo. RC: hay que usar `convertir`. SE: no se puede.',
      '`convertir` sobre una casilla SE es error aunque lo escribas.',
      '`convertir(planta) a roca` puede fallar al correr si el texto no es un número.',
    ],
  },
  {
    seccion: 'efectividades',
    titulo: 'convertir',
    pokemon: 'ditto',
    texto:
      'Cambia un valor de un tipo a otro, cuando la tabla lo permite. Ditto se transforma, pero no en cualquier cosa.',
    ejemplo:
      'agua precio = 9.99\nroca entero = convertir(precio) a roca\nplanta texto = convertir(entero) a planta',
    notas: [
      'De agua a roca corta los decimales: 9.99 queda en 9.',
      'Casi todo se puede convertir a planta para mostrarlo.',
    ],
  },

  // ─── Operadores ───────────────────────────────────────────────────────────
  {
    seccion: 'operadores',
    titulo: 'Qué se evalúa primero',
    pokemon: 'alakazam',
    especial: 'precedencia',
    texto: 'De arriba abajo, del que menos pega al que más: los de abajo se evalúan primero.',
    notas: [
      'Los paréntesis siempre ganan: `(2 + 3) * 4` da 20.',
      '`no rival igual fantasma` se lee `no (rival igual fantasma)`.',
      'Las comparaciones no se encadenan: `a > b > c` es un error de sintaxis.',
    ],
  },
  {
    seccion: 'operadores',
    titulo: 'Aritmética',
    pokemon: 'machamp',
    texto: 'Suma, resta, multiplicación, división y resto entre números.',
    ejemplo: 'roca total = 7 + 2 * 3\nagua mitad = 7 / 2\nroca sobra = 7 resto 2',
    notas: [
      'roca con agua da agua: el resultado se ensancha.',
      '`/` siempre da agua. `resto` solo entre rocas.',
      'Dividir entre 0 (o sacar resto) es un error.',
    ],
  },
  {
    seccion: 'operadores',
    titulo: 'Comparaciones',
    pokemon: 'alakazam',
    texto: 'Comparan dos valores y dan un eléctrico.',
    ejemplo: 'electrico fuerte = vida > 50\nelectrico mismo = nombre igual "Pikachu"',
    notas: [
      '`>` `<` `>=` `<=` entre números o entre fuegos.',
      '`igual` y `diferente` entre valores del mismo tipo.',
      '`contiene` pregunta si algo está adentro: `equipo contiene 5`, `"hola" contiene \'h\'`.',
    ],
  },
  {
    seccion: 'operadores',
    titulo: 'Lógica: y, o, no',
    pokemon: 'slowpoke',
    texto:
      'Combinan eléctricos. Se evalúan en cortocircuito: si con la primera parte basta, la segunda ni se mira.',
    ejemplo: 'si vida > 0 y no dormido\n    gritar "¡Puede atacar!"\nfin',
  },

  // ─── Decisiones y ciclos ──────────────────────────────────────────────────
  {
    seccion: 'control',
    titulo: 'si · sino si · sino',
    pokemon: 'psyduck',
    texto: 'Elige qué hacer según una condición. Cada bloque se cierra con `fin`.',
    ejemplo:
      'si vida > 50\n    gritar "¡En plena forma!"\nsino si vida > 0\n    gritar "Aguanta…"\nsino\n    gritar "Se debilitó."\nfin',
    notas: ['`sino si` se escribe con espacio y cuenta como una sola palabra.'],
  },
  {
    seccion: 'control',
    titulo: 'segun',
    pokemon: 'eevee',
    texto: 'Elige entre muchos casos a la vez, como las evoluciones de Eevee.',
    ejemplo:
      'segun estado\n    SANO                entonces gritar "Puede atacar"\n    DORMIDO, PARALIZADO entonces gritar "Podría no atacar"\n    ENVENENADO          entonces gritar "Pierde vida"\nfin',
    notas: [
      'Sobre una especie o un eléctrico: hay que cubrir todos los casos, y el compilador te dice cuáles faltan.',
      'Sobre números o textos: la rama `otro entonces …` es obligatoria.',
      'Cada caso lleva una sola instrucción después de `entonces`.',
    ],
  },
  {
    seccion: 'control',
    titulo: 'mientras',
    pokemon: 'tauros',
    texto: 'Repite mientras la condición se cumpla. Revisa antes de cada vuelta.',
    ejemplo:
      'roca vida = 30\nmientras vida > 0\n    vida = vida - 12\n    gritar "Vida: ", vida\nfin',
    notas: ['Si la condición no depende de nada que cambie adentro, el ciclo nunca termina.'],
  },
  {
    seccion: 'control',
    titulo: 'recorrer',
    pokemon: 'tauros',
    texto: 'Pasa por cada número de un rango o por cada elemento de una colección.',
    ejemplo:
      'recorrer turno de 1 hasta 3\n    gritar "Turno ", turno\nfin\n\nrecorrer nivel en niveles\n    gritar nivel\nfin',
    notas: [
      'El rango incluye los dos extremos. Si el primero es mayor, no entra ni una vez.',
      'La variable de `recorrer` es de solo lectura y desaparece al salir.',
      'No se puede cambiar una colección mientras se la recorre.',
    ],
  },
  {
    seccion: 'control',
    titulo: 'huir y siguiente',
    pokemon: 'abra',
    texto: '`huir` sale del ciclo; `siguiente` salta a la próxima vuelta. Abra se teletransporta.',
    ejemplo:
      'recorrer n de 1 hasta 10\n    si n resto 2 igual 0\n        siguiente\n    fin\n    si n > 7\n        huir\n    fin\n    gritar n\nfin',
    notas: ['Solo sirven dentro de un ciclo.'],
  },

  // ─── Movimientos ──────────────────────────────────────────────────────────
  {
    seccion: 'movimientos',
    titulo: 'Declarar un movimiento',
    pokemon: 'machamp',
    texto: 'Los movimientos son las funciones. Si llevan un tipo, entregan un valor de ese tipo.',
    ejemplo:
      'movimiento roca doble(roca n)\n    entregar n * 2\nfin\n\nmovimiento saludar(planta nombre)\n    gritar "¡Hola, ", nombre, "!"\nfin',
    notas: [
      'Un movimiento con tipo debe `entregar` por todos los caminos.',
      'No hay sobrecarga: dos movimientos no pueden llamarse igual.',
      'Llamar a uno con tipo sin usar su valor da un aviso.',
    ],
  },
  {
    seccion: 'movimientos',
    titulo: 'Paso por valor',
    pokemon: 'ditto',
    texto:
      'Un movimiento recibe una copia de todo, incluso de equipos y fichas. Lo que cambie adentro no toca tus datos.',
    ejemplo:
      'movimiento vaciar(equipo de roca lista)\n    quitar lista[1]\nfin\n\ncombate\n    equipo de roca mios = [1, 2]\n    vaciar(mios)\n    gritar tamaño(mios)\nfin',
    notas: ['El ejemplo muestra 2: `mios` no cambió.'],
  },
  {
    seccion: 'movimientos',
    titulo: 'Recursión',
    pokemon: 'slowpoke',
    texto:
      'Un movimiento puede llamarse a sí mismo. Hasta 1000 llamadas anidadas; después, el programa falla.',
    ejemplo:
      'movimiento roca factorial(roca n)\n    si n <= 1\n        entregar 1\n    fin\n    entregar n * factorial(n - 1)\nfin',
  },

  // ─── Equipo y mochila ─────────────────────────────────────────────────────
  {
    seccion: 'colecciones',
    titulo: 'sumar y quitar',
    pokemon: 'kangaskhan',
    texto: '`sumar` agrega al final de un equipo; `quitar` saca una posición o una clave.',
    ejemplo: 'sumar 50 a niveles\nquitar niveles[1]\nquitar bolsa["pocion"]',
  },
  {
    seccion: 'colecciones',
    titulo: 'Recorrer una mochila',
    pokemon: 'kangaskhan',
    texto:
      'Con dos nombres, `recorrer` da la clave y el valor de cada compartimento, en el orden en que entraron.',
    ejemplo: 'recorrer objeto, cantidad en bolsa\n    gritar objeto, ": ", cantidad\nfin',
  },
  {
    seccion: 'colecciones',
    titulo: 'contiene y tamaño',
    pokemon: 'snorlax',
    texto: 'Pregunta si algo está adentro o cuántos elementos hay.',
    ejemplo:
      'si bolsa contiene "pocion"\n    gritar "Hay pociones"\nfin\ngritar "Equipo de ", tamaño(niveles)',
  },

  // ─── Funciones incorporadas ───────────────────────────────────────────────
  {
    seccion: 'incorporadas',
    titulo: 'gritar',
    pokemon: 'pikachu',
    texto:
      'Muestra en la salida todo lo que le pases, junto y sin espacios extra, y salta de línea al final.',
    ejemplo: 'gritar "Vida: ", vida, " de ", VIDA_MAXIMA',
  },
  {
    seccion: 'incorporadas',
    titulo: 'capturar',
    pokemon: 'chansey',
    texto:
      'Espera a que escribas algo y lo guarda en un dato. Si la respuesta no sirve para ese tipo, lo explica y vuelve a preguntar.',
    ejemplo: 'roca edad\ncapturar(edad, "¿Cuántos años tienes? ")',
    notas: ['Es una instrucción: no devuelve nada.'],
  },
  {
    seccion: 'incorporadas',
    titulo: 'aleatorio',
    pokemon: 'porygon',
    texto: 'Un número al azar entre dos rocas, las dos incluidas.',
    ejemplo: 'roca golpe = aleatorio(85, 100)',
    notas: ['Si el primero es mayor que el segundo, el programa falla.'],
  },
  {
    seccion: 'incorporadas',
    titulo: 'redondear',
    pokemon: 'squirtle',
    texto: 'Convierte un agua en la roca más cercana. La mitad exacta se aleja del cero.',
    ejemplo: 'roca arriba = redondear(2.5)\nroca abajo = redondear(-2.5)',
    notas: ['En el ejemplo, `arriba` vale 3 y `abajo` vale -3.'],
  },
  {
    seccion: 'incorporadas',
    titulo: 'tamaño',
    pokemon: 'snorlax',
    texto: 'Cuántos elementos tiene un equipo o una mochila, o cuántas letras tiene una planta.',
    ejemplo: 'roca letras = tamaño("Pikachu")',
  },

  // ─── Proyectos ────────────────────────────────────────────────────────────
  {
    seccion: 'proyectos',
    titulo: 'enseñar … desde',
    pokemon: 'abra',
    texto: 'Trae movimientos, especies, fichas o medallas de otro archivo del proyecto.',
    ejemplo:
      'enseñar calcular_dano desde "operaciones.pks"\nenseñar Estado, Pokemon desde "tipos.pks"',
    notas: [
      'Las importaciones van al principio del archivo.',
      'No se importan datos que cambian: no hay variables globales.',
      'Si `a.pks` enseña de `b.pks` y `b.pks` de `a.pks`, el compilador muestra la cadena completa del ciclo.',
    ],
  },
  {
    seccion: 'proyectos',
    titulo: 'El archivo principal',
    pokemon: 'charizard',
    texto:
      'Un proyecto es una carpeta de archivos .pks. Uno es el principal: el único con el bloque `combate`.',
    ejemplo:
      '// principal.pks\nenseñar VIDA_MAXIMA desde "constantes.pks"\n\ncombate\n    gritar "Vida máxima: ", VIDA_MAXIMA\nfin',
    notas: [
      'Orden dentro de un archivo: importaciones, declaraciones y, al final, el combate.',
      'Primero se evalúan las medallas; después empieza el combate.',
    ],
  },

  // ─── Mensajes de error ────────────────────────────────────────────────────
  {
    seccion: 'errores',
    titulo: '¡Es superefectivo!',
    pokemon: 'pikachu',
    texto: 'El programa compiló sin errores. Puede que haya avisos, pero corre.',
  },
  {
    seccion: 'errores',
    titulo: 'No es muy efectivo…',
    pokemon: 'mewtwo',
    texto:
      'Un error de tipos o una regla del lenguaje que no se cumple. Revisa la tabla de efectividades.',
    ejemplo: 'roca x = 1 + verdadero',
  },
  {
    seccion: 'errores',
    titulo: '¡Se escapó!',
    pokemon: 'magikarp',
    texto:
      'Algo mal escrito: una palabra fuera de lugar o un bloque sin su `fin`. El mensaje dice en qué línea empezó el bloque.',
    ejemplo: 'combate\n    si vida > 0\n        gritar "sigue"\nfin',
  },
  {
    seccion: 'errores',
    titulo: '¡No pasó nada!',
    pokemon: 'psyduck',
    texto:
      'Un nombre que no existe o un dato que se lee antes de tener valor. El asistente sugiere el nombre que quisiste escribir.',
    ejemplo: 'roca vida = 10\ngritar vdia',
  },
  {
    seccion: 'errores',
    titulo: 'No se encontró la ruta',
    pokemon: 'abra',
    texto:
      'Un problema al importar: el archivo no existe, el nombre no está en él o las importaciones forman un ciclo.',
  },
  {
    seccion: 'errores',
    titulo: '¡Falló el ataque!',
    pokemon: 'snorlax',
    texto:
      'El programa se cayó mientras corría: división entre cero, índice fuera de rango, clave inexistente, recursión excesiva…',
  },
  {
    seccion: 'errores',
    titulo: '¿Seguro que quieres hacer eso?',
    pokemon: 'porygon',
    texto: 'Un aviso: el programa corre, pero algo se ve raro, como un dato que nunca se usa.',
  },
];

// La escalera de precedencia (sección 2), del que se evalúa último al primero.
export const PRECEDENCIA = [
  ['o', 'uno de los dos'],
  ['y', 'los dos'],
  ['no', 'lo contrario'],
  ['igual · diferente', 'comparar valores'],
  ['contiene', 'estar adentro'],
  ['> < >= <=', 'ordenar'],
  ['+ -', 'sumar y restar'],
  ['* / resto', 'multiplicar y dividir'],
  ['- (unario)', 'negativo'],
  ['sino', 'valor de respaldo'],
  ['x[i] · x.campo', 'acceder'],
];

// Qué significa cada casilla de la tabla de efectividades.
export const EFECTIVIDAD = {
  MT: { nombre: 'Mismo tipo', clase: 'mt', texto: 'Es el mismo tipo: pasa tal cual.' },
  EF: { nombre: 'Automática', clase: 'ef', texto: '¡Es superefectivo! Se convierte sola.' },
  RC: {
    nombre: 'Requiere convertir',
    clase: 'rc',
    texto: 'Se puede, pero escribiendo `convertir(…) a …`.',
  },
  SE: {
    nombre: 'Sin efecto',
    clase: 'se',
    texto: 'No es muy efectivo… No se puede, ni con convertir.',
  },
};
