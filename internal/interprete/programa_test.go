package interprete

import (
	"context"
	"math/rand/v2"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/keylorpineda/PokeScript/internal/ast/astprueba"
)

func correrSeccion10(t *testing.T, semilla uint64) *ESMemoria {
	t.Helper()
	es := NuevaESMemoria("Chispo")
	in := Nuevo(es)
	in.Azar = rand.New(rand.NewPCG(semilla, 0))
	if err := in.Ejecutar(context.Background(), astprueba.Seccion10UnArchivo()); err != nil {
		t.Fatalf("semilla %d: %v", semilla, err)
	}
	return es
}

// TestProgramaSeccion10 es la prueba de integración del plan (T7.4),
// adelantada con el fixture de astprueba, que ya se comprobó contra los .pks
// reales. Revisa la estructura de la salida y
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
