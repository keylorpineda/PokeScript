package interprete

import "github.com/keylorpineda/PokeScript/internal/ast"

// Variable es un dato con nombre durante la ejecución.
type Variable struct {
	Valor Value
	// Tipo es el tipo declarado. Hace falta en ejecución para ensanchar roca
	// a agua al guardar, para saber qué leer en capturar y para construir
	// los literales { }. Es nil en las variables de recorrer.
	Tipo        *ast.TipoExpr
	Asignada    bool // falso mientras un dato declarado sin valor no reciba uno
	Medalla     bool // constante: no se reasigna ni se modifica su contenido
	SoloLectura bool // variable de recorrer
}

// Entorno es un alcance: un bloque, el cuerpo de un movimiento o el archivo.
// Cada uno conoce a su padre, así que buscar un nombre sube por la cadena.
type Entorno struct {
	vars  map[string]*Variable
	padre *Entorno
	// mov es el movimiento que se está ejecutando; solo se llena en el
	// alcance raíz de su cuerpo. entregar lo busca para saber el tipo de
	// retorno.
	mov *ast.DeclMovimiento
}

// NuevoEntorno crea un alcance hijo de padre (nil para el alcance de archivo).
func NuevoEntorno(padre *Entorno) *Entorno {
	return &Entorno{vars: map[string]*Variable{}, padre: padre}
}

// Declarar agrega una variable al alcance actual. Devuelve false si el
// nombre ya existe en este mismo alcance.
func (e *Entorno) Declarar(nombre string, v *Variable) bool {
	if _, existe := e.vars[nombre]; existe {
		return false
	}
	e.vars[nombre] = v
	return true
}

// Buscar devuelve la variable con ese nombre en este alcance o en alguno de
// sus padres, o nil si no existe.
func (e *Entorno) Buscar(nombre string) *Variable {
	for a := e; a != nil; a = a.padre {
		if v, ok := a.vars[nombre]; ok {
			return v
		}
	}
	return nil
}

// movimiento devuelve el movimiento en ejecución, o nil si se está en
// combate o en el alcance de archivo.
func (e *Entorno) movimiento() *ast.DeclMovimiento {
	for a := e; a != nil; a = a.padre {
		if a.mov != nil {
			return a.mov
		}
	}
	return nil
}
