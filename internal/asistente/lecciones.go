package asistente

import (
	"sort"

	"github.com/keylorpineda/PokeScript/internal/diag"
)

// Leccion es la explicación que muestra el panel del asistente para un
// diagnóstico: qué significa, en palabras simples, y un ejemplo correcto.
type Leccion struct {
	Titulo  string `json:"titulo"`
	Texto   string `json:"texto"`
	Ejemplo string `json:"ejemplo,omitempty"` // un programa completo que compila
}

// Explicar devuelve la lección de un diagnóstico, buscando primero por su
// código y, si no hay una específica, por su categoría (sección 7.1:
// «un mapa de Category + subcódigo → plantilla de explicación»).
func Explicar(d diag.Diagnostic) Leccion {
	if l, ok := lecciones[d.Code]; ok {
		return l
	}
	if l, ok := porCategoria[d.Category]; ok {
		return l
	}
	return Leccion{Titulo: d.Heading, Texto: d.Desc}
}

// CodigosConLeccion devuelve, en orden, los códigos que tienen una lección
// propia. El menú de consulta del IDE los puede listar.
func CodigosConLeccion() []string {
	codigos := make([]string, 0, len(lecciones))
	for c := range lecciones {
		codigos = append(codigos, c)
	}
	sort.Strings(codigos)
	return codigos
}

var porCategoria = map[diag.Categoria]Leccion{
	diag.Lexico: {
		Titulo: "Algo en el texto no se reconoce",
		Texto:  "Antes de entender tu programa, PokeScript lo separa en palabras, números, textos y símbolos. Aquí encontró algo que no forma parte del lenguaje.",
	},
	diag.Sintactico: {
		Titulo: "La forma de la instrucción no es la esperada",
		Texto:  "Cada instrucción tiene una forma fija, como una frase con su orden. Revisa que no falte ni sobre ninguna palabra o símbolo.",
	},
	diag.Semantico: {
		Titulo: "La instrucción está bien escrita, pero no tiene sentido",
		Texto:  "La forma es correcta, pero combina cosas que no encajan: tipos que no se llevan, nombres que no existen o reglas del lenguaje que no se cumplen.",
	},
	diag.Importacion: {
		Titulo: "Un problema entre archivos",
		Texto:  "Con «enseñar … desde» un archivo usa lo que otro declara. El archivo y el nombre tienen que existir, y dos archivos no pueden depender uno del otro en círculo.",
	},
	diag.Ejecucion: {
		Titulo: "El programa falló mientras corría",
		Texto:  "El programa estaba bien escrito, pero al ejecutarse encontró un valor con el que no podía seguir. Revisa qué valor tenían los datos en ese momento.",
	},
}

var lecciones = map[string]Leccion{
	// ─── Léxicos y sintácticos ─────────────────────────────────────────
	"cadena-sin-cerrar": {
		Titulo:  "Un texto sin su comilla de cierre",
		Texto:   "Un texto (planta) empieza y termina con comillas dobles en la misma línea. Sin la comilla final, PokeScript no sabe dónde termina.",
		Ejemplo: "combate\n    gritar \"¡Hola!\"\nfin\n",
	},
	"simbolo-de-otro-lenguaje": {
		Titulo:  "Ese símbolo es de otro lenguaje",
		Texto:   "PokeScript usa palabras para comparar y combinar: «igual», «diferente», «y», «o» y «resto», en lugar de ==, !=, &&, || y %.",
		Ejemplo: "combate\n    roca vida = 10\n    si vida igual 10 y vida resto 2 igual 0\n        gritar \"par\"\n    fin\nfin\n",
	},
	"bloque-sin-cerrar": {
		Titulo:  "Un bloque se quedó sin su fin",
		Texto:   "Cada combate, movimiento, si, segun, mientras y recorrer abre un bloque que se cierra con su propio «fin». La sangría ayuda a ver qué fin cierra cada bloque: ponlo a la misma altura que la línea que lo abre.",
		Ejemplo: "combate\n    roca n = 0\n    mientras n < 3\n        si n igual 1\n            gritar n\n        fin\n        n = n + 1\n    fin\nfin\n",
	},
	"operador-no-encadenable": {
		Titulo:  "Dos comparaciones seguidas",
		Texto:   "En matemática se escribe 1 < x < 10, pero en PokeScript cada comparación va sola. Para pedir las dos cosas, únelas con «y».",
		Ejemplo: "combate\n    roca x = 5\n    si x > 1 y x < 10\n        gritar \"en rango\"\n    fin\nfin\n",
	},
	"dos-instrucciones": {
		Titulo:  "Una instrucción por línea",
		Texto:   "En PokeScript el salto de línea termina cada instrucción. Si pones dos en la misma línea, no sabe dónde termina la primera.",
		Ejemplo: "combate\n    roca x = 1\n    gritar x\nfin\n",
	},
	"rango-sin-hasta": {
		Titulo:  "Un rango se escribe con «hasta»",
		Texto:   "Para contar de un número a otro se usa «recorrer n de inicio hasta fin». La palabra «a» se usa en otros lugares, como en «sumar valor a equipo».",
		Ejemplo: "combate\n    recorrer n de 1 hasta 5\n        gritar n\n    fin\nfin\n",
	},
	"asignacion-esperada": {
		Titulo:  "Un valor suelto no hace nada",
		Texto:   "Una línea que empieza con un nombre debe guardar algo (x = …), declarar un dato (Tipo nombre) o llamar a un movimiento (nombre(…)). Si empezaba con una palabra del lenguaje, revisa cómo está escrita.",
		Ejemplo: "combate\n    roca vida = 10\n    vida = vida + 1\n    gritar vida\nfin\n",
	},
	"instruccion-fuera-de-bloque": {
		Titulo:  "Las instrucciones van dentro de un bloque",
		Texto:   "Fuera de los bloques solo se declaran cosas: medallas, especies, fichas, movimientos y el combate. Lo que el programa hace va dentro de combate o de un movimiento.",
		Ejemplo: "combate\n    gritar \"¡Hola!\"\nfin\n",
	},

	// ─── Tipos ─────────────────────────────────────────────────────────
	"tipo-incompatible": {
		Titulo:  "Estos tipos no se llevan",
		Texto:   "Como en los combates, no todos los tipos se combinan. La tabla de efectividades dice qué conversiones son automáticas, cuáles necesitan «convertir» y cuáles no se pueden hacer.",
		Ejemplo: "combate\n    roca nivel = 5\n    planta texto = \"Nivel \" + convertir(nivel) a planta\n    gritar texto\nfin\n",
	},
	"asignacion-incompatible": {
		Titulo:  "El valor no cabe en ese dato",
		Texto:   "Cada dato guarda un solo tipo. Un roca se guarda automáticamente en un agua, pero al revés hay que usar «convertir» o «redondear».",
		Ejemplo: "combate\n    agua precision = 7.8\n    roca redondo = redondear(precision)\n    gritar redondo\nfin\n",
	},
	"condicion-no-electrico": {
		Titulo:  "Una condición es verdadera o falsa",
		Texto:   "si y mientras necesitan un electrico. PokeScript no adivina: «si vida» no dice nada, pero «si vida > 0» sí.",
		Ejemplo: "combate\n    roca vida = 10\n    si vida > 0\n        gritar \"sigue en pie\"\n    fin\nfin\n",
	},
	"segun-incompleto": {
		Titulo:  "Al segun le faltan casos",
		Texto:   "Cuando un segun elige entre los valores de una especie, tiene que decir qué hacer con cada uno. Así ningún caso se escapa.",
		Ejemplo: "especie Clima\n    SOL, LLUVIA\nfin\ncombate\n    Clima hoy = SOL\n    segun hoy\n        SOL    entonces gritar \"¡Lanzallamas!\"\n        LLUVIA entonces gritar \"¡Hidrobomba!\"\n    fin\nfin\n",
	},
	"segun-sin-otro": {
		Titulo:  "Un segun sobre números necesita «otro»",
		Texto:   "Un roca, un fuego o un texto pueden tener infinitos valores. La rama «otro entonces …» dice qué hacer con todos los que no nombraste.",
		Ejemplo: "combate\n    roca dado = aleatorio(1, 6)\n    segun dado\n        6    entonces gritar \"¡Crítico!\"\n        otro entonces gritar \"Golpe normal\"\n    fin\nfin\n",
	},
	"posible-sin-comprobar": {
		Titulo:  "Un posible puede estar vacío",
		Texto:   "Un dato posible puede valer fantasma. Antes de usar su valor, compruébalo con «si dato diferente fantasma», o da un valor de respaldo con «sino».",
		Ejemplo: "combate\n    posible planta rival = fantasma\n    si rival diferente fantasma\n        gritar rival\n    fin\n    gritar rival sino \"nadie\"\nfin\n",
	},

	// ─── Nombres y datos ───────────────────────────────────────────────
	"nombre-no-declarado": {
		Titulo:  "Ese nombre no existe aquí",
		Texto:   "Antes de usar un dato hay que declararlo con su tipo. Si está en otro archivo, se trae con «enseñar … desde». Revisa también que esté bien escrito: vida y Vida son nombres distintos.",
		Ejemplo: "combate\n    roca vida = 10\n    gritar vida\nfin\n",
	},
	"dato-sin-valor": {
		Titulo:  "El dato todavía no tiene valor",
		Texto:   "Un dato declarado sin valor necesita recibir uno (con = o con capturar) antes de leerlo, por todos los caminos: si solo lo asignas dentro de un si, puede que nunca se asigne.",
		Ejemplo: "combate\n    planta nombre\n    capturar(nombre, \"¿Nombre? \")\n    gritar nombre\nfin\n",
	},
	"medalla-reasignada": {
		Titulo:  "Una medalla no cambia",
		Texto:   "Una medalla es un valor fijo, como las medallas que ganas: una vez obtenida, no cambia. Si el valor tiene que cambiar, declara un dato común.",
		Ejemplo: "medalla roca VIDA_MAXIMA = 100\ncombate\n    roca vida = VIDA_MAXIMA\n    vida = vida - 10\n    gritar vida\nfin\n",
	},
	"medalla-modificada": {
		Titulo:  "El contenido de una medalla tampoco cambia",
		Texto:   "Si una medalla guarda un equipo o una ficha, no se le pueden agregar, quitar ni cambiar elementos. Copia su valor a un dato común y cambia ese.",
		Ejemplo: "medalla equipo de planta INICIALES = [\"Bulbasaur\", \"Charmander\"]\ncombate\n    equipo de planta mio = INICIALES\n    sumar \"Squirtle\" a mio\n    gritar mio\nfin\n",
	},
	"nombre-ocultado": {
		Titulo:  "Ese nombre ya está en uso",
		Texto:   "Dentro de un bloque no se puede declarar un nombre que ya existe afuera: quedaría escondido y no se sabría a cuál se refiere cada uso. Elige un nombre nuevo que diga qué guarda.",
		Ejemplo: "combate\n    roca vida = 10\n    si vida > 5\n        roca extra = 2\n        gritar vida + extra\n    fin\nfin\n",
	},
	"corte-fuera-de-ciclo": {
		Titulo:  "huir y siguiente son para ciclos",
		Texto:   "huir sale del ciclo más cercano y siguiente salta a su próxima vuelta. Fuera de un mientras o un recorrer no tienen de dónde salir.",
		Ejemplo: "combate\n    recorrer n de 1 hasta 10\n        si n igual 5\n            huir\n        fin\n        gritar n\n    fin\nfin\n",
	},
	"coleccion-en-recorrido": {
		Titulo:  "No se cambia lo que se está recorriendo",
		Texto:   "Si una colección cambia mientras la recorres, no se sabe qué elementos quedan por visitar. Junta los cambios en otra colección.",
		Ejemplo: "combate\n    equipo de roca niveles = [5, 12, 8]\n    equipo de roca altos = []\n    recorrer n en niveles\n        si n > 6\n            sumar n a altos\n        fin\n    fin\n    gritar altos\nfin\n",
	},
	"variable-de-recorrido": {
		Titulo:  "La variable del recorrido es de solo lectura",
		Texto:   "En cada vuelta, recorrer le da a su variable el valor del elemento que toca. Cambiarla no cambia la colección; si necesitas otro valor, cópialo a un dato nuevo.",
		Ejemplo: "combate\n    recorrer n de 1 hasta 3\n        roca doble = n * 2\n        gritar doble\n    fin\nfin\n",
	},

	// ─── Movimientos ───────────────────────────────────────────────────
	"falta-entregar": {
		Titulo:  "El movimiento no siempre entrega su valor",
		Texto:   "Un movimiento con tipo promete un valor. Si hay un camino (por ejemplo, un si sin sino) que llega al fin sin «entregar», esa promesa no se cumple.",
		Ejemplo: "movimiento roca mayor(roca x, roca z)\n    si x > z\n        entregar x\n    fin\n    entregar z\nfin\ncombate\n    gritar mayor(3, 7)\nfin\n",
	},
	"argumento-incompatible": {
		Titulo:  "El movimiento recibe otro tipo",
		Texto:   "Cada parámetro de un movimiento tiene un tipo, y cada valor que le pasas debe encajar en él, en el mismo orden.",
		Ejemplo: "movimiento roca doble(roca x)\n    entregar x * 2\nfin\ncombate\n    gritar doble(21)\nfin\n",
	},

	// ─── Importaciones ─────────────────────────────────────────────────
	"archivo-inexistente": {
		Titulo: "No se encontró el archivo",
		Texto:  "«enseñar … desde» busca el archivo en la misma carpeta del proyecto, con el nombre exacto y la extensión .pks.",
	},
	"nombre-inexistente": {
		Titulo: "El archivo no declara ese nombre",
		Texto:  "Solo se importan movimientos, especies, fichas y medallas que el otro archivo declara. Los valores de una especie llegan solos al importar la especie.",
	},
	"importacion-circular": {
		Titulo: "Dos archivos se necesitan en círculo",
		Texto:  "Si a.pks usa algo de b.pks y b.pks usa algo de a.pks, ninguno puede cargarse primero. Mueve lo que ambos comparten a un tercer archivo.",
	},

	// ─── Ejecución ─────────────────────────────────────────────────────
	"division-cero": {
		Titulo:  "No se puede dividir entre cero",
		Texto:   "Dividir entre cero no tiene resultado. Antes de dividir, comprueba que el divisor no sea 0.",
		Ejemplo: "combate\n    roca total = 10\n    roca partes = 0\n    si partes diferente 0\n        gritar total / partes\n    sino\n        gritar \"no hay partes\"\n    fin\nfin\n",
	},
	"indice-fuera-rango": {
		Titulo:  "Esa posición no existe",
		Texto:   "Los equipos se cuentan desde 1 hasta tamaño(equipo). Una posición fuera de ese rango no tiene elemento.",
		Ejemplo: "combate\n    equipo de planta e = [\"Pikachu\", \"Eevee\"]\n    gritar e[tamaño(e)]\nfin\n",
	},
	"clave-inexistente": {
		Titulo:  "La mochila no tiene esa clave",
		Texto:   "Leer una clave que no está en la mochila es un error. Compruébalo antes con «contiene».",
		Ejemplo: "combate\n    mochila de planta a roca bolsa = {\"Poción\": 2}\n    si bolsa contiene \"Poción\"\n        gritar bolsa[\"Poción\"]\n    fin\nfin\n",
	},
	"desbordamiento": {
		Titulo: "El número es demasiado grande",
		Texto:  "Un roca guarda enteros hasta unos nueve trillones. Si una cuenta se pasa de ese límite, el programa se detiene en lugar de dar un resultado incorrecto.",
	},
	"recursion-excedida": {
		Titulo:  "Demasiadas llamadas anidadas",
		Texto:   "Un movimiento que se llama a sí mismo necesita un caso en el que deja de hacerlo. Sin él, las llamadas nunca terminan.",
		Ejemplo: "movimiento roca factorial(roca n)\n    si n <= 1\n        entregar 1\n    fin\n    entregar n * factorial(n - 1)\nfin\ncombate\n    gritar factorial(5)\nfin\n",
	},
}
