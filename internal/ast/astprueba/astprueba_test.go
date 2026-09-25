package astprueba

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/lexer"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// Mientras no exista el parser, esta prueba comprueba que el fixture
// corresponde a los .pks reales: los nombres y literales del árbol armado a
// mano deben ser exactamente los tokens IDENT y *_LIT que produce el lexer,
// en el mismo orden.
func TestSeccion10CoincideConLosArchivos(t *testing.T) {
	for archivo, programa := range Seccion10() {
		fuente, err := os.ReadFile(filepath.Join("..", "..", "lexer", "testdata", archivo))
		if err != nil {
			t.Fatal(err)
		}
		res := lexer.Analizar(archivo, string(fuente))
		if len(res.Diagnosticos) > 0 {
			t.Fatalf("%s tiene errores léxicos: %v", archivo, res.Diagnosticos)
		}
		var tokens []string
		for _, tk := range res.Tokens {
			switch tk.Kind {
			case token.IDENT, token.ROCA_LIT, token.AGUA_LIT, token.FUEGO_LIT, token.PLANTA_LIT:
				tokens = append(tokens, tk.Lexeme)
			}
		}
		hojas := Hojas(programa)
		if !reflect.DeepEqual(hojas, tokens) {
			t.Errorf("%s: el fixture no coincide con el archivo\n  fixture: %q\n  lexer:   %q", archivo, hojas, tokens)
		}
	}
}

func TestSeccion10UnArchivo(t *testing.T) {
	p := Seccion10UnArchivo()
	if len(p.Importaciones) != 0 {
		t.Error("la versión de un archivo no debe tener importaciones")
	}
	// 2 medallas + especie + ficha + calcular_dano + describir + combate
	if len(p.Declaraciones) != 7 {
		t.Errorf("tiene %d declaraciones, want 7", len(p.Declaraciones))
	}
}

func TestDiferenciaIgnoraPosiciones(t *testing.T) {
	a := Bin(token.PLUS, Id("x"), Roca(1))
	b := Bin(token.PLUS, Id("x"), Roca(1))
	b.Pos = ast.Pos{Line: 9, Col: 4, Len: 5}
	b.OpPos = ast.Pos{Line: 9, Col: 6, Len: 1}
	b.Izq.(*ast.Ident).Pos = ast.Pos{Line: 9, Col: 4, Len: 1}
	if d := Diferencia(a, b); d != "" {
		t.Errorf("solo cambian posiciones, pero Diferencia dijo: %s", d)
	}
}

func TestDiferenciaEncuentraElPrimerCambio(t *testing.T) {
	casos := []struct {
		nombre   string
		a, b     any
		contiene string
	}{
		{"operador", Bin(token.PLUS, Id("x"), Roca(1)), Bin(token.MINUS, Id("x"), Roca(1)), "Op"},
		{"nombre", Dato(TRoca(), "a", nil), Dato(TRoca(), "b", nil), "Nombre.Nombre"},
		{"valor que falta", Dato(TRoca(), "a", Roca(1)), Dato(TRoca(), "a", nil), "Valor"},
		{"tipo de nodo", Entregar(Roca(1)), Entregar(Agua(1)), "Valor"},
		{"cantidad de elementos", Gritar(Roca(1)), Gritar(Roca(1), Roca(2)), "Args"},
		{
			"dentro de un bloque",
			Combate(Gritar(Planta("a")), Gritar(Planta("b"))),
			Combate(Gritar(Planta("a")), Gritar(Planta("c"))),
			"Cuerpo[1].Args[0].Valor",
		},
	}
	for _, c := range casos {
		d := Diferencia(c.a, c.b)
		if d == "" || !strings.Contains(d, c.contiene) {
			t.Errorf("%s: Diferencia = %q, debería mencionar %q", c.nombre, d, c.contiene)
		}
	}
}

func TestDiferenciaSliceVacioYNil(t *testing.T) {
	a := &ast.Si{Pos: p0, Sino: nil}
	b := &ast.Si{Pos: p0, Sino: []ast.Instr{}}
	if d := Diferencia(a, b); d != "" {
		t.Errorf("un sino nil y uno vacío deberían ser iguales: %s", d)
	}
}

func TestSiSinSinoDejaSliceVacio(t *testing.T) {
	s := Si(Electrico(true), nil, nil, nil)
	if s.Sino == nil || len(s.Sino) != 0 || s.TieneSino {
		t.Error("sin sino, Sino debe ser un slice vacío y TieneSino falso")
	}
}

func TestLosTiposNoSeComparten(t *testing.T) {
	a := TRoca()
	TPosible(a)
	if a.Posible {
		t.Error("TPosible no debe modificar el tipo original")
	}
}
