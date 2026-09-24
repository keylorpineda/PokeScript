package interprete

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Texto convierte un valor en el texto que muestra gritar (decisión H1).
//
// Dentro de una colección los valores se muestran como literales de
// PokeScript. Sueltos, fuego y planta se muestran como texto plano, sin
// comillas, porque gritar los concatena: gritar "¡", nombre, "!" escribe
// ¡Bulbi!.
func Texto(v Value) string {
	switch x := v.(type) {
	case rune:
		return string(x)
	case string:
		return x
	default:
		return literal(v)
	}
}

// literal escribe un valor con la forma de un literal de PokeScript.
func literal(v Value) string {
	switch x := v.(type) {
	case nil:
		return "fantasma"
	case int64:
		return strconv.FormatInt(x, 10)
	case float64:
		return textoAgua(x)
	case rune:
		return "'" + escapar(string(x), '\'') + "'"
	case string:
		return `"` + escapar(x, '"') + `"`
	case bool:
		if x {
			return "verdadero"
		}
		return "falso"
	case EspecieVal:
		return x.Valor
	case *Equipo:
		partes := make([]string, len(x.elems))
		for i, e := range x.elems {
			partes[i] = literal(e)
		}
		return "[" + strings.Join(partes, ", ") + "]"
	case *Mochila:
		partes := make([]string, len(x.claves))
		for i, k := range x.claves {
			partes[i] = literal(k) + ": " + literal(x.valores[i])
		}
		return "{" + strings.Join(partes, ", ") + "}"
	case *Ficha:
		partes := make([]string, len(x.campos))
		for i, c := range x.campos {
			partes[i] = c + ": " + literal(x.valores[c])
		}
		return "{" + strings.Join(partes, ", ") + "}"
	default:
		panic(fmt.Sprintf("error interno: valor de tipo %T no pertenece a PokeScript", v))
	}
}

// textoAgua escribe un agua siempre con punto decimal, para que se distinga
// de una roca: 12.0, 3.14, -0.5. Usa la menor cantidad de decimales que
// representa el valor exacto y nunca notación científica.
func textoAgua(f float64) string {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		// Las operaciones de agua detienen la ejecución antes de producir
		// infinito o NaN (decisión H9), así que nunca debería llegar aquí.
		panic(fmt.Sprintf("error interno: agua no finita: %v", f))
	}
	if f == 0 {
		f = 0 // -0.0 se muestra como 0.0
	}
	s := strconv.FormatFloat(f, 'f', -1, 64)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}

// escapar aplica los escapes de los literales (sección 1.4 y registro de
// decisiones): la comilla que cierra el literal, \n y \\.
func escapar(s string, comilla rune) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case comilla:
			b.WriteRune('\\')
			b.WriteRune(comilla)
		case '\n':
			b.WriteString(`\n`)
		case '\\':
			b.WriteString(`\\`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
