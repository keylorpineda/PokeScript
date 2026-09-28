package interprete

import (
	"errors"
	"math"

	"github.com/keylorpineda/PokeScript/internal/token"
)

// Errores de las operaciones. El evaluador los convierte en diagnósticos
// con la posición y los valores involucrados.
var (
	errDesbordamiento = errors.New("el resultado no cabe en el tipo")
	errDivisionCero   = errors.New("división entre cero")
	errOperandos      = errors.New("operandos que la tabla de operaciones no admite")
)

// operarAritmetica aplica + - * / resto según la tabla 3.3:
//
//   - roca con roca da roca y se revisa el desbordamiento de int64.
//   - Si participa un agua, la roca se ensancha y el resultado es agua.
//   - / siempre da agua; resto solo acepta rocas.
//   - planta + planta concatena.
//   - Dividir entre 0 o 0.0 es error (sección 6 y decisión H9).
//   - Un agua que resulte infinita es desbordamiento, así que NaN nunca
//     aparece (decisión H9).
func operarAritmetica(op token.Kind, a, b Value) (Value, error) {
	if op == token.PLUS {
		if x, ok := a.(string); ok {
			if y, ok := b.(string); ok {
				return x + y, nil
			}
			return nil, errOperandos
		}
	}
	x, xEntero, ok := numero(a)
	if !ok {
		return nil, errOperandos
	}
	y, yEntero, ok := numero(b)
	if !ok {
		return nil, errOperandos
	}

	switch op {
	case token.SLASH:
		if esCero(b) {
			return nil, errDivisionCero
		}
		return aguaFinita(aFlotante(a) / aFlotante(b))
	case token.RESTO:
		if !xEntero || !yEntero {
			return nil, errOperandos
		}
		if y == 0 {
			return nil, errDivisionCero
		}
		return x % y, nil // MinInt64 resto -1 da 0 en Go, sin desbordar
	}

	if xEntero && yEntero {
		return operarRocas(op, x, y)
	}
	fa, fb := aFlotante(a), aFlotante(b)
	switch op {
	case token.PLUS:
		return aguaFinita(fa + fb)
	case token.MINUS:
		return aguaFinita(fa - fb)
	case token.STAR:
		return aguaFinita(fa * fb)
	}
	return nil, errOperandos
}

// operarRocas suma, resta o multiplica dos int64 revisando antes si el
// resultado se sale del rango.
func operarRocas(op token.Kind, x, y int64) (Value, error) {
	switch op {
	case token.PLUS:
		if (y > 0 && x > math.MaxInt64-y) || (y < 0 && x < math.MinInt64-y) {
			return nil, errDesbordamiento
		}
		return x + y, nil
	case token.MINUS:
		if (y < 0 && x > math.MaxInt64+y) || (y > 0 && x < math.MinInt64+y) {
			return nil, errDesbordamiento
		}
		return x - y, nil
	case token.STAR:
		if x == 0 || y == 0 {
			return int64(0), nil
		}
		r := x * y
		// Si hubo desbordamiento, dividir de vuelta no recupera el factor.
		// -1 × MinInt64 es el único caso que esa prueba no detecta.
		if r/y != x || (x == -1 && y == math.MinInt64) || (y == -1 && x == math.MinInt64) {
			return nil, errDesbordamiento
		}
		return r, nil
	}
	return nil, errOperandos
}

// negar aplica el - unario a una roca o un agua.
func negar(a Value) (Value, error) {
	switch x := a.(type) {
	case int64:
		if x == math.MinInt64 {
			return nil, errDesbordamiento
		}
		return -x, nil
	case float64:
		return -x, nil
	}
	return nil, errOperandos
}

// comparar aplica > < >= <= a dos números (roca o agua, ensanchando si se
// mezclan) o a dos fuego.
func comparar(op token.Kind, a, b Value) (bool, error) {
	var c int
	if x, ok := a.(rune); ok {
		y, ok := b.(rune)
		if !ok {
			return false, errOperandos
		}
		c = compararOrden(x, y)
	} else {
		x, xEntero, ok := numero(a)
		if !ok {
			return false, errOperandos
		}
		y, yEntero, ok := numero(b)
		if !ok {
			return false, errOperandos
		}
		if xEntero && yEntero {
			c = compararOrden(x, y)
		} else {
			c = compararOrden(aFlotante(a), aFlotante(b))
		}
	}
	switch op {
	case token.GT:
		return c > 0, nil
	case token.LT:
		return c < 0, nil
	case token.GE:
		return c >= 0, nil
	case token.LE:
		return c <= 0, nil
	}
	return false, errOperandos
}

func compararOrden[T int64 | float64 | rune](x, y T) int {
	switch {
	case x < y:
		return -1
	case x > y:
		return 1
	}
	return 0
}

// numero indica si v es roca o agua. Devuelve el valor como int64 cuando es
// roca (entero = true).
func numero(v Value) (n int64, entero bool, ok bool) {
	switch x := v.(type) {
	case int64:
		return x, true, true
	case float64:
		return 0, false, true
	}
	return 0, false, false
}

func aFlotante(v Value) float64 {
	if x, ok := v.(int64); ok {
		return float64(x)
	}
	return v.(float64)
}

func esCero(v Value) bool {
	switch x := v.(type) {
	case int64:
		return x == 0
	case float64:
		return x == 0
	}
	return false
}

func aguaFinita(f float64) (Value, error) {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return nil, errDesbordamiento
	}
	return f, nil
}

// aRoca convierte un agua ya truncada o redondeada en roca, revisando que
// quepa en int64 (decisión H10).
func aRoca(f float64) (int64, error) {
	// 2^63 se representa exacto en float64; -2^63 también.
	if math.IsNaN(f) || f >= 9223372036854775808.0 || f < -9223372036854775808.0 {
		return 0, errDesbordamiento
	}
	return int64(f), nil
}
