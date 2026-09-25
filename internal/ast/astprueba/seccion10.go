package astprueba

import (
	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// Seccion10 devuelve el programa de ejemplo de la sección 10 de la
// especificación, un árbol por archivo, tal como el parser debe producirlo
// al leer internal/lexer/testdata/*.pks. Cada llamada arma árboles nuevos.
func Seccion10() map[string]*ast.Programa {
	return map[string]*ast.Programa{
		"constantes.pks":  constantes(),
		"tipos.pks":       tiposPks(),
		"operaciones.pks": operaciones(),
		"principal.pks":   principal(),
	}
}

// Seccion10UnArchivo junta los cuatro archivos en uno, sin importaciones,
// para ejecutarlo mientras no existe internal/proyecto (hito 9).
func Seccion10UnArchivo() *ast.Programa {
	p := Programa("principal.pks", nil)
	for _, nombre := range []string{"constantes.pks", "tipos.pks", "operaciones.pks", "principal.pks"} {
		p.Declaraciones = append(p.Declaraciones, Seccion10()[nombre].Declaraciones...)
	}
	return p
}

func constantes() *ast.Programa {
	return Programa("constantes.pks", nil,
		Medalla(TRoca(), "VIDA_MAXIMA", Roca(100)),
		Medalla(TRoca(), "NIVEL", Roca(25)),
	)
}

func tiposPks() *ast.Programa {
	return Programa("tipos.pks", nil,
		Especie("Estado", "SANO", "ENVENENADO", "DORMIDO", "PARALIZADO"),
		Ficha("Pokemon", TPlanta(), "nombre", TRoca(), "vida", TNombre("Estado"), "estado"),
	)
}

func operaciones() *ast.Programa {
	return Programa("operaciones.pks",
		[]*ast.Importacion{Importar("constantes.pks", "NIVEL")},
		Movimiento(TRoca(), "calcular_dano", []any{TRoca(), "poder"},
			Dato(TAgua(), "base", Bin(token.SLASH, Bin(token.STAR, Id("poder"), Id("NIVEL")), Roca(50))),
			Dato(TRoca(), "variacion", Aleatorio(Roca(85), Roca(100))),
			Dato(TAgua(), "total", Bin(token.SLASH, Bin(token.STAR, Id("base"), Id("variacion")), Roca(100))),
			Entregar(Bin(token.PLUS, Redondear(Id("total")), Roca(2))),
		),
	)
}

func principal() *ast.Programa {
	vida := func(quien string) ast.Expr { return Campo(Id(quien), "vida") }
	nombre := func(quien string) ast.Expr { return Campo(Id(quien), "nombre") }

	describir := Movimiento(nil, "describir", []any{TNombre("Estado"), "actual"},
		Segun(Id("actual"), nil,
			Alternativa(Gritar(Planta("  Puede atacar")), Id("SANO")),
			Alternativa(Gritar(Planta("  Pierde vida cada turno")), Id("ENVENENADO")),
			Alternativa(Gritar(Planta("  Podría no atacar")), Id("DORMIDO"), Id("PARALIZADO")),
		))

	combate := Combate(
		Dato(TPlanta(), "nombre", nil),
		Capturar(Id("nombre"), "¿Cómo se llama tu Pokemon? "),

		Dato(TNombre("Pokemon"), "mio", Llaves(Id("nombre"), Id("nombre"), Id("vida"), Id("VIDA_MAXIMA"), Id("estado"), Id("SANO"))),
		Dato(TNombre("Pokemon"), "rival", Llaves(Id("nombre"), Planta("Bulbi"), Id("vida"), Id("VIDA_MAXIMA"), Id("estado"), Id("SANO"))),

		Gritar(Planta("¡"), nombre("mio"), Planta(" entra en combate!")),
		Instr(Llamada("describir", Campo(Id("mio"), "estado"))),

		Rango("turno", Roca(1), Roca(20),
			Gritar(Planta("--- Turno "), Id("turno"), Planta(" ---")),

			Dato(TRoca(), "dano", Llamada("calcular_dano", Roca(40))),
			Asignacion(vida("rival"), Bin(token.MINUS, vida("rival"), Id("dano"))),
			Gritar(nombre("mio"), Planta(" ataca y hace "), Id("dano"), Planta(" de daño")),

			Si(Bin(token.LE, vida("rival"), Roca(0)), []ast.Instr{
				Gritar(nombre("rival"), Planta(" se debilitó. ¡Ganaste!")),
				Huir(),
			}, nil, nil),

			Dato(TRoca(), "contra", Llamada("calcular_dano", Roca(35))),
			Asignacion(vida("mio"), Bin(token.MINUS, vida("mio"), Id("contra"))),
			Gritar(nombre("rival"), Planta(" contraataca: "), Id("contra"), Planta(" de daño")),
			Gritar(Planta("Vida de "), nombre("mio"), Planta(": "), vida("mio")),

			Si(Bin(token.LE, vida("mio"), Roca(0)), []ast.Instr{
				Gritar(nombre("mio"), Planta(" se debilitó. Perdiste.")),
				Huir(),
			}, nil, nil),
		),
	)

	return Programa("principal.pks",
		[]*ast.Importacion{
			Importar("operaciones.pks", "calcular_dano"),
			Importar("tipos.pks", "Estado", "Pokemon"),
			Importar("constantes.pks", "VIDA_MAXIMA"),
		},
		describir,
		combate,
	)
}
