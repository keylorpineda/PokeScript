package interprete

import (
	"context"
	"math/rand/v2"
	"testing"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/ast/astprueba"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// correrProyecto ejecuta varios archivos en el orden dado.
func correrProyecto(t *testing.T, archivos []*ast.Programa, entradas ...string) (*ESMemoria, error) {
	t.Helper()
	es := NuevaESMemoria(entradas...)
	in := Nuevo(es)
	in.Azar = rand.New(rand.NewPCG(1, 2))
	return es, in.EjecutarProyecto(context.Background(), archivos)
}

// La sección 10 con sus cuatro archivos reales, cada uno con su alcance e
// importaciones, en el orden de internal/proyecto.
func TestProyectoSeccion10PorArchivo(t *testing.T) {
	s := astprueba.Seccion10()
	archivos := []*ast.Programa{s["constantes.pks"], s["tipos.pks"], s["operaciones.pks"], s["principal.pks"]}
	es, err := correrProyecto(t, archivos, "Chispo")
	if err != nil {
		t.Fatal(err)
	}
	if len(es.Salida) < 4 || es.Salida[0] != "¡Chispo entra en combate!" || es.Salida[1] != "  Puede atacar" {
		t.Fatalf("salida = %q", es.Salida)
	}
	ultima := es.Salida[len(es.Salida)-1]
	if ultima != "Bulbi se debilitó. ¡Ganaste!" && ultima != "Chispo se debilitó. Perdiste." {
		t.Errorf("el combate no terminó con un ganador: %q", ultima)
	}
}

// Un error dentro de un movimiento importado señala el archivo del
// movimiento, no el del combate.
func TestErrorEnOtroArchivo(t *testing.T) {
	ops := astprueba.Programa("ops.pks", nil,
		astprueba.Movimiento(tRoca, "partir", []any{tRoca, "n"},
			astprueba.Entregar(astprueba.Bin(token.RESTO, id("n"), roca(0)))),
	)
	principal := astprueba.Programa("principal.pks",
		[]*ast.Importacion{astprueba.Importar("ops.pks", "partir")},
		combate(gritar(llamada("partir", roca(7)))),
	)
	_, err := correrProyecto(t, []*ast.Programa{ops, principal})
	d := errorEjecucion(t, err)
	if d.File != "ops.pks" || d.Code != CodigoDivisionCero {
		t.Errorf("error en %s (%s), want ops.pks (%s)", d.File, d.Code, CodigoDivisionCero)
	}

	// Un error en combate sigue señalando el principal.
	principal2 := astprueba.Programa("principal.pks", nil, combate(gritar(bin(token.RESTO, roca(1), roca(0)))))
	_, err = correrProyecto(t, []*ast.Programa{ops, principal2})
	if d := errorEjecucion(t, err); d.File != "principal.pks" {
		t.Errorf("error en %s, want principal.pks", d.File)
	}
}

// Cada archivo ve sus medallas y las que importa; las importadas ya tienen
// valor porque su archivo se preparó antes.
func TestMedallasPorArchivo(t *testing.T) {
	constantes := astprueba.Programa("constantes.pks", nil, medalla(tRoca, "NIVEL", roca(25)))
	medidas := astprueba.Programa("medidas.pks",
		[]*ast.Importacion{astprueba.Importar("constantes.pks", "NIVEL")},
		medalla(tRoca, "DOBLE", bin(token.STAR, id("NIVEL"), roca(2))),
		astprueba.Movimiento(tRoca, "triple", nil, astprueba.Entregar(bin(token.STAR, id("NIVEL"), roca(3)))),
	)
	principal := astprueba.Programa("principal.pks",
		[]*ast.Importacion{astprueba.Importar("medidas.pks", "DOBLE", "triple")},
		combate(gritar(id("DOBLE"), astprueba.Planta(" "), llamada("triple"))),
	)
	es, err := correrProyecto(t, []*ast.Programa{constantes, medidas, principal})
	if err != nil {
		t.Fatal(err)
	}
	if len(es.Salida) != 1 || es.Salida[0] != "50 75" {
		t.Errorf("salida = %q, want [\"50 75\"]", es.Salida)
	}
}

// Un movimiento no ve las medallas de otro archivo que su archivo no
// importa: el alcance es el de su archivo, no el de quien llama.
func TestAlcancePorArchivo(t *testing.T) {
	otro := astprueba.Programa("otro.pks", nil,
		astprueba.Movimiento(tRoca, "leer", nil, astprueba.Entregar(id("SECRETO"))))
	principal := astprueba.Programa("principal.pks",
		[]*ast.Importacion{astprueba.Importar("otro.pks", "leer")},
		medalla(tRoca, "SECRETO", roca(7)),
		combate(gritar(llamada("leer"))),
	)
	_, err := correrProyecto(t, []*ast.Programa{otro, principal})
	if d := errorEjecucion(t, err); d.Code != CodigoInterno || d.File != "otro.pks" {
		t.Errorf("código %s en %s, want %s en otro.pks", d.Code, d.File, CodigoInterno)
	}
}
