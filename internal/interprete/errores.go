package interprete

import (
	"errors"
	"fmt"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/diag"
)

// ErrorEjecucion es un error que detiene el programa (sección 6). Lleva el
// diagnóstico completo, con la posición de la instrucción y los valores
// involucrados, en el mismo formato que los errores de compilación.
type ErrorEjecucion struct {
	Diag diag.Diagnostic
}

func (e *ErrorEjecucion) Error() string { return e.Diag.String() }

// Códigos estables de los errores de ejecución. El asistente los usa junto
// con la categoría para buscar su explicación.
const (
	CodigoDivisionCero      = "division-cero"
	CodigoDesbordamiento    = "desbordamiento"
	CodigoIndiceFueraRango  = "indice-fuera-rango"
	CodigoClaveInexistente  = "clave-inexistente"
	CodigoConversion        = "conversion-invalida"
	CodigoRecursion         = "recursion-excedida"
	CodigoAleatorioRango    = "aleatorio-rango"
	CodigoInterno           = "error-interno"
	limiteLlamadasAnidadas  = 1000
	descripcionRangoRoca    = "un roca guarda enteros entre -9223372036854775808 y 9223372036854775807."
	sugerenciaDivisionCero  = "comprueba que el divisor sea diferente de 0 antes de dividir."
	causaErrorInterno       = "el analizador debió rechazar este programa antes de ejecutarlo."
	sugerenciaErrorInterno  = "avisa al equipo de PokeScript con el programa que produjo este error."
	sugerenciaIndiceDesde1  = "los índices empiezan en 1; usa tamaño() para saber cuántos elementos hay."
	sugerenciaClaveContiene = "antes de leer una clave, comprueba que exista con contiene."
)

// fallo arma un error de ejecución en la posición de un nodo.
func (in *Interprete) fallo(nodo ast.Nodo, codigo, desc, causa, sugerencia string) error {
	p := nodo.Posicion()
	return &ErrorEjecucion{Diag: diag.Diagnostic{
		Severity: diag.Error,
		Category: diag.Ejecucion,
		Code:     codigo,
		Heading:  diag.EncabezadoEjecucion,
		File:     in.archivo,
		Line:     p.Line,
		Col:      p.Col,
		Len:      p.Len,
		Desc:     desc,
		Cause:    causa,
		Suggest:  sugerencia,
	}}
}

// interno reporta algo que un programa ya validado nunca debería hacer.
// Mientras no exista el analizador (hito 5), aparece con programas que
// tienen errores de tipos o de nombres.
func (in *Interprete) interno(nodo ast.Nodo, formato string, args ...any) error {
	return in.fallo(nodo, CodigoInterno,
		"error interno del intérprete: "+fmt.Sprintf(formato, args...),
		causaErrorInterno, sugerenciaErrorInterno)
}

// falloAcceso traduce los errores de equipo, mochila y planta.
func (in *Interprete) falloAcceso(nodo ast.Nodo, coleccion ast.Expr, err error) error {
	var fuera *ErrFueraDeRango
	if errors.As(err, &fuera) {
		return in.fallo(nodo, CodigoIndiceFueraRango,
			fmt.Sprintf("no existe la posición %d en %s.", fuera.Indice, nombreDe(coleccion, fuera.Coleccion)),
			err.Error()+".", sugerenciaIndiceDesde1)
	}
	var falta *ErrClaveInexistente
	if errors.As(err, &falta) {
		return in.fallo(nodo, CodigoClaveInexistente,
			fmt.Sprintf("la clave %s no existe en %s.", literal(falta.Clave), nombreDe(coleccion, "la mochila")),
			"solo se puede leer o quitar una clave que ya se guardó.", sugerenciaClaveContiene)
	}
	return err
}

// falloDesbordamiento explica un resultado que no cabe en roca o agua.
func (in *Interprete) falloDesbordamiento(nodo ast.Nodo, operacion string) error {
	return in.fallo(nodo, CodigoDesbordamiento,
		fmt.Sprintf("el resultado de %s es demasiado grande.", operacion),
		descripcionRangoRoca+" Un agua no puede llegar a infinito.",
		"revisa si el cálculo crece sin control, por ejemplo dentro de un ciclo.")
}

// nombreDe devuelve el nombre de la colección si es un dato; si es otra
// expresión, una descripción genérica.
func nombreDe(e ast.Expr, generico string) string {
	if id, ok := e.(*ast.Ident); ok {
		return id.Nombre
	}
	if generico == "equipo" || generico == "texto" {
		return "el " + generico
	}
	return generico
}

// describir muestra un operando con su nombre y su valor, para que el
// mensaje diga "vida (120)" y no solo "120".
func describir(e ast.Expr, v Value) string {
	if id, ok := e.(*ast.Ident); ok {
		return fmt.Sprintf("%s (%s)", id.Nombre, literal(v))
	}
	return literal(v)
}
