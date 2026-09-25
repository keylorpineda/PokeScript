package parser

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/keylorpineda/PokeScript/internal/ast"
	. "github.com/keylorpineda/PokeScript/internal/ast/astprueba"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// TestContratoSeccion10 es la prueba de contrato entre el parser y el
// intérprete: los .pks reales de la sección 10 deben dar exactamente el
// árbol que el intérprete ya sabe ejecutar (astprueba.Seccion10).
func TestContratoSeccion10(t *testing.T) {
	for archivo, want := range Seccion10() {
		t.Run(archivo, func(t *testing.T) {
			fuente, err := os.ReadFile(filepath.Join("..", "lexer", "testdata", archivo))
			if err != nil {
				t.Fatal(err)
			}
			r := Analizar(archivo, string(fuente))
			for _, d := range r.Diagnosticos {
				t.Errorf("%d:%d %s: %s", d.Line, d.Col, d.Code, d.Desc)
			}
			if d := Diferencia(r.Programa, want); d != "" {
				t.Errorf("el árbol no coincide con astprueba.Seccion10: %s", d)
			}
		})
	}
}

func cuerpo(instrucciones ...ast.Instr) *ast.Programa {
	return Programa("prueba.pks", nil, Combate(instrucciones...))
}

func codigosDe(r Resultado) []string {
	var cs []string
	for _, d := range r.Diagnosticos {
		cs = append(cs, d.Code)
	}
	return cs
}

func TestInstrucciones(t *testing.T) {
	casos := []struct {
		nombre string
		fuente string
		want   *ast.Programa
	}{
		{"declaraciones",
			"combate\n roca vida = 100\n planta nombre\n posible planta rival = fantasma\n equipo de roca niveles = [1, 2]\n mochila de planta a roca bolsa = {}\n Estado e = DORMIDO\n medalla roca TOPE = 5\nfin",
			cuerpo(
				Dato(TRoca(), "vida", Roca(100)),
				Dato(TPlanta(), "nombre", nil),
				Dato(TPosible(TPlanta()), "rival", Fantasma()),
				Dato(TEquipo(TRoca()), "niveles", Equipo(Roca(1), Roca(2))),
				Dato(TMochila(TPlanta(), TRoca()), "bolsa", Llaves()),
				Dato(TNombre("Estado"), "e", Id("DORMIDO")),
				MedallaLocal(TRoca(), "TOPE", Roca(5)),
			)},
		{"asignaciones",
			"combate\n vida = vida - 10\n e[1] = 3\n m[\"Poción\"] = 2\n mio.vida = 0\n e[1][2].x = 1\nfin",
			cuerpo(
				Asignacion(Id("vida"), Bin(token.MINUS, Id("vida"), Roca(10))),
				Asignacion(Indice(Id("e"), Roca(1)), Roca(3)),
				Asignacion(Indice(Id("m"), Planta("Poción")), Roca(2)),
				Asignacion(Campo(Id("mio"), "vida"), Roca(0)),
				Asignacion(Campo(Indice(Indice(Id("e"), Roca(1)), Roca(2)), "x"), Roca(1)),
			)},
		{"entrada, salida y colecciones",
			"combate\n gritar \"Hola\", x, 3\n capturar(nombre, \"¿Nombre? \")\n capturar(e[2], \"¿Otro? \")\n sumar \"Bulbasaur\" a mi_equipo\n quitar mi_equipo[2]\n saludar(\"Ash\")\nfin",
			cuerpo(
				Gritar(Planta("Hola"), Id("x"), Roca(3)),
				Capturar(Id("nombre"), "¿Nombre? "),
				Capturar(Indice(Id("e"), Roca(2)), "¿Otro? "),
				Sumar(Planta("Bulbasaur"), "mi_equipo"),
				Quitar("mi_equipo", Roca(2)),
				Instr(Llamada("saludar", Planta("Ash"))),
			)},
		{"si con sino si y sino",
			"combate\n si vida > 50\n  gritar 1\n sino si vida > 0\n  gritar 2\n sino si vida igual 0\n  gritar 3\n sino\n  gritar 4\n fin\nfin",
			cuerpo(Si(Bin(token.GT, Id("vida"), Roca(50)), []ast.Instr{Gritar(Roca(1))},
				[]*ast.RamaSi{
					Rama(Bin(token.GT, Id("vida"), Roca(0)), Gritar(Roca(2))),
					Rama(Bin(token.IGUAL, Id("vida"), Roca(0)), Gritar(Roca(3))),
				},
				[]ast.Instr{Gritar(Roca(4))}))},
		{"si sin sino deja Sino vacío",
			"combate\n si listo\n  huir\n fin\nfin",
			cuerpo(Si(Id("listo"), []ast.Instr{Huir()}, nil, nil))},
		{"sino vacío",
			"combate\n si listo\n  huir\n sino\n fin\nfin",
			cuerpo(Si(Id("listo"), []ast.Instr{Huir()}, nil, []ast.Instr{}))},
		{"ciclos",
			"combate\n mientras vida > 0\n  vida = vida - 1\n  siguiente\n fin\n recorrer n de 1 hasta 5\n  gritar n\n fin\n recorrer p en equipo_ash\n  gritar p\n fin\n recorrer baya, cantidad en bayas\n  gritar baya, cantidad\n fin\nfin",
			cuerpo(
				Mientras(Bin(token.GT, Id("vida"), Roca(0)), Asignacion(Id("vida"), Bin(token.MINUS, Id("vida"), Roca(1))), Siguiente()),
				Rango("n", Roca(1), Roca(5), Gritar(Id("n"))),
				Recorrido("p", "", Id("equipo_ash"), Gritar(Id("p"))),
				Recorrido("baya", "cantidad", Id("bayas"), Gritar(Id("baya"), Id("cantidad"))),
			)},
		{"segun con otro",
			"combate\n segun ataque\n  'F' entonces gritar \"¡Lanzallamas!\"\n  'A', 'H' entonces gritar \"¡Hidrobomba!\"\n  otro entonces gritar \"Placaje\"\n fin\nfin",
			cuerpo(Segun(Id("ataque"), Gritar(Planta("Placaje")),
				Alternativa(Gritar(Planta("¡Lanzallamas!")), Fuego('F')),
				Alternativa(Gritar(Planta("¡Hidrobomba!")), Fuego('A'), Fuego('H')),
			))},
		{"líneas vacías y comentarios dentro de bloques",
			"combate\n\n  // comentario\n  gritar 1\n\nfin\n",
			cuerpo(Gritar(Roca(1)))},
		{"movimientos",
			"movimiento roca doble(roca x, agua y_extra)\n entregar x * 2\nfin\nmovimiento saludar()\n entregar\nfin\nmovimiento equipo de roca lista(equipo de roca e)\n entregar e\nfin",
			Programa("prueba.pks", nil,
				Movimiento(TRoca(), "doble", []any{TRoca(), "x", TAgua(), "y_extra"}, Entregar(Bin(token.STAR, Id("x"), Roca(2)))),
				Movimiento(nil, "saludar", nil, Entregar(nil)),
				Movimiento(TEquipo(TRoca()), "lista", []any{TEquipo(TRoca()), "e"}, Entregar(Id("e"))),
			)},
		{"especie en varias líneas",
			"especie Clima\n SOLEADO,\n LLUVIA, GRANIZO\nfin",
			Programa("prueba.pks", nil, Especie("Clima", "SOLEADO", "LLUVIA", "GRANIZO"))},
		{"importaciones con varios nombres",
			"enseñar Estado, Pokemon desde \"tipos.pks\"\ncombate\nfin",
			Programa("prueba.pks", []*ast.Importacion{Importar("tipos.pks", "Estado", "Pokemon")}, Combate())},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			r := Analizar("prueba.pks", c.fuente)
			if len(r.Diagnosticos) > 0 {
				d := r.Diagnosticos[0]
				t.Fatalf("diagnósticos inesperados %v; el primero en %d:%d: %s", codigosDe(r), d.Line, d.Col, d.Desc)
			}
			if d := Diferencia(r.Programa, c.want); d != "" {
				t.Errorf("árbol distinto: %s", d)
			}
		})
	}
}

func TestSinoVacioNoEsNil(t *testing.T) {
	for _, fuente := range []string{"combate\n si x\n fin\nfin", "combate\n si x\n sino\n fin\nfin"} {
		r := Analizar("prueba.pks", fuente)
		s := r.Programa.Declaraciones[0].(*ast.Combate).Cuerpo[0].(*ast.Si)
		if s.Sino == nil {
			t.Errorf("%q: Si.Sino es nil; el contrato pide un slice vacío", fuente)
		}
	}
}

func TestPosicionesDeInstrucciones(t *testing.T) {
	r := Analizar("prueba.pks", "combate\n    mientras vida > 0\n        vida = vida - 1\n    fin\nfin")
	c := r.Programa.Declaraciones[0].(*ast.Combate)
	if want := (ast.Pos{Line: 1, Col: 1, Len: 7}); c.Pos != want {
		t.Errorf("Combate.Pos = %+v, want %+v", c.Pos, want)
	}
	m := c.Cuerpo[0].(*ast.Mientras)
	if want := (ast.Pos{Line: 2, Col: 5, Len: 17}); m.Pos != want {
		t.Errorf("Mientras.Pos = %+v, want %+v", m.Pos, want)
	}
	a := m.Cuerpo[0].(*ast.Asignacion)
	if want := (ast.Pos{Line: 3, Col: 9, Len: 15}); a.Pos != want {
		t.Errorf("Asignacion.Pos = %+v, want %+v", a.Pos, want)
	}

	r = Analizar("prueba.pks", "movimiento roca f()\n entregar 1\nfin")
	mv := r.Programa.Declaraciones[0].(*ast.DeclMovimiento)
	if want := (ast.Pos{Line: 3, Col: 1, Len: 3}); mv.FinPos != want {
		t.Errorf("FinPos = %+v, want %+v", mv.FinPos, want)
	}
}

// ─── Errores ───────────────────────────────────────────────────────────────

func TestBloqueSinCerrar(t *testing.T) {
	// Caso sintáctico 1 de la sección 11: el mensaje dice en qué línea se
	// abrió el bloque, y la sangría señala cuál es el fin que falta.
	fuente := strings.Join([]string{
		"combate",                         // 1
		"    planta nombre",               // 2
		"    roca intentos = 0",           // 3
		"    mientras intentos < 3",       // 4
		"        intentos = intentos + 1", // 5
		"        si intentos igual 3",     // 6
		"            gritar nombre",       // 7
		"        fin",                     // 8
		"fin",                             // 9: cierra el mientras, no el combate
	}, "\n")
	r := Analizar("captura.pks", fuente)
	if got := codigosDe(r); !reflect.DeepEqual(got, []string{"bloque-sin-cerrar"}) {
		t.Fatalf("códigos = %v", got)
	}
	d := r.Diagnosticos[0]
	if d.Line != 1 || d.Col != 1 || d.Len != 7 {
		t.Errorf("el error debe señalar el combate de la línea 1, señala %d:%d len %d", d.Line, d.Col, d.Len)
	}
	if !strings.Contains(d.Desc, "línea 1") {
		t.Errorf("la descripción debe decir dónde se abrió: %q", d.Desc)
	}
	if !strings.Contains(d.Cause, "«mientras»") || !strings.Contains(d.Cause, "línea 4") {
		t.Errorf("la causa debe señalar el mientras de la línea 4: %q", d.Cause)
	}
	if d.Heading != "¡Se escapó!" {
		t.Errorf("Heading = %q", d.Heading)
	}
}

func TestBloquesSinCerrarAntesDeOtraDeclaracion(t *testing.T) {
	r := Analizar("prueba.pks", "movimiento f()\n si x\n  gritar 1\nmovimiento g()\nfin")
	if got := codigosDe(r); !reflect.DeepEqual(got, []string{"bloque-sin-cerrar", "bloque-sin-cerrar"}) {
		t.Fatalf("códigos = %v", got)
	}
	if r.Diagnosticos[0].Line != 2 || r.Diagnosticos[1].Line != 1 {
		t.Errorf("deben reportarse el si (línea 2) y el movimiento (línea 1): %d y %d", r.Diagnosticos[0].Line, r.Diagnosticos[1].Line)
	}
	// El segundo movimiento se sigue leyendo.
	if n := len(r.Programa.Declaraciones); n != 2 {
		t.Errorf("se esperaban 2 declaraciones, hay %d", n)
	}
}

func TestErroresDeInstrucciones(t *testing.T) {
	casos := []struct {
		nombre  string
		fuente  string
		codigos []string
		linea   int
		col     int
	}{
		// Casos sintácticos de la sección 11.
		{"operador no encadenable", "combate\n si p > q > r\n  huir\n fin\nfin", []string{"operador-no-encadenable"}, 2, 11},
		{"dos instrucciones en una línea", "combate\n roca x = 1 gritar x\nfin", []string{"dos-instrucciones"}, 2, 13},
		{"cadena sin cerrar", "combate\n gritar \"hola\nfin", []string{"cadena-sin-cerrar"}, 2, 9},
		{"recorrer sin hasta", "combate\n recorrer n de 1 a 10\n  gritar n\n fin\nfin", []string{"rango-sin-hasta"}, 2, 18},

		// Otros.
		{"fin sobrante", "combate\nfin\nfin", []string{"fin-sobrante"}, 3, 1},
		{"instrucción fuera de bloque", "gritar \"hola\"", []string{"instruccion-fuera-de-bloque"}, 1, 1},
		{"sino sin si", "combate\n sino\nfin", []string{"sino-sin-si"}, 2, 2},
		{"otro sin segun", "combate\n otro entonces huir\nfin", []string{"otro-sin-segun"}, 2, 2},
		{"valor suelto", "combate\n vida + 1\nfin", []string{"asignacion-esperada"}, 2, 7},
		{"importación después de declaración", "combate\nfin\nenseñar x desde \"a.pks\"", []string{"importacion-fuera-de-lugar"}, 3, 1},
		{"medalla sin valor", "medalla roca TOPE", []string{"token-esperado"}, 1, 18},
		{"rama después de otro", "combate\n segun x\n  otro entonces huir\n  1 entonces huir\n fin\nfin", []string{"rama-despues-de-otro"}, 4, 3},
		{"patrón inválido", "combate\n segun x\n  p + 1 entonces huir\n  otro entonces huir\n fin\nfin", []string{"token-esperado"}, 3, 5},
		{"patrón con operador", "combate\n segun x\n  (1) entonces huir\n fin\nfin", []string{"patron-invalido"}, 3, 3},
		{"capturar sin mensaje", "combate\n capturar(x)\nfin", []string{"token-esperado"}, 2, 12},
		{"sumar sin a", "combate\n sumar 1 lista\nfin", []string{"token-esperado"}, 2, 10},
		{"recorrer sin en", "combate\n recorrer x equipo_rojo\n fin\nfin", []string{"token-esperado"}, 2, 13},
		{"movimiento sin paréntesis", "movimiento f\nfin", []string{"token-esperado"}, 1, 13},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			r := Analizar("prueba.pks", c.fuente)
			if got := codigosDe(r); !reflect.DeepEqual(got, c.codigos) {
				t.Fatalf("códigos = %v, want %v", got, c.codigos)
			}
			d := r.Diagnosticos[0]
			if d.Line != c.linea || d.Col != c.col {
				t.Errorf("posición = %d:%d, want %d:%d", d.Line, d.Col, c.linea, c.col)
			}
			if d.Desc == "" || d.Suggest == "" || d.Heading == "" {
				t.Errorf("diagnóstico incompleto: %+v", d)
			}
		})
	}
}

func TestRecuperacionReportaTodosLosErrores(t *testing.T) {
	// Tres líneas malas: se reportan las tres y el resto del bloque se lee.
	r := Analizar("prueba.pks", "combate\n roca = 1\n gritar\n x y\n gritar \"bien\"\nfin")
	if got := codigosDe(r); len(got) != 3 {
		t.Fatalf("se esperaban 3 errores, hubo %v", got)
	}
	c := r.Programa.Declaraciones[0].(*ast.Combate)
	if len(c.Cuerpo) != 1 {
		t.Errorf("la línea buena debe quedar en el árbol: cuerpo con %d instrucciones", len(c.Cuerpo))
	}
}

func TestErrorEnCabeceraNoDescuadraLosFin(t *testing.T) {
	// La condición está mal, pero el cuerpo y su fin se leen igual: el
	// único error es el de la condición, no un "bloque sin cerrar".
	r := Analizar("prueba.pks", "combate\n si p > \n  gritar 1\n fin\n gritar 2\nfin")
	if got := codigosDe(r); !reflect.DeepEqual(got, []string{"expresion-esperada"}) {
		t.Errorf("códigos = %v", got)
	}
}

func TestLimiteDeVeinteErrores(t *testing.T) {
	lineas := []string{"combate"}
	for i := 0; i < 40; i++ {
		lineas = append(lineas, " roca = 1")
	}
	lineas = append(lineas, "fin")
	r := Analizar("prueba.pks", strings.Join(lineas, "\n"))
	if n := len(r.Diagnosticos); n != MaxErrores {
		t.Errorf("se esperaban %d diagnósticos, hubo %d", MaxErrores, n)
	}
}

func TestResultadoTieneErrores(t *testing.T) {
	if Analizar("prueba.pks", "combate\nfin").TieneErrores() {
		t.Error("un programa válido no debe tener errores")
	}
	if !Analizar("prueba.pks", "combate").TieneErrores() {
		t.Error("un combate sin fin debe tener errores")
	}
}
