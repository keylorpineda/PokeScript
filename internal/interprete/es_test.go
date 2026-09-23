package interprete

import (
	"errors"
	"reflect"
	"testing"
)

var _ ES = (*ESMemoria)(nil)

func TestESMemoria(t *testing.T) {
	es := NuevaESMemoria("Pikachu", "12")

	es.Escribir("¡Hola!")
	nombre, err := es.Leer("¿Nombre? ")
	if err != nil || nombre != "Pikachu" {
		t.Fatalf("Leer = %q, %v", nombre, err)
	}
	nivel, err := es.Leer("¿Nivel? ")
	if err != nil || nivel != "12" {
		t.Fatalf("Leer = %q, %v", nivel, err)
	}
	if _, err := es.Leer("¿Otro? "); !errors.Is(err, ErrEntradaCerrada) {
		t.Errorf("sin entradas, Leer debería devolver ErrEntradaCerrada, devolvió %v", err)
	}

	if want := []string{"¡Hola!"}; !reflect.DeepEqual(es.Salida, want) {
		t.Errorf("Salida = %q, want %q", es.Salida, want)
	}
	if want := []string{"¿Nombre? ", "¿Nivel? ", "¿Otro? "}; !reflect.DeepEqual(es.Prompts, want) {
		t.Errorf("Prompts = %q, want %q", es.Prompts, want)
	}
}
