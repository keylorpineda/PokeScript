package lexer

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/keylorpineda/PokeScript/internal/token"
)

// K abrevia token.Kind en las tablas de pruebas.
type K = token.Kind

func kinds(r Resultado) []K {
	ks := make([]K, len(r.Tokens))
	for i, t := range r.Tokens {
		ks[i] = t.Kind
	}
	return ks
}

func codigos(r Resultado) []string {
	cs := make([]string, len(r.Diagnosticos))
	for i, d := range r.Diagnosticos {
		cs[i] = d.Code
	}
	return cs
}

func TestKinds(t *testing.T) {
	const NL, EOF = token.NEWLINE, token.EOF
	casos := []struct {
		nombre  string
		entrada string
		want    []K
	}{
		{"vacío", "", []K{EOF}},
		{"solo espacios y líneas vacías", "   \n\n\t\n", []K{EOF}},
		{"solo comentario", "// hola\n", []K{EOF}},
		{"declaración", "roca vida = 100", []K{token.ROCA, token.IDENT, token.ASSIGN, token.ROCA_LIT, NL, EOF}},
		{"agua", "agua x = 6.9", []K{token.AGUA, token.IDENT, token.ASSIGN, token.AGUA_LIT, NL, EOF}},
		{"acceso a campo", "chispo.vida", []K{token.IDENT, token.DOT, token.IDENT, NL, EOF}},
		{"número seguido de punto sin dígito", "6.", []K{token.ROCA_LIT, token.DOT, NL, EOF}},
		{"punto seguido de número", ".5", []K{token.DOT, token.ROCA_LIT, NL, EOF}},
		{"menos es operador", "-5", []K{token.MINUS, token.ROCA_LIT, NL, EOF}},
		{"sino si", "sino si x", []K{token.SINO_SI, token.IDENT, NL, EOF}},
		{"sino si con varios espacios", "sino \t si x", []K{token.SINO_SI, token.IDENT, NL, EOF}},
		{"sino solo", "sino", []K{token.SINO, NL, EOF}},
		{"sino seguido de siguiente", "sino siguiente", []K{token.SINO, token.SIGUIENTE, NL, EOF}},
		{"sino seguido de identificador que empieza con si", "x sino si_algo", []K{token.IDENT, token.SINO, token.IDENT, NL, EOF}},
		{"sinosi pegado es identificador", "sinosi", []K{token.IDENT, NL, EOF}},
		{"sino y si en líneas distintas", "sino\nsi x", []K{token.SINO, NL, token.SI, token.IDENT, NL, EOF}},
		{"comparaciones", "p >= q <= r > s < t", []K{token.IDENT, token.GE, token.IDENT, token.LE, token.IDENT, token.GT, token.IDENT, token.LT, token.IDENT, NL, EOF}},
		{"aritmética", "p + q - r * s / t resto u", []K{token.IDENT, token.PLUS, token.IDENT, token.MINUS, token.IDENT, token.STAR, token.IDENT, token.SLASH, token.IDENT, token.RESTO, token.IDENT, NL, EOF}},
		{"agrupadores", "( ) [ ] { } , :", []K{token.LPAREN, token.RPAREN, token.LBRACKET, token.RBRACKET, token.LBRACE, token.RBRACE, token.COMMA, token.COLON, NL, EOF}},
		{"comentario al final de la línea", "gritar x // saludo", []K{token.GRITAR, token.IDENT, NL, EOF}},
		{"división no es comentario", "p / q", []K{token.IDENT, token.SLASH, token.IDENT, NL, EOF}},
		{"líneas en blanco no duplican NEWLINE", "p\n\n\nq\n", []K{token.IDENT, NL, token.IDENT, NL, EOF}},
		{"línea de comentario entre instrucciones", "p\n// nota\nq", []K{token.IDENT, NL, token.IDENT, NL, EOF}},
		{"finales de línea de Windows", "p\r\nq\r\n", []K{token.IDENT, NL, token.IDENT, NL, EOF}},
		{"identificadores con ñ y tildes", "año número Ñandú", []K{token.IDENT, token.IDENT, token.IDENT, NL, EOF}},
		{"palabras con ñ", "enseñar tamaño", []K{token.ENSENAR, token.TAMANO, NL, EOF}},
		{"guion bajo y dígitos", "calcular_dano2", []K{token.IDENT, NL, EOF}},
		{"distingue mayúsculas", "Roca roca", []K{token.IDENT, token.ROCA, NL, EOF}},
		{"literales electrico y fantasma", "verdadero falso fantasma", []K{token.VERDADERO, token.FALSO, token.FANTASMA, NL, EOF}},
		{"texto y carácter", `"hola" 'x'`, []K{token.PLANTA_LIT, token.FUEGO_LIT, NL, EOF}},
		{"marca BOM al inicio", "\uFEFFcombate", []K{token.COMBATE, NL, EOF}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			r := Analizar("prueba.pks", c.entrada)
			if len(r.Diagnosticos) > 0 {
				t.Fatalf("diagnósticos inesperados: %v", codigos(r))
			}
			if got := kinds(r); !reflect.DeepEqual(got, c.want) {
				t.Errorf("kinds =\n  %v\nwant\n  %v", got, c.want)
			}
		})
	}
}

func TestPosiciones(t *testing.T) {
	r := Analizar("prueba.pks", "combate\n    roca año = 5\n    sino si x\nfin")
	want := []token.Token{
		{Kind: token.COMBATE, Lexeme: "combate", Line: 1, Col: 1, Len: 7},
		{Kind: token.NEWLINE, Lexeme: "\n", Line: 1, Col: 8, Len: 1},
		{Kind: token.ROCA, Lexeme: "roca", Line: 2, Col: 5, Len: 4},
		{Kind: token.IDENT, Lexeme: "año", Line: 2, Col: 10, Len: 3}, // en runas, no en bytes
		{Kind: token.ASSIGN, Lexeme: "=", Line: 2, Col: 14, Len: 1},
		{Kind: token.ROCA_LIT, Lexeme: "5", Line: 2, Col: 16, Len: 1},
		{Kind: token.NEWLINE, Lexeme: "\n", Line: 2, Col: 17, Len: 1},
		{Kind: token.SINO_SI, Lexeme: "sino si", Line: 3, Col: 5, Len: 7},
		{Kind: token.IDENT, Lexeme: "x", Line: 3, Col: 13, Len: 1},
		{Kind: token.NEWLINE, Lexeme: "\n", Line: 3, Col: 14, Len: 1},
		{Kind: token.FIN, Lexeme: "fin", Line: 4, Col: 1, Len: 3},
		{Kind: token.NEWLINE, Lexeme: "", Line: 4, Col: 4, Len: 0},
		{Kind: token.EOF, Line: 4, Col: 4},
	}
	if !reflect.DeepEqual(r.Tokens, want) {
		for i := range max(len(r.Tokens), len(want)) {
			var g, w token.Token
			if i < len(r.Tokens) {
				g = r.Tokens[i]
			}
			if i < len(want) {
				w = want[i]
			}
			if g != w {
				t.Errorf("token %d = %+v, want %+v", i, g, w)
			}
		}
	}
}

func TestSangrias(t *testing.T) {
	r := Analizar("prueba.pks", "combate\n    roca x = 1\n\n  // comentario\n\tgritar x\nfin\n")
	want := map[int]int{1: 1, 2: 5, 5: 2, 6: 1}
	if !reflect.DeepEqual(r.Sangrias, want) {
		t.Errorf("Sangrias = %v, want %v", r.Sangrias, want)
	}
}

func TestValoresDeLiterales(t *testing.T) {
	casos := []struct {
		entrada string
		kind    K
		lexema  string
		largo   int
	}{
		{`"hola"`, token.PLANTA_LIT, "hola", 6},
		{`""`, token.PLANTA_LIT, "", 2},
		{`"dijo \"hola\""`, token.PLANTA_LIT, `dijo "hola"`, 15},
		{`"línea\nnueva"`, token.PLANTA_LIT, "línea\nnueva", 14},
		{`"barra \\"`, token.PLANTA_LIT, `barra \`, 10},
		{`"¡Hola, Pokémon!"`, token.PLANTA_LIT, "¡Hola, Pokémon!", 17},
		{`'P'`, token.FUEGO_LIT, "P", 3},
		{`'ñ'`, token.FUEGO_LIT, "ñ", 3},
		{`'\''`, token.FUEGO_LIT, "'", 4},
		{`'\n'`, token.FUEGO_LIT, "\n", 4},
		{`'"'`, token.FUEGO_LIT, `"`, 3},
		{"42", token.ROCA_LIT, "42", 2},
		{"0.95", token.AGUA_LIT, "0.95", 4},
		{"9223372036854775807", token.ROCA_LIT, "9223372036854775807", 19},
	}
	for _, c := range casos {
		r := Analizar("prueba.pks", c.entrada)
		if len(r.Diagnosticos) > 0 {
			t.Errorf("%s: diagnósticos inesperados %v", c.entrada, codigos(r))
			continue
		}
		tk := r.Tokens[0]
		if tk.Kind != c.kind || tk.Lexeme != c.lexema || tk.Len != c.largo {
			t.Errorf("%s: token = {%v %q Len:%d}, want {%v %q Len:%d}", c.entrada, tk.Kind, tk.Lexeme, tk.Len, c.kind, c.lexema, c.largo)
		}
	}
}

func TestErrores(t *testing.T) {
	casos := []struct {
		nombre  string
		entrada string
		codigos []string
		col     int // columna del primer diagnóstico
		largo   int // Len del primer diagnóstico
	}{
		// Caso sintáctico 4 de la sección 11.
		{"cadena sin cerrar", `gritar "hola`, []string{"cadena-sin-cerrar"}, 8, 5},
		{"cadena cortada por salto de línea", "gritar \"hola\ngritar x", []string{"cadena-sin-cerrar"}, 8, 5},
		{"escape desconocido", `"a\tb"`, []string{"escape-desconocido"}, 3, 2},
		{"dos escapes desconocidos", `"\q\w"`, []string{"escape-desconocido", "escape-desconocido"}, 2, 2},
		{"carácter vacío", "''", []string{"caracter-vacio"}, 1, 2},
		{"carácter con dos letras", "'ab'", []string{"caracter-largo"}, 1, 4},
		{"carácter sin cerrar", "'a", []string{"caracter-sin-cerrar"}, 1, 2},
		{"identificador que empieza con número", "roca 3vidas = 1", []string{"identificador-con-numero"}, 6, 6},
		{"número demasiado grande", "9223372036854775808", []string{"numero-muy-grande"}, 1, 19},
		{"punto y coma", "p = 1;", []string{"punto-y-coma"}, 6, 1},
		{"doble igual", "si p == q", []string{"simbolo-de-otro-lenguaje"}, 6, 2},
		{"distinto", "si p != q", []string{"simbolo-de-otro-lenguaje"}, 6, 2},
		{"y lógico", "p && q", []string{"simbolo-de-otro-lenguaje"}, 3, 2},
		{"o lógico", "p || q", []string{"simbolo-de-otro-lenguaje"}, 3, 2},
		{"porcentaje", "p % q", []string{"simbolo-de-otro-lenguaje"}, 3, 1},
		{"símbolo desconocido", "p @ q", []string{"simbolo-desconocido"}, 3, 1},
		{"guion bajo al inicio", "_x", []string{"simbolo-desconocido"}, 1, 1},
		{"sigue después de un error", "p @ q\nr # s", []string{"simbolo-desconocido", "simbolo-desconocido"}, 3, 1},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			r := Analizar("prueba.pks", c.entrada)
			if got := codigos(r); !reflect.DeepEqual(got, c.codigos) {
				t.Fatalf("códigos = %v, want %v", got, c.codigos)
			}
			d := r.Diagnosticos[0]
			if d.Col != c.col || d.Len != c.largo || d.Line != 1 {
				t.Errorf("posición = línea %d col %d len %d, want línea 1 col %d len %d", d.Line, d.Col, d.Len, c.col, c.largo)
			}
			if d.Category != "lexico" || d.Severity != "error" || d.File != "prueba.pks" {
				t.Errorf("diagnóstico mal clasificado: %+v", d)
			}
			if d.Heading == "" || d.Desc == "" || d.Suggest == "" {
				t.Errorf("falta encabezado, descripción o sugerencia: %+v", d)
			}
		})
	}
}

func TestSimbolosDeOtrosLenguajesSeReemplazan(t *testing.T) {
	r := Analizar("prueba.pks", "si p == q && r != s || t % u")
	want := []K{
		token.SI, token.IDENT, token.IGUAL, token.IDENT, token.Y, token.IDENT,
		token.DIFERENTE, token.IDENT, token.O, token.IDENT, token.RESTO, token.IDENT,
		token.NEWLINE, token.EOF,
	}
	if got := kinds(r); !reflect.DeepEqual(got, want) {
		t.Errorf("kinds =\n  %v\nwant\n  %v", got, want)
	}
	if len(r.Diagnosticos) != 5 {
		t.Errorf("se esperaban 5 diagnósticos, hubo %d", len(r.Diagnosticos))
	}
}

func TestPuntoYComaNoDejaToken(t *testing.T) {
	r := Analizar("prueba.pks", "p = 1; q = 2")
	want := []K{token.IDENT, token.ASSIGN, token.ROCA_LIT, token.IDENT, token.ASSIGN, token.ROCA_LIT, token.NEWLINE, token.EOF}
	if got := kinds(r); !reflect.DeepEqual(got, want) {
		t.Errorf("kinds = %v, want %v", got, want)
	}
}

// TestEjemplosDeLaEspecificacion analiza los cuatro archivos del programa de
// la sección 10: no deben producir ningún diagnóstico léxico.
func TestEjemplosDeLaEspecificacion(t *testing.T) {
	archivos, err := filepath.Glob(filepath.Join("testdata", "*.pks"))
	if err != nil || len(archivos) == 0 {
		t.Fatalf("no hay archivos en testdata: %v", err)
	}
	for _, ruta := range archivos {
		t.Run(filepath.Base(ruta), func(t *testing.T) {
			fuente, err := os.ReadFile(ruta)
			if err != nil {
				t.Fatal(err)
			}
			r := Analizar(filepath.Base(ruta), string(fuente))
			for _, d := range r.Diagnosticos {
				t.Errorf("%d:%d %s: %s", d.Line, d.Col, d.Code, d.Desc)
			}
			for _, tk := range r.Tokens {
				if tk.Kind == token.ILLEGAL {
					t.Errorf("token ILLEGAL en %d:%d: %q", tk.Line, tk.Col, tk.Lexeme)
				}
			}
			// Cada línea con código termina en exactamente un NEWLINE.
			lineas := 0
			for _, l := range strings.Split(string(fuente), "\n") {
				s := strings.TrimSpace(l)
				if s != "" && !strings.HasPrefix(s, "//") {
					lineas++
				}
			}
			nl := 0
			for _, tk := range r.Tokens {
				if tk.Kind == token.NEWLINE {
					nl++
				}
			}
			if nl != lineas {
				t.Errorf("hay %d NEWLINE y %d líneas con código", nl, lineas)
			}
		})
	}
}
