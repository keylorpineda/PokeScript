package lexer

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// Resultado es lo que produce el análisis léxico de un archivo.
type Resultado struct {
	Tokens       []token.Token
	Diagnosticos []diag.Diagnostic
	// Sangrias guarda, para cada línea con código, la columna (base 1) de su
	// primer token. El analizador la usa para la advertencia de indentación
	// inconsistente (sección 1.5); el parser no la necesita.
	Sangrias map[int]int
}

// Analizar convierte el texto de un archivo .pks en tokens.
//
// Reglas principales (sección 1 de la especificación):
//   - El salto de línea es significativo: cada línea con código termina en un
//     NEWLINE. Las líneas vacías y las de solo comentario no emiten NEWLINE.
//   - "sino" seguido de "si" en la misma línea se une en un único SINO_SI.
//   - Tras un número, un punto seguido de dígito continúa un literal agua;
//     en cualquier otro caso es el token DOT.
//   - En PLANTA_LIT y FUEGO_LIT, Lexeme guarda el valor ya decodificado (sin
//     comillas y con los escapes resueltos); Col y Len siguen apuntando al
//     texto original, comillas incluidas.
//
// El análisis no se detiene en el primer error: registra el diagnóstico y
// sigue, para reportar todos los problemas de una vez. Los errores ya
// reportados dejan un token ILLEGAL; el parser debe saltarlo sin volver a
// reportarlo. Los símbolos de otros lenguajes (==, !=, &&, ||, %) se reportan
// y se reemplazan por su palabra en PokeScript para que el parser siga.
func Analizar(archivo, fuente string) Resultado {
	l := &lexer{
		archivo:  archivo,
		src:      []rune(strings.TrimPrefix(fuente, "\uFEFF")),
		linea:    1,
		col:      1,
		sangrias: map[int]int{},
	}
	l.analizar()
	return Resultado{Tokens: l.tokens, Diagnosticos: l.diags.Items(), Sangrias: l.sangrias}
}

type lexer struct {
	archivo string
	src     []rune
	pos     int // índice en src
	linea   int // base 1
	col     int // base 1

	tokens         []token.Token
	diags          diag.Lista
	sangrias       map[int]int
	lineaConCodigo bool // ya se emitió algún token en la línea actual
}

func (l *lexer) actual() rune { return l.ver(0) }

func (l *lexer) ver(n int) rune {
	if l.pos+n < len(l.src) {
		return l.src[l.pos+n]
	}
	return 0
}

func (l *lexer) fin() bool { return l.pos >= len(l.src) }

// avanzar consume una runa que no es salto de línea.
func (l *lexer) avanzar() rune {
	r := l.src[l.pos]
	l.pos++
	l.col++
	return r
}

func (l *lexer) emitir(k token.Kind, lexema string, linea, col, largo int) {
	if !l.lineaConCodigo {
		l.lineaConCodigo = true
		l.sangrias[linea] = col
	}
	l.tokens = append(l.tokens, token.Token{Kind: k, Lexeme: lexema, Line: linea, Col: col, Len: largo})
}

func (l *lexer) error(codigo string, linea, col, largo int, desc, causa, sugerencia string) {
	l.diags.Agregar(diag.Diagnostic{
		Severity: diag.Error,
		Category: diag.Lexico,
		Code:     codigo,
		Heading:  diag.EncabezadoSintaxis,
		File:     l.archivo,
		Line:     linea,
		Col:      col,
		Len:      largo,
		Desc:     desc,
		Cause:    causa,
		Suggest:  sugerencia,
	})
}

func (l *lexer) analizar() {
	for !l.fin() {
		r := l.actual()
		switch {
		case r == '\n':
			l.saltoDeLinea()
		case r == ' ' || r == '\t' || r == '\r':
			l.avanzar()
		case r == '/' && l.ver(1) == '/':
			for !l.fin() && l.actual() != '\n' {
				l.avanzar()
			}
		case unicode.IsLetter(r):
			l.identificador()
		case esDigito(r):
			l.numero()
		case r == '"':
			l.texto()
		case r == '\'':
			l.caracter()
		default:
			l.simbolo()
		}
	}
	if l.lineaConCodigo {
		l.emitir(token.NEWLINE, "", l.linea, l.col, 0)
	}
	l.tokens = append(l.tokens, token.Token{Kind: token.EOF, Line: l.linea, Col: l.col})
}

func (l *lexer) saltoDeLinea() {
	if l.lineaConCodigo {
		l.emitir(token.NEWLINE, "\n", l.linea, l.col, 1)
	}
	l.pos++
	l.linea++
	l.col = 1
	l.lineaConCodigo = false
}

func esDigito(r rune) bool { return r >= '0' && r <= '9' }

func esParteDeIdentificador(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

func (l *lexer) identificador() {
	inicio, linea, col := l.pos, l.linea, l.col
	for !l.fin() && esParteDeIdentificador(l.actual()) {
		l.avanzar()
	}
	palabra := string(l.src[inicio:l.pos])
	kind := token.Buscar(palabra)

	if kind == token.SINO && l.siguePalabraSi() {
		for l.actual() == ' ' || l.actual() == '\t' {
			l.avanzar()
		}
		l.avanzar() // s
		l.avanzar() // i
		l.emitir(token.SINO_SI, string(l.src[inicio:l.pos]), linea, col, l.pos-inicio)
		return
	}
	l.emitir(kind, palabra, linea, col, l.pos-inicio)
}

// siguePalabraSi mira, sin consumir, si después de espacios en la misma
// línea viene la palabra "si" completa.
func (l *lexer) siguePalabraSi() bool {
	i := l.pos
	for i < len(l.src) && (l.src[i] == ' ' || l.src[i] == '\t') {
		i++
	}
	if i == l.pos || i+1 >= len(l.src) || l.src[i] != 's' || l.src[i+1] != 'i' {
		return false
	}
	return i+2 >= len(l.src) || !esParteDeIdentificador(l.src[i+2])
}

func (l *lexer) numero() {
	inicio, linea, col := l.pos, l.linea, l.col
	for esDigito(l.actual()) {
		l.avanzar()
	}
	kind := token.ROCA_LIT
	if l.actual() == '.' && esDigito(l.ver(1)) {
		kind = token.AGUA_LIT
		l.avanzar()
		for esDigito(l.actual()) {
			l.avanzar()
		}
	}

	// "3vidas": un identificador no puede empezar con un número.
	if esParteDeIdentificador(l.actual()) {
		for !l.fin() && esParteDeIdentificador(l.actual()) {
			l.avanzar()
		}
		texto := string(l.src[inicio:l.pos])
		l.error("identificador-con-numero", linea, col, l.pos-inicio,
			fmt.Sprintf("«%s» no es un nombre válido.", texto),
			"los nombres de datos y movimientos deben empezar con una letra.",
			"mueve el número al final del nombre o cámbialo por una palabra.")
		l.emitir(token.ILLEGAL, texto, linea, col, l.pos-inicio)
		return
	}

	texto := string(l.src[inicio:l.pos])
	if kind == token.ROCA_LIT {
		if _, err := strconv.ParseInt(texto, 10, 64); err != nil {
			l.error("numero-muy-grande", linea, col, l.pos-inicio,
				fmt.Sprintf("el número %s es demasiado grande para un roca.", texto),
				"un roca guarda enteros de hasta 9223372036854775807.",
				"usa un número más pequeño o un agua si necesitas valores enormes.")
		}
	}
	l.emitir(kind, texto, linea, col, l.pos-inicio)
}

// texto lee un literal planta: "…" con los escapes \" \n \\.
func (l *lexer) texto() {
	inicio, linea, col := l.pos, l.linea, l.col
	l.avanzar() // comilla de apertura
	var valor strings.Builder
	for {
		if l.fin() || l.actual() == '\n' {
			l.error("cadena-sin-cerrar", linea, col, l.pos-inicio,
				"el texto que empieza aquí no tiene comilla de cierre.",
				"un texto debe abrir y cerrar con \" en la misma línea.",
				"agrega \" al final del texto.")
			l.emitir(token.ILLEGAL, string(l.src[inicio:l.pos]), linea, col, l.pos-inicio)
			return
		}
		r := l.avanzar()
		if r == '"' {
			break
		}
		if r == '\\' {
			valor.WriteString(l.escape(map[rune]string{'"': "\"", 'n': "\n", '\\': "\\"}, `\" \n \\`))
			continue
		}
		valor.WriteRune(r)
	}
	l.emitir(token.PLANTA_LIT, valor.String(), linea, col, l.pos-inicio)
}

// caracter lee un literal fuego: exactamente un carácter entre comillas simples.
func (l *lexer) caracter() {
	inicio, linea, col := l.pos, l.linea, l.col
	l.avanzar() // comilla de apertura
	var valor []rune
	for {
		if l.fin() || l.actual() == '\n' {
			l.error("caracter-sin-cerrar", linea, col, l.pos-inicio,
				"el carácter que empieza aquí no tiene comilla de cierre.",
				"un fuego se escribe entre comillas simples, como 'P'.",
				"agrega ' después del carácter.")
			l.emitir(token.ILLEGAL, string(l.src[inicio:l.pos]), linea, col, l.pos-inicio)
			return
		}
		r := l.avanzar()
		if r == '\'' {
			break
		}
		if r == '\\' {
			valor = append(valor, []rune(l.escape(map[rune]string{'\'': "'", 'n': "\n", '\\': "\\"}, `\' \n \\`))...)
			continue
		}
		valor = append(valor, r)
	}
	largo := l.pos - inicio
	switch {
	case len(valor) == 0:
		l.error("caracter-vacio", linea, col, largo,
			"un fuego no puede estar vacío.",
			"un fuego guarda exactamente un carácter.",
			"escribe un carácter entre las comillas, como 'A'.")
		l.emitir(token.ILLEGAL, "", linea, col, largo)
	case len(valor) > 1:
		l.error("caracter-largo", linea, col, largo,
			fmt.Sprintf("'%s' tiene %d caracteres y un fuego guarda solo uno.", string(valor), len(valor)),
			"las comillas simples son para un solo carácter.",
			fmt.Sprintf("si querías un texto, usa comillas dobles: \"%s\".", string(valor)))
		l.emitir(token.ILLEGAL, string(valor), linea, col, largo)
	default:
		l.emitir(token.FUEGO_LIT, string(valor), linea, col, largo)
	}
}

// escape procesa la secuencia que sigue a una barra invertida ya consumida.
// Una secuencia desconocida es error léxico y se conserva tal cual.
func (l *lexer) escape(validos map[rune]string, lista string) string {
	linea, col := l.linea, l.col-1
	if l.fin() || l.actual() == '\n' {
		return "\\"
	}
	r := l.avanzar()
	if v, ok := validos[r]; ok {
		return v
	}
	l.error("escape-desconocido", linea, col, 2,
		fmt.Sprintf("la secuencia \\%c no se reconoce.", r),
		"después de \\ solo se permiten estas secuencias: "+lista+".",
		"si querías escribir una barra invertida, usa \\\\.")
	return "\\" + string(r)
}

// otrosLenguajes traduce símbolos que el estudiante trae de otros lenguajes
// a su equivalente en PokeScript. Se reporta el error y se emite el token
// correcto para que el parser pueda seguir.
var otrosLenguajes = map[string]token.Kind{
	"==": token.IGUAL,
	"!=": token.DIFERENTE,
	"&&": token.Y,
	"||": token.O,
}

var simbolos = map[rune]token.Kind{
	'(': token.LPAREN,
	')': token.RPAREN,
	'[': token.LBRACKET,
	']': token.RBRACKET,
	'{': token.LBRACE,
	'}': token.RBRACE,
	',': token.COMMA,
	':': token.COLON,
	'.': token.DOT,
	'=': token.ASSIGN,
	'+': token.PLUS,
	'-': token.MINUS,
	'*': token.STAR,
	'/': token.SLASH,
	'>': token.GT,
	'<': token.LT,
}

func (l *lexer) simbolo() {
	linea, col := l.linea, l.col
	r, sig := l.actual(), l.ver(1)
	par := string([]rune{r, sig})

	if k, ok := otrosLenguajes[par]; ok {
		l.avanzar()
		l.avanzar()
		l.error("simbolo-de-otro-lenguaje", linea, col, 2,
			fmt.Sprintf("«%s» no existe en PokeScript.", par),
			fmt.Sprintf("en PokeScript esta operación se escribe con la palabra «%s».", k),
			fmt.Sprintf("cambia «%s» por «%s».", par, k))
		l.emitir(k, par, linea, col, 2)
		return
	}
	if par == ">=" || par == "<=" {
		l.avanzar()
		l.avanzar()
		k := token.GE
		if r == '<' {
			k = token.LE
		}
		l.emitir(k, par, linea, col, 2)
		return
	}
	if k, ok := simbolos[r]; ok {
		l.avanzar()
		l.emitir(k, string(r), linea, col, 1)
		return
	}

	l.avanzar()
	switch r {
	case ';':
		l.error("punto-y-coma", linea, col, 1,
			"PokeScript no usa punto y coma.",
			"cada instrucción termina con el salto de línea.",
			"borra el ; y, si había otra instrucción después, pásala a la línea siguiente.")
		return
	case '%':
		l.error("simbolo-de-otro-lenguaje", linea, col, 1,
			"«%» no existe en PokeScript.",
			"el residuo de una división se escribe con la palabra «resto».",
			"cambia «%» por «resto».")
		l.emitir(token.RESTO, "%", linea, col, 1)
		return
	default:
		l.error("simbolo-desconocido", linea, col, 1,
			fmt.Sprintf("el símbolo «%c» no se reconoce.", r),
			"ese carácter no forma parte del lenguaje.",
			"revisa si sobra o si querías escribir otro símbolo.")
	}
	l.emitir(token.ILLEGAL, string(r), linea, col, 1)
}
