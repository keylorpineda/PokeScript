package ast_test

import (
	"reflect"
	"testing"

	"github.com/keylorpineda/PokeScript/internal/ast"
	. "github.com/keylorpineda/PokeScript/internal/ast/astprueba"
	"github.com/keylorpineda/PokeScript/internal/parser"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// TestInspeccionarVisitaTodasLasHojas compara las hojas que encuentra
// Inspeccionar con las que calcula astprueba.Hojas, que a su vez se
// comprueba contra los tokens del lexer: así se sabe que no se salta nodos.
func TestInspeccionarVisitaTodasLasHojas(t *testing.T) {
	for archivo, prog := range Seccion10() {
		var got []string
		ast.InspeccionarPrograma(prog, func(n ast.Nodo) bool {
			switch x := n.(type) {
			case *ast.Ident:
				got = append(got, x.Nombre)
			case *ast.TipoExpr:
				if x.Forma == ast.TipoNombrado {
					got = append(got, x.Nombre)
				}
			case *ast.Importacion:
				got = append(got, x.Ruta)
			case *ast.LitRoca, *ast.LitAgua, *ast.LitFuego, *ast.LitPlanta:
				got = append(got, Hojas(x)...)
			}
			return true
		})
		// Hojas pone la ruta de la importación después de sus nombres.
		if want := Hojas(prog); len(got) != len(want) {
			t.Errorf("%s: Inspeccionar vio %d hojas y Hojas cuenta %d\n got: %v\nwant: %v", archivo, len(got), len(want), got, want)
		}
	}
}

func TestInspeccionarSinEntrarEnLosHijos(t *testing.T) {
	// Un si con una división adentro: si f devuelve false en el Si, la
	// división no se visita.
	prog := Programa("p.pks", nil, Combate(
		Si(Electrico(true), []ast.Instr{Gritar(Bin(token.SLASH, Roca(1), Roca(0)))}, nil, nil),
		Gritar(Bin(token.SLASH, Roca(2), Roca(0))),
	))
	var divisiones int
	ast.InspeccionarPrograma(prog, func(n ast.Nodo) bool {
		if _, esSi := n.(*ast.Si); esSi {
			return false
		}
		if b, ok := n.(*ast.Binaria); ok && b.Op == token.SLASH {
			divisiones++
		}
		return true
	})
	if divisiones != 1 {
		t.Errorf("se esperaba ver 1 división fuera del si, se vieron %d", divisiones)
	}
}

func TestInspeccionarSaltaHijosNil(t *testing.T) {
	var tipos []string
	mov := Movimiento(nil, "f", nil, Entregar(nil), Recorrido("x", "", Id("e")))
	ast.Inspeccionar(mov, func(n ast.Nodo) bool {
		tipos = append(tipos, reflect.TypeOf(n).String())
		return true
	})
	want := []string{"*ast.DeclMovimiento", "*ast.Ident", "*ast.Entregar", "*ast.RecorrerColeccion", "*ast.Ident", "*ast.Ident"}
	if !reflect.DeepEqual(tipos, want) {
		t.Errorf("visitados = %v\nwant       %v", tipos, want)
	}
	ast.InspeccionarPrograma(nil, func(ast.Nodo) bool { t.Error("no debe visitar nada"); return true })
}

// todo usa todas las construcciones del lenguaje.
const todo = `enseñar g desde "otro.pks"
medalla roca M = 1
especie E
    A, B
fin
ficha F
    roca x
fin
movimiento roca f(equipo de roca e, mochila de planta a roca m)
    entregar tamaño(e)
fin
combate
    posible planta p = fantasma
    roca n
    n = -1 + aleatorio(1, 2) * redondear(2.5)
    gritar p sino "x", 'c', verdadero, [1], {"k": 1}, convertir(n) a planta
    capturar(n, "?")
    f([1], {})
    sumar 1 a lista
    quitar lista[1]
    si no n > 1 y n contiene 2
        huir
    sino si n igual 3
        siguiente
    sino
        entregar
    fin
    segun n
        1, 2 entonces gritar 1
        otro entonces gritar 2
    fin
    mientras n < 3
        n = n + 1
    fin
    recorrer i de 1 hasta 3
        gritar i
    fin
    recorrer k, v en m
        r.x[1] = k
    fin
fin`

func TestInspeccionarVisitaTodosLosTiposDeNodo(t *testing.T) {
	r := parser.Analizar("todo.pks", todo)
	if len(r.Diagnosticos) > 0 {
		t.Fatalf("el programa de prueba tiene errores: %+v", r.Diagnosticos[0])
	}
	vistos := map[string]bool{}
	ast.InspeccionarPrograma(r.Programa, func(n ast.Nodo) bool {
		vistos[reflect.TypeOf(n).Elem().Name()] = true
		return true
	})
	for _, tipo := range []string{
		"Importacion", "TipoExpr", "DeclMedalla", "DeclEspecie", "DeclFicha", "Campo", "DeclMovimiento", "Param", "Combate",
		"DeclDato", "Asignacion", "Gritar", "Capturar", "LlamadaInstr", "Sumar", "Quitar", "Si", "RamaSi", "Segun",
		"Alternativa", "Mientras", "RecorrerColeccion", "RecorrerRango", "Huir", "Siguiente", "Entregar",
		"Ident", "LitRoca", "LitAgua", "LitFuego", "LitPlanta", "LitElectrico", "LitFantasma", "LitEquipo", "LitLlaves",
		"Par", "Binaria", "Unaria", "Respaldo", "Indice", "CampoAcceso", "Llamada", "Convertir", "Tamano", "Aleatorio", "Redondear",
	} {
		if !vistos[tipo] {
			t.Errorf("Inspeccionar no visitó ningún %s", tipo)
		}
	}
}
