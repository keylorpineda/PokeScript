package asistente

import (
	"sort"
	"unicode/utf8"
)

// Distancia es la distancia de edición de Levenshtein entre a y b: cuántas
// letras hay que insertar, borrar o cambiar para pasar de una a otra. Cuenta
// letras (runas), no bytes, así que «año» y «ano» están a distancia 1.
func Distancia(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	previa := make([]int, len(rb)+1)
	actual := make([]int, len(rb)+1)
	for j := range previa {
		previa[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		actual[0] = i
		for j := 1; j <= len(rb); j++ {
			costo := 1
			if ra[i-1] == rb[j-1] {
				costo = 0
			}
			actual[j] = min(previa[j]+1, actual[j-1]+1, previa[j-1]+costo)
		}
		previa, actual = actual, previa
	}
	return previa[len(rb)]
}

// Parecido devuelve el candidato más cercano a palabra, si está lo bastante
// cerca para ser un error de escritura (sección 7.1): distancia hasta 2, o
// hasta un tercio de la longitud en palabras largas. Nunca propone cambiar
// todas las letras, así que para palabras de una o dos letras exige
// distancia 1. Si hay empate, gana el primero en orden alfabético, para que
// la sugerencia sea siempre la misma.
func Parecido(palabra string, candidatos []string) (string, bool) {
	largo := utf8.RuneCountInString(palabra)
	limite := max(2, largo/3)
	if largo <= 2 {
		limite = 1
	}
	ordenados := append([]string{}, candidatos...)
	sort.Strings(ordenados)

	mejor, mejorDist := "", limite+1
	for _, c := range ordenados {
		if c == palabra {
			continue
		}
		if d := Distancia(palabra, c); d < mejorDist {
			mejor, mejorDist = c, d
		}
	}
	return mejor, mejor != ""
}
