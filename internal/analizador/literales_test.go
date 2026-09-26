package analizador

import (
	"strings"
	"testing"

	"github.com/keylorpineda/PokeScript/internal/ast"
)

const fichaPokemon = "especie Estado\n    SANO, DORMIDO\nfin\n" +
	"ficha Pokemon\n    planta nombre\n    roca vida\n    Estado estado\nfin\n"

func conFicha(cuerpo string) map[string]string {
	return map[string]string{"principal.pks": fichaPokemon + "combate\n" + cuerpo + "\nfin\n"}
}

func TestLiteralesValidos(t *testing.T) {
	casos := map[string]string{
		"equipo":                   "    equipo de roca e = [1, 2]",
		"equipo de agua ensancha":  "    equipo de agua e = [1, 2.5]",
		"equipo vacío":             "    equipo de planta e = []",
		"equipo anidado":           "    equipo de equipo de roca e = [[1], [2, 3]]",
		"mochila":                  "    mochila de planta a roca m = {\"a\": 1, \"b\": 2}",
		"mochila vacía":            "    mochila de planta a roca m = {}",
		"mochila de equipos":       "    mochila de planta a equipo de roca m = {\"a\": [1]}",
		"ficha completa":           "    Pokemon p = {nombre: \"Bulbi\", vida: 45, estado: SANO}",
		"ficha en otro orden":      "    Pokemon p = {estado: DORMIDO, vida: 1, nombre: \"x\"}",
		"equipo de fichas":         "    equipo de Pokemon e = [{nombre: \"a\", vida: 1, estado: SANO}]",
		"reasignar con un literal": "    equipo de roca e = []\n    e = [4, 5]",
	}
	for nombre, cuerpo := range casos {
		r := analizar(t, conFicha(cuerpo))
		if len(r.Diagnosticos) != 0 {
			t.Errorf("%s: diagnósticos inesperados: %v", nombre, codigos(r.Diagnosticos))
		}
	}
}

func TestLiteralComoArgumentoYEntrega(t *testing.T) {
	r := analizar(t, map[string]string{"principal.pks": fichaPokemon +
		"movimiento roca total(equipo de roca e)\n    entregar tamaño(e)\nfin\n" +
		"movimiento Pokemon nuevo(planta n)\n    entregar {nombre: n, vida: 10, estado: SANO}\nfin\n" +
		"combate\n    roca t = total([1, 2, 3])\n    Pokemon p = nuevo(\"a\")\nfin\n"})
	if len(r.Diagnosticos) != 0 {
		t.Errorf("diagnósticos inesperados: %v", codigos(r.Diagnosticos))
	}
}

func TestLiteralesErrores(t *testing.T) {
	casos := []struct {
		nombre string
		cuerpo string
		codigo string
	}{
		{"equipo suelto en gritar", "    gritar [1, 2]", "literal-sin-tipo"},
		{"llaves sueltas", "    gritar {\"a\": 1}", "literal-sin-tipo"},
		{"equipo con un elemento de otro tipo", "    equipo de roca e = [1, \"dos\"]", "tipo-incompatible"},
		{"equipo donde va una roca", "    roca x = [1]", "literal-incompatible"},
		{"llaves donde va un equipo", "    equipo de roca e = {}", "literal-incompatible"},
		{"clave de mochila de otro tipo", "    mochila de planta a roca m = {1: 1}", "tipo-incompatible"},
		{"valor de mochila de otro tipo", "    mochila de planta a roca m = {\"a\": \"b\"}", "tipo-incompatible"},
		{"campo ajeno", "    Pokemon p = {nombre: \"a\", vida: 1, estado: SANO, nivel: 5}", "campo-ajeno"},
		{"campo repetido", "    Pokemon p = {nombre: \"a\", nombre: \"b\", vida: 1, estado: SANO}", "campo-repetido-en-literal"},
		{"campo de otro tipo", "    Pokemon p = {nombre: 1, vida: 1, estado: SANO}", "tipo-incompatible"},
		{"clave calculada en ficha", "    Pokemon p = {\"nombre\": \"a\", vida: 1, estado: SANO}", "clave-de-ficha-invalida"},
		{"ficha incompleta", "    Pokemon p = {nombre: \"a\"}", "ficha-incompleta"},
	}
	for _, c := range casos {
		r := analizar(t, conFicha(c.cuerpo))
		if got := conCodigo(r.Diagnosticos, c.codigo); len(got) != 1 {
			t.Errorf("%s: esperaba un %s, llegó %v", c.nombre, c.codigo, codigos(r.Diagnosticos))
		}
	}
}

func TestFichaIncompletaNombraLosFaltantes(t *testing.T) {
	r := analizar(t, conFicha("    Pokemon p = {vida: 1}"))
	ds := conCodigo(r.Diagnosticos, "ficha-incompleta")
	if len(ds) != 1 || !strings.Contains(ds[0].Desc, "faltan los campos nombre, estado") {
		t.Errorf("diagnósticos = %+v", ds)
	}
}

// El analizador anota qué es cada { } para el intérprete.
func TestLlavesResueltas(t *testing.T) {
	archivos := conFicha("    Pokemon p = {nombre: \"a\", vida: 1, estado: SANO}\n" +
		"    mochila de planta a roca m = {\"x\": 1}")
	p := cargarProyecto(t, archivos)
	if ds := Analizar(p).Diagnosticos; len(ds) != 0 {
		t.Fatalf("diagnósticos: %v", codigos(ds))
	}
	var formas []ast.FormaLlaves
	ast.InspeccionarPrograma(p.Archivos["principal.pks"].Programa, func(n ast.Nodo) bool {
		if l, ok := n.(*ast.LitLlaves); ok {
			formas = append(formas, l.Resuelto)
		}
		return true
	})
	if len(formas) != 2 || formas[0] != ast.LlavesFicha || formas[1] != ast.LlavesMochila {
		t.Errorf("formas = %v", formas)
	}
}

// Un tipo que no existe ya tiene su error: el literal no suma otro.
func TestLiteralSinCascada(t *testing.T) {
	r := analizar(t, programa("    Desconocido d = {campo: 1}"))
	if got := codigos(r.Diagnosticos); len(got) != 1 || got[0] != "tipo-desconocido" {
		t.Errorf("diagnósticos = %v", got)
	}
}
