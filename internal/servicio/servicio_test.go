// Las palabras mal escritas de estas pruebas son a propósito: son los
// errores que el asistente debe corregir.
// cspell:words combte curra Estdo gritra nivle Pokemno vdia verdadro calcular_dan

package servicio

import (
	"context"
	"errors"
	"math/rand/v2"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/keylorpineda/PokeScript/internal/interprete"
)

func compilarMapa(t *testing.T, archivos map[string]string) Compilacion {
	t.Helper()
	fsys := fstest.MapFS{}
	for n, c := range archivos {
		fsys[n] = &fstest.MapFile{Data: []byte(c)}
	}
	c, err := CompilarFS(fsys, "prueba")
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCompilarYEjecutarSeccion10(t *testing.T) {
	c, err := Compilar(filepath.Join("..", "..", "ejemplos", "combate"))
	if err != nil {
		t.Fatal(err)
	}
	if c.TieneErrores() || c.Tabla == nil {
		t.Fatalf("diagnósticos: %+v", c.Diagnosticos)
	}
	es := interprete.NuevaESMemoria("Pikachu")
	in := interprete.Nuevo(es)
	in.Azar = rand.New(rand.NewPCG(1, 2))
	if err := Ejecutar(context.Background(), c, in); err != nil {
		t.Fatal(err)
	}
	salida := strings.Join(es.Salida, "\n")
	if !strings.Contains(salida, "¡Pikachu entra en combate!") {
		t.Errorf("salida inesperada:\n%s", salida)
	}
}

func TestCompilarArchivoSuelto(t *testing.T) {
	c, err := Compilar(filepath.Join("..", "..", "ejemplos", "hola.pks"))
	if err != nil || c.TieneErrores() {
		t.Fatalf("err = %v, diagnósticos = %+v", err, c.Diagnosticos)
	}
	if _, err := Compilar(filepath.Join(t.TempDir(), "no-existe.pks")); err == nil {
		t.Error("un archivo inexistente debe dar error")
	}
}

func TestLosDiagnosticosLleganConSugerencias(t *testing.T) {
	c := compilarMapa(t, map[string]string{
		"principal.pks":   "enseñar calcular_dano desde \"operaciones.pks\"\ncombate\n    roca vida = 10\n    gritar calcular_dan(vdia)\nfin\n",
		"operaciones.pks": "movimiento roca calcular_dano(roca poder)\n    entregar poder\nfin\n",
	})
	var fixes []string
	for _, d := range c.Diagnosticos {
		if d.Fix != nil {
			fixes = append(fixes, d.Fix.Replacement)
		}
	}
	// Un nombre importado y un dato local, cada uno con su corrección.
	if strings.Join(fixes, ",") != "calcular_dano,vida" {
		t.Errorf("correcciones = %v, diagnósticos %+v", fixes, c.Diagnosticos)
	}
}

func TestSugerenciasSinAnalizador(t *testing.T) {
	// Con un error de importación el analizador no corre, pero el asistente
	// igual corrige el nombre del archivo.
	c := compilarMapa(t, map[string]string{
		"principal.pks": "enseñar f desde \"util.pk\"\ncombate\nfin\n",
		"util.pks":      "movimiento f()\nfin\n",
	})
	if c.Tabla != nil {
		t.Error("con errores de importación el analizador no debe correr")
	}
	if len(c.Diagnosticos) != 1 || c.Diagnosticos[0].Fix == nil || c.Diagnosticos[0].Fix.Replacement != `"util.pks"` {
		t.Errorf("diagnósticos = %+v", c.Diagnosticos)
	}
}

func TestNoEjecutaConErrores(t *testing.T) {
	c := compilarMapa(t, map[string]string{"principal.pks": "combate\n    gritar x\nfin\n"})
	err := Ejecutar(context.Background(), c, interprete.Nuevo(interprete.NuevaESMemoria()))
	if !errors.Is(err, ErrHayErrores) {
		t.Errorf("err = %v, want ErrHayErrores", err)
	}
}

func TestCompilarFSInvalido(t *testing.T) {
	if _, err := CompilarFS(fstest.MapFS{}, "vacío"); err == nil {
		t.Error("un proyecto sin archivo principal debe dar error")
	}
}

func TestSugerenciaDeTipo(t *testing.T) {
	c := compilarMapa(t, map[string]string{
		"principal.pks": "especie Estado\n    SANO\nfin\nficha Pokemon\n    roca vida\nfin\ncombate\n    Estdo e = SANO\n    gritar e\nfin\n",
	})
	if len(c.Diagnosticos) != 1 || c.Diagnosticos[0].Fix == nil || c.Diagnosticos[0].Fix.Replacement != "Estado" {
		t.Errorf("diagnósticos = %+v", c.Diagnosticos)
	}
}
