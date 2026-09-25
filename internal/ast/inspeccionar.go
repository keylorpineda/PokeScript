package ast

// Inspeccionar recorre el árbol en profundidad, en el orden en que aparece
// en el código, y llama a f con cada nodo. Si f devuelve false, no entra en
// los hijos de ese nodo. Es el equivalente de go/ast.Inspect.
//
// Sirve para las validaciones que no necesitan llevar un ámbito, como
// buscar divisiones entre el literal 0 o todos los segun de un cuerpo:
//
//	ast.Inspeccionar(mov, func(n ast.Nodo) bool {
//	    if b, ok := n.(*ast.Binaria); ok && b.Op == token.SLASH { … }
//	    return true
//	})
//
// Los hijos nil (un Si sin sino, un entregar sin valor, una cabecera que
// no se pudo leer) se saltan.
func Inspeccionar(n Nodo, f func(Nodo) bool) { //nolint:gocyclo // un caso por cada tipo de nodo
	if esNil(n) || !f(n) {
		return
	}
	visitar := func(hijos ...Nodo) {
		for _, h := range hijos {
			Inspeccionar(h, f)
		}
	}
	switch x := n.(type) {
	case *Importacion:
		for _, id := range x.Nombres {
			visitar(id)
		}
	case *TipoExpr:
		visitar(x.Clave, x.Elem)

	case *DeclMedalla:
		visitar(x.Tipo, x.Nombre, x.Valor)
	case *DeclEspecie:
		visitar(x.Nombre)
		for _, v := range x.Valores {
			visitar(v)
		}
	case *DeclFicha:
		visitar(x.Nombre)
		for _, c := range x.Campos {
			visitar(c)
		}
	case *Campo:
		visitar(x.Tipo, x.Nombre)
	case *DeclMovimiento:
		visitar(x.Retorno, x.Nombre)
		for _, p := range x.Params {
			visitar(p)
		}
		visitarInstrucciones(x.Cuerpo, f)
	case *Param:
		visitar(x.Tipo, x.Nombre)
	case *Combate:
		visitarInstrucciones(x.Cuerpo, f)

	case *DeclDato:
		visitar(x.Tipo, x.Nombre, x.Valor)
	case *Asignacion:
		visitar(x.Destino, x.Valor)
	case *Gritar:
		visitarExprs(x.Args, f)
	case *Capturar:
		visitar(x.Destino, x.Mensaje)
	case *LlamadaInstr:
		visitar(x.Llamada)
	case *Sumar:
		visitar(x.Valor, x.Coleccion)
	case *Quitar:
		visitar(x.Coleccion, x.Indice)
	case *Si:
		for _, r := range x.Ramas {
			visitar(r)
		}
		visitarInstrucciones(x.Sino, f)
	case *RamaSi:
		visitar(x.Cond)
		visitarInstrucciones(x.Cuerpo, f)
	case *Segun:
		visitar(x.Valor)
		for _, a := range x.Alternativas {
			visitar(a)
		}
		visitar(x.Otro)
	case *Alternativa:
		visitarExprs(x.Patrones, f)
		visitar(x.Cuerpo)
	case *Mientras:
		visitar(x.Cond)
		visitarInstrucciones(x.Cuerpo, f)
	case *RecorrerColeccion:
		visitar(x.Var, x.Var2, x.Coleccion)
		visitarInstrucciones(x.Cuerpo, f)
	case *RecorrerRango:
		visitar(x.Var, x.Desde, x.Hasta)
		visitarInstrucciones(x.Cuerpo, f)
	case *Entregar:
		visitar(x.Valor)

	case *LitEquipo:
		visitarExprs(x.Elems, f)
	case *LitLlaves:
		for _, p := range x.Pares {
			visitar(p)
		}
	case *Par:
		visitar(x.Clave, x.Valor)
	case *Binaria:
		visitar(x.Izq, x.Der)
	case *Unaria:
		visitar(x.Operando)
	case *Respaldo:
		visitar(x.Valor, x.Reemplazo)
	case *Indice:
		visitar(x.Coleccion, x.Indice)
	case *CampoAcceso:
		visitar(x.Objeto, x.Nombre)
	case *Llamada:
		visitar(x.Nombre)
		visitarExprs(x.Args, f)
	case *Convertir:
		visitar(x.Valor, x.Destino)
	case *Tamano:
		visitar(x.Valor)
	case *Aleatorio:
		visitar(x.Min, x.Max)
	case *Redondear:
		visitar(x.Valor)
	}
	// Ident, literales simples, Huir y Siguiente no tienen hijos.
}

// InspeccionarPrograma aplica Inspeccionar a cada importación y
// declaración de un archivo, en orden.
func InspeccionarPrograma(p *Programa, f func(Nodo) bool) {
	if p == nil {
		return
	}
	for _, im := range p.Importaciones {
		Inspeccionar(im, f)
	}
	for _, d := range p.Declaraciones {
		Inspeccionar(d, f)
	}
}

func visitarInstrucciones(is []Instr, f func(Nodo) bool) {
	for _, i := range is {
		Inspeccionar(i, f)
	}
}

func visitarExprs(es []Expr, f func(Nodo) bool) {
	for _, e := range es {
		Inspeccionar(e, f)
	}
}

// esNil detecta tanto la interfaz nil como un puntero nil guardado en la
// interfaz, por ejemplo un *Ident nil en Var2 o un *TipoExpr nil en Retorno.
func esNil(n Nodo) bool {
	if n == nil {
		return true
	}
	switch x := n.(type) {
	case *Ident:
		return x == nil
	case *TipoExpr:
		return x == nil
	case *Llamada:
		return x == nil
	}
	return false
}
