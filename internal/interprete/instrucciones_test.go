package interprete

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/token"
)

func TestInstrucciones(t *testing.T) {
	casos := []struct {
		nombre string
		cuerpo []ast.Instr
		want   []string
	}{
		{
			"gritar concatena sin separador",
			[]ast.Instr{
				dato(tPlanta, "nombre", planta("Bulbi")),
				gritar(planta("¡"), id("nombre"), planta(" tiene "), roca(100), planta(" de vida!")),
			},
			[]string{"¡Bulbi tiene 100 de vida!"},
		},
		{
			"declarar roca en un agua la ensancha",
			[]ast.Instr{dato(tAgua, "x", roca(5)), gritar(id("x"))},
			[]string{"5.0"},
		},
		{
			"asignar reemplaza el valor",
			[]ast.Instr{dato(tRoca, "x", nil), asignacion(id("x"), roca(1)), asignacion(id("x"), bin(token.PLUS, id("x"), roca(1))), gritar(id("x"))},
			[]string{"2"},
		},
		{
			"si, sino si y sino",
			[]ast.Instr{
				dato(tRoca, "n", roca(5)),
				si(bin(token.GT, id("n"), roca(10)), []ast.Instr{gritar(planta("grande"))},
					[]*ast.RamaSi{rama(bin(token.GT, id("n"), roca(3)), gritar(planta("mediano")))},
					[]ast.Instr{gritar(planta("pequeño"))}),
			},
			[]string{"mediano"},
		},
		{
			"si sin sino que no se cumple",
			[]ast.Instr{si(electrico(false), []ast.Instr{gritar(planta("no"))}, nil, nil), gritar(planta("sigue"))},
			[]string{"sigue"},
		},
		{
			"mientras con huir y siguiente",
			[]ast.Instr{
				dato(tRoca, "i", roca(0)),
				mientras(electrico(true),
					asignacion(id("i"), bin(token.PLUS, id("i"), roca(1))),
					si(bin(token.IGUAL, id("i"), roca(2)), []ast.Instr{siguiente()}, nil, nil),
					si(bin(token.GT, id("i"), roca(4)), []ast.Instr{huir()}, nil, nil),
					gritar(id("i")),
				),
			},
			[]string{"1", "3", "4"},
		},
		{
			"recorrer un rango incluye ambos extremos",
			[]ast.Instr{rango("n", roca(1), roca(3), gritar(id("n")))},
			[]string{"1", "2", "3"},
		},
		{
			"recorrer un rango vacío no entra",
			[]ast.Instr{rango("n", roca(3), roca(1), gritar(id("n"))), gritar(planta("fin"))},
			[]string{"fin"},
		},
		{
			"los extremos del rango se evalúan una sola vez",
			[]ast.Instr{
				dato(tRoca, "tope", roca(2)),
				rango("n", roca(1), id("tope"), asignacion(id("tope"), roca(10)), gritar(id("n"))),
			},
			[]string{"1", "2"},
		},
		{
			"recorrer hasta el máximo de roca no desborda",
			[]ast.Instr{rango("n", roca(9223372036854775806), roca(9223372036854775807), gritar(id("n")))},
			[]string{"9223372036854775806", "9223372036854775807"},
		},
		{
			"recorrer un equipo",
			[]ast.Instr{
				dato(tEquipo(tPlanta), "e", equipoLit(planta("a"), planta("b"))),
				recorrido("p", "", id("e"), gritar(id("p"))),
			},
			[]string{"a", "b"},
		},
		{
			"recorrer una mochila con una variable da las claves",
			[]ast.Instr{
				dato(tMochila(tPlanta, tRoca), "m", llaves(planta("x"), roca(1), planta("y"), roca(2))),
				recorrido("k", "", id("m"), gritar(id("k"))),
			},
			[]string{"x", "y"},
		},
		{
			"recorrer una mochila con dos variables",
			[]ast.Instr{
				dato(tMochila(tPlanta, tRoca), "m", llaves(planta("x"), roca(1), planta("y"), roca(2))),
				recorrido("k", "v", id("m"), gritar(id("k"), planta("="), id("v"))),
			},
			[]string{"x=1", "y=2"},
		},
		{
			"recorrer un texto letra por letra",
			[]ast.Instr{recorrido("c", "", planta("Éxito"), gritar(id("c")))},
			[]string{"É", "x", "i", "t", "o"},
		},
		{
			"huir sale solo del ciclo más interno",
			[]ast.Instr{rango("i", roca(1), roca(2), rango("j", roca(1), roca(5),
				si(bin(token.GT, id("j"), roca(1)), []ast.Instr{huir()}, nil, nil),
				gritar(id("i"), planta(","), id("j"))))},
			[]string{"1,1", "2,1"},
		},
		{
			"sumar, quitar y asignar a un equipo",
			[]ast.Instr{
				dato(tEquipo(tAgua), "e", equipoLit(roca(1), agua(2.5))),
				sumar(roca(3), "e"),
				quitar("e", roca(1)),
				asignacion(idx(id("e"), roca(1)), roca(7)),
				gritar(id("e")),
			},
			[]string{"[7.0, 3.0]"},
		},
		{
			"asignar a una clave nueva de mochila la agrega al final",
			[]ast.Instr{
				dato(tMochila(tPlanta, tRoca), "m", llaves(planta("a"), roca(1))),
				asignacion(idx(id("m"), planta("b")), roca(2)),
				asignacion(idx(id("m"), planta("a")), roca(10)),
				gritar(id("m")),
			},
			[]string{`{"a": 10, "b": 2}`},
		},
		{
			"asignar a un destino anidado",
			[]ast.Instr{
				dato(tEquipo(tEquipo(tRoca)), "tabla", equipoLit(equipoLit(roca(1), roca(2)))),
				asignacion(idx(idx(id("tabla"), roca(1)), roca(2)), roca(9)),
				gritar(id("tabla")),
			},
			[]string{"[[1, 9]]"},
		},
		{
			"asignar copia: cambiar la copia no toca el original",
			[]ast.Instr{
				dato(tEquipo(tRoca), "a", equipoLit(roca(1))),
				dato(tEquipo(tRoca), "b", id("a")),
				sumar(roca(2), "b"),
				gritar(id("a"), planta(" "), id("b")),
			},
			[]string{"[1] [1, 2]"},
		},
		{
			"cada vuelta tiene su propio alcance",
			[]ast.Instr{rango("n", roca(1), roca(2), dato(tRoca, "doble", bin(token.STAR, id("n"), roca(2))), gritar(id("doble")))},
			[]string{"2", "4"},
		},
		{
			"medalla local",
			[]ast.Instr{medallaLocal(tRoca, "MAX", roca(3)), gritar(id("MAX"))},
			[]string{"3"},
		},
	}
	for _, c := range casos {
		got, err := soloCombate(t, c.cuerpo...)
		if err != nil {
			t.Errorf("%s: error inesperado: %v", c.nombre, err)
			continue
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s:\n  salida = %q\n  want     %q", c.nombre, got, c.want)
		}
	}
}

func TestSegun(t *testing.T) {
	describir := movimiento(nil, "describir", []any{tNombre("Estado"), "actual"},
		segun(id("actual"), nil,
			alternativa(gritar(planta("Puede atacar")), id("SANO")),
			alternativa(gritar(planta("Pierde vida")), id("ENVENENADO")),
			alternativa(gritar(planta("Podría no atacar")), id("DORMIDO"), id("PARALIZADO")),
		))
	nivel := segun(id("n"), gritar(planta("otro")),
		alternativa(gritar(planta("uno")), roca(1)),
		alternativa(gritar(planta("dos o tres")), roca(2), roca(3)),
	)
	p := programa(
		especie("Estado", "SANO", "ENVENENADO", "DORMIDO", "PARALIZADO"),
		describir,
		combate(
			instr(llamada("describir", id("SANO"))),
			instr(llamada("describir", id("PARALIZADO"))),
			rango("n", roca(1), roca(4), nivel),
		),
	)
	es, err := correr(t, p)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"Puede atacar", "Podría no atacar", "uno", "dos o tres", "dos o tres", "otro"}
	if !reflect.DeepEqual(es.Salida, want) {
		t.Errorf("salida = %q\nwant     %q", es.Salida, want)
	}
}

func TestMovimientos(t *testing.T) {
	factorial := movimiento(tRoca, "factorial", []any{tRoca, "n"},
		si(bin(token.LE, id("n"), roca(1)), []ast.Instr{entregar(roca(1))}, nil, nil),
		entregar(bin(token.STAR, id("n"), llamada("factorial", bin(token.MINUS, id("n"), roca(1))))),
	)
	// Recibe una copia: modificarla no cambia el equipo de quien llama.
	agregar := movimiento(tRoca, "agregar", []any{tEquipo(tRoca), "e"},
		sumar(roca(99), "e"),
		entregar(&ast.Tamano{Valor: id("e")}),
	)
	// Un movimiento con tipo agua que entrega una roca la ensancha.
	mitad := movimiento(tAgua, "mitad", []any{tAgua, "x"}, entregar(bin(token.SLASH, id("x"), roca(2))))
	entero := movimiento(tAgua, "entero", nil, entregar(roca(4)))
	saludar := movimiento(nil, "saludar", []any{tPlanta, "a"}, gritar(planta("hola "), id("a")), entregar(nil), gritar(planta("no llega")))
	// No ve los datos de combate: solo las medallas del archivo y sus parámetros.
	usarMedalla := movimiento(tRoca, "doble_nivel", nil, entregar(bin(token.STAR, id("NIVEL"), roca(2))))

	p := programa(
		medalla(tRoca, "NIVEL", roca(25)),
		factorial, agregar, mitad, entero, saludar, usarMedalla,
		combate(
			dato(tEquipo(tRoca), "mio", equipoLit(roca(1))),
			gritar(llamada("factorial", roca(10))),
			gritar(llamada("agregar", id("mio")), planta(" "), id("mio")),
			gritar(llamada("mitad", roca(5))),
			gritar(llamada("entero")),
			instr(llamada("saludar", planta("Bulbi"))),
			gritar(llamada("doble_nivel")),
		),
	)
	es, err := correr(t, p)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"3628800", "2 [1]", "2.5", "4.0", "hola Bulbi", "50"}
	if !reflect.DeepEqual(es.Salida, want) {
		t.Errorf("salida = %q\nwant     %q", es.Salida, want)
	}
}

func TestRecursionLimitada(t *testing.T) {
	infinito := movimiento(tRoca, "caer", []any{tRoca, "n"}, entregar(llamada("caer", bin(token.PLUS, id("n"), roca(1)))))
	hasta := movimiento(tRoca, "hasta", []any{tRoca, "n"},
		si(bin(token.IGUAL, id("n"), roca(0)), []ast.Instr{entregar(roca(0))}, nil, nil),
		entregar(bin(token.PLUS, roca(1), llamada("hasta", bin(token.MINUS, id("n"), roca(1))))),
	)

	// 1000 llamadas anidadas se permiten.
	es, err := correr(t, programa(hasta, combate(gritar(llamada("hasta", roca(999))))))
	if err != nil || es.Salida[0] != "999" {
		t.Fatalf("1000 llamadas anidadas deberían funcionar: %v %v", es.Salida, err)
	}
	_, err = correr(t, programa(infinito, combate(gritar(llamada("caer", roca(0))))))
	if d := errorEjecucion(t, err); d.Code != CodigoRecursion {
		t.Errorf("código = %s, want %s", d.Code, CodigoRecursion)
	}
}

func TestCapturar(t *testing.T) {
	casos := []struct {
		nombre   string
		tipo     *ast.TipoExpr
		entradas []string
		salida   []string
	}{
		{"roca a la primera", tRoca, []string{"42"}, []string{"42"}},
		{"roca con espacios y signo", tRoca, []string{"  -7 "}, []string{"-7"}},
		{
			"roca vuelve a preguntar",
			tRoca,
			[]string{"doce", "12.5", "12"},
			[]string{
				"«doce» no es un roca. Escribe un número entero, por ejemplo 42.",
				"«12.5» no es un roca. Escribe un número entero, por ejemplo 42.",
				"12",
			},
		},
		{"roca que no cabe", tRoca, []string{"99999999999999999999", "1"}, []string{
			"«99999999999999999999» no cabe en un roca. " + descripcionRangoRoca, "1",
		}},
		{"agua acepta un entero", tAgua, []string{"5"}, []string{"5.0"}},
		{"agua con coma no sirve", tAgua, []string{"1,5", "1.5"}, []string{
			"«1,5» no es un agua. Escribe un número con punto decimal, por ejemplo 3.5.", "1.5",
		}},
		{"fuego exactamente un carácter", tFuego, []string{"ab", "é"}, []string{
			"«ab» no es un fuego. Escribe exactamente un carácter.", "é",
		}},
		{"fuego acepta un espacio", tFuego, []string{" "}, []string{" "}},
		{"planta tal cual", tPlanta, []string{"  Mr. Mime "}, []string{"  Mr. Mime "}},
		{"electrico", tElectrico, []string{"si", "verdadero"}, []string{
			"«si» no es un electrico. Escribe verdadero o falso.", "verdadero",
		}},
		{"especie", tNombre("Estado"), []string{"sano", "SANO"}, []string{
			"«sano» no es un valor de Estado. Escribe uno de estos: SANO, DORMIDO.", "SANO",
		}},
		{"posible roca vacío es fantasma", tPosible(tRoca), []string{""}, []string{"fantasma"}},
		{"posible planta vacío es texto vacío", tPosible(tPlanta), []string{""}, []string{"[]"}},
	}
	for _, c := range casos {
		p := programa(
			especie("Estado", "SANO", "DORMIDO"),
			combate(
				dato(c.tipo, "x", nil),
				capturar(id("x"), "¿Valor? "),
				// Para ver el texto vacío se grita x dentro de un equipo,
				// excepto si x es fantasma.
				gritarCaptura(c.tipo),
			),
		)
		es, err := correr(t, p, c.entradas...)
		if err != nil {
			t.Errorf("%s: error inesperado: %v", c.nombre, err)
			continue
		}
		if !reflect.DeepEqual(es.Salida, c.salida) {
			t.Errorf("%s:\n  salida = %q\n  want     %q", c.nombre, es.Salida, c.salida)
		}
		if len(es.Prompts) != len(c.entradas) {
			t.Errorf("%s: preguntó %d veces, want %d", c.nombre, len(es.Prompts), len(c.entradas))
		}
	}
}

// gritarCaptura muestra la x capturada. En posible planta la muestra como
// el tamaño de un equipo vacío si x es "", para poder distinguirla.
func gritarCaptura(t *ast.TipoExpr) ast.Instr {
	if t.Posible && t.Simple == token.PLANTA {
		return si(bin(token.IGUAL, id("x"), planta("")), []ast.Instr{gritar(planta("[]"))}, nil, []ast.Instr{gritar(id("x"))})
	}
	return gritar(id("x"))
}

func TestCapturarEnUnCampoDeFicha(t *testing.T) {
	p := programa(
		ficha("Pokemon", tPlanta, "nombre", tRoca, "vida"),
		combate(
			dato(tNombre("Pokemon"), "p", llaves(id("nombre"), planta("?"), id("vida"), roca(1))),
			capturar(campo(id("p"), "vida"), "¿Vida? "),
			gritar(id("p")),
		),
	)
	es, err := correr(t, p, "80")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{`{nombre: "?", vida: 80}`}; !reflect.DeepEqual(es.Salida, want) {
		t.Errorf("salida = %q, want %q", es.Salida, want)
	}
}

func TestDetenerLaEjecucion(t *testing.T) {
	t.Run("entrada cerrada mientras espera capturar", func(t *testing.T) {
		_, err := correr(t, programa(combate(dato(tRoca, "x", nil), capturar(id("x"), "? "))))
		if !errors.Is(err, ErrEntradaCerrada) {
			t.Errorf("err = %v, want ErrEntradaCerrada", err)
		}
	})
	t.Run("ciclo infinito se detiene con el contexto", func(t *testing.T) {
		ctx, cancelar := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancelar()
		in := Nuevo(NuevaESMemoria())
		err := in.Ejecutar(ctx, programa(combate(mientras(electrico(true)))))
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("err = %v, want context.DeadlineExceeded", err)
		}
	})
}

func TestProgramaSinCombate(t *testing.T) {
	_, err := correr(t, programa(medalla(tRoca, "X", roca(1))))
	if d := errorEjecucion(t, err); d.Code != CodigoInterno {
		t.Errorf("código = %s", d.Code)
	}
}

// Lo que el analizador rechazará llega como error interno.
func TestInstruccionesInvalidas(t *testing.T) {
	casos := []struct {
		nombre string
		cuerpo []ast.Instr
	}{
		{"reasignar una medalla", []ast.Instr{medallaLocal(tRoca, "M", roca(1)), asignacion(id("M"), roca(2))}},
		{"asignar a la variable de recorrer", []ast.Instr{rango("n", roca(1), roca(2), asignacion(id("n"), roca(5)))}},
		{"leer un dato sin valor", []ast.Instr{dato(tRoca, "x", nil), gritar(id("x"))}},
		{"entregar dentro de combate", []ast.Instr{entregar(nil)}},
		{"huir fuera de un ciclo", []ast.Instr{huir()}},
		{"declarar dos veces en el mismo bloque", []ast.Instr{dato(tRoca, "x", roca(1)), dato(tRoca, "x", roca(2))}},
		{"condición que no es electrico", []ast.Instr{si(roca(1), nil, nil, nil)}},
		{"capturar a un equipo", []ast.Instr{dato(tEquipo(tRoca), "e", equipoLit()), capturar(id("e"), "?")}},
	}
	for _, c := range casos {
		_, err := soloCombate(t, c.cuerpo...)
		if d := errorEjecucion(t, err); d.Code != CodigoInterno {
			t.Errorf("%s: código = %s, want %s", c.nombre, d.Code, CodigoInterno)
		}
	}
}

func TestFichaYClaveInvalidasNoRompenLaEjecucion(t *testing.T) {
	p := programa(
		ficha("Pokemon", tPlanta, "nombre"),
		combate(
			dato(tNombre("Pokemon"), "p", llaves(id("nombre"), planta("a"))),
			gritar(campo(id("p"), "vida")),
		),
	)
	_, err := correr(t, p)
	if d := errorEjecucion(t, err); d.Code != CodigoInterno {
		t.Errorf("campo inexistente: código = %s", d.Code)
	}
	_, err = soloCombate(t,
		dato(tMochila(tAgua, tRoca), "m", llaves(agua(1.5), roca(1))),
	)
	if d := errorEjecucion(t, err); d.Code != CodigoInterno {
		t.Errorf("clave agua: código = %s", d.Code)
	}
}

// Si el analizador ya decidió qué es un { }, el intérprete lo respeta y
// detecta si no coincide con el tipo esperado.
func TestLlavesResueltasPorElAnalizador(t *testing.T) {
	comoMochila := llaves(planta("a"), roca(1))
	comoMochila.Resuelto = ast.LlavesMochila
	salida, err := soloCombate(t, dato(tMochila(tPlanta, tRoca), "m", comoMochila), gritar(id("m")))
	if err != nil || len(salida) != 1 || salida[0] != `{"a": 1}` {
		t.Errorf("mochila resuelta: %q, %v", salida, err)
	}

	contradictorio := llaves(planta("a"), roca(1))
	contradictorio.Resuelto = ast.LlavesFicha
	_, err = soloCombate(t, dato(tMochila(tPlanta, tRoca), "m", contradictorio))
	if d := errorEjecucion(t, err); d.Code != CodigoInterno {
		t.Errorf("código = %s, want %s", d.Code, CodigoInterno)
	}
}
