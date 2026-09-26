package analizador

import (
	"fmt"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/tipos"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// J5 · posible (sección 4.1 y validación 11):
//
//   - Un dato posible se puede comparar con igual o diferente, mostrar con
//     gritar, guardar en otro posible o reemplazar con sino. Para cualquier
//     otra cosa hay que comprobar antes que no es fantasma.
//   - Estrechamiento: «x diferente fantasma» comprueba x en la rama
//     afirmativa; «x igual fantasma», en el sino (y en los sino si que
//     siguen). Se propaga a través de y, no de o; no invierte lo que se
//     sabe. Se pierde si el dato se reasigna dentro del bloque.
//
// Se engancha al recorrido de tipos_expr.go, que lleva los ámbitos y aplica
// estos hechos a cada bloque.
func init() {
	hechosDe = hechosDeCondicion
	posibleSinComprobar = reportarPosible
}

// hechosDeCondicion devuelve los datos posible que quedan comprobados si la
// condición es verdadera y si es falsa.
func hechosDeCondicion(v *verificador, cond ast.Expr) (siVerdadera, siFalsa []string) {
	switch x := cond.(type) {
	case *ast.Binaria:
		switch x.Op {
		case token.DIFERENTE:
			if n := v.comparadoConFantasma(x); n != "" {
				return []string{n}, nil
			}
		case token.IGUAL:
			if n := v.comparadoConFantasma(x); n != "" {
				return nil, []string{n}
			}
		case token.Y:
			// Si «A y B» es verdadera, lo son las dos. Si es falsa, no se sabe
			// cuál falló.
			a, _ := hechosDeCondicion(v, x.Izq)
			b, _ := hechosDeCondicion(v, x.Der)
			return append(a, b...), nil
		}
	case *ast.Unaria:
		if x.Op == token.NO {
			verdad, falso := hechosDeCondicion(v, x.Operando)
			return falso, verdad
		}
	}
	return nil, nil
}

// comparadoConFantasma reconoce «x igual fantasma» o «fantasma igual x» con
// x un dato posible, y devuelve su nombre.
func (v *verificador) comparadoConFantasma(x *ast.Binaria) string {
	id, ok := x.Izq.(*ast.Ident)
	otro := x.Der
	if !ok {
		id, ok = x.Der.(*ast.Ident)
		otro = x.Izq
	}
	if _, esFantasma := otro.(*ast.LitFantasma); !ok || !esFantasma {
		return ""
	}
	if l := v.buscarLocal(id.Nombre); l != nil && l.tipo != nil && l.tipo.Opcional {
		return id.Nombre
	}
	return ""
}

// reportarPosible explica un posible usado donde hace falta su valor
// (validación 11).
func reportarPosible(v *verificador, e ast.Expr, t *tipos.Type) {
	quien, ejemplo := "este valor", "x"
	if id, ok := e.(*ast.Ident); ok {
		quien, ejemplo = "«"+id.Nombre+"»", id.Nombre
	}
	v.error(e.Posicion(), "posible-sin-comprobar", diag.EncabezadoSinValor,
		fmt.Sprintf("%s es de tipo %s y puede ser fantasma; hay que comprobarlo antes de usar su valor.", quien, t),
		"un dato posible puede no tener valor, y usarlo así fallaría justo cuando es fantasma.",
		fmt.Sprintf("compruébalo antes («si %s diferente fantasma … fin») o da un reemplazo («%s sino …»).", ejemplo, ejemplo))
}
