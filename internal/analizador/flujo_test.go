package analizador

import (
	"strings"
	"testing"

	"github.com/keylorpineda/PokeScript/internal/diag"
)

func soloDe(ds []diag.Diagnostic, codigos ...string) []string {
	var r []string
	for _, d := range ds {
		for _, c := range codigos {
			if d.Code == c {
				r = append(r, d.Code)
			}
		}
	}
	return r
}

var codigosFlujo = []string{"dato-sin-valor", "falta-entregar", "entregar-con-valor", "entregar-sin-valor", "entregar-fuera-de-movimiento"}

func TestAsignacionDefinida(t *testing.T) {
	casos := []struct {
		nombre string
		cuerpo string
		error  bool
	}{
		{"asignado antes de leer", "    roca x\n    x = 1\n    gritar x", false},
		{"capturar asigna", "    roca x\n    capturar(x, \"?\")\n    gritar x", false},
		{"leído sin valor", "    roca x\n    gritar x", true},
		{"se lee en su propia asignación", "    roca x\n    x = x + 1", true},
		{"si con sino que asigna en las dos ramas", "    roca x\n    si verdadero\n        x = 1\n    sino\n        x = 2\n    fin\n    gritar x", false},
		{"si con sino que asigna en una rama", "    roca x\n    si verdadero\n        x = 1\n    sino\n        gritar \"no\"\n    fin\n    gritar x", true},
		{"si sin sino", "    roca x\n    si verdadero\n        x = 1\n    fin\n    gritar x", true},
		{"sino si sin sino", "    roca x\n    si verdadero\n        x = 1\n    sino si falso\n        x = 2\n    fin\n    gritar x", true},
		{"sino si con sino", "    roca x\n    si verdadero\n        x = 1\n    sino si falso\n        x = 2\n    sino\n        x = 3\n    fin\n    gritar x", false},
		{"el cuerpo de mientras puede no ejecutarse", "    roca x\n    mientras falso\n        x = 1\n    fin\n    gritar x", true},
		{"el cuerpo de recorrer puede no ejecutarse", "    roca x\n    recorrer i de 1 hasta 3\n        x = i\n    fin\n    gritar x", true},
		{"dentro del ciclo sí cuenta", "    roca x\n    recorrer i de 1 hasta 3\n        x = i\n        gritar x\n    fin", false},
		{"segun con otro", "    roca x\n    segun 1\n        1 entonces x = 1\n        otro entonces x = 2\n    fin\n    gritar x", false},
		{"segun sin otro no exhaustivo", "    roca x\n    segun 1\n        1 entonces x = 1\n    fin\n    gritar x", true},
		{"segun exhaustivo sobre electrico", "    roca x\n    segun verdadero\n        verdadero entonces x = 1\n        falso entonces x = 2\n    fin\n    gritar x", false},
		{"leer un elemento lee el dato", "    equipo de roca e\n    gritar e[1]", true},
		{"asignar a un elemento lee el dato", "    equipo de roca e\n    e[1] = 5", true},
		{"un dato con valor inicial", "    roca x = 1\n    gritar x", false},
		{"un aviso por dato", "    roca x\n    gritar x\n    gritar x", true},
	}
	for _, c := range casos {
		r := analizar(t, programa(c.cuerpo))
		got := soloDe(r.Diagnosticos, "dato-sin-valor")
		if c.error && len(got) != 1 || !c.error && len(got) != 0 {
			t.Errorf("%s: dato-sin-valor = %d, todos = %v", c.nombre, len(got), codigos(r.Diagnosticos))
		}
	}
}

func TestAsignacionDefinidaConEspecie(t *testing.T) {
	const especie = "especie Estado\n    SANO, DORMIDO\nfin\n"
	completo := especie + "combate\n    Estado e = SANO\n    roca x\n    segun e\n        SANO entonces x = 1\n        DORMIDO entonces x = 2\n    fin\n    gritar x\nfin\n"
	incompleto := especie + "combate\n    Estado e = SANO\n    roca x\n    segun e\n        SANO entonces x = 1\n    fin\n    gritar x\nfin\n"
	if got := soloDe(analizar(t, map[string]string{"principal.pks": completo}).Diagnosticos, "dato-sin-valor"); len(got) != 0 {
		t.Errorf("segun exhaustivo sobre especie: %v", got)
	}
	if got := soloDe(analizar(t, map[string]string{"principal.pks": incompleto}).Diagnosticos, "dato-sin-valor"); len(got) != 1 {
		t.Errorf("segun incompleto sobre especie: %v", got)
	}
}

// Caso semántico 5 de la sección 11: dato leído sin valor asignado.
func TestCasoSemantico5(t *testing.T) {
	r := analizar(t, programa("    roca vida\n    si verdadero\n        vida = 10\n    fin\n    gritar vida"))
	var d *diag.Diagnostic
	for i := range r.Diagnosticos {
		if r.Diagnosticos[i].Code == "dato-sin-valor" {
			d = &r.Diagnosticos[i]
		}
	}
	if d == nil {
		t.Fatalf("diagnósticos = %v", codigos(r.Diagnosticos))
	}
	if d.Line != 6 || d.Col != 12 || d.Heading != diag.EncabezadoSinValor || !strings.Contains(d.Cause, "línea 2") {
		t.Errorf("diagnóstico = %+v", *d)
	}
}

func TestNoSonLecturas(t *testing.T) {
	archivos := map[string]string{"principal.pks": "ficha Pokemon\n    planta nombre\n    roca vida\nfin\n" +
		"movimiento roca vida_de(Pokemon p)\n    entregar p.vida\nfin\n" +
		"combate\n" +
		"    planta nombre\n    roca vida\n" +
		"    Pokemon p = {nombre: \"a\", vida: 1}\n" + // claves de ficha, no lecturas
		"    roca n = vida_de(p)\n" + // «vida» en p.vida tampoco
		"    gritar p.nombre\n" +
		"fin\n"}
	if got := soloDe(analizar(t, archivos).Diagnosticos, "dato-sin-valor"); len(got) != 0 {
		t.Errorf("diagnósticos = %v", got)
	}
}

func TestRetornoPorTodosLosCaminos(t *testing.T) {
	casos := []struct {
		nombre string
		cuerpo string
		error  bool
	}{
		{"entrega al final", "    entregar 1", false},
		{"no entrega", "    gritar \"hola\"", true},
		{"si con sino que entrega en las dos", "    si verdadero\n        entregar 1\n    sino\n        entregar 2\n    fin", false},
		{"si sin sino", "    si verdadero\n        entregar 1\n    fin", true},
		{"si sin sino y entrega después", "    si verdadero\n        entregar 1\n    fin\n    entregar 2", false},
		{"sino que no entrega", "    si verdadero\n        entregar 1\n    sino\n        gritar \"x\"\n    fin", true},
		{"solo dentro de un ciclo", "    mientras verdadero\n        entregar 1\n    fin", true},
		{"segun con otro", "    segun 1\n        1 entonces entregar 1\n        otro entonces entregar 2\n    fin", false},
		{"segun sin otro", "    segun 1\n        1 entonces entregar 1\n    fin", true},
	}
	for _, c := range casos {
		archivos := map[string]string{"principal.pks": "movimiento roca f()\n" + c.cuerpo + "\nfin\ncombate\n    roca x = f()\nfin\n"}
		got := soloDe(analizar(t, archivos).Diagnosticos, "falta-entregar")
		if c.error && len(got) != 1 || !c.error && len(got) != 0 {
			t.Errorf("%s: falta-entregar = %v", c.nombre, got)
		}
	}
}

func TestFaltaEntregarApuntaAlFin(t *testing.T) {
	r := analizar(t, map[string]string{"principal.pks": "movimiento roca f()\n    gritar \"x\"\nfin\ncombate\n    roca x = f()\nfin\n"})
	for _, d := range r.Diagnosticos {
		if d.Code == "falta-entregar" && (d.Line != 3 || d.Col != 1) {
			t.Errorf("falta-entregar en %d:%d, want 3:1", d.Line, d.Col)
		}
	}
}

func TestEntregarConYSinValor(t *testing.T) {
	casos := []struct {
		nombre, archivo, codigo string
	}{
		{"con valor en movimiento sin tipo", "movimiento f()\n    entregar 1\nfin\ncombate\n    f()\nfin\n", "entregar-con-valor"},
		{"sin valor en movimiento con tipo", "movimiento roca f()\n    entregar\nfin\ncombate\n    roca x = f()\nfin\n", "entregar-sin-valor"},
		{"en combate", "combate\n    entregar\nfin\n", "entregar-fuera-de-movimiento"},
	}
	for _, c := range casos {
		got := soloDe(analizar(t, map[string]string{"principal.pks": c.archivo}).Diagnosticos, codigosFlujo...)
		if len(got) != 1 || got[0] != c.codigo {
			t.Errorf("%s: %v, want [%s]", c.nombre, got, c.codigo)
		}
	}
	// Un entregar vacío en un movimiento sin tipo es válido.
	ok := "movimiento f()\n    si verdadero\n        entregar\n    fin\n    gritar \"x\"\nfin\ncombate\n    f()\nfin\n"
	if got := soloDe(analizar(t, map[string]string{"principal.pks": ok}).Diagnosticos, codigosFlujo...); len(got) != 0 {
		t.Errorf("entregar sin valor en movimiento sin tipo: %v", got)
	}
}

// Lo que sigue a entregar no se alcanza: no produce avisos.
func TestCodigoInalcanzableNoSeRevisa(t *testing.T) {
	archivos := map[string]string{"principal.pks": "movimiento roca f()\n    roca x\n    entregar 1\n    gritar x\nfin\ncombate\n    roca z = f()\nfin\n"}
	if got := soloDe(analizar(t, archivos).Diagnosticos, codigosFlujo...); len(got) != 0 {
		t.Errorf("diagnósticos = %v", got)
	}
}
