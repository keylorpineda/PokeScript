// Package consulta tiene el contenido del menú de consulta del IDE
// (sección 9): las palabras reservadas explicadas y la tabla de
// efectividades.
package consulta

import "github.com/keylorpineda/PokeScript/internal/tipos"

// PalabraDoc explica una palabra reservada.
type PalabraDoc struct {
	Palabra     string `json:"palabra"`
	Grupo       string `json:"grupo"`
	Descripcion string `json:"descripcion"`
	Ejemplo     string `json:"ejemplo"`
}

// Grupos de la tabla 1.1 de la especificación, en su orden.
const (
	GrupoDeclaracion = "Declaración de tipos"
	GrupoTipos       = "Tipos"
	GrupoConectores  = "Conectores de tipo"
	GrupoLiterales   = "Literales"
	GrupoEstructura  = "Estructura"
	GrupoControl     = "Control"
	GrupoDatos       = "E/S y datos"
	GrupoOperadores  = "Operadores palabra"
)

// PalabrasReservadas devuelve las 48 palabras reservadas más «sino si»,
// agrupadas y en el orden de la tabla 1.1.
func PalabrasReservadas() []PalabraDoc {
	return append([]PalabraDoc{}, palabras...)
}

// TablaEfectividades devuelve la tabla 3.2 lista para mostrar: la primera
// fila y la primera columna son los encabezados.
func TablaEfectividades() [][]string { return tipos.TablaEfectividades() }

var palabras = []PalabraDoc{
	// Declaración de tipos.
	{"especie", GrupoDeclaracion, "Declara una lista cerrada de valores posibles, como los estados de un Pokémon.", "especie Estado\n    SANO, DORMIDO\nfin"},
	{"ficha", GrupoDeclaracion, "Declara un registro: un tipo que agrupa varios datos con nombre.", "ficha Pokemon\n    planta nombre\n    roca   vida\nfin"},
	{"medalla", GrupoDeclaracion, "Declara un valor constante: una vez dado, no cambia nunca.", "medalla roca VIDA_MAXIMA = 100"},

	// Tipos.
	{"roca", GrupoTipos, "Números enteros, sin decimales.", "roca vida = 100"},
	{"agua", GrupoTipos, "Números con decimales.", "agua precision = 0.95"},
	{"fuego", GrupoTipos, "Un solo carácter, entre comillas simples.", "fuego inicial = 'P'"},
	{"planta", GrupoTipos, "Texto, entre comillas dobles.", "planta nombre = \"Pikachu\""},
	{"electrico", GrupoTipos, "Verdadero o falso.", "electrico listo = verdadero"},
	{"equipo", GrupoTipos, "Una lista de valores del mismo tipo; las posiciones empiezan en 1.", "equipo de roca niveles = [5, 12, 8]"},
	{"mochila", GrupoTipos, "Un diccionario: guarda valores por clave, en el orden en que se agregaron.", "mochila de planta a roca bolsa = {\"Poción\": 3}"},
	{"posible", GrupoTipos, "Marca un dato que puede no tener valor (fantasma). Hay que comprobarlo antes de usarlo.", "posible planta rival = fantasma"},

	// Conectores de tipo.
	{"de", GrupoConectores, "Une una colección con el tipo de lo que guarda, y marca el inicio de un rango.", "equipo de planta nombres = []"},
	{"a", GrupoConectores, "Separa la clave del valor en una mochila; también dice a qué tipo se convierte y a qué equipo se suma.", "mochila de planta a roca bolsa = {}"},

	// Literales.
	{"verdadero", GrupoLiterales, "El valor electrico que dice que algo se cumple.", "electrico listo = verdadero"},
	{"falso", GrupoLiterales, "El valor electrico que dice que algo no se cumple.", "electrico brillante = falso"},
	{"fantasma", GrupoLiterales, "La ausencia de valor. Solo se guarda en un dato posible.", "posible planta rival = fantasma"},

	// Estructura.
	{"combate", GrupoEstructura, "El bloque donde empieza el programa. Hay uno solo, en el archivo principal.", "combate\n    gritar \"¡Hola!\"\nfin"},
	{"movimiento", GrupoEstructura, "Declara una función: un bloque con nombre que recibe parámetros y puede entregar un valor.", "movimiento roca doble(roca x)\n    entregar x * 2\nfin"},
	{"entregar", GrupoEstructura, "Termina un movimiento y devuelve su valor.", "entregar x * 2"},
	{"fin", GrupoEstructura, "Cierra el bloque que se abrió más recientemente.", "si vida > 0\n    gritar \"sigue\"\nfin"},
	{"enseñar", GrupoEstructura, "Trae movimientos, especies, fichas o medallas de otro archivo del proyecto.", "enseñar calcular_dano desde \"operaciones.pks\""},
	{"desde", GrupoEstructura, "Indica de qué archivo se importa.", "enseñar Estado, Pokemon desde \"tipos.pks\""},

	// Control.
	{"si", GrupoControl, "Ejecuta un bloque solo si la condición es verdadera.", "si vida > 50\n    gritar \"en forma\"\nfin"},
	{"sino", GrupoControl, "La alternativa de un si. Dentro de una expresión, da un valor de respaldo para un posible.", "gritar rival sino \"nadie\""},
	{"sino si", GrupoControl, "Otra condición que se prueba si las anteriores fueron falsas.", "si vida > 50\n    gritar 1\nsino si vida > 0\n    gritar 2\nfin"},
	{"segun", GrupoControl, "Elige qué hacer según el valor de un dato, entre varios casos.", "segun estado\n    SANO entonces gritar \"bien\"\n    otro entonces gritar \"mal\"\nfin"},
	{"entonces", GrupoControl, "Separa los casos de un segun de lo que se hace en cada uno.", "SANO entonces gritar \"bien\""},
	{"otro", GrupoControl, "El caso de un segun que atrapa todos los valores no nombrados.", "otro entonces gritar \"otro caso\""},
	{"mientras", GrupoControl, "Repite un bloque mientras la condición sea verdadera.", "mientras vida > 0\n    vida = vida - 10\nfin"},
	{"recorrer", GrupoControl, "Repite un bloque por cada elemento de una colección o por cada número de un rango.", "recorrer n de 1 hasta 5\n    gritar n\nfin"},
	{"en", GrupoControl, "Indica qué colección recorre un recorrer.", "recorrer pokemon en equipo_ash\n    gritar pokemon\nfin"},
	{"hasta", GrupoControl, "Marca el final (incluido) de un rango.", "recorrer n de 1 hasta 10\n    gritar n\nfin"},
	{"huir", GrupoControl, "Sale del ciclo más cercano.", "si n igual 5\n    huir\nfin"},
	{"siguiente", GrupoControl, "Salta a la próxima vuelta del ciclo más cercano.", "si n igual 2\n    siguiente\nfin"},

	// E/S y datos.
	{"gritar", GrupoDatos, "Muestra valores en la salida y salta de línea.", "gritar \"Vida: \", vida"},
	{"capturar", GrupoDatos, "Pide un dato al usuario y lo guarda. Si no es del tipo correcto, vuelve a preguntar.", "capturar(nombre, \"¿Cómo se llama? \")"},
	{"convertir", GrupoDatos, "Pasa un valor de un tipo a otro cuando la tabla de efectividades lo permite.", "roca n = convertir(texto) a roca"},
	{"tamaño", GrupoDatos, "Cuántos elementos tiene un equipo o una mochila, o cuántas letras un texto.", "gritar tamaño(equipo_ash)"},
	{"aleatorio", GrupoDatos, "Un roca al azar entre dos extremos, ambos incluidos.", "roca dado = aleatorio(1, 6)"},
	{"redondear", GrupoDatos, "Redondea un agua al roca más cercano.", "roca redondo = redondear(7.5)"},
	{"sumar", GrupoDatos, "Agrega un valor al final de un equipo.", "sumar \"Bulbasaur\" a equipo_ash"},
	{"quitar", GrupoDatos, "Saca un elemento de un equipo por su posición, o de una mochila por su clave.", "quitar equipo_ash[2]"},
	{"contiene", GrupoDatos, "Dice si un equipo tiene un valor, si una mochila tiene una clave o si un texto contiene otro.", "si equipo_ash contiene \"Pikachu\"\n    gritar \"¡Ahí está!\"\nfin"},

	// Operadores palabra.
	{"resto", GrupoOperadores, "El residuo de una división entera.", "gritar 10 resto 3"},
	{"igual", GrupoOperadores, "Compara si dos valores son iguales.", "si vida igual 0\n    gritar \"debilitado\"\nfin"},
	{"diferente", GrupoOperadores, "Compara si dos valores son distintos.", "si rival diferente fantasma\n    gritar rival\nfin"},
	{"y", GrupoOperadores, "Verdadero si las dos condiciones lo son.", "si vida > 0 y nivel > 5\n    gritar \"listo\"\nfin"},
	{"o", GrupoOperadores, "Verdadero si al menos una de las condiciones lo es.", "si vida igual 0 o dormido\n    gritar \"no puede atacar\"\nfin"},
	{"no", GrupoOperadores, "Invierte una condición.", "si no listo\n    gritar \"espera\"\nfin"},
}
