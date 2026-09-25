// Package astprueba arma árboles de sintaxis a mano para las pruebas.
//
// Sirve para dos cosas. Mientras no exista el parser, el intérprete y el
// analizador se prueban con estos árboles. Cuando el parser exista, los
// fixtures (como Seccion10) se comparan contra lo que el parser produce al
// leer los .pks reales, con Diferencia: si el parser y el intérprete
// entienden el AST distinto, la prueba falla ahí y no en la demo.
//
// Todas las posiciones se ponen en 1:1; Diferencia las ignora.
package astprueba

import (
	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// En devuelve una posición de un carácter.
func En(linea, col int) ast.Pos { return ast.Pos{Line: linea, Col: col, Len: 1} }

var p0 = En(1, 1)

// ─── Expresiones ───────────────────────────────────────────────────────────

func Id(n string) *ast.Ident             { return &ast.Ident{Pos: p0, Nombre: n} }
func Roca(v int64) *ast.LitRoca          { return &ast.LitRoca{Pos: p0, Valor: v} }
func Agua(v float64) *ast.LitAgua        { return &ast.LitAgua{Pos: p0, Valor: v} }
func Fuego(v rune) *ast.LitFuego         { return &ast.LitFuego{Pos: p0, Valor: v} }
func Planta(v string) *ast.LitPlanta     { return &ast.LitPlanta{Pos: p0, Valor: v} }
func Electrico(v bool) *ast.LitElectrico { return &ast.LitElectrico{Pos: p0, Valor: v} }
func Fantasma() *ast.LitFantasma         { return &ast.LitFantasma{Pos: p0} }

// Equipo arma el literal [a, b, …].
func Equipo(elems ...ast.Expr) *ast.LitEquipo { return &ast.LitEquipo{Pos: p0, Elems: elems} }

// Llaves arma { clave: valor, … } a partir de pares alternados.
func Llaves(pares ...ast.Expr) *ast.LitLlaves {
	l := &ast.LitLlaves{Pos: p0}
	for i := 0; i+1 < len(pares); i += 2 {
		l.Pares = append(l.Pares, &ast.Par{Pos: p0, Clave: pares[i], Valor: pares[i+1]})
	}
	return l
}

func Bin(op token.Kind, a, b ast.Expr) *ast.Binaria {
	return &ast.Binaria{Pos: p0, Op: op, OpPos: p0, Izq: a, Der: b}
}
func Un(op token.Kind, a ast.Expr) *ast.Unaria { return &ast.Unaria{Pos: p0, Op: op, Operando: a} }
func Respaldo(v, r ast.Expr) *ast.Respaldo     { return &ast.Respaldo{Pos: p0, Valor: v, Reemplazo: r} }
func Indice(c, i ast.Expr) *ast.Indice         { return &ast.Indice{Pos: p0, Coleccion: c, Indice: i} }
func Campo(o ast.Expr, n string) *ast.CampoAcceso {
	return &ast.CampoAcceso{Pos: p0, Objeto: o, Nombre: Id(n)}
}
func Llamada(n string, args ...ast.Expr) *ast.Llamada {
	return &ast.Llamada{Pos: p0, Nombre: Id(n), Args: args}
}
func Convertir(v ast.Expr, t *ast.TipoExpr) *ast.Convertir {
	return &ast.Convertir{Pos: p0, Valor: v, Destino: t}
}
func Tamano(v ast.Expr) *ast.Tamano       { return &ast.Tamano{Pos: p0, Valor: v} }
func Redondear(v ast.Expr) *ast.Redondear { return &ast.Redondear{Pos: p0, Valor: v} }
func Aleatorio(a, b ast.Expr) *ast.Aleatorio {
	return &ast.Aleatorio{Pos: p0, Min: a, Max: b}
}

// ─── Tipos ─────────────────────────────────────────────────────────────────

func TSimple(k token.Kind) *ast.TipoExpr {
	return &ast.TipoExpr{Pos: p0, Forma: ast.TipoSimple, Simple: k}
}
func TRoca() *ast.TipoExpr      { return TSimple(token.ROCA) }
func TAgua() *ast.TipoExpr      { return TSimple(token.AGUA) }
func TFuego() *ast.TipoExpr     { return TSimple(token.FUEGO) }
func TPlanta() *ast.TipoExpr    { return TSimple(token.PLANTA) }
func TElectrico() *ast.TipoExpr { return TSimple(token.ELECTRICO) }
func TNombre(n string) *ast.TipoExpr {
	return &ast.TipoExpr{Pos: p0, Forma: ast.TipoNombrado, Nombre: n}
}
func TEquipo(e *ast.TipoExpr) *ast.TipoExpr {
	return &ast.TipoExpr{Pos: p0, Forma: ast.TipoEquipo, Elem: e}
}
func TMochila(k, v *ast.TipoExpr) *ast.TipoExpr {
	return &ast.TipoExpr{Pos: p0, Forma: ast.TipoMochila, Clave: k, Elem: v}
}

// TPosible devuelve una copia de t marcada como posible.
func TPosible(t *ast.TipoExpr) *ast.TipoExpr {
	c := *t
	c.Posible = true
	return &c
}

// ─── Instrucciones ─────────────────────────────────────────────────────────

func Dato(t *ast.TipoExpr, n string, v ast.Expr) *ast.DeclDato {
	return &ast.DeclDato{Pos: p0, Tipo: t, Nombre: Id(n), Valor: v}
}
func MedallaLocal(t *ast.TipoExpr, n string, v ast.Expr) *ast.DeclDato {
	d := Dato(t, n, v)
	d.Medalla = true
	return d
}
func Asignacion(dest, v ast.Expr) *ast.Asignacion {
	return &ast.Asignacion{Pos: p0, Destino: dest, Valor: v}
}
func Gritar(args ...ast.Expr) *ast.Gritar { return &ast.Gritar{Pos: p0, Args: args} }
func Capturar(dest ast.Expr, msg string) *ast.Capturar {
	return &ast.Capturar{Pos: p0, Destino: dest, Mensaje: Planta(msg)}
}
func Sumar(v ast.Expr, col string) *ast.Sumar {
	return &ast.Sumar{Pos: p0, Valor: v, Coleccion: Id(col)}
}
func Quitar(col string, i ast.Expr) *ast.Quitar {
	return &ast.Quitar{Pos: p0, Coleccion: Id(col), Indice: i}
}
func Instr(l *ast.Llamada) *ast.LlamadaInstr { return &ast.LlamadaInstr{Pos: p0, Llamada: l} }
func Entregar(v ast.Expr) *ast.Entregar      { return &ast.Entregar{Pos: p0, Valor: v} }
func Huir() *ast.Huir                        { return &ast.Huir{Pos: p0} }
func Siguiente() *ast.Siguiente              { return &ast.Siguiente{Pos: p0} }

// Si arma un si con sus ramas. sino == nil significa que no hay sino; en
// ese caso Sino queda como slice vacío, como lo deja el parser.
func Si(cond ast.Expr, cuerpo []ast.Instr, sinoSi []*ast.RamaSi, sino []ast.Instr) *ast.Si {
	s := &ast.Si{Pos: p0, Ramas: append([]*ast.RamaSi{{Pos: p0, Cond: cond, Cuerpo: cuerpo}}, sinoSi...), Sino: []ast.Instr{}}
	if sino != nil {
		s.Sino, s.TieneSino, s.SinoPos = sino, true, p0
	}
	return s
}
func Rama(cond ast.Expr, cuerpo ...ast.Instr) *ast.RamaSi {
	return &ast.RamaSi{Pos: p0, Cond: cond, Cuerpo: cuerpo}
}
func Mientras(cond ast.Expr, cuerpo ...ast.Instr) *ast.Mientras {
	return &ast.Mientras{Pos: p0, Cond: cond, Cuerpo: cuerpo}
}
func Rango(v string, a, b ast.Expr, cuerpo ...ast.Instr) *ast.RecorrerRango {
	return &ast.RecorrerRango{Pos: p0, Var: Id(v), Desde: a, Hasta: b, Cuerpo: cuerpo}
}

// Recorrido arma recorrer v [, v2] en c; v2 vacío significa una sola variable.
func Recorrido(v, v2 string, c ast.Expr, cuerpo ...ast.Instr) *ast.RecorrerColeccion {
	r := &ast.RecorrerColeccion{Pos: p0, Var: Id(v), Coleccion: c, Cuerpo: cuerpo}
	if v2 != "" {
		r.Var2 = Id(v2)
	}
	return r
}
func Alternativa(cuerpo ast.Instr, patrones ...ast.Expr) *ast.Alternativa {
	return &ast.Alternativa{Pos: p0, Patrones: patrones, Cuerpo: cuerpo}
}

// Segun arma un segun; otro == nil significa que no hay otro.
func Segun(v ast.Expr, otro ast.Instr, alts ...*ast.Alternativa) *ast.Segun {
	s := &ast.Segun{Pos: p0, Valor: v, Alternativas: alts, Otro: otro}
	if otro != nil {
		s.OtroPos = p0
	}
	return s
}

// ─── Declaraciones ─────────────────────────────────────────────────────────

func Importar(ruta string, nombres ...string) *ast.Importacion {
	im := &ast.Importacion{Pos: p0, Ruta: ruta, RutaPos: p0}
	for _, n := range nombres {
		im.Nombres = append(im.Nombres, Id(n))
	}
	return im
}

func Especie(n string, valores ...string) *ast.DeclEspecie {
	d := &ast.DeclEspecie{Pos: p0, Nombre: Id(n)}
	for _, v := range valores {
		d.Valores = append(d.Valores, Id(v))
	}
	return d
}

// Ficha arma una declaración de ficha a partir de pares tipo, nombre.
func Ficha(n string, campos ...any) *ast.DeclFicha {
	d := &ast.DeclFicha{Pos: p0, Nombre: Id(n)}
	for i := 0; i+1 < len(campos); i += 2 {
		d.Campos = append(d.Campos, &ast.Campo{Pos: p0, Tipo: campos[i].(*ast.TipoExpr), Nombre: Id(campos[i+1].(string))})
	}
	return d
}

func Medalla(t *ast.TipoExpr, n string, v ast.Expr) *ast.DeclMedalla {
	return &ast.DeclMedalla{Pos: p0, Tipo: t, Nombre: Id(n), Valor: v}
}

// Movimiento arma una declaración; params va en pares tipo, nombre.
func Movimiento(retorno *ast.TipoExpr, n string, params []any, cuerpo ...ast.Instr) *ast.DeclMovimiento {
	d := &ast.DeclMovimiento{Pos: p0, Retorno: retorno, Nombre: Id(n), Cuerpo: cuerpo, FinPos: p0}
	for i := 0; i+1 < len(params); i += 2 {
		d.Params = append(d.Params, &ast.Param{Pos: p0, Tipo: params[i].(*ast.TipoExpr), Nombre: Id(params[i+1].(string))})
	}
	return d
}

func Combate(cuerpo ...ast.Instr) *ast.Combate { return &ast.Combate{Pos: p0, Cuerpo: cuerpo} }

// Programa arma un archivo con sus importaciones y declaraciones.
func Programa(archivo string, imports []*ast.Importacion, decls ...ast.Decl) *ast.Programa {
	return &ast.Programa{Archivo: archivo, Importaciones: imports, Declaraciones: decls}
}
