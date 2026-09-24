package interprete

import (
	"context"
	"math/rand/v2"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// programaSeccion10 arma a mano el programa de ejemplo de la sección 10 de
// la especificación. Como todavía no hay importaciones (hito 9), junta los
// cuatro archivos en uno solo; las instrucciones son las mismas.
func programaSeccion10() *ast.Programa {
	// operaciones.pks
	calcularDano := movimiento(tRoca, "calcular_dano", []any{tRoca, "poder"},
		dato(tAgua, "base", bin(token.SLASH, bin(token.STAR, id("poder"), id("NIVEL")), roca(50))),
		dato(tRoca, "variacion", &ast.Aleatorio{Min: roca(85), Max: roca(100)}),
		dato(tAgua, "total", bin(token.SLASH, bin(token.STAR, id("base"), id("variacion")), roca(100))),
		entregar(bin(token.PLUS, &ast.Redondear{Valor: id("total")}, roca(2))),
	)

	// principal.pks
	describir := movimiento(nil, "describir", []any{tNombre("Estado"), "actual"},
		segun(id("actual"), nil,
			alternativa(gritar(planta("  Puede atacar")), id("SANO")),
			alternativa(gritar(planta("  Pierde vida cada turno")), id("ENVENENADO")),
			alternativa(gritar(planta("  Podría no atacar")), id("DORMIDO"), id("PARALIZADO")),
		))
	vidaDe := func(quien string) ast.Expr { return campo(id(quien), "vida") }
	nombreDe := func(quien string) ast.Expr { return campo(id(quien), "nombre") }

	return programa(
		// constantes.pks
		medalla(tRoca, "VIDA_MAXIMA", roca(100)),
		medalla(tRoca, "NIVEL", roca(25)),
		// tipos.pks
		especie("Estado", "SANO", "ENVENENADO", "DORMIDO", "PARALIZADO"),
		ficha("Pokemon", tPlanta, "nombre", tRoca, "vida", tNombre("Estado"), "estado"),
		calcularDano,
		describir,
		combate(
			dato(tPlanta, "nombre", nil),
			capturar(id("nombre"), "¿Cómo se llama tu Pokemon? "),

			dato(tNombre("Pokemon"), "mio", llaves(id("nombre"), id("nombre"), id("vida"), id("VIDA_MAXIMA"), id("estado"), id("SANO"))),
			dato(tNombre("Pokemon"), "rival", llaves(id("nombre"), planta("Bulbi"), id("vida"), id("VIDA_MAXIMA"), id("estado"), id("SANO"))),

			gritar(planta("¡"), nombreDe("mio"), planta(" entra en combate!")),
			instr(llamada("describir", campo(id("mio"), "estado"))),

			rango("turno", roca(1), roca(20),
				gritar(planta("--- Turno "), id("turno"), planta(" ---")),

				dato(tRoca, "dano", llamada("calcular_dano", roca(40))),
				asignacion(vidaDe("rival"), bin(token.MINUS, vidaDe("rival"), id("dano"))),
				gritar(nombreDe("mio"), planta(" ataca y hace "), id("dano"), planta(" de daño")),

				si(bin(token.LE, vidaDe("rival"), roca(0)), []ast.Instr{
					gritar(nombreDe("rival"), planta(" se debilitó. ¡Ganaste!")),
					huir(),
				}, nil, nil),

				dato(tRoca, "contra", llamada("calcular_dano", roca(35))),
				asignacion(vidaDe("mio"), bin(token.MINUS, vidaDe("mio"), id("contra"))),
				gritar(nombreDe("rival"), planta(" contraataca: "), id("contra"), planta(" de daño")),
				gritar(planta("Vida de "), nombreDe("mio"), planta(": "), vidaDe("mio")),

				si(bin(token.LE, vidaDe("mio"), roca(0)), []ast.Instr{
					gritar(nombreDe("mio"), planta(" se debilitó. Perdiste.")),
					huir(),
				}, nil, nil),
			),
		),
	)
}

func correrSeccion10(t *testing.T, semilla uint64) *ESMemoria {
	t.Helper()
	es := NuevaESMemoria("Chispo")
	in := Nuevo(es)
	in.Azar = rand.New(rand.NewPCG(semilla, 0))
	if err := in.Ejecutar(context.Background(), programaSeccion10()); err != nil {
		t.Fatalf("semilla %d: %v", semilla, err)
	}
	return es
}

// TestProgramaSeccion10 es la prueba de integración del plan (T7.4),
// adelantada con el AST armado a mano. Revisa la estructura de la salida y
// que cada daño esté en el rango que dicta la fórmula:
//
//	poder 40: base 20.0, total entre 17.0 y 20.0 → daño de 19 a 22
//	poder 35: base 17.5, total entre 14.875 y 17.5 → daño de 17 a 20
func TestProgramaSeccion10(t *testing.T) {
	ataque := regexp.MustCompile(`^Chispo ataca y hace (\d+) de daño$`)
	contra := regexp.MustCompile(`^Bulbi contraataca: (\d+) de daño$`)

	for semilla := uint64(0); semilla < 30; semilla++ {
		es := correrSeccion10(t, semilla)
		s := es.Salida

		if !reflect.DeepEqual(es.Prompts, []string{"¿Cómo se llama tu Pokemon? "}) {
			t.Fatalf("prompts = %q", es.Prompts)
		}
		if len(s) < 4 || s[0] != "¡Chispo entra en combate!" || s[1] != "  Puede atacar" || s[2] != "--- Turno 1 ---" {
			t.Fatalf("semilla %d: inicio inesperado: %q", semilla, s)
		}
		ultima := s[len(s)-1]
		if ultima != "Bulbi se debilitó. ¡Ganaste!" && ultima != "Chispo se debilitó. Perdiste." {
			t.Fatalf("semilla %d: el combate no terminó con un ganador: %q", semilla, ultima)
		}

		vidaRival, vidaMia := 100, 100
		for _, linea := range s {
			if m := ataque.FindStringSubmatch(linea); m != nil {
				n, _ := strconv.Atoi(m[1])
				if n < 19 || n > 22 {
					t.Errorf("semilla %d: daño %d fuera de 19..22", semilla, n)
				}
				vidaRival -= n
			}
			if m := contra.FindStringSubmatch(linea); m != nil {
				n, _ := strconv.Atoi(m[1])
				if n < 17 || n > 20 {
					t.Errorf("semilla %d: contraataque %d fuera de 17..20", semilla, n)
				}
				vidaMia -= n
			}
			if strings.HasPrefix(linea, "Vida de Chispo: ") && linea != "Vida de Chispo: "+strconv.Itoa(vidaMia) {
				t.Errorf("semilla %d: %q, want vida %d", semilla, linea, vidaMia)
			}
		}
		if ganaste := ultima == "Bulbi se debilitó. ¡Ganaste!"; ganaste != (vidaRival <= 0) {
			t.Errorf("semilla %d: el resultado no coincide con las vidas (rival %d, mía %d)", semilla, vidaRival, vidaMia)
		}
	}
}

func TestProgramaSeccion10EsReproducible(t *testing.T) {
	a := correrSeccion10(t, 7).Salida
	b := correrSeccion10(t, 7).Salida
	if !reflect.DeepEqual(a, b) {
		t.Error("con la misma semilla la salida debería ser la misma")
	}
}
