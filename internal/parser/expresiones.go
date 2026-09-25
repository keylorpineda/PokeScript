package parser

import (
	"fmt"
	"strconv"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// La precedencia está en la jerarquía de funciones, de menor a mayor
// (sección 2 de la especificación):
//
//	expresion → o → y → no → igualdad → pertenencia → orden
//	          → suma → producto → unaria → respaldo → acceso → primaria

func (p *parser) expresion() ast.Expr { return p.exprO() }

// exprO: expr_y { "o" expr_y }
func (p *parser) exprO() ast.Expr {
	izq := p.exprY()
	for p.es(token.O) {
		izq = p.binaria(izq, p.avanzar(), p.exprY())
	}
	return izq
}

// exprY: expr_no { "y" expr_no }
func (p *parser) exprY() ast.Expr {
	izq := p.exprNo()
	for p.es(token.Y) {
		izq = p.binaria(izq, p.avanzar(), p.exprNo())
	}
	return izq
}

// exprNo: [ "no" ] expr_igualdad. El "no" queda por debajo de las
// comparaciones: "no rival igual fantasma" es "no (rival igual fantasma)".
func (p *parser) exprNo() ast.Expr {
	if !p.es(token.NO) {
		return p.exprIgualdad()
	}
	ini := p.avanzar()
	operando := p.exprIgualdad()
	return &ast.Unaria{Pos: p.desde(ini), Op: token.NO, Operando: operando}
}

// exprIgualdad: expr_pertenencia [ ("igual" | "diferente") expr_pertenencia ]
func (p *parser) exprIgualdad() ast.Expr {
	izq := p.exprPertenencia()
	if p.es(token.IGUAL) || p.es(token.DIFERENTE) {
		izq = p.binaria(izq, p.avanzar(), p.exprPertenencia())
		p.noEncadenable(token.IGUAL, token.DIFERENTE)
	}
	return izq
}

// exprPertenencia: expr_orden [ "contiene" expr_orden ]
func (p *parser) exprPertenencia() ast.Expr {
	izq := p.exprOrden()
	if p.es(token.CONTIENE) {
		izq = p.binaria(izq, p.avanzar(), p.exprOrden())
		p.noEncadenable(token.CONTIENE)
	}
	return izq
}

// exprOrden: expr_suma [ (">" | "<" | ">=" | "<=") expr_suma ]
func (p *parser) exprOrden() ast.Expr {
	izq := p.exprSuma()
	if esOrden(p.actual().Kind) {
		izq = p.binaria(izq, p.avanzar(), p.exprSuma())
		p.noEncadenable(token.GT, token.LT, token.GE, token.LE)
	}
	return izq
}

func esOrden(k token.Kind) bool {
	return k == token.GT || k == token.LT || k == token.GE || k == token.LE
}

// noEncadenable falla si, tras una comparación, viene otra del mismo nivel:
// "a > b > c" es error de sintaxis, no de tipos (sección 3.3).
func (p *parser) noEncadenable(ops ...token.Kind) {
	t := p.actual()
	for _, k := range ops {
		if t.Kind == k {
			p.fallar(t, "operador-no-encadenable",
				fmt.Sprintf("«%s» no se puede encadenar con otra comparación del mismo tipo.", t.Kind),
				"cada expresión admite una sola comparación de orden, una sola igualdad y un solo contiene.",
				"separa las comparaciones con «y»: en vez de «a > b > c», escribe «a > b y b > c».")
		}
	}
}

// exprSuma: expr_producto { ("+" | "-") expr_producto }
func (p *parser) exprSuma() ast.Expr {
	izq := p.exprProducto()
	for p.es(token.PLUS) || p.es(token.MINUS) {
		izq = p.binaria(izq, p.avanzar(), p.exprProducto())
	}
	return izq
}

// exprProducto: expr_unaria { ("*" | "/" | "resto") expr_unaria }
func (p *parser) exprProducto() ast.Expr {
	izq := p.exprUnaria()
	for p.es(token.STAR) || p.es(token.SLASH) || p.es(token.RESTO) {
		izq = p.binaria(izq, p.avanzar(), p.exprUnaria())
	}
	return izq
}

// exprUnaria: [ "-" ] expr_respaldo
func (p *parser) exprUnaria() ast.Expr {
	if !p.es(token.MINUS) {
		return p.exprRespaldo()
	}
	ini := p.avanzar()
	operando := p.exprRespaldo()
	return &ast.Unaria{Pos: p.desde(ini), Op: token.MINUS, Operando: operando}
}

// exprRespaldo: expr_acceso [ "sino" expr_acceso ]
func (p *parser) exprRespaldo() ast.Expr {
	valor := p.exprAcceso()
	if !p.es(token.SINO) {
		return valor
	}
	p.avanzar()
	reemplazo := p.exprAcceso()
	return &ast.Respaldo{Pos: entre(valor.Posicion(), reemplazo.Posicion()), Valor: valor, Reemplazo: reemplazo}
}

// exprAcceso: primaria { "[" expresion "]" | "." identificador }
func (p *parser) exprAcceso() ast.Expr {
	e := p.primaria()
	for {
		switch {
		case p.aceptar(token.LBRACKET):
			indice := p.expresion()
			p.esperarCierre(token.RBRACKET, "]", e.Posicion())
			e = &ast.Indice{Pos: entre(e.Posicion(), ast.DesdeToken(p.anterior())), Coleccion: e, Indice: indice}
		case p.aceptar(token.DOT):
			nombre := p.identificador("el nombre de un campo")
			e = &ast.CampoAcceso{Pos: entre(e.Posicion(), nombre.Pos), Objeto: e, Nombre: nombre}
		default:
			return e
		}
	}
}

func (p *parser) binaria(izq ast.Expr, op token.Token, der ast.Expr) *ast.Binaria {
	return &ast.Binaria{
		Pos:   entre(izq.Posicion(), der.Posicion()),
		Op:    op.Kind,
		OpPos: ast.DesdeToken(op),
		Izq:   izq,
		Der:   der,
	}
}

// ─── Primarias ─────────────────────────────────────────────────────────────

func (p *parser) primaria() ast.Expr {
	t := p.actual()
	switch t.Kind {
	case token.ROCA_LIT:
		p.avanzar()
		v, _ := strconv.ParseInt(t.Lexeme, 10, 64) // el lexer ya validó el rango
		return &ast.LitRoca{Pos: ast.DesdeToken(t), Valor: v}
	case token.AGUA_LIT:
		p.avanzar()
		v, _ := strconv.ParseFloat(t.Lexeme, 64)
		return &ast.LitAgua{Pos: ast.DesdeToken(t), Valor: v}
	case token.FUEGO_LIT:
		p.avanzar()
		return &ast.LitFuego{Pos: ast.DesdeToken(t), Valor: []rune(t.Lexeme)[0]}
	case token.PLANTA_LIT:
		p.avanzar()
		return &ast.LitPlanta{Pos: ast.DesdeToken(t), Valor: t.Lexeme}
	case token.VERDADERO, token.FALSO:
		p.avanzar()
		return &ast.LitElectrico{Pos: ast.DesdeToken(t), Valor: t.Kind == token.VERDADERO}
	case token.FANTASMA:
		p.avanzar()
		return &ast.LitFantasma{Pos: ast.DesdeToken(t)}
	case token.IDENT:
		if p.ver(1).Kind == token.LPAREN {
			return p.llamada()
		}
		p.avanzar()
		return &ast.Ident{Pos: ast.DesdeToken(t), Nombre: t.Lexeme}
	case token.LPAREN:
		p.avanzar()
		e := p.expresion()
		p.esperarCierre(token.RPAREN, ")", ast.DesdeToken(t))
		return e
	case token.LBRACKET:
		return p.literalEquipo()
	case token.LBRACE:
		return p.literalLlaves()
	case token.CONVERTIR:
		return p.convertir()
	case token.TAMANO:
		ini := p.avanzar()
		args := p.argumentosFijos("tamaño", 1)
		return &ast.Tamano{Pos: p.desde(ini), Valor: args[0]}
	case token.REDONDEAR:
		ini := p.avanzar()
		args := p.argumentosFijos("redondear", 1)
		return &ast.Redondear{Pos: p.desde(ini), Valor: args[0]}
	case token.ALEATORIO:
		ini := p.avanzar()
		args := p.argumentosFijos("aleatorio", 2)
		return &ast.Aleatorio{Pos: p.desde(ini), Min: args[0], Max: args[1]}
	}
	p.fallar(t, "expresion-esperada",
		fmt.Sprintf("se esperaba un valor y se encontró %s.", describir(t)),
		"a esta expresión le falta un valor: un número, un texto, un nombre o una llamada.",
		"revisa que no falte nada después del último operador.")
	return nil
}

// identificador consume un nombre o falla.
func (p *parser) identificador(que string) *ast.Ident {
	t := p.esperar(token.IDENT, que)
	return &ast.Ident{Pos: ast.DesdeToken(t), Nombre: t.Lexeme}
}

// esperarCierre consume el símbolo que cierra lo que se abrió en abierto.
func (p *parser) esperarCierre(k token.Kind, simbolo string, abierto ast.Pos) {
	if p.aceptar(k) {
		return
	}
	t := p.actual()
	if t.Kind == token.ILLEGAL {
		p.fallar(t, "", "", "", "")
	}
	p.error(abierto, "falta-cierre",
		fmt.Sprintf("falta «%s» para cerrar lo que se abrió aquí; se encontró %s.", simbolo, describir(t)),
		"cada (, [ y { necesita su cierre en la misma línea.",
		fmt.Sprintf("agrega «%s» donde termina la expresión.", simbolo))
	panic(errSintaxis{})
}

// llamada: identificador "(" [ argumentos ] ")"
func (p *parser) llamada() *ast.Llamada {
	nombre := p.identificador("el nombre del movimiento")
	abre := p.esperar(token.LPAREN, "«(»")
	args := p.listaHasta(token.RPAREN, ")", ast.DesdeToken(abre))
	return &ast.Llamada{Pos: entre(nombre.Pos, ast.DesdeToken(p.anterior())), Nombre: nombre, Args: args}
}

// listaHasta lee expresiones separadas por coma hasta el cierre, que consume.
func (p *parser) listaHasta(cierre token.Kind, simbolo string, abierto ast.Pos) []ast.Expr {
	var lista []ast.Expr
	if p.aceptar(cierre) {
		return lista
	}
	for {
		lista = append(lista, p.expresion())
		if !p.aceptar(token.COMMA) {
			break
		}
	}
	p.esperarCierre(cierre, simbolo, abierto)
	return lista
}

// argumentosFijos lee "(" e1, …, en ")" para las funciones incorporadas.
func (p *parser) argumentosFijos(nombre string, n int) []ast.Expr {
	abre := p.esperar(token.LPAREN, fmt.Sprintf("«(» después de «%s»", nombre))
	args := p.listaHasta(token.RPAREN, ")", ast.DesdeToken(abre))
	if len(args) != n {
		p.error(entre(ast.DesdeToken(abre), ast.DesdeToken(p.anterior())), "cantidad-de-argumentos",
			fmt.Sprintf("«%s» recibe %d %s y aquí tiene %d.", nombre, n, plural(n, "valor", "valores"), len(args)),
			fmt.Sprintf("la forma es %s.", formaIncorporada(nombre)),
			"revisa los valores entre paréntesis.")
		panic(errSintaxis{})
	}
	return args
}

func plural(n int, uno, varios string) string {
	if n == 1 {
		return uno
	}
	return varios
}

func formaIncorporada(nombre string) string {
	switch nombre {
	case "aleatorio":
		return "aleatorio(mínimo, máximo)"
	default:
		return nombre + "(valor)"
	}
}

// literalEquipo: "[" [ expresion { "," expresion } ] "]"
func (p *parser) literalEquipo() *ast.LitEquipo {
	abre := p.avanzar()
	elems := p.listaHasta(token.RBRACKET, "]", ast.DesdeToken(abre))
	return &ast.LitEquipo{Pos: p.desde(abre), Elems: elems}
}

// literalLlaves: "{" [ par { "," par } ] "}". El parser no decide si es
// mochila o ficha: lo resuelve el analizador (sección 2.1).
func (p *parser) literalLlaves() *ast.LitLlaves {
	abre := p.avanzar()
	l := &ast.LitLlaves{}
	if !p.aceptar(token.RBRACE) {
		for {
			clave := p.expresion()
			p.esperar(token.COLON, "«:» entre la clave y el valor")
			valor := p.expresion()
			l.Pares = append(l.Pares, &ast.Par{Pos: entre(clave.Posicion(), valor.Posicion()), Clave: clave, Valor: valor})
			if !p.aceptar(token.COMMA) {
				break
			}
		}
		p.esperarCierre(token.RBRACE, "}", ast.DesdeToken(abre))
	}
	l.Pos = p.desde(abre)
	return l
}

// convertir: "convertir" "(" expresion ")" "a" tipo
func (p *parser) convertir() *ast.Convertir {
	ini := p.avanzar()
	args := p.argumentosFijos("convertir", 1)
	if !p.es(token.A) {
		t := p.actual()
		p.fallar(t, "convertir-sin-tipo",
			fmt.Sprintf("después de «convertir(…)» se esperaba «a» y el tipo de destino; se encontró %s.", describir(t)),
			"convertir necesita saber a qué tipo pasar el valor.",
			"completa la conversión, por ejemplo: convertir(texto) a roca.")
	}
	p.avanzar()
	destino := p.tipo()
	return &ast.Convertir{Pos: p.desde(ini), Valor: args[0], Destino: destino}
}
