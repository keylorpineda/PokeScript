package interprete

import (
	"context"
	"math/rand/v2"
	"testing"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// Constructores de AST para las pruebas. Mientras no exista el parser, los
// programas se arman a mano con estas funciones (T2.6 y T3.1 del plan).

func en(linea, col int) ast.Pos { return ast.Pos{Line: linea, Col: col, Len: 1} }

func id(n string) *ast.Ident                 { return &ast.Ident{Pos: en(1, 1), Nombre: n} }
func roca(v int64) *ast.LitRoca              { return &ast.LitRoca{Pos: en(1, 1), Valor: v} }
func agua(v float64) *ast.LitAgua            { return &ast.LitAgua{Pos: en(1, 1), Valor: v} }
func fuego(v rune) *ast.LitFuego             { return &ast.LitFuego{Pos: en(1, 1), Valor: v} }
func planta(v string) *ast.LitPlanta         { return &ast.LitPlanta{Pos: en(1, 1), Valor: v} }
func electrico(v bool) *ast.LitElectrico     { return &ast.LitElectrico{Pos: en(1, 1), Valor: v} }
func fantasma() *ast.LitFantasma             { return &ast.LitFantasma{Pos: en(1, 1)} }
func equipoLit(e ...ast.Expr) *ast.LitEquipo { return &ast.LitEquipo{Pos: en(1, 1), Elems: e} }

func bin(op token.Kind, a, b ast.Expr) *ast.Binaria {
	return &ast.Binaria{Pos: en(1, 1), Op: op, OpPos: en(1, 1), Izq: a, Der: b}
}
func un(op token.Kind, a ast.Expr) *ast.Unaria {
	return &ast.Unaria{Pos: en(1, 1), Op: op, Operando: a}
}
func idx(c, i ast.Expr) *ast.Indice { return &ast.Indice{Pos: en(1, 1), Coleccion: c, Indice: i} }
func campo(o ast.Expr, n string) *ast.CampoAcceso {
	return &ast.CampoAcceso{Pos: en(1, 1), Objeto: o, Nombre: id(n)}
}
func llamada(n string, args ...ast.Expr) *ast.Llamada {
	return &ast.Llamada{Pos: en(1, 1), Nombre: id(n), Args: args}
}

// llaves arma { clave: valor, … } a partir de pares alternados.
func llaves(pares ...ast.Expr) *ast.LitLlaves {
	l := &ast.LitLlaves{Pos: en(1, 1)}
	for i := 0; i+1 < len(pares); i += 2 {
		l.Pares = append(l.Pares, &ast.Par{Pos: en(1, 1), Clave: pares[i], Valor: pares[i+1]})
	}
	return l
}

func tSimple(k token.Kind) *ast.TipoExpr { return &ast.TipoExpr{Forma: ast.TipoSimple, Simple: k} }
func tNombre(n string) *ast.TipoExpr     { return &ast.TipoExpr{Forma: ast.TipoNombrado, Nombre: n} }
func tEquipo(e *ast.TipoExpr) *ast.TipoExpr {
	return &ast.TipoExpr{Forma: ast.TipoEquipo, Elem: e}
}
func tMochila(k, v *ast.TipoExpr) *ast.TipoExpr {
	return &ast.TipoExpr{Forma: ast.TipoMochila, Clave: k, Elem: v}
}
func tPosible(t *ast.TipoExpr) *ast.TipoExpr { c := *t; c.Posible = true; return &c }

var (
	tRoca      = tSimple(token.ROCA)
	tAgua      = tSimple(token.AGUA)
	tFuego     = tSimple(token.FUEGO)
	tPlanta    = tSimple(token.PLANTA)
	tElectrico = tSimple(token.ELECTRICO)
)

func dato(t *ast.TipoExpr, n string, v ast.Expr) *ast.DeclDato {
	return &ast.DeclDato{Pos: en(1, 1), Tipo: t, Nombre: id(n), Valor: v}
}
func medallaLocal(t *ast.TipoExpr, n string, v ast.Expr) *ast.DeclDato {
	d := dato(t, n, v)
	d.Medalla = true
	return d
}
func asignacion(dest, v ast.Expr) *ast.Asignacion {
	return &ast.Asignacion{Pos: en(1, 1), Destino: dest, Valor: v}
}
func gritar(args ...ast.Expr) *ast.Gritar { return &ast.Gritar{Pos: en(1, 1), Args: args} }
func capturar(dest ast.Expr, msg string) *ast.Capturar {
	return &ast.Capturar{Pos: en(1, 1), Destino: dest, Mensaje: planta(msg)}
}
func sumar(v ast.Expr, col string) *ast.Sumar {
	return &ast.Sumar{Pos: en(1, 1), Valor: v, Coleccion: id(col)}
}
func quitar(col string, i ast.Expr) *ast.Quitar {
	return &ast.Quitar{Pos: en(1, 1), Coleccion: id(col), Indice: i}
}
func instr(l *ast.Llamada) *ast.LlamadaInstr { return &ast.LlamadaInstr{Pos: en(1, 1), Llamada: l} }
func entregar(v ast.Expr) *ast.Entregar      { return &ast.Entregar{Pos: en(1, 1), Valor: v} }
func huir() *ast.Huir                        { return &ast.Huir{Pos: en(1, 1)} }
func siguiente() *ast.Siguiente              { return &ast.Siguiente{Pos: en(1, 1)} }

// si arma un si con sus ramas; sino == nil significa que no hay sino.
func si(cond ast.Expr, cuerpo []ast.Instr, sinoSi []*ast.RamaSi, sino []ast.Instr) *ast.Si {
	s := &ast.Si{Pos: en(1, 1), Ramas: append([]*ast.RamaSi{{Pos: en(1, 1), Cond: cond, Cuerpo: cuerpo}}, sinoSi...)}
	if sino != nil {
		s.Sino, s.TieneSino = sino, true
	}
	return s
}
func rama(cond ast.Expr, cuerpo ...ast.Instr) *ast.RamaSi {
	return &ast.RamaSi{Pos: en(1, 1), Cond: cond, Cuerpo: cuerpo}
}
func mientras(cond ast.Expr, cuerpo ...ast.Instr) *ast.Mientras {
	return &ast.Mientras{Pos: en(1, 1), Cond: cond, Cuerpo: cuerpo}
}
func rango(v string, a, b ast.Expr, cuerpo ...ast.Instr) *ast.RecorrerRango {
	return &ast.RecorrerRango{Pos: en(1, 1), Var: id(v), Desde: a, Hasta: b, Cuerpo: cuerpo}
}
func recorrido(v, v2 string, c ast.Expr, cuerpo ...ast.Instr) *ast.RecorrerColeccion {
	r := &ast.RecorrerColeccion{Pos: en(1, 1), Var: id(v), Coleccion: c, Cuerpo: cuerpo}
	if v2 != "" {
		r.Var2 = id(v2)
	}
	return r
}
func alternativa(cuerpo ast.Instr, patrones ...ast.Expr) *ast.Alternativa {
	return &ast.Alternativa{Pos: en(1, 1), Patrones: patrones, Cuerpo: cuerpo}
}
func segun(v ast.Expr, otro ast.Instr, alts ...*ast.Alternativa) *ast.Segun {
	return &ast.Segun{Pos: en(1, 1), Valor: v, Alternativas: alts, Otro: otro}
}

func especie(n string, valores ...string) *ast.DeclEspecie {
	d := &ast.DeclEspecie{Pos: en(1, 1), Nombre: id(n)}
	for _, v := range valores {
		d.Valores = append(d.Valores, id(v))
	}
	return d
}

// ficha arma una declaración de ficha a partir de pares tipo, nombre.
func ficha(n string, campos ...any) *ast.DeclFicha {
	d := &ast.DeclFicha{Pos: en(1, 1), Nombre: id(n)}
	for i := 0; i+1 < len(campos); i += 2 {
		d.Campos = append(d.Campos, &ast.Campo{Pos: en(1, 1), Tipo: campos[i].(*ast.TipoExpr), Nombre: id(campos[i+1].(string))})
	}
	return d
}

func medalla(t *ast.TipoExpr, n string, v ast.Expr) *ast.DeclMedalla {
	return &ast.DeclMedalla{Pos: en(1, 1), Tipo: t, Nombre: id(n), Valor: v}
}

// movimiento arma una declaración; params va en pares tipo, nombre.
func movimiento(retorno *ast.TipoExpr, n string, params []any, cuerpo ...ast.Instr) *ast.DeclMovimiento {
	d := &ast.DeclMovimiento{Pos: en(1, 1), Retorno: retorno, Nombre: id(n), Cuerpo: cuerpo, FinPos: en(1, 1)}
	for i := 0; i+1 < len(params); i += 2 {
		d.Params = append(d.Params, &ast.Param{Pos: en(1, 1), Tipo: params[i].(*ast.TipoExpr), Nombre: id(params[i+1].(string))})
	}
	return d
}

func combate(cuerpo ...ast.Instr) *ast.Combate { return &ast.Combate{Pos: en(1, 1), Cuerpo: cuerpo} }

func programa(decls ...ast.Decl) *ast.Programa {
	return &ast.Programa{Archivo: "principal.pks", Declaraciones: decls}
}

// correr ejecuta un programa con entradas preparadas y azar fijo, y
// devuelve lo que escribió.
func correr(t *testing.T, p *ast.Programa, entradas ...string) (*ESMemoria, error) {
	t.Helper()
	es := NuevaESMemoria(entradas...)
	in := Nuevo(es)
	in.Azar = rand.New(rand.NewPCG(1, 2))
	return es, in.Ejecutar(context.Background(), p)
}

// soloCombate ejecuta instrucciones sueltas dentro de combate.
func soloCombate(t *testing.T, cuerpo ...ast.Instr) ([]string, error) {
	t.Helper()
	es, err := correr(t, programa(combate(cuerpo...)))
	return es.Salida, err
}
