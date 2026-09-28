package explorador

import (
	"os"
	"path/filepath"
	"testing"
)

func crear(t *testing.T, ruta string, contenido string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ruta, []byte(contenido), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestListar(t *testing.T) {
	raiz := t.TempDir()
	crear(t, filepath.Join(raiz, "combate", "principal.pks"), "combate\nfin\n")
	crear(t, filepath.Join(raiz, "combate", "tipos.pks"), "")
	crear(t, filepath.Join(raiz, "solo_json", "proyecto.json"), "{}")
	crear(t, filepath.Join(raiz, "fotos", "pikachu.png"), "")
	crear(t, filepath.Join(raiz, ".oculta", "nada.pks"), "")
	crear(t, filepath.Join(raiz, "Suelto.PKS"), "")
	crear(t, filepath.Join(raiz, "notas.txt"), "")

	c, err := Listar(raiz)
	if err != nil {
		t.Fatal(err)
	}
	if c.Nombre != filepath.Base(raiz) || c.Padre != filepath.Dir(raiz) || !c.Proyecto {
		t.Errorf("carpeta = %+v", c)
	}
	casos := []Subcarpeta{
		{"combate", filepath.Join(raiz, "combate"), true, 2},
		{"fotos", filepath.Join(raiz, "fotos"), false, 0},
		{"solo_json", filepath.Join(raiz, "solo_json"), true, 0},
	}
	if len(c.Carpetas) != len(casos) {
		t.Fatalf("carpetas = %+v (la oculta no debe aparecer)", c.Carpetas)
	}
	for i, want := range casos {
		if c.Carpetas[i] != want {
			t.Errorf("carpeta %d = %+v, want %+v", i, c.Carpetas[i], want)
		}
	}
	wantArchivos := []Archivo{{"notas.txt", false}, {"Suelto.PKS", true}}
	if len(c.Archivos) != 2 || c.Archivos[0] != wantArchivos[0] || c.Archivos[1] != wantArchivos[1] {
		t.Errorf("archivos = %+v, want %+v", c.Archivos, wantArchivos)
	}

	if _, err := Listar(filepath.Join(raiz, "no-existe")); err == nil {
		t.Error("una carpeta inexistente debe dar error")
	}
}

func TestLugares(t *testing.T) {
	lugares := Lugares()
	if len(lugares) == 0 || lugares[0].Tipo != "personal" {
		t.Fatalf("lugares = %+v: el primero debe ser la carpeta personal", lugares)
	}
	for _, l := range lugares {
		if info, err := os.Stat(l.Ruta); err != nil || !info.IsDir() {
			t.Errorf("%s (%s) no es una carpeta", l.Nombre, l.Ruta)
		}
	}
}
