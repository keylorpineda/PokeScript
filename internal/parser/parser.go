package parser

import (
	"fmt"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/lexer"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// MaxErrores es el máximo de diagnósticos sintácticos por archivo (sección 2.1).
const MaxErrores = 20

// parser guarda el estado del descenso recursivo sobre los tokens de un archivo.
type parser struct {
	archivo string
	tokens  []token.Token
	pos     int
	diags   diag.Lista

	// sospechosos son bloques cerrados por un fin con menos sangría que su
	// apertura: pistas para explicar un bloque sin cerrar.
	sospechosos []token.Token
}

func nuevo(archivo string, tokens []token.Token) *parser {
	return &parser{archivo: archivo, tokens: tokens, diags: diag.Lista{Max: MaxErrores}}
}

// errSintaxis se lanza con panic al encontrar un error. La instrucción que
// se está leyendo lo recupera y el parser se sincroniza en el siguiente
// NEWLINE (sección 2.1).
type errSintaxis struct{}

// ─── Movimiento sobre los tokens ───────────────────────────────────────────

func (p *parser) actual() token.Token { return p.ver(0) }

// ver mira n tokens adelante sin consumir. Más allá del final devuelve EOF.
func (p *parser) ver(n int) token.Token {
	if i := p.pos + n; i < len(p.tokens) {
		return p.tokens[i]
	}
	return p.tokens[len(p.tokens)-1]
}

func (p *parser) es(k token.Kind) bool { return p.actual().Kind == k }

func (p *parser) avanzar() token.Token {
	t := p.actual()
	if t.Kind != token.EOF {
		p.pos++
	}
	return t
}

// aceptar consume el token actual si es de tipo k.
func (p *parser) aceptar(k token.Kind) bool {
	if p.es(k) {
		p.avanzar()
		return true
	}
	return false
}

// anterior es el último token consumido.
func (p *parser) anterior() token.Token {
	if p.pos == 0 {
		return p.tokens[0]
	}
	return p.tokens[p.pos-1]
}

// ─── Posiciones ────────────────────────────────────────────────────────────

// desde devuelve la posición que va del token ini hasta el último token
// consumido. Todas las expresiones caben en una línea, porque el salto de
// línea termina la instrucción.
func (p *parser) desde(ini token.Token) ast.Pos {
	fin := p.anterior()
	largo := fin.Col + fin.Len - ini.Col
	if fin.Line != ini.Line || largo < ini.Len {
		largo = ini.Len
	}
	return ast.Pos{Line: ini.Line, Col: ini.Col, Len: largo}
}

// entre devuelve la posición que cubre desde el inicio de a hasta el final de b.
func entre(a, b ast.Pos) ast.Pos {
	if a.Line != b.Line || b.Col+b.Len < a.Col {
		return a
	}
	return ast.Pos{Line: a.Line, Col: a.Col, Len: b.Col + b.Len - a.Col}
}

// ─── Errores ───────────────────────────────────────────────────────────────

// fallar registra un error sintáctico en la posición de t y corta la
// instrucción actual. Si t es ILLEGAL, el lexer ya lo reportó y no se repite.
func (p *parser) fallar(t token.Token, codigo, desc, causa, sugerencia string) {
	if t.Kind != token.ILLEGAL {
		p.error(ast.DesdeToken(t), codigo, desc, causa, sugerencia)
	}
	panic(errSintaxis{})
}

func (p *parser) error(pos ast.Pos, codigo, desc, causa, sugerencia string) {
	if pos.Len < 1 {
		pos.Len = 1
	}
	p.diags.Agregar(diag.Diagnostic{
		Severity: diag.Error,
		Category: diag.Sintactico,
		Code:     codigo,
		Heading:  diag.EncabezadoSintaxis,
		File:     p.archivo,
		Line:     pos.Line,
		Col:      pos.Col,
		Len:      pos.Len,
		Desc:     desc,
		Cause:    causa,
		Suggest:  sugerencia,
	})
}

// esperar consume un token de tipo k o falla explicando qué faltó.
func (p *parser) esperar(k token.Kind, que string) token.Token {
	if p.es(k) {
		return p.avanzar()
	}
	t := p.actual()
	p.fallar(t, "token-esperado",
		fmt.Sprintf("se esperaba %s y se encontró %s.", que, describir(t)),
		"la instrucción está incompleta o tiene algo de más en este punto.",
		fmt.Sprintf("revisa que aquí vaya %s.", que))
	return t
}

// describir nombra un token para los mensajes de error.
func describir(t token.Token) string {
	switch t.Kind {
	case token.NEWLINE:
		return "el final de la línea"
	case token.EOF:
		return "el final del archivo"
	case token.IDENT:
		return fmt.Sprintf("el nombre «%s»", t.Lexeme)
	case token.ROCA_LIT, token.AGUA_LIT:
		return fmt.Sprintf("el número %s", t.Lexeme)
	case token.PLANTA_LIT:
		return "un texto"
	case token.FUEGO_LIT:
		return "un carácter"
	}
	if t.Kind.EsPalabraReservada() {
		return fmt.Sprintf("la palabra «%s»", t.Kind)
	}
	return fmt.Sprintf("«%s»", t.Kind)
}

// ─── Entrada para expresiones sueltas ──────────────────────────────────────

// ResultadoExpresion es el resultado de ParsearExpresion.
type ResultadoExpresion struct {
	Expr         ast.Expr // nil si hubo un error sintáctico
	Diagnosticos []diag.Diagnostic
}

// ParsearExpresion lee una sola expresión que ocupa todo el texto. Incluye
// los diagnósticos léxicos. Sirve para las pruebas y para evaluar
// expresiones sueltas desde el IDE.
func ParsearExpresion(archivo, fuente string) ResultadoExpresion {
	lex := lexer.Analizar(archivo, fuente)
	p := nuevo(archivo, lex.Tokens)
	var e ast.Expr
	func() {
		defer p.recuperar()
		e = p.expresion()
		if !p.es(token.NEWLINE) && !p.es(token.EOF) {
			t := p.actual()
			p.fallar(t, "sobra-despues-de-expresion",
				fmt.Sprintf("después de la expresión sobra %s.", describir(t)),
				"una expresión termina donde termina la línea.",
				"borra lo que sobra o pásalo a otra línea.")
		}
	}()
	if p.diags.TieneErrores() {
		e = nil
	}
	diags := append(append([]diag.Diagnostic{}, lex.Diagnosticos...), p.diags.Items()...)
	if len(lex.Diagnosticos) > 0 {
		e = nil
	}
	return ResultadoExpresion{Expr: e, Diagnosticos: diags}
}

// recuperar atrapa un errSintaxis; cualquier otro panic se deja pasar.
func (p *parser) recuperar() {
	if r := recover(); r != nil {
		if _, ok := r.(errSintaxis); !ok {
			panic(r)
		}
	}
}
