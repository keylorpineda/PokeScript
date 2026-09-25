package astprueba

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
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
		tokens := hojasDelLexer(t, res.Tokens)
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
	// En general, nil y vacío son iguales.
	if d := Diferencia(Gritar(), &ast.Gritar{Pos: p0, Args: []ast.Expr{}}); d != "" {
		t.Errorf("unos argumentos nil y unos vacíos deberían ser iguales: %s", d)
	}
	// Pero Si.Sino nil rompe el contrato del AST.
	a := &ast.Si{Pos: p0, Sino: nil}
	b := &ast.Si{Pos: p0, Sino: []ast.Instr{}}
	if d := Diferencia(a, b); !strings.Contains(d, "Sino") {
		t.Errorf("un sino nil debería reportarse, Diferencia dijo %q", d)
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

// hojasDelLexer toma los tokens IDENT y *_LIT. Los números se pasan a su
// forma canónica (007 → 7, 1.50 → 1.5), la misma que usa Hojas, porque el
// árbol guarda el valor y no cómo se escribió.
func hojasDelLexer(t *testing.T, tokens []token.Token) []string {
	t.Helper()
	var hojas []string
	for _, tk := range tokens {
		switch tk.Kind {
		case token.IDENT, token.FUEGO_LIT, token.PLANTA_LIT:
			hojas = append(hojas, tk.Lexeme)
		case token.ROCA_LIT:
			n, err := strconv.ParseInt(tk.Lexeme, 10, 64)
			if err != nil {
				t.Fatalf("roca inválida %q: %v", tk.Lexeme, err)
			}
			hojas = append(hojas, strconv.FormatInt(n, 10))
		case token.AGUA_LIT:
			f, err := strconv.ParseFloat(tk.Lexeme, 64)
			if err != nil {
				t.Fatalf("agua inválida %q: %v", tk.Lexeme, err)
			}
			hojas = append(hojas, strconv.FormatFloat(f, 'f', -1, 64))
		}
	}
	return hojas
}

func TestHojasConLiteralesEscritosDistinto(t *testing.T) {
	res := lexer.Analizar("prueba.pks", "gritar 1.0, 1.50, 007\n")
	got := hojasDelLexer(t, res.Tokens)
	want := Hojas(Gritar(Agua(1.0), Agua(1.5), Roca(7)))
	if !reflect.DeepEqual(got, want) {
		t.Errorf("lexer %q, fixture %q", got, want)
	}
}
