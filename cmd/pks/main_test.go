package main

import (
	"bytes"
	"context"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/keylorpineda/PokeScript/internal/interprete"
)

// ejecutarPrueba corre el comando con una semilla fija para aleatorio.
func ejecutarPrueba(t *testing.T, ruta, entrada string) (codigo int, salida, errores string) {
	t.Helper()
	var out, errs bytes.Buffer
	in := interprete.Nuevo(nil)
	in.Azar = rand.New(rand.NewPCG(1, 2))
	codigo = correr(context.Background(), []string{ruta}, strings.NewReader(entrada), &out, &errs, in)
	return codigo, out.String(), errs.String()
}

func TestEjemplos(t *testing.T) {
	casos := []struct {
		ruta    string
		entrada string
		salida  string
	}{
		{"hola.pks", "", "¡Hola, mundo Pokémon!\n"},
		{"entrenamiento.pks", "", "Tu equipo tiene 4 Pokémon\n  Aranja x3\n  Meloc x5\nNivel final: 11\n"},
		{"estados.pks", "", "Turno 1: Bulbasaur tiene 35 PS\nTurno 2: Bulbasaur tiene 25 PS\nTurno 3: Bulbasaur tiene 15 PS\n"},
		{"centro.pks", "Mewtwo\n70\n", "¿Qué Pokémon atrapaste? ¿De qué nivel? ¡Increíble! Mewtwo es legendario\n"},
	}
	for _, c := range casos {
		t.Run(c.ruta, func(t *testing.T) {
			codigo, salida, errores := ejecutarPrueba(t, filepath.Join("..", "..", "ejemplos", c.ruta), c.entrada)
			if codigo != salidaOK || errores != "" {
				t.Fatalf("código %d, errores:\n%s", codigo, errores)
			}
			if salida != c.salida {
				t.Errorf("salida =\n%q\nwant\n%q", salida, c.salida)
			}
		})
	}
}

// TestProyectoSeccion10 corre el proyecto de cuatro archivos de la sección
// 10: el hito 4 de punta a punta, desde el texto hasta la salida.
func TestProyectoSeccion10(t *testing.T) {
	codigo, salida, errores := ejecutarPrueba(t, filepath.Join("..", "..", "ejemplos", "combate"), "Pikachu\n")
	if codigo != salidaOK || errores != "" {
		t.Fatalf("código %d, errores:\n%s", codigo, errores)
	}
	for _, parte := range []string{"¿Cómo se llama tu Pokemon? ¡Pikachu entra en combate!", "  Puede atacar", "--- Turno 1 ---"} {
		if !strings.Contains(salida, parte) {
			t.Errorf("la salida no contiene %q:\n%s", parte, salida)
		}
	}
	if !strings.Contains(salida, "¡Ganaste!") && !strings.Contains(salida, "Perdiste.") {
		t.Errorf("el combate debe terminar con un ganador:\n%s", salida)
	}
}

func TestErroresDeCompilacionYEjecucion(t *testing.T) {
	dir := t.TempDir()
	escribir := func(nombre, fuente string) string {
		ruta := filepath.Join(dir, nombre)
		if err := os.WriteFile(ruta, []byte(fuente), 0o600); err != nil {
			t.Fatal(err)
		}
		return ruta
	}

	codigo, salida, errores := ejecutarPrueba(t, escribir("sintaxis.pks", "combate\n    gritar 1 gritar 2\nfin\n"), "")
	if codigo != salidaCompilar || salida != "" || !strings.Contains(errores, "¡Se escapó!") {
		t.Errorf("sintaxis: código %d, salida %q, errores:\n%s", codigo, salida, errores)
	}

	codigo, _, errores = ejecutarPrueba(t, escribir("division.pks", "combate\n    roca x = 0\n    gritar 1 / x\nfin\n"), "")
	if codigo != salidaEjecucion || !strings.Contains(errores, "¡Falló el ataque!") {
		t.Errorf("ejecución: código %d, errores:\n%s", codigo, errores)
	}

	codigo, _, errores = ejecutarPrueba(t, escribir("entrada.pks", "combate\n    planta x\n    capturar(x, \"? \")\nfin\n"), "")
	if codigo != salidaInterrumpe || !strings.Contains(errores, "detenida") {
		t.Errorf("sin entrada: código %d, errores:\n%s", codigo, errores)
	}

	codigo, _, _ = ejecutarPrueba(t, filepath.Join(dir, "no-existe.pks"), "")
	if codigo != salidaUso {
		t.Errorf("archivo inexistente: código %d", codigo)
	}
}

func TestUso(t *testing.T) {
	var out, errs bytes.Buffer
	if c := correr(context.Background(), nil, strings.NewReader(""), &out, &errs, nil); c != salidaUso {
		t.Errorf("sin argumentos: código %d", c)
	}
	if !strings.Contains(errs.String(), "Uso:") {
		t.Errorf("falta el mensaje de uso: %q", errs.String())
	}
}

func TestProyectoJSONInvalido(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "proyecto.json"), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if c, _, errores := ejecutarPrueba(t, dir, ""); c != salidaUso || !strings.Contains(errores, "proyecto.json") {
		t.Errorf("código %d, errores %q", c, errores)
	}
}
