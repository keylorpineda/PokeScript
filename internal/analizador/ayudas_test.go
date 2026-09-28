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
//
// Deja afuera las advertencias de estilo 20 a 23 (advertencias.go): los
// fragmentos de prueba declaran datos solo para revisar otra regla, y avisar
// que no se usan sería ruido. Esas advertencias se prueban con
// analizarConAdvertencias.
func analizar(t *testing.T, archivos map[string]string) Resultado {
	t.Helper()
	r := analizarConAdvertencias(t, archivos)
	var ds []diag.Diagnostic
	for _, d := range r.Diagnosticos {
		if !advertenciasDeEstilo[d.Code] {
			ds = append(ds, d)
		}
	}
	r.Diagnosticos = ds
	return r
}

// analizarConAdvertencias corre las dos pasadas y devuelve todo.
func analizarConAdvertencias(t *testing.T, archivos map[string]string) Resultado {
	t.Helper()
	return Analizar(cargarProyecto(t, archivos))
}

var advertenciasDeEstilo = map[string]bool{
	"dato-sin-usar":         true,
	"ciclo-sin-cambio":      true,
	"sangria-inconsistente": true,
	"convencion-de-nombres": true,
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
