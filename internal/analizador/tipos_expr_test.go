package analizador

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/proyecto"
)

// soloCodigo devuelve los códigos de los diagnósticos que tienen el prefijo
// dado; así cada prueba mira solo lo suyo aunque haya otras verificaciones.
func conCodigo(ds []diag.Diagnostic, codigo string) []diag.Diagnostic {
	var r []diag.Diagnostic
	for _, d := range ds {
		if d.Code == codigo {
			r = append(r, d)
		}
	}
	return r
}

func TestTiposSinErrores(t *testing.T) {
	casos := map[string]string{
		"ensanchar roca a agua":      "    agua x = 5\n    x = x + 1",
		"división da agua":           "    agua x = 7 / 2",
		"concatenar planta":          "    planta s = \"a\" + \"b\"",
		"comparar":                   "    electrico e = 1 < 2.5 y 'a' < 'b'",
		"contiene en texto":          "    electrico e = \"Pikachu\" contiene 'k'",
		"convertir con la tabla":     "    roca n = convertir(\"12\") a roca\n    planta s = convertir(3.5) a planta",
		"redondear y aleatorio":      "    roca n = redondear(2.5) + aleatorio(1, 6)",
		"índice y tamaño":            "    equipo de roca e = [1, 2]\n    roca n = e[1] + tamaño(e)",
		"recorrer un equipo":         "    equipo de planta e = [\"a\"]\n    recorrer p en e\n        planta q = p + \"!\"\n    fin",
		"recorrer mochila":           "    mochila de planta a roca m = {\"a\": 1}\n    recorrer k, n en m\n        planta s = k\n        roca x = n\n    fin",
		"recorrer texto":             "    recorrer c en \"hola\"\n        fuego f = c\n    fin",
		"rango":                      "    recorrer i de 1 hasta 3\n        roca x = i * 2\n    fin",
		"sumar y quitar":             "    equipo de agua e = []\n    sumar 1 a e\n    quitar e[1]",
		"capturar":                   "    roca n\n    capturar(n, \"¿? \")",
		"posible con igual":          "    posible planta r = fantasma\n    electrico e = r igual fantasma",
		"posible con sino":           "    posible planta r = fantasma\n    planta s = r sino \"nadie\"",
		"gritar cualquier cosa":      "    gritar 1, 2.5, 'a', \"b\", verdadero",
		"mientras con condición":     "    roca n = 0\n    mientras n < 3\n        n = n + 1\n    fin",
		"no con electrico":           "    electrico e = no (1 > 2)",
		"menos unario":               "    agua x = -2.5",
		"asignar a posible un valor": "    posible roca x = fantasma\n    x = 5",
	}
	for nombre, cuerpo := range casos {
		r := analizar(t, programa(cuerpo))
		if len(r.Diagnosticos) != 0 {
			t.Errorf("%s: diagnósticos inesperados: %v", nombre, codigos(r.Diagnosticos))
		}
	}
}

func TestTiposErrores(t *testing.T) {
	casos := []struct {
		nombre string
		cuerpo string
		codigo string
	}{
		{"roca más planta", "    roca x = 1 + \"a\"", "tipo-incompatible"},
		{"resto con agua", "    roca x = 5.5 resto 2", "tipo-incompatible"},
		{"comparar textos con >", "    electrico e = \"a\" > \"b\"", "tipo-incompatible"},
		{"roca igual agua", "    electrico e = 1 igual 1.0", "tipo-incompatible"},
		{"y con roca", "    electrico e = verdadero y 1", "tipo-incompatible"},
		{"no con roca", "    electrico e = no 1", "tipo-incompatible"},
		{"tamaño de roca", "    roca n = tamaño(5)", "tipo-incompatible"},
		{"índice de roca", "    roca x = 5\n    roca z = x[1]", "tipo-incompatible"},
		{"índice agua en equipo", "    equipo de roca e = [1]\n    roca z = e[1.5]", "tipo-incompatible"},
		{"agua a roca", "    roca x = 2.5", "asignacion-incompatible"},
		{"planta a roca pide convertir", "    roca x = \"12\"", "asignacion-incompatible"},
		{"reasignar con otro tipo", "    roca x = 1\n    x = verdadero", "asignacion-incompatible"},
		{"fantasma en un dato no posible", "    roca x = fantasma", "asignacion-incompatible"},
		{"condición roca en si", "    si 1\n        gritar \"a\"\n    fin", "condicion-no-electrico"},
		{"condición planta en mientras", "    mientras \"a\"\n    fin", "condicion-no-electrico"},
		{"condición en sino si", "    si verdadero\n    sino si 2\n    fin", "condicion-no-electrico"},
		{"nombre no declarado", "    roca x = vida + 1", "nombre-no-declarado"},
		{"movimiento no declarado", "    curar(1)", "nombre-no-declarado"},
		{"división entre 0", "    agua x = 5 / 0", "division-entre-cero"},
		{"resto entre 0", "    roca x = 5 resto 0", "division-entre-cero"},
		{"división entre 0.0", "    agua x = 5.0 / 0.0", "division-entre-cero"},
		{"convertir imposible", "    roca x = convertir(verdadero) a roca", "conversion-imposible"},
		{"capturar un equipo", "    equipo de roca e = []\n    capturar(e, \"?\")", "capturar-invalido"},
		{"sumar a una mochila", "    mochila de planta a roca m = {}\n    sumar 1 a m", "sumar-invalido"},
		{"sumar otro tipo", "    equipo de roca e = []\n    sumar \"a\" a e", "tipo-incompatible"},
		{"quitar de una roca", "    roca x = 1\n    quitar x[1]", "quitar-invalido"},
		{"recorrer una roca", "    recorrer x en 5\n    fin", "recorrer-invalido"},
		{"recorrer equipo con dos variables", "    equipo de roca e = []\n    recorrer p, q en e\n    fin", "recorrer-invalido"},
		{"extremo de rango agua", "    recorrer i de 1 hasta 2.5\n    fin", "tipo-incompatible"},
		{"aleatorio con agua", "    roca n = aleatorio(1, 2.5)", "tipo-incompatible"},
		{"redondear un planta", "    roca n = redondear(\"a\")", "tipo-incompatible"},
		{"sino sin posible", "    planta s = \"a\" sino \"b\"", "tipo-incompatible"},
	}
	for _, c := range casos {
		r := analizar(t, programa(c.cuerpo))
		if got := conCodigo(r.Diagnosticos, c.codigo); len(got) != 1 {
			t.Errorf("%s: esperaba un %s, llegó %v", c.nombre, c.codigo, codigos(r.Diagnosticos))
		}
	}
}

// Caso semántico 1 de la sección 11: roca + electrico remite a la tabla.
func TestCasoSemantico1(t *testing.T) {
	r := analizar(t, programa("    roca vida = 10\n    roca x = vida + verdadero"))
	ds := conCodigo(r.Diagnosticos, "tipo-incompatible")
	if len(ds) != 1 {
		t.Fatalf("diagnósticos = %v", codigos(r.Diagnosticos))
	}
	d := ds[0]
	want := diag.Diagnostic{
		Severity: diag.Error, Category: diag.Semantico, Code: "tipo-incompatible",
		Heading: diag.EncabezadoTipos, File: "principal.pks", Line: 3, Col: 14, Len: 16,
		Desc:    "no es posible sumar un valor de tipo roca con uno de tipo electrico.",
		Cause:   "la combinación de estos dos tipos no tiene efecto según la tabla de efectividades.",
		Suggest: "consultar la tabla de efectividades desde el menú del entorno.",
	}
	if !reflect.DeepEqual(d, want) {
		t.Errorf("diagnóstico =\n%#v\nwant\n%#v", d, want)
	}
}

func TestLlamadas(t *testing.T) {
	const declaraciones = "movimiento roca curar(roca vida, planta nombre)\n    entregar vida\nfin\n" +
		"movimiento saludar()\n    gritar \"hola\"\nfin\n"
	casos := []struct {
		nombre string
		cuerpo string
		codigo string
	}{
		{"argumento de otro tipo", "    roca x = curar(\"a\", \"b\")", "argumento-incompatible"},
		{"faltan argumentos", "    roca x = curar(1)", "cantidad-de-argumentos"},
		{"sobran argumentos", "    saludar(1)", "cantidad-de-argumentos"},
		{"movimiento sin tipo en una expresión", "    roca x = saludar()", "movimiento-sin-valor"},
		{"llamar a algo que no es movimiento", "    roca n = 1\n    n(2)", "no-es-movimiento"},
		{"movimiento usado como valor", "    roca n = curar", "no-es-un-valor"},
		{"valor descartado", "    curar(1, \"a\")", "valor-descartado"},
	}
	for _, c := range casos {
		r := analizar(t, map[string]string{"principal.pks": declaraciones + "combate\n" + c.cuerpo + "\nfin\n"})
		if got := conCodigo(r.Diagnosticos, c.codigo); len(got) != 1 {
			t.Errorf("%s: esperaba un %s, llegó %v", c.nombre, c.codigo, codigos(r.Diagnosticos))
		}
	}
	r := analizar(t, map[string]string{"principal.pks": declaraciones + "combate\n    curar(1, \"a\")\nfin\n"})
	if d := conCodigo(r.Diagnosticos, "valor-descartado"); len(d) != 1 || d[0].Severity != diag.Advertencia {
		t.Errorf("el valor descartado debe ser advertencia: %+v", d)
	}
}

func TestEntregaYMedallas(t *testing.T) {
	r := analizar(t, map[string]string{"principal.pks": "medalla roca MAX = \"cien\"\n" +
		"movimiento roca f()\n    entregar \"a\"\nfin\n" +
		"movimiento agua g()\n    entregar 1\nfin\n" +
		"combate\n    agua x = g()\nfin\n"})
	if got := codigos(r.Diagnosticos); !reflect.DeepEqual(got, []string{"asignacion-incompatible", "entrega-incompatible"}) {
		t.Errorf("diagnósticos = %v", got)
	}
}

func TestCampos(t *testing.T) {
	const ficha = "ficha Pokemon\n    planta nombre\n    roca vida\nfin\n"
	r := analizar(t, map[string]string{"principal.pks": ficha +
		"movimiento roca vida_de(Pokemon p)\n    entregar p.vida\nfin\n" +
		"movimiento roca mala(Pokemon p)\n    entregar p.nivel\nfin\n" +
		"combate\n    roca x = 5\n    roca z = x.vida\nfin\n"})
	if got := codigos(r.Diagnosticos); !reflect.DeepEqual(got, []string{"campo-inexistente", "no-es-ficha"}) {
		t.Errorf("diagnósticos = %v", got)
	}
	if d := conCodigo(r.Diagnosticos, "campo-inexistente"); len(d) == 1 && !strings.Contains(d[0].Cause, "nombre, vida") {
		t.Errorf("la causa debe listar los campos: %q", d[0].Cause)
	}
}

// Caso de importación 3 de la sección 11: argumentos de tipo incorrecto a
// un movimiento importado. El error va en el archivo que llama.
func TestCasoImportacion3(t *testing.T) {
	r := analizar(t, map[string]string{
		"operaciones.pks": "movimiento roca calcular_dano(roca poder)\n    entregar poder * 2\nfin\n",
		"principal.pks": "enseñar calcular_dano desde \"operaciones.pks\"\n" +
			"combate\n    roca d = calcular_dano(\"fuerte\")\nfin\n",
	})
	ds := conCodigo(r.Diagnosticos, "argumento-incompatible")
	if len(ds) != 1 {
		t.Fatalf("diagnósticos = %v", codigos(r.Diagnosticos))
	}
	d := ds[0]
	if d.File != "principal.pks" || d.Line != 3 || !strings.Contains(d.Desc, "declarado en operaciones.pks") {
		t.Errorf("diagnóstico = %s %d: %s", d.File, d.Line, d.Desc)
	}
}

func TestNoDeclaradoSugiereImportar(t *testing.T) {
	r := analizar(t, map[string]string{
		"constantes.pks": "medalla roca NIVEL = 25\n",
		"principal.pks":  "enseñar NIVEL desde \"constantes.pks\"\ncombate\n    roca x = NIVEL\nfin\n",
		"otro.pks":       "movimiento roca f()\n    entregar NIVEL\nfin\n",
	})
	if len(r.Diagnosticos) != 0 {
		t.Errorf("otro.pks no es alcanzable desde el principal y no debe revisarse: %v", codigos(r.Diagnosticos))
	}
	r = analizar(t, map[string]string{
		"constantes.pks": "medalla roca NIVEL = 25\n",
		"principal.pks":  "enseñar NIVEL desde \"constantes.pks\"\nenseñar f desde \"otro.pks\"\ncombate\n    roca x = f()\nfin\n",
		"otro.pks":       "movimiento roca f()\n    entregar NIVEL\nfin\n",
	})
	ds := conCodigo(r.Diagnosticos, "nombre-no-declarado")
	if len(ds) != 1 || ds[0].File != "otro.pks" || !strings.Contains(ds[0].Suggest, "enseñar NIVEL desde \"constantes.pks\"") {
		t.Errorf("diagnósticos = %+v", ds)
	}
}

func TestParametrosYAlcances(t *testing.T) {
	r := analizar(t, map[string]string{"principal.pks": "" +
		"movimiento roca doble(roca n)\n    entregar n * 2\nfin\n" +
		"combate\n" +
		"    si verdadero\n        roca dentro = 1\n    fin\n" +
		"    roca x = dentro\n" + // murió con el bloque
		"    roca n = doble(3)\n" +
		"fin\n"})
	ds := conCodigo(r.Diagnosticos, "nombre-no-declarado")
	if len(ds) != 1 || ds[0].Line != 8 {
		t.Errorf("diagnósticos = %+v", r.Diagnosticos)
	}
}

// Los ejemplos del repositorio y el proyecto de la sección 10 no deben dar
// ningún diagnóstico: cmd/pks exige una salida de errores vacía.
func TestEjemplosSinDiagnosticos(t *testing.T) {
	carpeta := filepath.Join("..", "..", "ejemplos")
	entradas, err := os.ReadDir(carpeta)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entradas {
		ruta := filepath.Join(carpeta, e.Name())
		var res proyecto.Resultado
		if e.IsDir() {
			res, err = proyecto.Cargar(ruta)
		} else {
			res, err = proyecto.CargarArchivo(ruta)
		}
		if err != nil || res.TieneErrores() {
			t.Fatalf("%s no carga: %v %v", e.Name(), err, res.Diagnosticos)
		}
		if ds := Analizar(res.Proyecto).Diagnosticos; len(ds) != 0 {
			for _, d := range ds {
				t.Errorf("%s: %s %d:%d %s: %s", e.Name(), d.File, d.Line, d.Col, d.Code, d.Desc)
			}
		}
	}
}
