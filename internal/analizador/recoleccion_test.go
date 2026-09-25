package analizador

import (
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/proyecto"
	"github.com/keylorpineda/PokeScript/internal/tipos"
)

// recolectar carga un proyecto en memoria y corre la pasada 1. Falla si
// el proyecto tiene errores de sintaxis o de importación, para que cada
// prueba mida solo la recolección.
func recolectar(t *testing.T, archivos map[string]string) Resultado {
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
		t.Fatalf("el proyecto de prueba tiene errores previos: %v", codigos(cargado.Diagnosticos))
	}
	return Recolectar(cargado.Proyecto)
}

func codigos(ds []diag.Diagnostic) []string {
	var cs []string
	for _, d := range ds {
		cs = append(cs, d.Code)
	}
	return cs
}

func TestTablaSeccion10(t *testing.T) {
	cargado, err := proyecto.Cargar(filepath.Join("..", "..", "ejemplos", "combate"))
	if err != nil {
		t.Fatal(err)
	}
	r := Recolectar(cargado.Proyecto)
	for _, d := range r.Diagnosticos {
		t.Errorf("%s %d:%d %s: %s", d.File, d.Line, d.Col, d.Code, d.Desc)
	}
	tb := r.Tabla

	estado := tb.Especies["Estado"]
	if estado == nil || estado.Archivo != "tipos.pks" || !reflect.DeepEqual(estado.Valores, []string{"SANO", "ENVENENADO", "DORMIDO", "PARALIZADO"}) {
		t.Errorf("especie Estado = %+v", estado)
	}
	if tb.Valores["DORMIDO"] != estado {
		t.Error("DORMIDO debe apuntar a la especie Estado")
	}

	pokemon := tb.Fichas["Pokemon"]
	if pokemon == nil || len(pokemon.Campos) != 3 {
		t.Fatalf("ficha Pokemon = %+v", pokemon)
	}
	if c := pokemon.Campo("estado"); c == nil || !tipos.Iguales(c.Tipo, tipos.Especie("Estado")) {
		t.Errorf("campo estado = %+v", c)
	}
	if pokemon.Campo("ataque") != nil {
		t.Error("Campo de un nombre inexistente debe ser nil")
	}

	if m := tb.Medallas["NIVEL"]; m == nil || !tipos.Iguales(m.Tipo, tipos.Roca()) || m.Archivo != "constantes.pks" {
		t.Errorf("medalla NIVEL = %+v", m)
	}

	dano := tb.Movimientos["calcular_dano"]
	if dano == nil || !tipos.Iguales(dano.Retorno, tipos.Roca()) || len(dano.Params) != 1 || !tipos.Iguales(dano.Params[0].Tipo, tipos.Roca()) {
		t.Errorf("movimiento calcular_dano = %+v", dano)
	}
	describir := tb.Movimientos["describir"]
	if describir == nil || describir.Retorno != nil || !tipos.Iguales(describir.Params[0].Tipo, tipos.Especie("Estado")) {
		t.Errorf("movimiento describir = %+v", describir)
	}
}

func TestVisibilidad(t *testing.T) {
	cargado, err := proyecto.Cargar(filepath.Join("..", "..", "ejemplos", "combate"))
	if err != nil {
		t.Fatal(err)
	}
	tb := Recolectar(cargado.Proyecto).Tabla
	casos := []struct {
		archivo, nombre string
		clase           Clase
		visible         bool
	}{
		{"principal.pks", "calcular_dano", ClaseMovimiento, true}, // importado
		{"principal.pks", "Pokemon", ClaseFicha, true},
		{"principal.pks", "SANO", ClaseValorEspecie, true}, // visible porque se importa Estado
		{"principal.pks", "VIDA_MAXIMA", ClaseMedalla, true},
		{"principal.pks", "describir", ClaseMovimiento, true}, // propio
		{"principal.pks", "NIVEL", 0, false},                  // existe en el proyecto, pero no se importa
		{"operaciones.pks", "NIVEL", ClaseMedalla, true},
		{"operaciones.pks", "Estado", 0, false},
		{"operaciones.pks", "SANO", 0, false}, // su especie no es visible aquí
		{"tipos.pks", "DORMIDO", ClaseValorEspecie, true},
		{"principal.pks", "Mewtwo", 0, false},
	}
	for _, c := range casos {
		s := tb.Buscar(c.archivo, c.nombre)
		if (s != nil) != c.visible {
			t.Errorf("Buscar(%s, %s) visible = %v, want %v", c.archivo, c.nombre, s != nil, c.visible)
			continue
		}
		if s != nil && s.Clase != c.clase {
			t.Errorf("Buscar(%s, %s).Clase = %v, want %v", c.archivo, c.nombre, s.Clase, c.clase)
		}
	}

	visibles := tb.Visibles("operaciones.pks")
	sort.Strings(visibles)
	if want := []string{"NIVEL", "calcular_dano"}; !reflect.DeepEqual(visibles, want) {
		t.Errorf("Visibles(operaciones.pks) = %v, want %v", visibles, want)
	}
}

func TestErroresDeRecoleccion(t *testing.T) {
	casos := []struct {
		nombre   string
		archivos map[string]string
		codigos  []string
		archivo  string
		linea    int
		contiene string
	}{
		{"movimiento duplicado en otro archivo",
			map[string]string{
				"principal.pks": "enseñar g desde \"a.pks\"\nmovimiento f()\nfin\ncombate\nfin\n",
				"a.pks":         "movimiento f()\nfin\nmovimiento g()\nfin\n",
			},
			[]string{"movimiento-duplicado"}, "principal.pks", 2, "está declarado en a.pks, línea 1"},
		{"movimiento duplicado",
			map[string]string{"principal.pks": "movimiento f()\nfin\nmovimiento f()\nfin\ncombate\nfin\n"},
			[]string{"movimiento-duplicado"}, "principal.pks", 3, "no hay sobrecarga"},
		{"nombre duplicado entre clases",
			map[string]string{"principal.pks": "medalla roca Estado = 1\nespecie Estado\n    A\nfin\ncombate\nfin\n"},
			[]string{"nombre-duplicado"}, "principal.pks", 2, "la medalla «Estado» está declarado en principal.pks, línea 1"},
		{"valor de especie repetido entre especies",
			map[string]string{"principal.pks": "especie Estado\n    SANO\nfin\nespecie Clima\n    SANO\nfin\ncombate\nfin\n"},
			[]string{"valor-de-especie-repetido"}, "principal.pks", 5, "SANO"},
		{"campo repetido",
			map[string]string{"principal.pks": "ficha P\n    roca vida\n    agua vida\nfin\ncombate\nfin\n"},
			[]string{"campo-repetido"}, "principal.pks", 3, "línea 2"},
		{"parámetro repetido",
			map[string]string{"principal.pks": "movimiento f(roca x, agua x)\nfin\ncombate\nfin\n"},
			[]string{"parametro-repetido"}, "principal.pks", 1, "«x»"},
		{"sin combate",
			map[string]string{"principal.pks": "movimiento f()\nfin\n"},
			[]string{"sin-combate"}, "principal.pks", 1, "no tiene un bloque combate"},
		{"combate repetido",
			map[string]string{"principal.pks": "combate\nfin\ncombate\nfin\n"},
			[]string{"combate-repetido"}, "principal.pks", 3, "más de un bloque combate"},
		{"combate fuera del principal",
			map[string]string{
				"principal.pks": "enseñar f desde \"a.pks\"\ncombate\nfin\n",
				"a.pks":         "movimiento f()\nfin\ncombate\nfin\n",
			},
			[]string{"combate-fuera-del-principal"}, "a.pks", 3, "«principal.pks»"},
		{"tipo desconocido",
			map[string]string{"principal.pks": "ficha P\n    Clima c\nfin\ncombate\nfin\n"},
			[]string{"tipo-desconocido"}, "principal.pks", 2, "declara la especie"},
		{"tipo no importado",
			map[string]string{
				"principal.pks": "enseñar f desde \"a.pks\"\nmovimiento g(Estado e)\nfin\ncombate\nfin\n",
				"a.pks":         "especie Estado\n    SANO\nfin\nmovimiento f()\nfin\n",
			},
			[]string{"tipo-desconocido"}, "principal.pks", 2, "enseñar Estado desde \"a.pks\""},
		{"posible no permitido",
			map[string]string{"principal.pks": "ficha P\n    roca x\nfin\nmovimiento f(posible P p)\nfin\ncombate\nfin\n"},
			[]string{"tipo-invalido"}, "principal.pks", 4, "posible solo se aplica"},
		{"clave de mochila ficha",
			map[string]string{"principal.pks": "ficha P\n    roca x\nfin\nmedalla mochila de P a roca M = {}\ncombate\nfin\n"},
			[]string{"tipo-invalido"}, "principal.pks", 4, "clave de una mochila"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			r := recolectar(t, c.archivos)
			if got := codigos(r.Diagnosticos); !reflect.DeepEqual(got, c.codigos) {
				t.Fatalf("códigos = %v, want %v", got, c.codigos)
			}
			d := r.Diagnosticos[0]
			if d.File != c.archivo || d.Line != c.linea {
				t.Errorf("posición = %s:%d, want %s:%d", d.File, d.Line, c.archivo, c.linea)
			}
			if d.Category != diag.Semantico || d.Heading == "" || d.Suggest == "" {
				t.Errorf("diagnóstico incompleto: %+v", d)
			}
			if !strings.Contains(d.Desc+" "+d.Cause+" "+d.Suggest, c.contiene) {
				t.Errorf("el mensaje no contiene %q:\n%s\n%s\n%s", c.contiene, d.Desc, d.Cause, d.Suggest)
			}
		})
	}
}

func TestResolverTipo(t *testing.T) {
	r := recolectar(t, map[string]string{
		"principal.pks": "especie Estado\n    SANO\nfin\nficha P\n    roca x\nfin\n" +
			"medalla equipo de mochila de Estado a equipo de P M = []\ncombate\nfin\n",
	})
	if len(r.Diagnosticos) > 0 {
		t.Fatalf("diagnósticos: %v", codigos(r.Diagnosticos))
	}
	want := tipos.Equipo(tipos.Mochila(tipos.Especie("Estado"), tipos.Equipo(tipos.Ficha("P"))))
	if got := r.Tabla.Medallas["M"].Tipo; !tipos.Iguales(got, want) {
		t.Errorf("tipo = %s, want %s", got, want)
	}
}
