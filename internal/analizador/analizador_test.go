package analizador

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/proyecto"
)

// conVerificaciones reemplaza las verificaciones registradas mientras dura
// la prueba.
func conVerificaciones(t *testing.T, vs ...verificacion) {
	t.Helper()
	antes := verificaciones
	verificaciones = nil
	for _, v := range vs {
		registrar(v.nombre, v.revisar)
	}
	t.Cleanup(func() { verificaciones = antes })
}

func TestAnalizarSeccion10(t *testing.T) {
	cargado, err := proyecto.Cargar(filepath.Join("..", "..", "ejemplos", "combate"))
	if err != nil {
		t.Fatal(err)
	}
	r := Analizar(cargado.Proyecto)
	for _, d := range r.Diagnosticos {
		t.Errorf("%s %d:%d %s: %s", d.File, d.Line, d.Col, d.Code, d.Desc)
	}
	if r.Tabla == nil || r.Tabla.Movimientos["describir"] == nil {
		t.Error("Analizar debe devolver la tabla de la pasada 1")
	}
}

func TestVerificacionesEnOrdenYDiagnosticosOrdenados(t *testing.T) {
	var llamadas []string
	marcar := func(nombre string, linea int) verificacion {
		return verificacion{nombre: nombre, revisar: func(c *Contexto) {
			llamadas = append(llamadas, nombre)
			c.Error("principal.pks", ast.Pos{Line: linea, Col: 1, Len: 1}, nombre, diag.EncabezadoTipos, "d", "c", "s")
		}}
	}
	conVerificaciones(t, marcar("zeta", 1), marcar("alfa", 3), marcar("medio", 2))

	r := analizar(t, programa("    gritar 1"))
	if want := []string{"alfa", "medio", "zeta"}; !reflect.DeepEqual(llamadas, want) {
		t.Errorf("orden de las verificaciones = %v, want %v", llamadas, want)
	}
	if want := []string{"zeta", "medio", "alfa"}; !reflect.DeepEqual(codigos(r.Diagnosticos), want) {
		t.Errorf("los diagnósticos deben quedar por línea: %v", codigos(r.Diagnosticos))
	}
}

func TestContexto(t *testing.T) {
	conVerificaciones(t, verificacion{nombre: "prueba", revisar: func(c *Contexto) {
		var nombres []string
		for _, a := range c.Archivos() {
			nombres = append(nombres, a.Nombre)
		}
		if want := []string{"util.pks", "principal.pks"}; !reflect.DeepEqual(nombres, want) {
			t.Errorf("Archivos() = %v, want %v", nombres, want)
		}
		c.Advertencia("util.pks", ast.Pos{Line: 1, Col: 1}, "aviso", "d", "c", "s")
	}})
	r := analizar(t, map[string]string{
		"principal.pks": "enseñar f desde \"util.pks\"\ncombate\n    f()\nfin\n",
		"util.pks":      "movimiento f()\nfin\n",
	})
	if len(r.Diagnosticos) != 1 {
		t.Fatalf("diagnósticos = %v", codigos(r.Diagnosticos))
	}
	d := r.Diagnosticos[0]
	if d.Severity != diag.Advertencia || d.Heading != diag.EncabezadoAdvertencia || d.Len != 1 {
		t.Errorf("advertencia = %+v", d)
	}
}
