package proyecto

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/keylorpineda/PokeScript/internal/diag"
)

func mapa(archivos map[string]string) fstest.MapFS {
	fsys := fstest.MapFS{}
	for n, contenido := range archivos {
		fsys[n] = &fstest.MapFile{Data: []byte(contenido)}
	}
	return fsys
}

func cargarMapa(t *testing.T, archivos map[string]string) Resultado {
	t.Helper()
	r, err := CargarFS(mapa(archivos), "prueba")
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func codigos(ds []diag.Diagnostic) []string {
	var cs []string
	for _, d := range ds {
		cs = append(cs, d.Code)
	}
	return cs
}

func TestProyectoSeccion10(t *testing.T) {
	r, err := Cargar(filepath.Join("..", "..", "ejemplos", "combate"))
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range r.Diagnosticos {
		t.Errorf("%s %d:%d %s: %s", d.File, d.Line, d.Col, d.Code, d.Desc)
	}
	p := r.Proyecto
	if p.Nombre != "Combate por turnos" || p.Principal != "principal.pks" {
		t.Errorf("metadatos = %q, %q", p.Nombre, p.Principal)
	}
	if want := []string{"constantes.pks", "operaciones.pks", "tipos.pks", "principal.pks"}; !reflect.DeepEqual(p.Orden, want) {
		t.Errorf("Orden = %v, want %v", p.Orden, want)
	}
	if want := []string{"operaciones.pks", "tipos.pks", "constantes.pks"}; !reflect.DeepEqual(p.Archivos["principal.pks"].Dependencias, want) {
		t.Errorf("Dependencias del principal = %v", p.Archivos["principal.pks"].Dependencias)
	}
}

func TestSinProyectoJSONUsaPrincipalPorDefecto(t *testing.T) {
	r := cargarMapa(t, map[string]string{"principal.pks": "combate\nfin\n"})
	if r.Proyecto.Principal != "principal.pks" || r.Proyecto.Nombre != "prueba" || len(r.Diagnosticos) != 0 {
		t.Errorf("proyecto = %+v, diagnósticos %v", r.Proyecto, codigos(r.Diagnosticos))
	}
}

func TestProyectoJSON(t *testing.T) {
	r := cargarMapa(t, map[string]string{
		"proyecto.json": `{"nombre": "Mi proyecto", "principal": "juego.pks"}`,
		"juego.pks":     "combate\nfin\n",
	})
	if r.Proyecto.Nombre != "Mi proyecto" || r.Proyecto.Principal != "juego.pks" {
		t.Errorf("metadatos = %q, %q", r.Proyecto.Nombre, r.Proyecto.Principal)
	}
}

func TestErroresDeProyecto(t *testing.T) {
	casos := []struct {
		nombre   string
		archivos map[string]string
		contiene string
	}{
		{"json inválido", map[string]string{"proyecto.json": "{", "principal.pks": ""}, "no es válido"},
		{"json sin principal", map[string]string{"proyecto.json": `{"nombre": "x"}`}, "no dice cuál es el archivo principal"},
		{"principal inexistente", map[string]string{"otro.pks": ""}, "no se encontró el archivo principal"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := CargarFS(mapa(c.archivos), "prueba")
			if err == nil || !strings.Contains(err.Error(), c.contiene) {
				t.Errorf("err = %v, want que contenga %q", err, c.contiene)
			}
		})
	}
	if _, err := Cargar(filepath.Join(t.TempDir(), "no-existe")); err == nil {
		t.Error("una carpeta inexistente debe dar error")
	}
}

func TestErroresDeImportacion(t *testing.T) {
	tipos := "especie Estado\n    SANO, DORMIDO\nfin\nficha Pokemon\n    roca vida\nfin\n"
	casos := []struct {
		nombre    string
		principal string
		extra     map[string]string
		codigos   []string
		linea     int
		col       int
		contiene  string
	}{
		// Casos de importación 1 y 2 de la sección 11.
		{"archivo inexistente", `enseñar x desde "tipo.pks"` + "\ncombate\nfin", map[string]string{"tipos.pks": tipos},
			[]string{"archivo-inexistente"}, 1, 17, "tipos.pks"},
		{"nombre inexistente", `enseñar Entrenador desde "tipos.pks"` + "\ncombate\nfin", map[string]string{"tipos.pks": tipos},
			[]string{"nombre-inexistente"}, 1, 9, "«Estado» y «Pokemon»"},
		{"valor de especie en vez de la especie", `enseñar SANO desde "tipos.pks"` + "\ncombate\nfin", map[string]string{"tipos.pks": tipos},
			[]string{"nombre-inexistente"}, 1, 9, "importa «Estado»"},
		{"ruta con carpeta", `enseñar x desde "sub/tipos.pks"` + "\ncombate\nfin", nil,
			[]string{"ruta-invalida"}, 1, 17, "solo el nombre"},
		{"colisión con declaración local", `enseñar Estado desde "tipos.pks"` + "\nespecie Estado\n    A\nfin\ncombate\nfin", map[string]string{"tipos.pks": tipos},
			[]string{"colision-con-importacion"}, 1, 9, "cambia el nombre"},
		{"importación repetida", `enseñar Estado, Estado desde "tipos.pks"` + "\ncombate\nfin", map[string]string{"tipos.pks": tipos},
			[]string{"importacion-repetida"}, 1, 17, "borra"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			archivos := map[string]string{"principal.pks": c.principal}
			for n, f := range c.extra {
				archivos[n] = f
			}
			r := cargarMapa(t, archivos)
			if got := codigos(r.Diagnosticos); !reflect.DeepEqual(got, c.codigos) {
				t.Fatalf("códigos = %v, want %v", got, c.codigos)
			}
			d := r.Diagnosticos[0]
			if d.File != "principal.pks" || d.Line != c.linea || d.Col != c.col {
				t.Errorf("posición = %s %d:%d, want principal.pks %d:%d", d.File, d.Line, d.Col, c.linea, c.col)
			}
			if d.Category != diag.Importacion || d.Heading != diag.EncabezadoImportacion {
				t.Errorf("clasificación = %s / %q", d.Category, d.Heading)
			}
			if !strings.Contains(d.Desc+" "+d.Suggest, c.contiene) {
				t.Errorf("el mensaje no contiene %q:\n%s\n%s", c.contiene, d.Desc, d.Suggest)
			}
		})
	}
}

func TestImportacionCircular(t *testing.T) {
	// Caso de importación 4 de la sección 11: se reporta la cadena completa.
	r := cargarMapa(t, map[string]string{
		"principal.pks": "enseñar f desde \"a.pks\"\ncombate\nfin\n",
		"a.pks":         "enseñar g desde \"b.pks\"\nmovimiento f()\nfin\n",
		"b.pks":         "enseñar f desde \"a.pks\"\nmovimiento g()\nfin\n",
	})
	if got := codigos(r.Diagnosticos); !reflect.DeepEqual(got, []string{"importacion-circular"}) {
		t.Fatalf("códigos = %v", got)
	}
	d := r.Diagnosticos[0]
	if !strings.Contains(d.Desc, "a.pks → b.pks → a.pks") {
		t.Errorf("la descripción debe tener la cadena completa: %q", d.Desc)
	}
	if d.File != "b.pks" || d.Line != 1 || d.Col != 17 {
		t.Errorf("el error debe señalar la importación que cierra el ciclo; señala %s %d:%d", d.File, d.Line, d.Col)
	}
	// Aunque haya un ciclo, todos los archivos quedan en el orden una vez.
	if len(r.Proyecto.Orden) != 3 {
		t.Errorf("Orden = %v", r.Proyecto.Orden)
	}
}

func TestSeImportaASiMismo(t *testing.T) {
	r := cargarMapa(t, map[string]string{"principal.pks": "enseñar f desde \"principal.pks\"\nmovimiento f()\nfin\ncombate\nfin\n"})
	if got := codigos(r.Diagnosticos); !reflect.DeepEqual(got, []string{"importacion-circular"}) {
		t.Fatalf("códigos = %v", got)
	}
	if !strings.Contains(r.Diagnosticos[0].Desc, "a sí mismo") {
		t.Errorf("Desc = %q", r.Diagnosticos[0].Desc)
	}
}

func TestArchivoNoAlcanzableSeLeePeroNoSeOrdena(t *testing.T) {
	r := cargarMapa(t, map[string]string{
		"principal.pks": "combate\nfin\n",
		"suelto.pks":    "combate\n    gritar\nfin\n",
	})
	if _, ok := r.Proyecto.Archivos["suelto.pks"]; !ok {
		t.Error("todos los .pks de la carpeta deben quedar en Archivos")
	}
	if !reflect.DeepEqual(r.Proyecto.Orden, []string{"principal.pks"}) {
		t.Errorf("Orden = %v", r.Proyecto.Orden)
	}
	// Los errores de sintaxis de todos los archivos se reportan.
	if got := codigos(r.Diagnosticos); len(got) != 1 || r.Diagnosticos[0].File != "suelto.pks" {
		t.Errorf("diagnósticos = %v", got)
	}
	if !r.TieneErrores() {
		t.Error("TieneErrores debe ser true")
	}
}

func TestCargarArchivo(t *testing.T) {
	r, err := CargarArchivo(filepath.Join("..", "..", "ejemplos", "hola.pks"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Proyecto.Principal != "hola.pks" || !reflect.DeepEqual(r.Proyecto.Orden, []string{"hola.pks"}) || len(r.Diagnosticos) != 0 {
		t.Errorf("proyecto = %+v, diagnósticos %v", r.Proyecto, codigos(r.Diagnosticos))
	}
	if _, err := CargarArchivo(filepath.Join(t.TempDir(), "no-existe.pks")); err == nil {
		t.Error("un archivo inexistente debe dar error")
	}
}

func TestCargarArchivoEnCarpetaActual(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "solo.pks"), []byte("combate\nfin\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	antes, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(antes) })
	r, err := CargarArchivo("solo.pks")
	if err != nil || r.Proyecto.Principal != "solo.pks" {
		t.Errorf("r = %+v, err = %v", r.Proyecto, err)
	}
}

func TestArchivoSueltoSoloLeeLoQueImporta(t *testing.T) {
	dir := t.TempDir()
	escribir := func(n, contenido string) {
		if err := os.WriteFile(filepath.Join(dir, n), []byte(contenido), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	escribir("juego.pks", "enseñar f desde \"util.pks\"\ncombate\n    f()\nfin\n")
	escribir("util.pks", "movimiento f()\n    gritar 1\nfin\n")
	escribir("roto.pks", "combate\n    gritar\n")

	r, err := CargarArchivo(filepath.Join(dir, "juego.pks"))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Diagnosticos) != 0 {
		t.Errorf("los errores de roto.pks no deben aparecer: %v", codigos(r.Diagnosticos))
	}
	if _, cargado := r.Proyecto.Archivos["roto.pks"]; cargado {
		t.Error("roto.pks no se importa, así que no debe leerse")
	}
	if want := []string{"util.pks", "juego.pks"}; !reflect.DeepEqual(r.Proyecto.Orden, want) {
		t.Errorf("Orden = %v, want %v", r.Proyecto.Orden, want)
	}
}
