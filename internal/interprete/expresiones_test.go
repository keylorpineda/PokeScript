package interprete

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// grito evalúa una expresión dentro de combate con gritar y devuelve el texto.
func grito(t *testing.T, antes []ast.Instr, e ast.Expr) (string, error) {
	t.Helper()
	salida, err := soloCombate(t, append(antes, gritar(e))...)
	if err != nil {
		return "", err
	}
	return strings.Join(salida, "\n"), nil
}

// errorEjecucion comprueba que err sea un error de ejecución y lo devuelve.
func errorEjecucion(t *testing.T, err error) diag.Diagnostic {
	t.Helper()
	var ee *ErrorEjecucion
	if !errors.As(err, &ee) {
		t.Fatalf("esperaba un error de ejecución, llegó %v", err)
	}
	if ee.Diag.Category != diag.Ejecucion || ee.Diag.Heading != diag.EncabezadoEjecucion {
		t.Errorf("categoría o encabezado incorrectos: %+v", ee.Diag)
	}
	return ee.Diag
}

func TestExpresiones(t *testing.T) {
	vida := []ast.Instr{dato(tRoca, "vida", roca(120))}
	equipo := []ast.Instr{dato(tEquipo(tPlanta), "e", equipoLit(planta("Pikachu"), planta("Eevee")))}
	mochila := []ast.Instr{dato(tMochila(tPlanta, tRoca), "m", llaves(planta("agua"), roca(3)))}
	posibleVacio := []ast.Instr{dato(tPosible(tPlanta), "rival", fantasma())}

	casos := []struct {
		nombre string
		antes  []ast.Instr
		expr   ast.Expr
		want   string
	}{
		{"precedencia 2 + 3 * 4", nil, bin(token.PLUS, roca(2), bin(token.STAR, roca(3), roca(4))), "14"},
		{"agrupación (2 + 3) * 4", nil, bin(token.STAR, bin(token.PLUS, roca(2), roca(3)), roca(4)), "20"},
		{"división da agua", nil, bin(token.SLASH, roca(7), roca(2)), "3.5"},
		{"roca más agua", nil, bin(token.PLUS, roca(1), agua(0.5)), "1.5"},
		{"menos unario", vida, un(token.MINUS, id("vida")), "-120"},
		{"concatenar", nil, bin(token.PLUS, planta("Pika"), planta("chu")), "Pikachu"},
		{"comparación", vida, bin(token.GE, id("vida"), roca(120)), "verdadero"},
		{"comparar fuego", nil, bin(token.LT, fuego('a'), fuego('b')), "verdadero"},
		{"igual", nil, bin(token.IGUAL, planta("a"), planta("a")), "verdadero"},
		{"diferente", nil, bin(token.DIFERENTE, roca(1), roca(1)), "falso"},
		{"no", nil, un(token.NO, electrico(false)), "verdadero"},
		{"y", nil, bin(token.Y, electrico(true), electrico(false)), "falso"},
		{"o", nil, bin(token.O, electrico(false), electrico(true)), "verdadero"},
		{
			"y en cortocircuito no evalúa el lado derecho",
			nil,
			bin(token.Y, electrico(false), bin(token.GT, bin(token.SLASH, roca(1), roca(0)), roca(0))),
			"falso",
		},
		{
			"o en cortocircuito no evalúa el lado derecho",
			nil,
			bin(token.O, electrico(true), bin(token.GT, bin(token.SLASH, roca(1), roca(0)), roca(0))),
			"verdadero",
		},
		{"fantasma igual fantasma", posibleVacio, bin(token.IGUAL, id("rival"), fantasma()), "verdadero"},
		{"respaldo con fantasma", posibleVacio, &ast.Respaldo{Valor: id("rival"), Reemplazo: planta("nadie")}, "nadie"},
		{"índice desde 1", equipo, idx(id("e"), roca(2)), "Eevee"},
		{"índice de texto", nil, idx(planta("Pokémon"), roca(4)), "é"},
		{"clave de mochila", mochila, idx(id("m"), planta("agua")), "3"},
		{"contiene en equipo", equipo, bin(token.CONTIENE, id("e"), planta("Eevee")), "verdadero"},
		{"contiene en mochila busca claves", mochila, bin(token.CONTIENE, id("m"), planta("fuego")), "falso"},
		{"texto contiene texto", nil, bin(token.CONTIENE, planta("Pikachu"), planta("chu")), "verdadero"},
		{"texto no contiene texto", nil, bin(token.CONTIENE, planta("Pikachu"), planta("Chu")), "falso"},
		{"texto contiene letra", nil, bin(token.CONTIENE, planta("Pokémon"), fuego('é')), "verdadero"},
		{"texto no contiene letra", nil, bin(token.CONTIENE, planta("Pikachu"), fuego('z')), "falso"},
		{"el texto vacío siempre está contenido", nil, bin(token.CONTIENE, planta(""), planta("")), "verdadero"},
		{"tamaño de equipo", equipo, &ast.Tamano{Valor: id("e")}, "2"},
		{"tamaño de texto cuenta letras", nil, &ast.Tamano{Valor: planta("Pokémon")}, "7"},
		{"redondear la mitad se aleja de cero", nil, &ast.Redondear{Valor: agua(2.5)}, "3"},
		{"redondear negativo", nil, &ast.Redondear{Valor: agua(-2.5)}, "-3"},
		{"convertir agua a roca trunca", nil, &ast.Convertir{Valor: agua(2.9), Destino: tRoca}, "2"},
		{"convertir agua negativa trunca hacia cero", nil, &ast.Convertir{Valor: agua(-2.9), Destino: tRoca}, "-2"},
		{"convertir roca a agua", nil, &ast.Convertir{Valor: roca(12), Destino: tAgua}, "12.0"},
		{"convertir agua a planta", nil, bin(token.PLUS, &ast.Convertir{Valor: agua(12), Destino: tPlanta}, planta("!")), "12.0!"},
		{"convertir planta a roca", nil, &ast.Convertir{Valor: planta(" -42 "), Destino: tRoca}, "-42"},
		{"convertir planta entera a agua", nil, &ast.Convertir{Valor: planta("5"), Destino: tAgua}, "5.0"},
		{"convertir electrico a planta", nil, &ast.Convertir{Valor: electrico(true), Destino: tPlanta}, "verdadero"},
	}
	for _, c := range casos {
		got, err := grito(t, c.antes, c.expr)
		if err != nil {
			t.Errorf("%s: error inesperado: %v", c.nombre, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s: gritar = %q, want %q", c.nombre, got, c.want)
		}
	}
}

func TestAleatorio(t *testing.T) {
	for i := 0; i < 200; i++ {
		got, err := grito(t, nil, &ast.Aleatorio{Min: roca(85), Max: roca(100)})
		if err != nil {
			t.Fatal(err)
		}
		n, err := strconv.Atoi(got)
		if err != nil || n < 85 || n > 100 {
			t.Fatalf("aleatorio(85, 100) = %s, fuera del rango", got)
		}
	}
	got, err := grito(t, nil, &ast.Aleatorio{Min: roca(7), Max: roca(7)})
	if err != nil || got != "7" {
		t.Errorf("aleatorio(7, 7) = %s, %v", got, err)
	}
	_, err = grito(t, nil, &ast.Aleatorio{Min: roca(10), Max: roca(1)})
	if d := errorEjecucion(t, err); d.Code != CodigoAleatorioRango {
		t.Errorf("código = %s", d.Code)
	}
}

// Los errores de ejecución dicen dónde fallaron y con qué valores.
func TestErroresDeEjecucion(t *testing.T) {
	division := bin(token.SLASH, id("vida"), id("turnos"))
	division.Pos = ast.Pos{Line: 7, Col: 12, Len: 13}

	casos := []struct {
		nombre string
		antes  []ast.Instr
		expr   ast.Expr
		codigo string
		desc   string
	}{
		{
			"división entre cero con nombres y valores",
			[]ast.Instr{dato(tRoca, "vida", roca(120)), dato(tRoca, "turnos", roca(0))},
			division,
			CodigoDivisionCero,
			"no se puede dividir vida (120) entre turnos (0).",
		},
		{
			"resto entre cero",
			nil,
			bin(token.RESTO, roca(5), roca(0)),
			CodigoDivisionCero,
			"no se puede calcular el resto de 5 entre 0.",
		},
		{
			"desbordamiento",
			[]ast.Instr{dato(tRoca, "grande", roca(9223372036854775807))},
			bin(token.PLUS, id("grande"), roca(1)),
			CodigoDesbordamiento,
			"el resultado de grande (9223372036854775807) + 1 es demasiado grande.",
		},
		{
			"índice fuera de rango",
			[]ast.Instr{dato(tEquipo(tPlanta), "rivales", equipoLit(planta("a"), planta("b"), planta("c")))},
			idx(id("rivales"), roca(4)),
			CodigoIndiceFueraRango,
			"no existe la posición 4 en rivales.",
		},
		{
			"clave inexistente",
			[]ast.Instr{dato(tMochila(tPlanta, tRoca), "m", llaves())},
			idx(id("m"), planta("agua")),
			CodigoClaveInexistente,
			`la clave "agua" no existe en m.`,
		},
		{
			"convertir un texto que no es número",
			nil,
			&ast.Convertir{Valor: planta("doce"), Destino: tRoca},
			CodigoConversion,
			`no se puede convertir "doce" a roca.`,
		},
		{
			"convertir un agua que no cabe en roca",
			nil,
			&ast.Convertir{Valor: agua(1e20), Destino: tRoca},
			CodigoDesbordamiento,
			"el resultado de convertir(1.0e20) a roca es demasiado grande.",
		},
	}
	for _, c := range casos {
		_, err := grito(t, c.antes, c.expr)
		d := errorEjecucion(t, err)
		if d.Code != c.codigo || d.Desc != c.desc {
			t.Errorf("%s:\n  código = %s, want %s\n  desc   = %s\n  want     %s", c.nombre, d.Code, c.codigo, d.Desc, c.desc)
		}
	}

	_, err := grito(t, []ast.Instr{dato(tRoca, "vida", roca(1)), dato(tRoca, "turnos", roca(0))}, division)
	if d := errorEjecucion(t, err); d.Line != 7 || d.Col != 12 || d.Len != 13 || d.File != "principal.pks" {
		t.Errorf("posición = %s %d:%d (%d), want principal.pks 7:12 (13)", d.File, d.Line, d.Col, d.Len)
	}
}

// Sin analizador, lo que el analizador rechazaría llega como error interno
// en vez de romper el programa.
func TestErrorInterno(t *testing.T) {
	casos := []struct {
		nombre string
		expr   ast.Expr
	}{
		{"nombre no declarado", id("nadie")},
		{"roca más electrico", bin(token.PLUS, roca(1), electrico(true))},
		{"condición que no es electrico", bin(token.Y, roca(1), electrico(true))},
		{"llaves sin tipo esperado", llaves(planta("a"), roca(1))},
		{"movimiento inexistente", llamada("volar")},
		{"convertir fuera de la tabla", &ast.Convertir{Valor: electrico(true), Destino: tRoca}},
	}
	for _, c := range casos {
		_, err := grito(t, nil, c.expr)
		if d := errorEjecucion(t, err); d.Code != CodigoInterno {
			t.Errorf("%s: código = %s, want %s", c.nombre, d.Code, CodigoInterno)
		}
	}
}
