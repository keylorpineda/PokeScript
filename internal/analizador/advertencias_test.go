package analizador

import (
	"reflect"
	"strings"
	"testing"

	"github.com/keylorpineda/PokeScript/internal/diag"
)

func TestAdvertencias(t *testing.T) {
	casos := []struct {
		nombre   string
		archivos map[string]string
		codigos  []string
		lineas   []int
		contiene string
	}{
		// 20. Datos sin usar.
		{"dato sin usar",
			programa("    roca x = 1"),
			[]string{"dato-sin-usar"}, []int{2}, "nunca se usa"},
		{"asignar no es usar",
			programa("    roca x = 1\n    x = 2"),
			[]string{"dato-sin-usar"}, []int{2}, "«x»"},
		{"medalla local sin usar",
			programa("    medalla roca TOPE = 5"),
			[]string{"dato-sin-usar"}, []int{2}, "TOPE"},
		{"leer, sumar y cambiar un elemento son usos",
			programa("    roca x = 1\n    gritar x\n    equipo de roca e = []\n    sumar 1 a e\n    equipo de roca f = [1]\n    f[1] = 2"),
			nil, nil, ""},
		{"el mismo nombre en bloques hermanos",
			programa("    si verdadero\n        roca x = 1\n        gritar x\n    sino\n        roca x = 2\n    fin"),
			[]string{"dato-sin-usar"}, []int{6}, "«x»"},
		{"parámetros y variables de recorrido no se avisan",
			map[string]string{"principal.pks": "movimiento f(roca x)\nfin\ncombate\n    f(1)\n    recorrer n de 1 hasta 2\n        gritar \"hola\"\n    fin\nfin\n"},
			nil, nil, ""},

		// 21. Ciclos que no cambian.
		{"condición que no cambia",
			programa("    roca n = 0\n    mientras n < 3\n        gritar n\n    fin"),
			[]string{"ciclo-sin-cambio"}, []int{3}, "no cambia"},
		{"condición que sí cambia",
			programa("    roca n = 0\n    mientras n < 3\n        n = n + 1\n    fin"),
			nil, nil, ""},
		{"ciclo con huir",
			programa("    mientras verdadero\n        huir\n    fin"),
			nil, nil, ""},
		{"condición con aleatorio",
			programa("    mientras aleatorio(1, 6) < 6\n        gritar \"otra vez\"\n    fin"),
			nil, nil, ""},
		{"el huir de un ciclo interno no saca del externo",
			programa("    roca n = 0\n    mientras n < 3\n        recorrer i de 1 hasta 2\n            huir\n        fin\n    fin"),
			[]string{"ciclo-sin-cambio"}, []int{3}, "huir"},
		{"entregar dentro de un ciclo interno sí saca",
			map[string]string{"principal.pks": "movimiento roca f(roca n)\n    mientras n < 3\n        recorrer i de 1 hasta 2\n            entregar i\n        fin\n    fin\n    entregar 0\nfin\ncombate\n    gritar f(1)\nfin\n"},
			nil, nil, ""},

		// 22. Sangría.
		{"cuerpo sin sangría",
			map[string]string{"principal.pks": "combate\ngritar 1\nfin\n"},
			[]string{"sangria-inconsistente"}, []int{2}, "misma altura"},
		{"líneas desalineadas",
			programa("    gritar 1\n      gritar 2"),
			[]string{"sangria-inconsistente"}, []int{3}, "no está alineada"},
		{"sino y segun desalineados",
			programa("    si verdadero\n        gritar 1\n    sino\n    gritar 2\n    fin\n    segun 1\n        1 entonces gritar 1\n     otro entonces gritar 2\n    fin"),
			[]string{"sangria-inconsistente", "sangria-inconsistente"}, []int{5, 9}, "línea 4"},
		{"sangría con tabuladores y espacios parejos",
			map[string]string{"principal.pks": "combate\n\tsi verdadero\n\t\tgritar 1\n\tfin\nfin\n"},
			nil, nil, ""},

		// 23. Convención de nombres.
		{"especie, valores y ficha",
			map[string]string{"principal.pks": "especie estado\n    SANO, dormido\nfin\nficha pokemon\n    roca vida\nfin\ncombate\n    estado e = SANO\n    gritar e\nfin\n"},
			[]string{"convencion-de-nombres", "convencion-de-nombres", "convencion-de-nombres"}, []int{1, 2, 4}, "«Estado»"},
		{"medallas y datos",
			map[string]string{"principal.pks": "medalla roca tope = 5\nmovimiento f(roca PODER)\n    gritar PODER\nfin\ncombate\n    roca Vida = tope\n    gritar Vida\n    f(1)\nfin\n"},
			[]string{"convencion-de-nombres", "convencion-de-nombres", "convencion-de-nombres"}, []int{1, 2, 6}, "«TOPE»"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			r := analizarConAdvertencias(t, c.archivos)
			if got := codigos(r.Diagnosticos); !reflect.DeepEqual(got, c.codigos) {
				var detalle []string
				for _, d := range r.Diagnosticos {
					detalle = append(detalle, d.Desc)
				}
				t.Fatalf("códigos = %v, want %v\n%s", got, c.codigos, strings.Join(detalle, "\n"))
			}
			for i, d := range r.Diagnosticos {
				if d.Line != c.lineas[i] {
					t.Errorf("advertencia %d en la línea %d, want %d", i, d.Line, c.lineas[i])
				}
				if d.Severity != diag.Advertencia || d.Heading != diag.EncabezadoAdvertencia {
					t.Errorf("debe ser una advertencia: %+v", d)
				}
			}
			if len(r.Diagnosticos) > 0 {
				d := r.Diagnosticos[0]
				if !strings.Contains(d.Desc+" "+d.Cause+" "+d.Suggest, c.contiene) {
					t.Errorf("el mensaje no contiene %q:\n%s\n%s\n%s", c.contiene, d.Desc, d.Cause, d.Suggest)
				}
			}
		})
	}
}
