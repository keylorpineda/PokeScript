package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/keylorpineda/PokeScript/internal/servicio"
)

// TestCasosSeccion11 corre los casos de prueba prioritarios de la sección
// 11 de la especificación de punta a punta: desde los .pks hasta los
// diagnósticos, igual que cuando un estudiante corre su programa. Es el
// criterio para la etiqueta v0.2. La prueba completa (el programa de la
// sección 10) está en TestProyectoSeccion10.
func TestCasosSeccion11(t *testing.T) {
	casos := []struct {
		nombre   string
		archivos map[string]string // un solo archivo va como principal.pks
		codigo   string
		mensaje  string
	}{
		// Errores sintácticos (5).
		{"sintáctico 1: falta un fin",
			map[string]string{"principal.pks": "combate\n    si verdadero\n        gritar 1\n"},
			"bloque-sin-cerrar", "que abriste en la línea"},
		{"sintáctico 2: a > b > c",
			map[string]string{"principal.pks": "combate\n    roca p = 1\n    gritar p > 2 > 3\nfin\n"},
			"operador-no-encadenable", "no se puede encadenar"},
		{"sintáctico 3: dos instrucciones en una línea",
			map[string]string{"principal.pks": "combate\n    gritar 1 gritar 2\nfin\n"},
			"dos-instrucciones", "dos instrucciones en la misma línea"},
		{"sintáctico 4: cadena sin cerrar",
			map[string]string{"principal.pks": "combate\n    gritar \"hola\nfin\n"},
			"cadena-sin-cerrar", "comilla de cierre"},
		{"sintáctico 5: recorrer n de 1 a 10",
			map[string]string{"principal.pks": "combate\n    recorrer n de 1 a 10\n        gritar n\n    fin\nfin\n"},
			"rango-sin-hasta", "se usa «hasta»"},

		// Errores semánticos (5).
		{"semántico 1: roca + electrico",
			map[string]string{"principal.pks": "combate\n    roca x = 1 + verdadero\n    gritar x\nfin\n"},
			"tipo-incompatible", "tabla de efectividades"},
		{"semántico 2: segun sin cubrir un valor",
			map[string]string{"principal.pks": "especie Estado\n    SANO, DORMIDO\nfin\ncombate\n    Estado e = SANO\n    segun e\n        SANO entonces gritar 1\n    fin\nfin\n"},
			"segun-incompleto", "DORMIDO"},
		{"semántico 3: posible sin comprobar",
			map[string]string{"principal.pks": "combate\n    posible planta rival = fantasma\n    gritar rival + \"!\"\nfin\n"},
			"posible-sin-comprobar", "puede ser fantasma"},
		{"semántico 4: reasignación de medalla",
			map[string]string{"principal.pks": "combate\n    medalla roca TOPE = 5\n    TOPE = 6\nfin\n"},
			"medalla-reasignada", "es una medalla"},
		{"semántico 5: dato leído sin valor",
			map[string]string{"principal.pks": "combate\n    roca x\n    gritar x\nfin\n"},
			"dato-sin-valor", "antes de tener un valor"},

		// Importaciones (3 + 1 circular).
		{"importación 1: archivo inexistente",
			map[string]string{"principal.pks": "enseñar f desde \"util.pks\"\ncombate\nfin\n"},
			"archivo-inexistente", "no existe en el proyecto"},
		{"importación 2: nombre inexistente",
			map[string]string{
				"principal.pks": "enseñar g desde \"util.pks\"\ncombate\nfin\n",
				"util.pks":      "movimiento f()\nfin\n",
			},
			"nombre-inexistente", "no está declarado en «util.pks»"},
		{"importación 3: argumentos de tipo incorrecto",
			map[string]string{
				"principal.pks": "enseñar doble desde \"util.pks\"\ncombate\n    gritar doble(\"dos\")\nfin\n",
				"util.pks":      "movimiento roca doble(roca x)\n    entregar x * 2\nfin\n",
			},
			"argumento-incompatible", "doble"},
		{"importación 4: ciclo",
			map[string]string{
				"principal.pks": "enseñar f desde \"a.pks\"\ncombate\nfin\n",
				"a.pks":         "enseñar g desde \"b.pks\"\nmovimiento f()\nfin\n",
				"b.pks":         "enseñar f desde \"a.pks\"\nmovimiento g()\nfin\n",
			},
			"importacion-circular", "a.pks → b.pks → a.pks"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			dir := t.TempDir()
			for n, fuente := range c.archivos {
				if err := os.WriteFile(filepath.Join(dir, n), []byte(fuente), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			compilado, err := servicio.Compilar(dir)
			if err != nil {
				t.Fatal(err)
			}
			diags := compilado.Diagnosticos
			encontrado := false
			var vistos []string
			for _, d := range diags {
				vistos = append(vistos, d.Code)
				if d.Code == c.codigo && strings.Contains(d.Desc+" "+d.Cause+" "+d.Suggest, c.mensaje) {
					encontrado = true
				}
			}
			if !encontrado {
				t.Errorf("no se reportó %s con %q; se reportó %v", c.codigo, c.mensaje, vistos)
			}
			// Con cualquiera de estos errores, el programa no se ejecuta.
			codigo, salida, _ := ejecutarPrueba(t, dir, "")
			if codigo != salidaCompilar || salida != "" {
				t.Errorf("el programa no debería ejecutarse: código %d, salida %q", codigo, salida)
			}
		})
	}
}
