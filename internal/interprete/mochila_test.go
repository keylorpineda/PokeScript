package interprete

import (
	"errors"
	"reflect"
	"testing"
)

func TestMochilaConservaElOrden(t *testing.T) {
	m := NuevaMochila()
	m.Poner("Pikachu", int64(25))
	m.Poner("Eevee", int64(133))
	m.Poner("Abra", int64(63))

	if got, want := m.Claves(), []Value{"Pikachu", "Eevee", "Abra"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Claves = %v, want %v", got, want)
	}
}

func TestMochilaPoner(t *testing.T) {
	casos := []struct {
		nombre string
		pasos  func(t *testing.T, m *Mochila)
		want   string
	}{
		{
			"reemplazar conserva la posición",
			func(t *testing.T, m *Mochila) { m.Poner("b", int64(20)) },
			`{"a": 1, "b": 20, "c": 3}`,
		},
		{
			"clave nueva va al final",
			func(t *testing.T, m *Mochila) { m.Poner("d", int64(4)) },
			`{"a": 1, "b": 2, "c": 3, "d": 4}`,
		},
		{
			"quitar y volver a poner la manda al final",
			func(t *testing.T, m *Mochila) {
				if err := m.Quitar("a"); err != nil {
					t.Fatal(err)
				}
				m.Poner("a", int64(1))
			},
			`{"b": 2, "c": 3, "a": 1}`,
		},
		{
			"quitar del medio mantiene el resto en orden",
			func(t *testing.T, m *Mochila) {
				if err := m.Quitar("b"); err != nil {
					t.Fatal(err)
				}
				m.Poner("c", int64(30))
			},
			`{"a": 1, "c": 30}`,
		},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			m := NuevaMochila()
			m.Poner("a", int64(1))
			m.Poner("b", int64(2))
			m.Poner("c", int64(3))
			c.pasos(t, m)
			if got := literal(m); got != c.want {
				t.Errorf("mochila = %s, want %s", got, c.want)
			}
		})
	}
}

func TestMochilaClaveInexistente(t *testing.T) {
	m := NuevaMochila()
	m.Poner("agua", int64(1))

	_, err := m.Obtener("fuego")
	var falta *ErrClaveInexistente
	if !errors.As(err, &falta) {
		t.Fatalf("Obtener: esperaba ErrClaveInexistente, dio %v", err)
	}
	if got, want := err.Error(), `la clave "fuego" no existe en la mochila`; got != want {
		t.Errorf("mensaje = %q, want %q", got, want)
	}
	if err := m.Quitar("fuego"); !errors.As(err, &falta) {
		t.Errorf("Quitar: esperaba ErrClaveInexistente, dio %v", err)
	}
}

func TestMochilaClavesDeDistintoTipoNoChocan(t *testing.T) {
	m := NuevaMochila()
	m.Poner(int64(65), "roca")
	m.Poner('A', "fuego")
	m.Poner(EspecieVal{Tipo: "Estado", Valor: "SANO"}, "especie")

	if m.Tamano() != 3 {
		t.Fatalf("Tamano = %d, want 3", m.Tamano())
	}
	if v, _ := m.Obtener('A'); v != "fuego" {
		t.Errorf("Obtener('A') = %v, want fuego", v)
	}
	if !m.Contiene(EspecieVal{Tipo: "Estado", Valor: "SANO"}) {
		t.Error("no encontró la clave de especie")
	}
}

func TestMochilaClavesDevuelveCopia(t *testing.T) {
	m := NuevaMochila()
	m.Poner("a", int64(1))
	claves := m.Claves()
	m.Poner("b", int64(2))
	if len(claves) != 1 {
		t.Errorf("el recorrido cambió al modificar la mochila: %v", claves)
	}
}

func TestMochilaClaveInvalidaEsErrorInterno(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("una ficha como clave debería detener la ejecución (decisión H3)")
		}
	}()
	NuevaMochila().Poner(nuevoPokemon("Bulbi", 1), int64(1))
}
