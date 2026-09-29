// Las palabras mal escritas de estas pruebas son a propósito: son los
// errores que el asistente debe corregir.
// cspell:words combte curra Estdo gritra nivle Pokemno recorerer vdia verdadro calcular_dan

package asistente_test

import (
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/keylorpineda/PokeScript/internal/analizador"
	"github.com/keylorpineda/PokeScript/internal/asistente"
	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/proyecto"
)

func TestDistancia(t *testing.T) {
	casos := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"curar", "curar", 0},
		{"curra", "curar", 2},
		{"vdia", "vida", 2},
		{"gritra", "gritar", 2},
		{"mientas", "mientras", 1},
		{"año", "ano", 1}, // cuenta letras, no bytes
		{"", "fin", 3},
		{"Vida", "vida", 1},
	}
	for _, c := range casos {
		if got := asistente.Distancia(c.a, c.b); got != c.want {
			t.Errorf("Distancia(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
		if got := asistente.Distancia(c.b, c.a); got != c.want {
			t.Errorf("Distancia no es simétrica para %q y %q", c.a, c.b)
		}
	}
}

func TestParecido(t *testing.T) {
	nombres := []string{"curar", "vida", "calcular_dano", "nivel", "x"}
	casos := []struct {
		palabra string
		want    string
		ok      bool
	}{
		{"curra", "curar", true},
		{"vdia", "vida", true},
		{"calcular_daño", "calcular_dano", true},
		{"calcularDano", "calcular_dano", true}, // 2 cambios en una palabra larga
		{"nivle", "nivel", true},
		{"y", "x", true},       // una letra: se permite distancia 1
		{"zz", "", false},      // dos letras: distancia 2 sería cambiarlo todo
		{"ataque", "", false},  // nada se parece
		{"curar", "", false},   // igual a un candidato: no hay nada que corregir
		{"ciruela", "", false}, // demasiado lejos de «curar»
	}
	for _, c := range casos {
		got, ok := asistente.Parecido(c.palabra, nombres)
		if got != c.want || ok != c.ok {
			t.Errorf("Parecido(%q) = %q, %v; want %q, %v", c.palabra, got, ok, c.want, c.ok)
		}
	}
	// Con empate gana el primero en orden alfabético, siempre el mismo.
	if got, _ := asistente.Parecido("pata", []string{"pala", "papa", "gata"}); got != "gata" {
		t.Errorf("empate = %q, want gata", got)
	}
}

// analizar carga un proyecto en memoria, lo analiza y arma la Entrada del
// asistente como lo hará el servicio de compilación.
func analizar(t *testing.T, archivos map[string]string) ([]diag.Diagnostic, asistente.Entrada) {
	t.Helper()
	fsys := fstest.MapFS{}
	for n, c := range archivos {
		fsys[n] = &fstest.MapFile{Data: []byte(c)}
	}
	r, err := proyecto.CargarFS(fsys, "prueba")
	if err != nil {
		t.Fatal(err)
	}
	diags := r.Diagnosticos
	var tabla *analizador.Tabla
	if !r.TieneErrores() {
		a := analizador.Analizar(r.Proyecto)
		diags, tabla = append(diags, a.Diagnosticos...), a.Tabla
	}
	e := asistente.Entrada{
		Fuentes: archivos,
		Nombres: func(archivo string) []string {
			var ns []string
			if tabla != nil {
				ns = tabla.Visibles(archivo)
			}
			return append(ns, "vida", "nombre") // locales del programa de prueba
		},
		Tipos:      func(string) []string { return []string{"Estado", "Pokemon"} },
		Exportados: func(string) []string { return []string{"calcular_dano", "NIVEL"} },
		Archivos:   func() []string { return []string{"principal.pks", "tipos.pks", "operaciones.pks"} },
	}
	return diags, e
}

func TestAsistir(t *testing.T) {
	casos := []struct {
		nombre   string
		archivos map[string]string
		codigo   string
		fix      diag.Fix
		pregunta string
	}{
		{"nombre de movimiento mal escrito",
			map[string]string{"principal.pks": "movimiento roca curar(roca x)\n    entregar x + 20\nfin\ncombate\n    gritar curra(35)\nfin\n"},
			"nombre-no-declarado", diag.Fix{Line: 5, Col: 12, Len: 5, Replacement: "curar"}, "¿Quisiste decir «curar»?"},
		{"dato local mal escrito",
			map[string]string{"principal.pks": "combate\n    roca vida = 10\n    gritar vdia\nfin\n"},
			"nombre-no-declarado", diag.Fix{Line: 3, Col: 12, Len: 4, Replacement: "vida"}, "«vida»"},
		{"tipo mal escrito",
			map[string]string{"principal.pks": "combate\n    Pokemno p\nfin\n"},
			"tipo-desconocido", diag.Fix{Line: 2, Col: 5, Len: 7, Replacement: "Pokemon"}, "«Pokemon»"},
		{"archivo mal escrito",
			map[string]string{"principal.pks": "enseñar x desde \"tipo.pks\"\ncombate\nfin\n"},
			"archivo-inexistente", diag.Fix{Line: 1, Col: 17, Len: 10, Replacement: `"tipos.pks"`}, "«tipos.pks»"},
		{"nombre importado mal escrito",
			map[string]string{
				"principal.pks":   "enseñar calcular_dan desde \"operaciones.pks\"\ncombate\nfin\n",
				"operaciones.pks": "movimiento roca calcular_dano()\n    entregar 1\nfin\n",
			},
			"nombre-inexistente", diag.Fix{Line: 1, Col: 9, Len: 12, Replacement: "calcular_dano"}, "«calcular_dano»"},
		{"palabra reservada mal escrita en un bloque",
			map[string]string{"principal.pks": "combate\n    gritra \"hola\"\nfin\n"},
			"asignacion-esperada", diag.Fix{Line: 2, Col: 5, Len: 6, Replacement: "gritar"}, "«gritar»"},
		{"palabra reservada mal escrita fuera de un bloque",
			map[string]string{"principal.pks": "combte\n    gritar 1\nfin\n"},
			"instruccion-fuera-de-bloque", diag.Fix{Line: 1, Col: 1, Len: 6, Replacement: "combate"}, "«combate»"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			diags, e := analizar(t, c.archivos)
			asistidos := asistente.Asistir(diags, e)
			for i, d := range asistidos {
				if d.Code != c.codigo {
					continue
				}
				if d.Fix == nil || *d.Fix != c.fix {
					t.Fatalf("Fix = %+v, want %+v", d.Fix, c.fix)
				}
				if !strings.HasPrefix(d.Suggest, "¿Quisiste decir") || !strings.Contains(d.Suggest, c.pregunta) {
					t.Errorf("Suggest = %q", d.Suggest)
				}
				if diags[i].Fix != nil {
					t.Error("Asistir no debe modificar los diagnósticos originales")
				}
				return
			}
			t.Fatalf("no apareció %s; códigos %v", c.codigo, codigos(asistidos))
		})
	}
}

func TestAsistirSinSugerencia(t *testing.T) {
	// Nada se parece a «ataque», y «vida + 1» empieza con un nombre válido:
	// en ninguno de los dos casos se inventa una corrección.
	for _, fuente := range []string{
		"combate\n    gritar ataque\nfin\n",
		"combate\n    roca vida = 1\n    vida + 1\nfin\n",
	} {
		diags, e := analizar(t, map[string]string{"principal.pks": fuente})
		for _, d := range asistente.Asistir(diags, e) {
			if d.Fix != nil || strings.HasPrefix(d.Suggest, "¿Quisiste decir") {
				t.Errorf("%q: sugerencia inventada %+v", fuente, d)
			}
		}
	}
	if got := asistente.Asistir(nil, asistente.Entrada{}); len(got) != 0 {
		t.Error("sin diagnósticos no hay nada que asistir")
	}
}

func codigos(ds []diag.Diagnostic) []string {
	var cs []string
	for _, d := range ds {
		cs = append(cs, d.Code)
	}
	return cs
}

// TestLosEjemplosDeLasLeccionesCompilan garantiza que ninguna lección
// enseñe un programa que no funciona.
func TestLosEjemplosDeLasLeccionesCompilan(t *testing.T) {
	for _, codigo := range asistente.CodigosConLeccion() {
		l := asistente.Explicar(diag.Diagnostic{Code: codigo})
		if l.Titulo == "" || l.Texto == "" {
			t.Errorf("%s: lección sin título o sin texto", codigo)
		}
		if l.Ejemplo == "" {
			continue
		}
		t.Run(codigo, func(t *testing.T) {
			diags, _ := analizar(t, map[string]string{"principal.pks": l.Ejemplo})
			for _, d := range diags {
				t.Errorf("%d:%d %s: %s", d.Line, d.Col, d.Code, d.Desc)
			}
		})
	}
}

func TestExplicarBuscaPorCategoria(t *testing.T) {
	l := asistente.Explicar(diag.Diagnostic{Code: "algo-nuevo", Category: diag.Ejecucion})
	if !strings.Contains(l.Titulo, "falló") {
		t.Errorf("sin lección propia debe usar la de su categoría: %+v", l)
	}
	l = asistente.Explicar(diag.Diagnostic{Code: "x", Category: "otra", Heading: "Encabezado", Desc: "desc"})
	if !reflect.DeepEqual(l, asistente.Leccion{Titulo: "Encabezado", Texto: "desc"}) {
		t.Errorf("sin categoría conocida debe usar el propio diagnóstico: %+v", l)
	}
}

func TestPalabraReservadaLeidaComoTipoONombre(t *testing.T) {
	casos := []struct {
		fuente, codigo, reemplazo string
	}{
		// «gritra vida» es, para la gramática, un dato «vida» de tipo «gritra».
		{"combate\n    gritra \"hola\"\nfin\n", "asignacion-esperada", "gritar"},
		{"combate\n    roca vida = 1\n    gritra vida\nfin\n", "tipo-desconocido", "gritar"},
		{"combate\n    electrico listo = verdadro\n    gritar listo\nfin\n", "nombre-no-declarado", "verdadero"},
		// «recorerer nombre» es un dato de tipo «recorerer»; el error cae en
		// lo que sobra después.
		{"combate\n    planta nombre = \"a\"\n    recorerer nombre, nombre\nfin\n", "token-inesperado", "recorrer"},
		{"combate\n    equipo de roca l = [1]\n    recorerer x en l\n    fin\nfin\n", "token-inesperado", "recorrer"},
	}
	for _, c := range casos {
		diags, e := analizar(t, map[string]string{"principal.pks": c.fuente})
		encontrado := false
		for _, d := range asistente.Asistir(diags, e) {
			if d.Code == c.codigo && d.Fix != nil && d.Fix.Replacement == c.reemplazo {
				encontrado = true
				if !strings.Contains(d.Suggest, "? Si no, ") {
					t.Errorf("la sugerencia original debe seguir después de la pregunta: %q", d.Suggest)
				}
			}
		}
		if !encontrado {
			t.Errorf("%q: faltó sugerir %q en %s", c.fuente, c.reemplazo, c.codigo)
		}
	}
}

func TestNoSugierePalabrasReservadasParaNombresCortos(t *testing.T) {
	for _, fuente := range []string{"combate\n    gritar x\nfin\n", "combate\n    x\nfin\n"} {
		diags, e := analizar(t, map[string]string{"principal.pks": fuente})
		for _, d := range asistente.Asistir(diags, e) {
			if d.Fix != nil {
				t.Errorf("%q: no debe sugerir %q para una sola letra", fuente, d.Fix.Replacement)
			}
		}
	}
}
