package analizador

import (
	"reflect"
	"strings"
	"testing"
)

func TestNombresYAmbitos(t *testing.T) {
	casos := []struct {
		nombre   string
		archivos map[string]string
		codigos  []string
		lineas   []int
		contiene string
	}{
		// Caso semántico 4 de la sección 11.
		{"medalla local reasignada",
			programa("    medalla roca TOPE = 5\n    TOPE = 6\n    gritar TOPE"),
			[]string{"medalla-reasignada"}, []int{3}, "línea 2"},
		{"medalla de archivo reasignada",
			map[string]string{"principal.pks": "medalla roca TOPE = 5\ncombate\n    TOPE = 6\nfin\n"},
			[]string{"medalla-reasignada"}, []int{3}, "no se le puede asignar otro valor"},
		{"contenido de una medalla",
			programa("    medalla equipo de roca E = [1]\n    sumar 2 a E\n    E[1] = 3\n    quitar E[1]\n    gritar E"),
			[]string{"medalla-modificada", "medalla-modificada", "medalla-modificada"}, []int{3, 4, 5}, "contenido"},
		{"capturar en una medalla",
			programa("    medalla planta N = \"Ash\"\n    capturar(N, \"¿Nombre? \")"),
			[]string{"medalla-reasignada"}, []int{3}, "N"},

		{"huir fuera de ciclo",
			programa("    huir"),
			[]string{"corte-fuera-de-ciclo"}, []int{2}, "«huir»"},
		{"siguiente en un movimiento llamado desde un ciclo",
			map[string]string{"principal.pks": "movimiento f()\n    siguiente\nfin\ncombate\n    recorrer n de 1 hasta 3\n        f()\n    fin\nfin\n"},
			[]string{"corte-fuera-de-ciclo"}, []int{2}, "«siguiente»"},
		{"huir y siguiente dentro de ciclos",
			programa("    roca n = 0\n    mientras n < 3\n        n = n + 1\n        si n igual 2\n            siguiente\n        fin\n        segun n\n            3 entonces huir\n            otro entonces gritar n\n        fin\n    fin"),
			nil, nil, ""},

		{"dato que oculta otro de afuera",
			programa("    roca x = 1\n    si x > 0\n        roca x = 2\n        gritar x\n    fin"),
			[]string{"nombre-ocultado"}, []int{4}, "declarado en la línea 2"},
		{"dato repetido en el mismo bloque",
			programa("    roca x = 1\n    planta x = \"a\"\n    gritar x"),
			[]string{"nombre-ocultado"}, []int{3}, "un dato"},
		{"recorridos anidados con la misma variable",
			programa("    equipo de roca e = [1, 2]\n    recorrer p en e\n        recorrer p en e\n            gritar p\n        fin\n    fin"),
			[]string{"nombre-ocultado"}, []int{4}, "la variable de un recorrido"},
		{"dato que oculta un parámetro",
			map[string]string{"principal.pks": "movimiento f(roca x)\n    roca x = 2\n    gritar x\nfin\ncombate\n    f(1)\nfin\n"},
			[]string{"nombre-ocultado"}, []int{2}, "un parámetro"},
		{"dato con el nombre de un movimiento",
			map[string]string{"principal.pks": "movimiento f()\nfin\ncombate\n    roca f = 1\n    gritar f\nfin\n"},
			[]string{"nombre-ocultado"}, []int{4}, "un movimiento"},
		{"parámetro con el nombre de un valor de especie",
			map[string]string{"principal.pks": "especie Estado\n    SANO\nfin\nmovimiento f(roca SANO)\n    gritar SANO\nfin\ncombate\n    f(1)\nfin\n"},
			[]string{"nombre-ocultado"}, []int{4}, "un valor de la especie «Estado»"},
		{"el mismo nombre en bloques hermanos no se oculta",
			programa("    si verdadero\n        roca x = 1\n        gritar x\n    sino\n        roca x = 2\n        gritar x\n    fin"),
			nil, nil, ""},

		{"modificar la colección que se recorre",
			programa("    equipo de roca e = [1, 2]\n    recorrer p en e\n        sumar p a e\n        e[1] = 0\n        quitar e[1]\n    fin"),
			[]string{"coleccion-en-recorrido", "coleccion-en-recorrido", "coleccion-en-recorrido"}, []int{4, 5, 6}, "se está recorriendo"},
		{"modificar la colección de un recorrido de afuera",
			programa("    equipo de roca e = [1]\n    equipo de roca otra = [2]\n    recorrer p en e\n        recorrer q en otra\n            sumar q a e\n        fin\n    fin"),
			[]string{"coleccion-en-recorrido"}, []int{6}, "«e»"},
		{"modificar otra colección está permitido",
			programa("    equipo de roca e = [1]\n    equipo de roca copia = []\n    recorrer p en e\n        sumar p a copia\n    fin\n    gritar copia"),
			nil, nil, ""},

		{"variable de recorrido",
			programa("    equipo de roca e = [1]\n    recorrer p en e\n        p = 0\n    fin"),
			[]string{"variable-de-recorrido"}, []int{4}, "solo"},
		{"variable de un rango",
			programa("    recorrer n de 1 hasta 3\n        n = n + 1\n    fin"),
			[]string{"variable-de-recorrido"}, []int{3}, "«n»"},
		{"valor de un recorrido de mochila",
			programa("    mochila de planta a roca m = {\"a\": 1}\n    recorrer k, v en m\n        v = 0\n        gritar k\n    fin"),
			[]string{"variable-de-recorrido"}, []int{4}, "«v»"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			r := analizar(t, c.archivos)
			if got := codigos(r.Diagnosticos); !reflect.DeepEqual(got, c.codigos) {
				var detalle []string
				for _, d := range r.Diagnosticos {
					detalle = append(detalle, d.Desc)
				}
				t.Fatalf("códigos = %v, want %v\n%s", got, c.codigos, strings.Join(detalle, "\n"))
			}
			for i, d := range r.Diagnosticos {
				if d.Line != c.lineas[i] {
					t.Errorf("diagnóstico %d en la línea %d, want %d", i, d.Line, c.lineas[i])
				}
			}
			if len(r.Diagnosticos) > 0 {
				d := r.Diagnosticos[0]
				if !strings.Contains(d.Desc+" "+d.Cause+" "+d.Suggest, c.contiene) {
					t.Errorf("el mensaje no contiene %q:\n%s\n%s\n%s", c.contiene, d.Desc, d.Cause, d.Suggest)
				}
				if d.Heading == "" || d.Suggest == "" || d.Len < 1 {
					t.Errorf("diagnóstico incompleto: %+v", d)
				}
			}
		})
	}
}
