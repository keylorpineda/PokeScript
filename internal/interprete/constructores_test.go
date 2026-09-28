package interprete

import (
	"context"
	"math/rand/v2"
	"testing"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/ast/astprueba"
)

// Los árboles de las pruebas se arman con internal/ast/astprueba, el mismo
// paquete que usará la prueba de contrato del parser. Estos nombres cortos
// solo abrevian las llamadas.
var (
	id           = astprueba.Id
	roca         = astprueba.Roca
	agua         = astprueba.Agua
	fuego        = astprueba.Fuego
	planta       = astprueba.Planta
	electrico    = astprueba.Electrico
	fantasma     = astprueba.Fantasma
	equipoLit    = astprueba.Equipo
	llaves       = astprueba.Llaves
	bin          = astprueba.Bin
	un           = astprueba.Un
	idx          = astprueba.Indice
	campo        = astprueba.Campo
	llamada      = astprueba.Llamada
	tNombre      = astprueba.TNombre
	tEquipo      = astprueba.TEquipo
	tMochila     = astprueba.TMochila
	tPosible     = astprueba.TPosible
	dato         = astprueba.Dato
	medallaLocal = astprueba.MedallaLocal
	asignacion   = astprueba.Asignacion
	gritar       = astprueba.Gritar
	capturar     = astprueba.Capturar
	sumar        = astprueba.Sumar
	quitar       = astprueba.Quitar
	instr        = astprueba.Instr
	entregar     = astprueba.Entregar
	huir         = astprueba.Huir
	siguiente    = astprueba.Siguiente
	si           = astprueba.Si
	rama         = astprueba.Rama
	mientras     = astprueba.Mientras
	rango        = astprueba.Rango
	recorrido    = astprueba.Recorrido
	alternativa  = astprueba.Alternativa
	segun        = astprueba.Segun
	especie      = astprueba.Especie
	ficha        = astprueba.Ficha
	medalla      = astprueba.Medalla
	movimiento   = astprueba.Movimiento
	combate      = astprueba.Combate

	tRoca      = astprueba.TRoca()
	tAgua      = astprueba.TAgua()
	tFuego     = astprueba.TFuego()
	tPlanta    = astprueba.TPlanta()
	tElectrico = astprueba.TElectrico()
)

func programa(decls ...ast.Decl) *ast.Programa {
	return astprueba.Programa("principal.pks", nil, decls...)
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
