// «vdia» está mal escrito a propósito: es el error que el asistente corrige.
// cspell:words vdia

package servicio

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/keylorpineda/PokeScript/internal/diag"
)

func buscarSimbolo(ss []Simbolo, nombre, clase string) *Simbolo {
	for i := range ss {
		if ss[i].Nombre == nombre && ss[i].Clase == clase {
			return &ss[i]
		}
	}
	return nil
}

func TestResultadoSeccion10(t *testing.T) {
	c, err := Compilar(filepath.Join("..", "..", "ejemplos", "combate"))
	if err != nil {
		t.Fatal(err)
	}
	r := c.Resultado()
	if !r.Exito || r.Encabezado != diag.EncabezadoExito || r.Principal != "principal.pks" || len(r.Archivos) != 4 {
		t.Fatalf("resultado = %+v", r)
	}
	ss := r.Simbolos["principal.pks"]
	casos := []struct {
		nombre, clase, tipo, detalle string
	}{
		{"calcular_dano", SimboloMovimiento, "roca", "calcular_dano(roca poder)"},
		{"describir", SimboloMovimiento, "", "describir(Estado actual)"},
		{"Pokemon", SimboloFicha, "", "planta nombre, roca vida, Estado estado"},
		{"Estado", SimboloEspecie, "", "SANO, ENVENENADO, DORMIDO, PARALIZADO"},
		{"SANO", SimboloValor, "Estado", ""},
		{"VIDA_MAXIMA", SimboloMedalla, "roca", ""},
		{"mio", SimboloDato, "Pokemon", ""},
		{"turno", SimboloDato, "roca", ""},
		{"actual", SimboloParametro, "Estado", ""},
	}
	for _, c := range casos {
		s := buscarSimbolo(ss, c.nombre, c.clase)
		if s == nil {
			t.Errorf("falta el símbolo %s (%s)", c.nombre, c.clase)
			continue
		}
		if s.Tipo != c.tipo || s.Detalle != c.detalle {
			t.Errorf("%s = %+v, want tipo %q detalle %q", c.nombre, *s, c.tipo, c.detalle)
		}
	}
	if buscarSimbolo(ss, "NIVEL", SimboloMedalla) != nil {
		t.Error("NIVEL no se importa en principal.pks: no debe sugerirse")
	}
	for i := 1; i < len(ss); i++ {
		if ss[i-1].Nombre > ss[i].Nombre {
			t.Fatalf("los símbolos deben venir ordenados: %q antes de %q", ss[i-1].Nombre, ss[i].Nombre)
		}
	}
}

func TestResultadoConErroresComoJSON(t *testing.T) {
	c := compilarMapa(t, map[string]string{"principal.pks": "combate\n    roca vida = 1\n    gritar vida\n    gritar vdia\nfin\n"})
	r := c.Resultado()
	if r.Exito || r.Encabezado != diag.EncabezadoSinValor || len(r.Diagnosticos) != 1 {
		t.Fatalf("resultado = %+v", r)
	}
	d := r.Diagnosticos[0]
	if d.Leccion.Titulo == "" || d.Fix == nil {
		t.Errorf("el diagnóstico debe traer lección y corrección: %+v", d)
	}
	datos, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	s := string(datos)
	for _, campo := range []string{`"exito":false`, `"code":"nombre-no-declarado"`, `"leccion":{"titulo":`, `"fix":{"line":4`, `"simbolos":{"principal.pks":[`} {
		if !strings.Contains(s, campo) {
			t.Errorf("el JSON no tiene %s:\n%s", campo, s)
		}
	}
	// Aun con errores de sintaxis (sin tabla) se sugieren los datos locales.
	c = compilarMapa(t, map[string]string{"principal.pks": "combate\n    roca vida = 1\n    gritar vida +\nfin\n"})
	if c.Tabla != nil || buscarSimbolo(c.Simbolos("principal.pks"), "vida", SimboloDato) == nil {
		t.Errorf("sin tabla deben quedar los locales: %+v", c.Simbolos("principal.pks"))
	}
}
