package analizador

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/tipos"
)

// J4 · segun (sección 4, pasada 2):
//
//	9  sobre una especie o un electrico: exhaustivo, nombrando los valores
//	   que faltan
//	10 sobre roca, agua, fuego o planta: otro obligatorio
//	18 rama inalcanzable: patrón repetido u otro redundante (advertencia)
//
// Además, cada patrón debe ser del mismo tipo que el valor y constante: un
// literal, un valor de especie o una medalla.
//
// Se engancha al recorrido de tipos_expr.go, que conoce el tipo del valor.
func init() { revisarSegun = revisarSegunCompleto }

func revisarSegunCompleto(v *verificador, s *ast.Segun, t *tipos.Type) {
	cubiertos := map[string]bool{}
	vistos := map[string]int{} // patrón → línea donde apareció primero
	for _, alt := range s.Alternativas {
		for _, p := range alt.Patrones {
			clave, ok := v.patron(p, t)
			if !ok {
				continue
			}
			if linea, repetido := vistos[clave]; repetido {
				v.c.Advertencia(v.archivo, p.Posicion(), "rama-inalcanzable",
					fmt.Sprintf("el patrón %s ya aparece en la línea %d.", clave, linea),
					"segun toma la primera rama que coincide, así que este patrón nunca se usa.",
					"quita el patrón repetido.")
				continue
			}
			vistos[clave] = p.Posicion().Line
			cubiertos[clave] = true
		}
	}
	if t == nil {
		return
	}

	var todos []string
	switch t.Kind {
	case tipos.KEspecie:
		if e := v.c.Tabla.Especies[t.Nombre]; e != nil {
			todos = e.Valores
		}
	case tipos.KElectrico:
		todos = []string{"verdadero", "falso"}
	case tipos.KRoca, tipos.KAgua, tipos.KFuego, tipos.KPlanta:
		if s.Otro == nil {
			v.error(s.Pos, "segun-sin-otro", diag.EncabezadoTipos,
				fmt.Sprintf("un segun sobre un %s necesita una rama otro.", t),
				fmt.Sprintf("un %s puede tomar muchísimos valores y las ramas no los pueden nombrar todos.", t),
				"agrega al final: otro entonces …")
		}
		return
	default:
		v.error(s.Valor.Posicion(), "segun-tipo-invalido", diag.EncabezadoTipos,
			fmt.Sprintf("segun no se puede usar sobre un valor de tipo %s.", t),
			"segun compara contra literales o valores de especie, y un "+t.String()+" no tiene ninguno.",
			"usa si … sino si … para decidir según este valor.")
		return
	}

	var faltan []string
	for _, n := range todos {
		if !cubiertos[n] {
			faltan = append(faltan, n)
		}
	}
	switch {
	case len(faltan) > 0 && s.Otro == nil:
		que := "el valor"
		if len(faltan) > 1 {
			que = "los valores"
		}
		v.error(s.Pos, "segun-incompleto", diag.EncabezadoTipos,
			fmt.Sprintf("este segun no cubre %s %s de %s.", que, strings.Join(faltan, ", "), t),
			fmt.Sprintf("un segun sobre %s debe tener una rama para cada valor posible, o una rama otro.", t),
			fmt.Sprintf("agrega una rama para %s, o una rama otro.", strings.Join(faltan, ", ")))
	case len(faltan) == 0 && s.Otro != nil:
		v.c.Advertencia(v.archivo, s.OtroPos, "rama-inalcanzable",
			"la rama otro nunca se ejecuta: las demás ramas ya cubren todos los valores de "+t.String()+".",
			"otro solo se usa cuando ninguna rama coincide.",
			"quita la rama otro.")
	}
}

// patron revisa un patrón y devuelve cómo se escribe, para detectar
// repetidos y contar los valores cubiertos. Devuelve false si el patrón no
// sirve (el error ya se reportó).
func (v *verificador) patron(p ast.Expr, t *tipos.Type) (string, bool) {
	if id, ok := p.(*ast.Ident); ok && v.buscarLocal(id.Nombre) != nil {
		v.error(id.Pos, "patron-no-constante", diag.EncabezadoTipos,
			fmt.Sprintf("«%s» es un dato, y un patrón de segun tiene que ser constante.", id.Nombre),
			"los patrones son literales, valores de especie o medallas: se comparan tal como están escritos.",
			"usa si "+id.Nombre+" igual … para comparar contra un dato.")
		return "", false
	}
	pt := v.tipo(p)
	if pt == nil {
		return "", false
	}
	if t != nil && !tipos.Iguales(pt, t) {
		v.error(p.Posicion(), "patron-incompatible", diag.EncabezadoTipos,
			fmt.Sprintf("este patrón es de tipo %s, y el valor del segun es de tipo %s.", pt, t),
			"cada patrón se compara con igual contra el valor, y solo se comparan valores del mismo tipo.",
			sugerenciaPatron(t))
		return "", false
	}
	return textoPatron(p), true
}

func sugerenciaPatron(t *tipos.Type) string {
	switch t.Kind {
	case tipos.KAgua:
		return "escribe el número con punto decimal, por ejemplo 1.0."
	case tipos.KEspecie:
		return "usa uno de los valores de " + t.Nombre + "."
	}
	return "usa un literal del mismo tipo que el valor."
}

// textoPatron escribe un patrón como aparece en el código.
func textoPatron(p ast.Expr) string {
	switch x := p.(type) {
	case *ast.Ident:
		return x.Nombre
	case *ast.LitRoca:
		return strconv.FormatInt(x.Valor, 10)
	case *ast.LitAgua:
		return strconv.FormatFloat(x.Valor, 'f', -1, 64)
	case *ast.LitFuego:
		return "'" + string(x.Valor) + "'"
	case *ast.LitPlanta:
		return strconv.Quote(x.Valor)
	case *ast.LitElectrico:
		if x.Valor {
			return "verdadero"
		}
		return "falso"
	}
	return ""
}
