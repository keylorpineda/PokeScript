package analizador

import (
	"strings"
	"testing"

	"github.com/keylorpineda/PokeScript/internal/diag"
)

const especieEstado = "especie Estado\n    SANO, ENVENENADO, DORMIDO, PARALIZADO\nfin\n"

func conEstado(cuerpo string) map[string]string {
	return map[string]string{"principal.pks": especieEstado + "combate\n    Estado e = SANO\n" + cuerpo + "\nfin\n"}
}

// codigosSegunEn devuelve, en orden, los códigos de ds que están en cods.
func codigosSegunEn(ds []diag.Diagnostic, cods ...string) []string {
	var r []string
	for _, d := range ds {
		for _, c := range cods {
			if d.Code == c {
				r = append(r, c)
			}
		}
	}
	return r
}

var codigosSegun = []string{"segun-incompleto", "segun-sin-otro", "segun-tipo-invalido", "rama-inalcanzable", "patron-incompatible", "patron-no-constante"}

func TestSegun(t *testing.T) {
	casos := []struct {
		nombre string
		cuerpo string
		want   []string
	}{
		{
			"especie completa con agrupación",
			"    segun e\n        SANO entonces gritar 1\n        ENVENENADO entonces gritar 2\n        DORMIDO, PARALIZADO entonces gritar 3\n    fin",
			nil,
		},
		{
			"especie incompleta con otro",
			"    segun e\n        SANO entonces gritar 1\n        otro entonces gritar 2\n    fin",
			nil,
		},
		{
			"especie incompleta sin otro",
			"    segun e\n        SANO entonces gritar 1\n        DORMIDO entonces gritar 2\n    fin",
			[]string{"segun-incompleto"},
		},
		{
			"otro redundante",
			"    segun e\n        SANO, ENVENENADO, DORMIDO, PARALIZADO entonces gritar 1\n        otro entonces gritar 2\n    fin",
			[]string{"rama-inalcanzable"},
		},
		{
			"patrón repetido",
			"    segun e\n        SANO entonces gritar 1\n        SANO entonces gritar 2\n        otro entonces gritar 3\n    fin",
			[]string{"rama-inalcanzable"},
		},
		{
			"electrico completo",
			"    segun 1 > 2\n        verdadero entonces gritar 1\n        falso entonces gritar 2\n    fin",
			nil,
		},
		{
			"electrico incompleto",
			"    segun 1 > 2\n        verdadero entonces gritar 1\n    fin",
			[]string{"segun-incompleto"},
		},
		{"roca con otro", "    segun 5\n        1, 2 entonces gritar 1\n        otro entonces gritar 2\n    fin", nil},
		{"roca sin otro", "    segun 5\n        1 entonces gritar 1\n    fin", []string{"segun-sin-otro"}},
		{"planta sin otro", "    segun \"a\"\n        \"a\" entonces gritar 1\n    fin", []string{"segun-sin-otro"}},
		{
			"patrón de otro tipo",
			"    segun 5\n        \"cinco\" entonces gritar 1\n        otro entonces gritar 2\n    fin",
			[]string{"patron-incompatible"},
		},
		{
			"roca como patrón de agua",
			"    segun 2.5\n        1 entonces gritar 1\n        otro entonces gritar 2\n    fin",
			[]string{"patron-incompatible"},
		},
		{
			"patrón que es un dato",
			"    roca n = 1\n    segun 5\n        n entonces gritar 1\n        otro entonces gritar 2\n    fin",
			[]string{"patron-no-constante"},
		},
		{
			"segun sobre un equipo",
			"    equipo de roca l = [1]\n    segun l\n        otro entonces gritar 1\n    fin",
			[]string{"segun-tipo-invalido"},
		},
	}
	for _, c := range casos {
		r := analizar(t, conEstado(c.cuerpo))
		got := codigosSegunEn(r.Diagnosticos, codigosSegun...)
		if strings.Join(got, ",") != strings.Join(c.want, ",") {
			t.Errorf("%s: %v, want %v (todos: %v)", c.nombre, got, c.want, codigos(r.Diagnosticos))
		}
	}
}

func TestSegunConMedallaComoPatron(t *testing.T) {
	archivos := map[string]string{"principal.pks": "medalla roca TOPE = 10\ncombate\n" +
		"    segun 5\n        TOPE entonces gritar 1\n        otro entonces gritar 2\n    fin\nfin\n"}
	if got := codigosSegunEn(analizar(t, archivos).Diagnosticos, codigosSegun...); len(got) != 0 {
		t.Errorf("una medalla es un patrón constante: %v", got)
	}
}

// Caso semántico 2 de la sección 11: segun sobre especie con un valor sin
// cubrir → lo nombra.
func TestCasoSemantico2(t *testing.T) {
	r := analizar(t, conEstado("    segun e\n        SANO entonces gritar 1\n        ENVENENADO entonces gritar 2\n        DORMIDO entonces gritar 3\n    fin"))
	var d *diag.Diagnostic
	for i := range r.Diagnosticos {
		if r.Diagnosticos[i].Code == "segun-incompleto" {
			d = &r.Diagnosticos[i]
		}
	}
	if d == nil {
		t.Fatalf("diagnósticos = %v", codigos(r.Diagnosticos))
	}
	if d.Desc != "este segun no cubre el valor PARALIZADO de Estado." || d.Line != 6 || d.Severity != diag.Error {
		t.Errorf("diagnóstico = %+v", *d)
	}
	if !strings.Contains(d.Suggest, "PARALIZADO") {
		t.Errorf("la sugerencia debe nombrar el faltante: %q", d.Suggest)
	}
}

func TestSegunNombraVariosFaltantesEnOrden(t *testing.T) {
	r := analizar(t, conEstado("    segun e\n        DORMIDO entonces gritar 1\n    fin"))
	for _, d := range r.Diagnosticos {
		if d.Code == "segun-incompleto" && !strings.Contains(d.Desc, "los valores SANO, ENVENENADO, PARALIZADO") {
			t.Errorf("desc = %q", d.Desc)
		}
	}
}

func TestAdvertenciasDeSegun(t *testing.T) {
	r := analizar(t, conEstado("    segun e\n        SANO entonces gritar 1\n        SANO entonces gritar 2\n        otro entonces gritar 3\n    fin"))
	for _, d := range r.Diagnosticos {
		if d.Code == "rama-inalcanzable" && d.Severity != diag.Advertencia {
			t.Errorf("rama-inalcanzable debe ser advertencia: %+v", d)
		}
	}
}
