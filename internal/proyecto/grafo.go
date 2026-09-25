package proyecto

import (
	"fmt"
	"strings"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/diag"
)

// Estados del recorrido en profundidad.
const (
	sinVisitar = iota
	enCurso    // está en la pila del recorrido: volver a él es un ciclo
	terminado
)

// ordenar recorre las importaciones desde el principal. Devuelve los
// archivos alcanzables con cada uno después de los que importa, y un
// diagnóstico por cada ciclo encontrado, con la cadena completa
// (sección 4, pasada 1, punto 2).
func (p *Proyecto) ordenar() ([]string, []diag.Diagnostic) {
	estado := map[string]int{}
	var (
		orden []string
		pila  []string
		diags []diag.Diagnostic
	)

	var visitar func(n string)
	visitar = func(n string) {
		estado[n] = enCurso
		pila = append(pila, n)
		for _, dep := range p.Archivos[n].Dependencias {
			switch estado[dep] {
			case sinVisitar:
				visitar(dep)
			case enCurso:
				diags = append(diags, p.errorCiclo(pila, n, dep))
			}
		}
		pila = pila[:len(pila)-1]
		estado[n] = terminado
		orden = append(orden, n)
	}
	visitar(p.Principal)
	return orden, diags
}

// errorCiclo arma el diagnóstico de la importación desde que cierra el
// ciclo hacia dep. La cadena va desde dep, pasando por la pila, hasta dep.
func (p *Proyecto) errorCiclo(pila []string, desde, dep string) diag.Diagnostic {
	inicio := 0
	for i, n := range pila {
		if n == dep {
			inicio = i
			break
		}
	}
	cadena := append(append([]string{}, pila[inicio:]...), dep)

	pos := ast.Pos{Line: 1, Col: 1, Len: 1}
	for _, im := range p.Archivos[desde].Programa.Importaciones {
		if im.Ruta == dep {
			pos = im.RutaPos
			break
		}
	}

	desc := fmt.Sprintf("las importaciones forman un ciclo: %s.", strings.Join(cadena, " → "))
	if len(cadena) == 2 {
		desc = fmt.Sprintf("«%s» se importa a sí mismo.", dep)
	}
	return errorImportacion(desde, pos, "importacion-circular", desc,
		"cada archivo del ciclo necesita que el otro se cargue primero, así que ninguno puede cargarse.",
		"mueve lo que ambos necesitan a un tercer archivo que los dos importen, o quita una de las importaciones.")
}
