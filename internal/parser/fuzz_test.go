package parser

import (
	"testing"
	"unicode/utf8"
)

// FuzzParsearExpresion verifica que ninguna entrada haga fallar al parser:
// o devuelve un árbol, o devuelve al menos un diagnóstico.
//
//	go test -fuzz=FuzzParsearExpresion -fuzztime=30s ./internal/parser/
func FuzzParsearExpresion(f *testing.F) {
	for _, s := range []string{
		"2 + 3 * 4", "p > q > r", "no x igual fantasma", "f(1, [2, {3: 4}])[1].x sino 0",
		"convertir(x) a mochila de roca a equipo de planta", "((((", "[,]", "{:}", "- - x", "",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, fuente string) {
		if !utf8.ValidString(fuente) {
			t.Skip()
		}
		r := ParsearExpresion("fuzz.pks", fuente)
		if r.Expr == nil && len(r.Diagnosticos) == 0 {
			t.Fatalf("sin árbol y sin diagnósticos para %q", fuente)
		}
		for _, d := range r.Diagnosticos {
			if d.Line < 1 || d.Col < 1 || d.Len < 1 {
				t.Fatalf("diagnóstico con posición inválida: %+v", d)
			}
		}
	})
}
