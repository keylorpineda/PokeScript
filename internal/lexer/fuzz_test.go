package lexer

import (
	"testing"
	"unicode/utf8"

	"github.com/keylorpineda/PokeScript/internal/token"
)

// FuzzAnalizar verifica que ninguna entrada haga fallar o colgar al lexer.
// Con `go test` corre solo las semillas; para explorar más:
//
//	go test -fuzz=FuzzAnalizar -fuzztime=30s ./internal/lexer/
func FuzzAnalizar(f *testing.F) {
	semillas := []string{
		"", "combate\n    gritar \"hola\"\nfin", "sino si", "sino", "6.9", "6.", ".5",
		`"sin cerrar`, `"\q"`, "''", "'ab'", "'\\", "\"\\", "3vidas", "a == b", "; % @ #",
		"99999999999999999999", "// solo comentario", "\r\n\r\n", "\uFEFF", "ñ_1 Á",
	}
	for _, s := range semillas {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, fuente string) {
		if !utf8.ValidString(fuente) {
			t.Skip()
		}
		r := Analizar("fuzz.pks", fuente)
		if len(r.Tokens) == 0 || r.Tokens[len(r.Tokens)-1].Kind != token.EOF {
			t.Fatalf("el último token debe ser EOF: %v", r.Tokens)
		}
		for _, tk := range r.Tokens {
			if tk.Line < 1 || tk.Col < 1 || tk.Len < 0 {
				t.Fatalf("posición inválida: %+v", tk)
			}
		}
		for _, d := range r.Diagnosticos {
			if d.Line < 1 || d.Col < 1 || d.Len < 1 {
				t.Fatalf("diagnóstico con posición inválida: %+v", d)
			}
		}
	})
}
