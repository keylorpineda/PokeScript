package interprete

import (
	"errors"
	"math"
	"testing"

	"github.com/keylorpineda/PokeScript/internal/token"
)

func TestOperarAritmetica(t *testing.T) {
	casos := []struct {
		nombre string
		op     token.Kind
		a, b   Value
		want   Value
		err    error
	}{
		{"suma de rocas", token.PLUS, int64(2), int64(3), int64(5), nil},
		{"roca más agua se ensancha", token.PLUS, int64(2), 0.5, 2.5, nil},
		{"agua menos roca", token.MINUS, 1.5, int64(1), 0.5, nil},
		{"producto de rocas", token.STAR, int64(-4), int64(3), int64(-12), nil},
		{"división siempre da agua", token.SLASH, int64(7), int64(2), 3.5, nil},
		{"división exacta también da agua", token.SLASH, int64(8), int64(2), 4.0, nil},
		{"resto", token.RESTO, int64(7), int64(3), int64(1), nil},
		{"resto con negativo", token.RESTO, int64(-7), int64(3), int64(-1), nil},
		{"resto del mínimo entre -1", token.RESTO, int64(math.MinInt64), int64(-1), int64(0), nil},
		{"concatenar planta", token.PLUS, "Pika", "chu", "Pikachu", nil},

		{"dividir entre roca 0", token.SLASH, int64(1), int64(0), nil, errDivisionCero},
		{"dividir entre agua 0.0", token.SLASH, 1.5, 0.0, nil, errDivisionCero},
		{"resto entre 0", token.RESTO, int64(5), int64(0), nil, errDivisionCero},

		{"suma que desborda", token.PLUS, int64(math.MaxInt64), int64(1), nil, errDesbordamiento},
		{"suma negativa que desborda", token.PLUS, int64(math.MinInt64), int64(-1), nil, errDesbordamiento},
		{"suma justo en el límite", token.PLUS, int64(math.MaxInt64 - 1), int64(1), int64(math.MaxInt64), nil},
		{"resta que desborda", token.MINUS, int64(math.MinInt64), int64(1), nil, errDesbordamiento},
		{"resta de negativo que desborda", token.MINUS, int64(math.MaxInt64), int64(-1), nil, errDesbordamiento},
		{"producto que desborda", token.STAR, int64(math.MaxInt64/2 + 1), int64(2), nil, errDesbordamiento},
		{"-1 por el mínimo desborda", token.STAR, int64(-1), int64(math.MinInt64), nil, errDesbordamiento},
		{"el mínimo por -1 desborda", token.STAR, int64(math.MinInt64), int64(-1), nil, errDesbordamiento},
		{"producto por cero", token.STAR, int64(math.MinInt64), int64(0), int64(0), nil},
		{"agua que llega a infinito", token.STAR, 1e308, 10.0, nil, errDesbordamiento},
		{"división de agua que llega a infinito", token.SLASH, 1e308, 1e-10, nil, errDesbordamiento},

		{"resto con agua no existe", token.RESTO, 7.5, int64(2), nil, errOperandos},
		{"roca más electrico no existe", token.PLUS, int64(1), true, nil, errOperandos},
		{"planta menos planta no existe", token.MINUS, "a", "b", nil, errOperandos},
		{"planta más roca no existe", token.PLUS, "a", int64(1), nil, errOperandos},
	}
	for _, c := range casos {
		got, err := operarAritmetica(c.op, c.a, c.b)
		if !errors.Is(err, c.err) {
			t.Errorf("%s: error = %v, want %v", c.nombre, err, c.err)
			continue
		}
		if got != c.want {
			t.Errorf("%s: = %v (%T), want %v (%T)", c.nombre, got, got, c.want, c.want)
		}
	}
}

func TestNegar(t *testing.T) {
	if v, err := negar(int64(5)); err != nil || v != int64(-5) {
		t.Errorf("negar(5) = %v, %v", v, err)
	}
	if v, err := negar(2.5); err != nil || v != -2.5 {
		t.Errorf("negar(2.5) = %v, %v", v, err)
	}
	if _, err := negar(int64(math.MinInt64)); !errors.Is(err, errDesbordamiento) {
		t.Errorf("negar(MinInt64) debería desbordar, dio %v", err)
	}
	if _, err := negar("a"); !errors.Is(err, errOperandos) {
		t.Errorf("negar(planta) debería fallar, dio %v", err)
	}
}

func TestComparar(t *testing.T) {
	casos := []struct {
		op   token.Kind
		a, b Value
		want bool
	}{
		{token.GT, int64(3), int64(2), true},
		{token.LT, int64(3), int64(2), false},
		{token.GE, int64(2), int64(2), true},
		{token.LE, int64(2), 2.5, true},
		{token.GT, 2.5, int64(2), true},
		{token.LT, 'a', 'b', true},
		{token.GE, 'b', 'a', true},
		{token.GT, int64(math.MaxInt64), int64(math.MaxInt64 - 1), true},
	}
	for _, c := range casos {
		got, err := comparar(c.op, c.a, c.b)
		if err != nil || got != c.want {
			t.Errorf("%v %s %v = %v, %v; want %v", c.a, c.op, c.b, got, err, c.want)
		}
	}
	for _, par := range [][2]Value{{"a", "b"}, {'a', int64(1)}, {true, false}} {
		if _, err := comparar(token.GT, par[0], par[1]); !errors.Is(err, errOperandos) {
			t.Errorf("comparar %v y %v debería fallar", par[0], par[1])
		}
	}
}

func TestARoca(t *testing.T) {
	casos := []struct {
		f    float64
		want int64
		err  bool
	}{
		{2.0, 2, false},
		{-3.0, -3, false},
		{9223372036854774784.0, 9223372036854774784, false}, // el mayor float64 que cabe
		{9223372036854775808.0, 0, true},                    // 2^63 ya no cabe
		{-9223372036854775808.0, math.MinInt64, false},
		{-1e19, 0, true},
	}
	for _, c := range casos {
		got, err := aRoca(c.f)
		if (err != nil) != c.err || (!c.err && got != c.want) {
			t.Errorf("aRoca(%v) = %d, %v", c.f, got, err)
		}
	}
}
