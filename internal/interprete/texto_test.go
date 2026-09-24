package interprete

import "testing"

func TestTexto(t *testing.T) {
	mochila := NuevaMochila()
	mochila.Poner("agua", int64(3))
	mochila.Poner('x', 1.5)

	casos := []struct {
		nombre string
		valor  Value
		want   string
	}{
		{"roca", int64(-5), "-5"},
		{"agua entera lleva punto", 12.0, "12.0"},
		{"agua con decimales", 3.14, "3.14"},
		{"agua negativa", -0.5, "-0.5"},
		{"agua sin error de redondeo visible", 0.1, "0.1"},
		{"menos cero se muestra como cero", negativoCero(), "0.0"},
		{"agua grande sin notación científica", 1e21, "1000000000000000000000.0"},
		{"fuego suelto sin comillas", 'a', "a"},
		{"planta suelta sin comillas", "Pikachu", "Pikachu"},
		{"planta suelta no se escapa", "dijo \"hola\"\n", "dijo \"hola\"\n"},
		{"verdadero", true, "verdadero"},
		{"falso", false, "falso"},
		{"fantasma", nil, "fantasma"},
		{"especie", estado("DORMIDO"), "DORMIDO"},
		{"equipo de planta", NuevoEquipo("Pikachu", "Charizard"), `["Pikachu", "Charizard"]`},
		{"equipo vacío", NuevoEquipo(), "[]"},
		{"equipo de fuego", NuevoEquipo('a', '\''), `['a', '\'']`},
		{"escapes dentro de colección", NuevoEquipo("a\"b\n\\"), `["a\"b\n\\"]`},
		{"comilla simple en planta no se escapa", NuevoEquipo("it's"), `["it's"]`},
		{"mochila en orden de inserción", mochila, `{"agua": 3, 'x': 1.5}`},
		{"mochila vacía", NuevaMochila(), "{}"},
		{"ficha en orden de declaración", nuevoPokemon("Bulbi", 100), `{nombre: "Bulbi", vida: 100, estado: SANO}`},
		{"equipo de fichas", NuevoEquipo(nuevoPokemon("A", 1)), `[{nombre: "A", vida: 1, estado: SANO}]`},
		{"equipo con agua", NuevoEquipo(1.0, 2.5), "[1.0, 2.5]"},
	}
	for _, c := range casos {
		if got := Texto(c.valor); got != c.want {
			t.Errorf("%s: Texto = %q, want %q", c.nombre, got, c.want)
		}
	}
}

// negativoCero devuelve -0.0 sin que el compilador lo simplifique a 0.0.
func negativoCero() float64 {
	cero := 0.0
	return -cero
}
