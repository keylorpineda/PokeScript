package analizador

import (
	"testing"
	"testing/fstest"

	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/proyecto"
)

// cargarProyecto arma un proyecto en memoria. Falla si tiene errores de
// sintaxis o de importación, para que cada prueba mida solo el analizador.
func cargarProyecto(t *testing.T, archivos map[string]string) *proyecto.Proyecto {
	t.Helper()
	fsys := fstest.MapFS{}
	for n, c := range archivos {
		fsys[n] = &fstest.MapFile{Data: []byte(c)}
	}
	cargado, err := proyecto.CargarFS(fsys, "prueba")
	if err != nil {
		t.Fatal(err)
	}
	if cargado.TieneErrores() {
		d := cargado.Diagnosticos[0]
		t.Fatalf("el proyecto de prueba tiene errores previos: %s %d:%d %s: %s", d.File, d.Line, d.Col, d.Code, d.Desc)
	}
	return cargado.Proyecto
}

// recolectar corre solo la pasada 1.
func recolectar(t *testing.T, archivos map[string]string) Resultado {
	t.Helper()
	return Recolectar(cargarProyecto(t, archivos))
}

// analizar corre las dos pasadas. Es la ayuda que usan las pruebas de las
// verificaciones:
//
//	r := analizar(t, map[string]string{"principal.pks": "combate\n    …\nfin\n"})
//	if got := codigos(r.Diagnosticos); …
func analizar(t *testing.T, archivos map[string]string) Resultado {
	t.Helper()
	return Analizar(cargarProyecto(t, archivos))
}

// programa arma un proyecto de un solo archivo con el cuerpo dentro de combate.
func programa(cuerpo string) map[string]string {
	return map[string]string{"principal.pks": "combate\n" + cuerpo + "\nfin\n"}
}

func codigos(ds []diag.Diagnostic) []string {
	var cs []string
	for _, d := range ds {
		cs = append(cs, d.Code)
	}
	return cs
}
