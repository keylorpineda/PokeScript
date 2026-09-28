package parser

import (
	"reflect"
	"testing"

	"github.com/keylorpineda/PokeScript/internal/ast"
	. "github.com/keylorpineda/PokeScript/internal/ast/astprueba"
	"github.com/keylorpineda/PokeScript/internal/token"
)

func codigos(r ResultadoExpresion) []string {
	var cs []string
	for _, d := range r.Diagnosticos {
		cs = append(cs, d.Code)
	}
	return cs
}

func TestExpresiones(t *testing.T) {
	casos := []struct {
		entrada string
		want    ast.Expr
	}{
		// Precedencia aritmética (hito 2: "2 + 3 * 4").
		{"2 + 3 * 4", Bin(token.PLUS, Roca(2), Bin(token.STAR, Roca(3), Roca(4)))},
		{"(2 + 3) * 4", Bin(token.STAR, Bin(token.PLUS, Roca(2), Roca(3)), Roca(4))},
		{"p - q - r", Bin(token.MINUS, Bin(token.MINUS, Id("p"), Id("q")), Id("r"))},
		{"x / 2 resto 3", Bin(token.RESTO, Bin(token.SLASH, Id("x"), Roca(2)), Roca(3))},
		{"-x * 2", Bin(token.STAR, Un(token.MINUS, Id("x")), Roca(2))},
		{"-5", Un(token.MINUS, Roca(5))},
		{"poder * NIVEL / 50", Bin(token.SLASH, Bin(token.STAR, Id("poder"), Id("NIVEL")), Roca(50))},

		// Lógica y comparaciones.
		{"p y q o r", Bin(token.O, Bin(token.Y, Id("p"), Id("q")), Id("r"))},
		{"p o q y r", Bin(token.O, Id("p"), Bin(token.Y, Id("q"), Id("r")))},
		{"x > 1 y x < 10", Bin(token.Y, Bin(token.GT, Id("x"), Roca(1)), Bin(token.LT, Id("x"), Roca(10)))},
		{"no rival igual fantasma", Un(token.NO, Bin(token.IGUAL, Id("rival"), Fantasma()))},
		{"x + 1 >= q * 2", Bin(token.GE, Bin(token.PLUS, Id("x"), Roca(1)), Bin(token.STAR, Id("q"), Roca(2)))},
		{"equipo_rojo contiene \"Pikachu\"", Bin(token.CONTIENE, Id("equipo_rojo"), Planta("Pikachu"))},
		{"x > 1 igual verdadero", Bin(token.IGUAL, Bin(token.GT, Id("x"), Roca(1)), Electrico(true))},
		{"rival.vida <= 0", Bin(token.LE, Campo(Id("rival"), "vida"), Roca(0))},

		// Respaldo.
		{`rival sino "nadie"`, Respaldo(Id("rival"), Planta("nadie"))},
		{"-rival sino 0", Un(token.MINUS, Respaldo(Id("rival"), Roca(0)))},

		// Acceso.
		{"equipo_ash[1]", Indice(Id("equipo_ash"), Roca(1))},
		{"mio.nombre", Campo(Id("mio"), "nombre")},
		{`m["p"][2].x`, Campo(Indice(Indice(Id("m"), Planta("p")), Roca(2)), "x")},
		{"e[i + 1]", Indice(Id("e"), Bin(token.PLUS, Id("i"), Roca(1)))},

		// Llamadas y funciones incorporadas.
		{"calcular_dano(40)", Llamada("calcular_dano", Roca(40))},
		{"iniciar()", Llamada("iniciar")},
		{"f(p, q + 1)", Llamada("f", Id("p"), Bin(token.PLUS, Id("q"), Roca(1)))},
		{"redondear(total) + 2", Bin(token.PLUS, Redondear(Id("total")), Roca(2))},
		{"aleatorio(85, 100)", Aleatorio(Roca(85), Roca(100))},
		{"tamaño(mi_equipo)", Tamano(Id("mi_equipo"))},
		{"convertir(texto) a roca", Convertir(Id("texto"), TRoca())},
		{"convertir(x) a equipo de planta", Convertir(Id("x"), TEquipo(TPlanta()))},
		{"convertir(x) a Estado", Convertir(Id("x"), TNombre("Estado"))},
		{"f(p)[1].vida", Campo(Indice(Llamada("f", Id("p")), Roca(1)), "vida")},

		// Literales.
		{"6.9", Agua(6.9)},
		{"'P'", Fuego('P')},
		{`"¡Hola!"`, Planta("¡Hola!")},
		{"falso", Electrico(false)},
		{"[]", Equipo()},
		{"[1, 2, 3]", Equipo(Roca(1), Roca(2), Roca(3))},
		{"{}", Llaves()},
		{`{"Aranja": 3, "Zreza": 0}`, Llaves(Planta("Aranja"), Roca(3), Planta("Zreza"), Roca(0))},
		{`{nombre: "Bulbi", vida: VIDA_MAXIMA, estado: SANO}`,
			Llaves(Id("nombre"), Planta("Bulbi"), Id("vida"), Id("VIDA_MAXIMA"), Id("estado"), Id("SANO"))},
		{"[[1], [2, 3]]", Equipo(Equipo(Roca(1)), Equipo(Roca(2), Roca(3)))},
	}
	for _, c := range casos {
		t.Run(c.entrada, func(t *testing.T) {
			r := ParsearExpresion("prueba.pks", c.entrada)
			if len(r.Diagnosticos) > 0 {
				t.Fatalf("diagnósticos inesperados: %v (%s)", codigos(r), r.Diagnosticos[0].Desc)
			}
			if d := Diferencia(r.Expr, c.want); d != "" {
				t.Errorf("árbol distinto: %s", d)
			}
		})
	}
}

func TestPosiciones(t *testing.T) {
	r := ParsearExpresion("prueba.pks", "2 + 3 * 4")
	b := r.Expr.(*ast.Binaria)
	if want := (ast.Pos{Line: 1, Col: 1, Len: 9}); b.Pos != want {
		t.Errorf("Pos de la suma = %+v, want %+v", b.Pos, want)
	}
	if want := (ast.Pos{Line: 1, Col: 3, Len: 1}); b.OpPos != want {
		t.Errorf("OpPos de la suma = %+v, want %+v", b.OpPos, want)
	}
	producto := b.Der.(*ast.Binaria)
	if want := (ast.Pos{Line: 1, Col: 5, Len: 5}); producto.Pos != want {
		t.Errorf("Pos del producto = %+v, want %+v", producto.Pos, want)
	}

	casos := []struct {
		entrada string
		want    ast.Pos
	}{
		{"calcular_dano(40)", ast.Pos{Line: 1, Col: 1, Len: 17}},
		{"mio.nombre", ast.Pos{Line: 1, Col: 1, Len: 10}},
		{"equipo_ash[1]", ast.Pos{Line: 1, Col: 1, Len: 13}},
		{"convertir(t) a roca", ast.Pos{Line: 1, Col: 1, Len: 19}},
		{"[1, 2]", ast.Pos{Line: 1, Col: 1, Len: 6}},
		{"no x", ast.Pos{Line: 1, Col: 1, Len: 4}},
		{"(año)", ast.Pos{Line: 1, Col: 2, Len: 3}}, // los paréntesis no forman nodo
	}
	for _, c := range casos {
		r := ParsearExpresion("prueba.pks", c.entrada)
		if r.Expr == nil {
			t.Errorf("%s: sin árbol, diagnósticos %v", c.entrada, codigos(r))
			continue
		}
		if got := r.Expr.Posicion(); got != c.want {
			t.Errorf("%s: Pos = %+v, want %+v", c.entrada, got, c.want)
		}
	}
}

func TestErroresDeExpresion(t *testing.T) {
	casos := []struct {
		entrada string
		codigos []string
		col     int
	}{
		// Caso sintáctico 2 de la sección 11.
		{"p > q > r", []string{"operador-no-encadenable"}, 7},
		{"p < q >= r", []string{"operador-no-encadenable"}, 7},
		{"p igual q diferente r", []string{"operador-no-encadenable"}, 11},
		{"e contiene 1 contiene 2", []string{"operador-no-encadenable"}, 14},
		{"2 +", []string{"expresion-esperada"}, 4},
		{"* 2", []string{"expresion-esperada"}, 1},
		{"(2 + 3", []string{"falta-cierre"}, 1},
		{"[1, 2", []string{"falta-cierre"}, 1},
		{"{\"p\": 1", []string{"falta-cierre"}, 1},
		{"f(1,", []string{"expresion-esperada"}, 5},
		{"e[1", []string{"falta-cierre"}, 1},
		{"x.", []string{"token-esperado"}, 3},
		{"{nombre \"Bulbi\"}", []string{"token-esperado"}, 9},
		{"convertir(x)", []string{"convertir-sin-tipo"}, 13},
		{"convertir(x) a", []string{"tipo-esperado"}, 15},
		{"convertir(x) a posible equipo de roca", []string{"posible-no-permitido"}, 24},
		{"convertir(x) a mochila de agua a roca", []string{"clave-de-mochila-invalida"}, 27},
		{"convertir(x) a equipo de posible roca", []string{"posible-no-permitido"}, 26},
		{"aleatorio(1)", []string{"cantidad-de-argumentos"}, 10},
		{"tamaño(p, q)", []string{"cantidad-de-argumentos"}, 7},
		{"x y", []string{"expresion-esperada"}, 4},
		{"p q", []string{"sobra-despues-de-expresion"}, 3},
	}
	for _, c := range casos {
		t.Run(c.entrada, func(t *testing.T) {
			r := ParsearExpresion("prueba.pks", c.entrada)
			if got := codigos(r); !reflect.DeepEqual(got, c.codigos) {
				t.Fatalf("códigos = %v, want %v", got, c.codigos)
			}
			if r.Expr != nil {
				t.Error("con errores, Expr debería ser nil")
			}
			d := r.Diagnosticos[0]
			if d.Col != c.col || d.Line != 1 || d.Len < 1 {
				t.Errorf("posición = %d:%d len %d, want 1:%d", d.Line, d.Col, d.Len, c.col)
			}
			if d.Category != "sintactico" || d.Heading == "" || d.Desc == "" || d.Suggest == "" {
				t.Errorf("diagnóstico incompleto: %+v", d)
			}
		})
	}
}

func TestErroresLexicosNoSeDuplican(t *testing.T) {
	// El lexer ya reportó el símbolo; el parser no agrega otro error.
	r := ParsearExpresion("prueba.pks", "2 + @")
	if got := codigos(r); !reflect.DeepEqual(got, []string{"simbolo-desconocido"}) {
		t.Errorf("códigos = %v", got)
	}
	// "==" se reporta en el lexer pero el árbol se arma con igual.
	r = ParsearExpresion("prueba.pks", "p == q")
	if got := codigos(r); !reflect.DeepEqual(got, []string{"simbolo-de-otro-lenguaje"}) {
		t.Errorf("códigos = %v", got)
	}
}
