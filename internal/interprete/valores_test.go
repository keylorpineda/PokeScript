package interprete

import (
	"errors"
	"testing"
)

var estado = func(v string) EspecieVal { return EspecieVal{Tipo: "Estado", Valor: v} }

func nuevoPokemon(nombre string, vida int64) *Ficha {
	return NuevaFicha("Pokemon", []string{"nombre", "vida", "estado"}, map[string]Value{
		"nombre": nombre, "vida": vida, "estado": estado("SANO"),
	})
}

func TestEquipoIndicesDesdeUno(t *testing.T) {
	e := NuevoEquipo("Pikachu", "Charizard", "Squirtle")

	casos := []struct {
		indice int64
		want   Value
		error  bool
	}{
		{1, "Pikachu", false},
		{3, "Squirtle", false},
		{0, nil, true},
		{4, nil, true},
		{-1, nil, true},
	}
	for _, c := range casos {
		got, err := e.Obtener(c.indice)
		if c.error {
			var fuera *ErrFueraDeRango
			if !errors.As(err, &fuera) {
				t.Errorf("Obtener(%d): esperaba ErrFueraDeRango, dio %v", c.indice, err)
			}
			continue
		}
		if err != nil || got != c.want {
			t.Errorf("Obtener(%d) = %v, %v; want %v", c.indice, got, err, c.want)
		}
	}
}

func TestEquipoSumarYQuitar(t *testing.T) {
	e := NuevoEquipo(int64(10), int64(20), int64(30))
	e.Sumar(int64(40))
	if err := e.Quitar(2); err != nil {
		t.Fatal(err)
	}
	if got, want := literal(e), "[10, 30, 40]"; got != want {
		t.Errorf("tras sumar y quitar = %s, want %s", got, want)
	}
	if err := e.Poner(1, int64(99)); err != nil {
		t.Fatal(err)
	}
	if got, want := literal(e), "[99, 30, 40]"; got != want {
		t.Errorf("tras poner = %s, want %s", got, want)
	}
	if err := e.Quitar(4); err == nil {
		t.Error("Quitar(4) en un equipo de 3 debería fallar")
	}
}

func TestErrFueraDeRangoMensaje(t *testing.T) {
	casos := []struct {
		err  *ErrFueraDeRango
		want string
	}{
		{&ErrFueraDeRango{Indice: 4, Tamano: 3, Coleccion: "equipo"},
			"el índice 4 está fuera de rango: el equipo tiene 3 elementos, del 1 al 3"},
		{&ErrFueraDeRango{Indice: 1, Tamano: 0, Coleccion: "equipo"},
			"el índice 1 está fuera de rango: el equipo está vacío"},
		{&ErrFueraDeRango{Indice: 8, Tamano: 7, Coleccion: "texto"},
			"el índice 8 está fuera de rango: el texto tiene 7 letras, del 1 al 7"},
	}
	for _, c := range casos {
		if got := c.err.Error(); got != c.want {
			t.Errorf("Error() = %q, want %q", got, c.want)
		}
	}
}

func TestPlantaCuentaLetras(t *testing.T) {
	if got := TamanoPlanta("Pokémon"); got != 7 {
		t.Errorf("TamanoPlanta = %d, want 7", got)
	}
	letra, err := LetraDe("Pokémon", 4)
	if err != nil || letra != 'é' {
		t.Errorf("LetraDe(4) = %q, %v; want 'é'", letra, err)
	}
	if _, err := LetraDe("Pokémon", 8); err == nil {
		t.Error("LetraDe(8) debería estar fuera de rango")
	}
}

// La prueba central del paso por valor: modificar la copia no toca el
// original, a ningún nivel de anidación.
func TestCopiarEsProfunda(t *testing.T) {
	t.Run("equipo anidado", func(t *testing.T) {
		original := NuevoEquipo(NuevoEquipo(int64(1), int64(2)))
		copia := Copiar(original).(*Equipo)

		interno, _ := copia.Obtener(1)
		interno.(*Equipo).Sumar(int64(3))
		copia.Sumar(NuevoEquipo())

		if got, want := literal(original), "[[1, 2]]"; got != want {
			t.Errorf("el original cambió: %s, want %s", got, want)
		}
		if got, want := literal(copia), "[[1, 2, 3], []]"; got != want {
			t.Errorf("copia = %s, want %s", got, want)
		}
	})

	t.Run("mochila", func(t *testing.T) {
		original := NuevaMochila()
		original.Poner("equipo", NuevoEquipo("Pikachu"))
		copia := Copiar(original).(*Mochila)

		v, _ := copia.Obtener("equipo")
		v.(*Equipo).Sumar("Eevee")
		copia.Poner("nueva", int64(1))

		if got, want := literal(original), `{"equipo": ["Pikachu"]}`; got != want {
			t.Errorf("el original cambió: %s, want %s", got, want)
		}
		if original.Contiene("nueva") {
			t.Error("la clave nueva apareció en el original")
		}
	})

	t.Run("ficha", func(t *testing.T) {
		original := nuevoPokemon("Bulbi", 100)
		copia := Copiar(original).(*Ficha)
		copia.PonerCampo("vida", int64(40))

		if got := original.Campo("vida"); got != int64(100) {
			t.Errorf("la vida del original cambió a %v", got)
		}
	})

	t.Run("lo que entra a un contenedor se copia", func(t *testing.T) {
		interno := NuevoEquipo(int64(1))
		externo := NuevoEquipo()
		externo.Sumar(interno)
		interno.Sumar(int64(2))

		if got, want := literal(externo), "[[1]]"; got != want {
			t.Errorf("externo = %s, want %s", got, want)
		}
	})
}

func TestIgual(t *testing.T) {
	m1 := NuevaMochila()
	m1.Poner("agua", int64(3))
	m1.Poner("fuego", int64(1))
	m2 := NuevaMochila()
	m2.Poner("fuego", int64(1))
	m2.Poner("agua", int64(3))
	m3 := NuevaMochila()
	m3.Poner("agua", int64(3))

	casos := []struct {
		nombre string
		a, b   Value
		want   bool
	}{
		{"fantasma igual fantasma", nil, nil, true},
		{"valor contra fantasma", "Bulbi", nil, false},
		{"fantasma contra valor", nil, int64(0), false},
		{"rocas iguales", int64(5), int64(5), true},
		{"roca y fuego no chocan", int64(65), 'A', false},
		{"aguas", 2.5, 2.5, true},
		{"especie igual", estado("SANO"), estado("SANO"), true},
		{"especie distinta", estado("SANO"), estado("DORMIDO"), false},
		{"equipos iguales", NuevoEquipo("a", "b"), NuevoEquipo("a", "b"), true},
		{"equipo en otro orden", NuevoEquipo("a", "b"), NuevoEquipo("b", "a"), false},
		{"equipo de distinto tamaño", NuevoEquipo("a"), NuevoEquipo("a", "b"), false},
		{"equipos vacíos", NuevoEquipo(), NuevoEquipo(), true},
		{"mochila en otro orden", m1, m2, true},
		{"mochila con menos pares", m1, m3, false},
		{"fichas iguales", nuevoPokemon("Bulbi", 100), nuevoPokemon("Bulbi", 100), true},
		{"fichas distintas", nuevoPokemon("Bulbi", 100), nuevoPokemon("Bulbi", 99), false},
		{"equipo contra mochila", NuevoEquipo(), NuevaMochila(), false},
	}
	for _, c := range casos {
		if got := Igual(c.a, c.b); got != c.want {
			t.Errorf("%s: Igual = %v, want %v", c.nombre, got, c.want)
		}
	}
}

func TestNuevaFichaIncompletaEsErrorInterno(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("una ficha sin todos sus campos debería detener la ejecución")
		}
	}()
	NuevaFicha("Pokemon", []string{"nombre", "vida"}, map[string]Value{"nombre": "Bulbi"})
}
