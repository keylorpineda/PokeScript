package analizador

import (
	"testing"

	"github.com/keylorpineda/PokeScript/internal/diag"
)

func posiblesSinComprobar(ds []diag.Diagnostic) int {
	n := 0
	for _, d := range ds {
		if d.Code == "posible-sin-comprobar" {
			n++
		}
	}
	return n
}

func TestPosible(t *testing.T) {
	const rival = "    posible planta rival = fantasma\n"
	casos := []struct {
		nombre  string
		cuerpo  string
		errores int
	}{
		// Usos que no necesitan comprobar.
		{"comparar con igual", rival + "    electrico e = rival igual fantasma", 0},
		{"gritar un posible", rival + "    gritar rival", 0},
		{"reemplazo con sino", rival + "    planta s = rival sino \"nadie\"", 0},
		{"guardar en otro posible", rival + "    posible planta copia = rival", 0},
		{"asignar un valor", rival + "    rival = \"Bulbi\"", 0},

		// Usos sin comprobar.
		{"concatenar", rival + "    planta s = rival + \"!\"", 1},
		{"guardar en un planta", rival + "    planta s = rival", 1},
		{"tamaño", rival + "    roca n = tamaño(rival)", 1},
		{"aritmética con posible roca", "    posible roca x = 5\n    roca y2 = x + 1", 1},
		{"condición posible", "    posible electrico b = verdadero\n    si b\n    fin", 1},

		// Estrechamiento.
		{"diferente fantasma en la rama afirmativa", rival + "    si rival diferente fantasma\n        planta s = rival + \"!\"\n    fin", 0},
		{"fantasma diferente x también", rival + "    si fantasma diferente rival\n        planta s = rival\n    fin", 0},
		{"igual fantasma en el sino", rival + "    si rival igual fantasma\n        gritar \"nadie\"\n    sino\n        planta s = rival\n    fin", 0},
		{"igual fantasma: la rama afirmativa no", rival + "    si rival igual fantasma\n        planta s = rival\n    fin", 1},
		{"igual fantasma en un sino si", rival + "    si rival igual fantasma\n        gritar \"nadie\"\n    sino si tamaño(rival) > 3\n        planta s = rival\n    fin", 0},
		{"se propaga por y", rival + "    si rival diferente fantasma y tamaño(rival) > 3\n        planta s = rival\n    fin", 0},
		{"no se propaga por o", rival + "    si rival diferente fantasma o verdadero\n        planta s = rival\n    fin", 1},
		{"o no comprueba la derecha", rival + "    electrico e = rival diferente fantasma o tamaño(rival) > 3", 1},
		{"no invierte", rival + "    si no (rival igual fantasma)\n        planta s = rival\n    fin", 0},
		{"mientras", rival + "    mientras rival diferente fantasma\n        planta s = rival\n        rival = fantasma\n    fin", 0},
		{"fuera del bloque ya no", rival + "    si rival diferente fantasma\n        gritar rival\n    fin\n    planta s = rival", 1},
		{"se pierde al reasignar", rival + "    si rival diferente fantasma\n        rival = \"otro\"\n        planta s = rival\n    fin", 1},
		{"se pierde al capturar", rival + "    si rival diferente fantasma\n        capturar(rival, \"?\")\n        planta s = rival\n    fin", 1},
		{"antes de reasignar sí vale", rival + "    si rival diferente fantasma\n        planta s = rival\n        rival = \"otro\"\n    fin", 0},
	}
	for _, c := range casos {
		r := analizar(t, programa(c.cuerpo))
		if got := posiblesSinComprobar(r.Diagnosticos); got != c.errores {
			t.Errorf("%s: posible-sin-comprobar = %d, want %d (todos: %v)", c.nombre, got, c.errores, codigos(r.Diagnosticos))
		}
	}
}

func TestPosibleComoArgumento(t *testing.T) {
	cabecera := "movimiento f(planta p)\n    gritar p\nfin\nmovimiento g(posible planta p)\n    gritar p\nfin\n"
	r := analizar(t, map[string]string{"principal.pks": cabecera +
		"combate\n    posible planta rival = fantasma\n    f(rival)\n    g(rival)\nfin\n"})
	if got := posiblesSinComprobar(r.Diagnosticos); got != 1 {
		t.Errorf("solo f(rival) necesita comprobar: %v", codigos(r.Diagnosticos))
	}
}

// Caso semántico 3 de la sección 11: uso de posible sin comprobar.
func TestCasoSemantico3(t *testing.T) {
	r := analizar(t, programa("    posible planta rival = fantasma\n    gritar \"Rival: \" + rival"))
	var d *diag.Diagnostic
	for i := range r.Diagnosticos {
		if r.Diagnosticos[i].Code == "posible-sin-comprobar" {
			d = &r.Diagnosticos[i]
		}
	}
	if d == nil {
		t.Fatalf("diagnósticos = %v", codigos(r.Diagnosticos))
	}
	if d.Line != 3 || d.Col != 24 || d.Heading != diag.EncabezadoSinValor ||
		d.Desc != "«rival» es de tipo posible planta y puede ser fantasma; hay que comprobarlo antes de usar su valor." {
		t.Errorf("diagnóstico = %+v", *d)
	}
}

// Un posible sin comprobar produce un solo error, no otro de tipos.
func TestPosibleSinCascada(t *testing.T) {
	r := analizar(t, programa("    posible roca x = 5\n    roca z = x + 1"))
	if got := codigos(r.Diagnosticos); len(got) != 1 || got[0] != "posible-sin-comprobar" {
		t.Errorf("diagnósticos = %v", got)
	}
}
