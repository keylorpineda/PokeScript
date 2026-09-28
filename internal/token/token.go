// Package token define los tipos de token de PokeScript y la tabla de
// palabras reservadas (sección 1 de la especificación).
package token

// Kind identifica el tipo de un token.
type Kind int

const (
	ILLEGAL Kind = iota // carácter o secuencia no reconocida
	EOF                 // fin del archivo
	NEWLINE             // fin de instrucción: el salto de línea es significativo

	// Identificadores y literales.
	literalStart
	IDENT      // vida, Pokemon, calcular_dano
	ROCA_LIT   // 42 (sin signo: el - siempre es operador)
	AGUA_LIT   // 6.9 (dígitos obligatorios a ambos lados del punto)
	FUEGO_LIT  // 'c'
	PLANTA_LIT // "texto"
	literalEnd

	// Símbolos.
	simboloStart
	LPAREN   // (
	RPAREN   // )
	LBRACKET // [
	RBRACKET // ]
	LBRACE   // {
	RBRACE   // }
	COMMA    // ,
	COLON    // :
	DOT      // .
	ASSIGN   // =
	PLUS     // +
	MINUS    // -
	STAR     // *
	SLASH    // /
	GT       // >
	LT       // <
	GE       // >=
	LE       // <=
	simboloEnd

	// Palabras reservadas.
	palabraStart

	// Declaración de tipos.
	ESPECIE
	FICHA
	MEDALLA

	// Tipos.
	ROCA
	AGUA
	FUEGO
	PLANTA
	ELECTRICO
	EQUIPO
	MOCHILA
	POSIBLE

	// Conectores de tipo.
	DE
	A

	// Literales.
	VERDADERO
	FALSO
	FANTASMA

	// Estructura.
	COMBATE
	MOVIMIENTO
	ENTREGAR
	FIN
	ENSENAR
	DESDE

	// Control.
	SI
	SINO
	SINO_SI // único token de dos términos: el lexer une "sino" + "si"
	SEGUN
	ENTONCES
	OTRO
	MIENTRAS
	RECORRER
	EN
	HASTA
	HUIR
	SIGUIENTE

	// Entrada/salida y datos.
	GRITAR
	CAPTURAR
	CONVERTIR
	TAMANO
	ALEATORIO
	REDONDEAR
	SUMAR
	QUITAR
	CONTIENE

	// Operadores palabra.
	RESTO
	IGUAL
	DIFERENTE
	Y
	O
	NO

	palabraEnd
)

var nombres = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	NEWLINE: "NEWLINE",

	IDENT:      "IDENT",
	ROCA_LIT:   "ROCA_LIT",
	AGUA_LIT:   "AGUA_LIT",
	FUEGO_LIT:  "FUEGO_LIT",
	PLANTA_LIT: "PLANTA_LIT",

	LPAREN:   "(",
	RPAREN:   ")",
	LBRACKET: "[",
	RBRACKET: "]",
	LBRACE:   "{",
	RBRACE:   "}",
	COMMA:    ",",
	COLON:    ":",
	DOT:      ".",
	ASSIGN:   "=",
	PLUS:     "+",
	MINUS:    "-",
	STAR:     "*",
	SLASH:    "/",
	GT:       ">",
	LT:       "<",
	GE:       ">=",
	LE:       "<=",

	ESPECIE:    "especie",
	FICHA:      "ficha",
	MEDALLA:    "medalla",
	ROCA:       "roca",
	AGUA:       "agua",
	FUEGO:      "fuego",
	PLANTA:     "planta",
	ELECTRICO:  "electrico",
	EQUIPO:     "equipo",
	MOCHILA:    "mochila",
	POSIBLE:    "posible",
	DE:         "de",
	A:          "a",
	VERDADERO:  "verdadero",
	FALSO:      "falso",
	FANTASMA:   "fantasma",
	COMBATE:    "combate",
	MOVIMIENTO: "movimiento",
	ENTREGAR:   "entregar",
	FIN:        "fin",
	ENSENAR:    "enseñar",
	DESDE:      "desde",
	SI:         "si",
	SINO:       "sino",
	SINO_SI:    "sino si",
	SEGUN:      "segun",
	ENTONCES:   "entonces",
	OTRO:       "otro",
	MIENTRAS:   "mientras",
	RECORRER:   "recorrer",
	EN:         "en",
	HASTA:      "hasta",
	HUIR:       "huir",
	SIGUIENTE:  "siguiente",
	GRITAR:     "gritar",
	CAPTURAR:   "capturar",
	CONVERTIR:  "convertir",
	TAMANO:     "tamaño",
	ALEATORIO:  "aleatorio",
	REDONDEAR:  "redondear",
	SUMAR:      "sumar",
	QUITAR:     "quitar",
	CONTIENE:   "contiene",
	RESTO:      "resto",
	IGUAL:      "igual",
	DIFERENTE:  "diferente",
	Y:          "y",
	O:          "o",
	NO:         "no",
}

// String devuelve el lexema de un símbolo o palabra reservada, o el nombre
// del tipo de token en los demás casos.
func (k Kind) String() string {
	if k >= 0 && int(k) < len(nombres) && nombres[k] != "" {
		return nombres[k]
	}
	return "Kind(?)"
}

// palabras mapea cada palabra reservada de un solo término a su Kind.
// "sino si" no está aquí: el lexer lo arma al ver SINO seguido de SI.
var palabras = func() map[string]Kind {
	m := make(map[string]Kind)
	for k := palabraStart + 1; k < palabraEnd; k++ {
		if k != SINO_SI {
			m[nombres[k]] = k
		}
	}
	return m
}()

// Buscar devuelve el Kind de la palabra reservada ident, o IDENT si ident
// no es una palabra reservada. Distingue mayúsculas: "Roca" es un IDENT.
func Buscar(ident string) Kind {
	if k, ok := palabras[ident]; ok {
		return k
	}
	return IDENT
}

// PalabrasReservadas devuelve las palabras reservadas en el orden de la
// especificación, sin "sino si".
func PalabrasReservadas() []string {
	lista := make([]string, 0, len(palabras))
	for k := palabraStart + 1; k < palabraEnd; k++ {
		if k != SINO_SI {
			lista = append(lista, nombres[k])
		}
	}
	return lista
}

// EsPalabraReservada informa si k es una palabra reservada (incluye SINO_SI).
func (k Kind) EsPalabraReservada() bool { return k > palabraStart && k < palabraEnd }

// EsLiteral informa si k es un identificador o un literal numérico, de carácter o de texto.
func (k Kind) EsLiteral() bool { return k > literalStart && k < literalEnd }

// EsSimbolo informa si k es un símbolo de puntuación u operador.
func (k Kind) EsSimbolo() bool { return k > simboloStart && k < simboloEnd }

// EsTipoSimple informa si k es roca, agua, fuego, planta o electrico.
func (k Kind) EsTipoSimple() bool { return k >= ROCA && k <= ELECTRICO }

// Categoria agrupa los tokens para el resaltado del editor (sección 9).
type Categoria string

const (
	CatReservada     Categoria = "reservada"
	CatTipo          Categoria = "tipo"
	CatLiteral       Categoria = "literal"
	CatIdentificador Categoria = "identificador"
	CatOperador      Categoria = "operador"
	CatSimbolo       Categoria = "simbolo"
	CatOtro          Categoria = "otro"
)

// Categoria devuelve la categoría de resaltado de k.
func (k Kind) Categoria() Categoria {
	switch {
	case k == IDENT:
		return CatIdentificador
	case k.EsLiteral(), k == VERDADERO, k == FALSO, k == FANTASMA:
		return CatLiteral
	case k >= ROCA && k <= POSIBLE:
		return CatTipo
	case k >= RESTO && k <= NO, k == CONTIENE, k >= PLUS && k <= LE, k == ASSIGN:
		return CatOperador
	case k.EsPalabraReservada():
		return CatReservada
	case k.EsSimbolo():
		return CatSimbolo
	default:
		return CatOtro
	}
}

// Token es la unidad que produce el lexer. Line, Col y Len se propagan a
// cada nodo del AST y a cada diagnóstico: sin ellos no hay subrayado.
type Token struct {
	Kind   Kind
	Lexeme string
	Line   int // base 1
	Col    int // base 1, en runas
	Len    int // en runas
}
