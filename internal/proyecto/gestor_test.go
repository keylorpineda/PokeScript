package proyecto

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCrearYCargar(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "mi_proyecto")
	info, err := Crear(dir, "Mi primer proyecto")
	if err != nil {
		t.Fatal(err)
	}
	want := Info{Nombre: "Mi primer proyecto", Principal: "principal.pks", Archivos: []string{"principal.pks"}}
	if !reflect.DeepEqual(info, want) {
		t.Errorf("Info = %+v, want %+v", info, want)
	}
	// El proyecto nuevo compila tal cual: sin errores de ningún tipo.
	r, err := Cargar(dir)
	if err != nil || len(r.Diagnosticos) != 0 {
		t.Errorf("err = %v, diagnósticos = %v", err, codigos(r.Diagnosticos))
	}
	if _, err := Crear(dir, "otro"); err == nil || !strings.Contains(err.Error(), "no está vacía") {
		t.Errorf("crear sobre una carpeta con archivos: %v", err)
	}
	if _, err := Crear(filepath.Join(t.TempDir(), "x"), "  "); err == nil {
		t.Error("un proyecto sin nombre debe dar error")
	}
}

func TestArchivos(t *testing.T) {
	dir := t.TempDir()
	if _, err := Crear(dir, "p"); err != nil {
		t.Fatal(err)
	}
	if err := NuevoArchivo(dir, "tipos.pks"); err != nil {
		t.Fatal(err)
	}
	if err := NuevoArchivo(dir, "tipos.pks"); err == nil {
		t.Error("crear un archivo que ya existe debe dar error")
	}
	if err := Guardar(dir, "tipos.pks", "especie Clima\n    SOL\nfin\n"); err != nil {
		t.Fatal(err)
	}
	if contenido, err := LeerArchivo(dir, "tipos.pks"); err != nil || !strings.HasPrefix(contenido, "especie Clima") {
		t.Errorf("LeerArchivo = %q, %v", contenido, err)
	}

	if err := Renombrar(dir, "tipos.pks", "clima.pks"); err != nil {
		t.Fatal(err)
	}
	if err := Renombrar(dir, "clima.pks", "principal.pks"); err == nil {
		t.Error("renombrar sobre un archivo existente debe dar error")
	}
	info, _ := Leer(dir)
	if !reflect.DeepEqual(info.Archivos, []string{"clima.pks", "principal.pks"}) {
		t.Errorf("Archivos = %v", info.Archivos)
	}

	if err := Borrar(dir, "principal.pks"); err == nil || !strings.Contains(err.Error(), "archivo principal") {
		t.Errorf("borrar el principal: %v", err)
	}
	if err := Borrar(dir, "clima.pks"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "clima.pks")); !os.IsNotExist(err) {
		t.Error("clima.pks debería estar borrado")
	}
}

func TestPrincipal(t *testing.T) {
	dir := t.TempDir()
	if _, err := Crear(dir, "p"); err != nil {
		t.Fatal(err)
	}
	if err := NuevoArchivo(dir, "juego.pks"); err != nil {
		t.Fatal(err)
	}
	if err := MarcarPrincipal(dir, "juego.pks"); err != nil {
		t.Fatal(err)
	}
	if info, _ := Leer(dir); info.Principal != "juego.pks" {
		t.Errorf("Principal = %q", info.Principal)
	}
	// Renombrar el principal actualiza proyecto.json.
	if err := Renombrar(dir, "juego.pks", "inicio.pks"); err != nil {
		t.Fatal(err)
	}
	if info, _ := Leer(dir); info.Principal != "inicio.pks" {
		t.Errorf("después de renombrar, Principal = %q", info.Principal)
	}
	if err := MarcarPrincipal(dir, "no-existe.pks"); err == nil {
		t.Error("marcar un archivo inexistente debe dar error")
	}
}

func TestNombresInvalidos(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"", " a.pks", "../fuera.pks", `sub\x.pks`, "sub/x.pks", "notas.txt", ".pks", "sin_extension"} {
		if err := Guardar(dir, n, "x"); err == nil {
			t.Errorf("Guardar(%q) debe dar error", n)
		}
		if _, err := LeerArchivo(dir, n); err == nil {
			t.Errorf("LeerArchivo(%q) debe dar error", n)
		}
	}
	// Nada se escribió fuera de la carpeta.
	if _, err := os.Stat(filepath.Join(filepath.Dir(dir), "fuera.pks")); !os.IsNotExist(err) {
		t.Error("se escribió un archivo fuera del proyecto")
	}
}

func TestLeerSinProyectoJSON(t *testing.T) {
	dir := t.TempDir()
	if err := Guardar(dir, "principal.pks", "combate\nfin\n"); err != nil {
		t.Fatal(err)
	}
	info, err := Leer(dir)
	if err != nil || info.Principal != "principal.pks" || info.Nombre != filepath.Base(dir) {
		t.Errorf("Info = %+v, err = %v", info, err)
	}
	if err := os.WriteFile(filepath.Join(dir, ArchivoProyecto), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Leer(dir); err == nil {
		t.Error("un proyecto.json inválido debe dar error")
	}
}
