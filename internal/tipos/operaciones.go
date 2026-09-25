package tipos

// Operadores tal como los escribe token.Kind.String(), para que el analizador
// pase directamente el operador de un nodo Binaria o Unaria. Los tres
// últimos no son tokens: son las operaciones de [ ], tamaño() y el sino de
// respaldo.
const (
	OpSuma       = "+"
	OpResta      = "-"
	OpProducto   = "*"
	OpDivision   = "/"
	OpResto      = "resto"
	OpMayor      = ">"
	OpMenor      = "<"
	OpMayorIgual = ">="
	OpMenorIgual = "<="
	OpIgual      = "igual"
	OpDiferente  = "diferente"
	OpY          = "y"
	OpO          = "o"
	OpNo         = "no"
	OpContiene   = "contiene"
	OpIndice     = "[]"
	OpTamano     = "tamaño"
	OpRespaldo   = "sino"
)

// OpKey identifica una operación binaria entre dos tipos simples.
type OpKey struct {
	Op       string
	Izq, Der Kind
}

// Operaciones es la tabla 3.3 para los tipos simples: una entrada por
// (operador, izquierda, derecha) con el tipo del resultado. Lo que depende de
// la estructura del tipo (igual, [ ], contiene sobre colecciones, sino) lo
// resuelve ResultadoOp.
var Operaciones = map[OpKey]Kind{}

func init() {
	numericos := [][2]Kind{{KRoca, KRoca}, {KAgua, KAgua}, {KRoca, KAgua}, {KAgua, KRoca}}
	for _, op := range []string{OpSuma, OpResta, OpProducto} {
		for _, p := range numericos {
			r := KAgua // ensanchamiento: si participa un agua, el resultado es agua
			if p[0] == KRoca && p[1] == KRoca {
				r = KRoca
			}
			Operaciones[OpKey{op, p[0], p[1]}] = r
		}
	}
	for _, p := range numericos {
		Operaciones[OpKey{OpDivision, p[0], p[1]}] = KAgua // / siempre da agua
	}
	Operaciones[OpKey{OpSuma, KPlanta, KPlanta}] = KPlanta
	Operaciones[OpKey{OpResto, KRoca, KRoca}] = KRoca

	for _, op := range []string{OpMayor, OpMenor, OpMayorIgual, OpMenorIgual} {
		for _, p := range numericos {
			Operaciones[OpKey{op, p[0], p[1]}] = KElectrico
		}
		Operaciones[OpKey{op, KFuego, KFuego}] = KElectrico
	}
	Operaciones[OpKey{OpY, KElectrico, KElectrico}] = KElectrico
	Operaciones[OpKey{OpO, KElectrico, KElectrico}] = KElectrico

	// Subcadena (decisión H16): se puede buscar un texto o una sola letra.
	Operaciones[OpKey{OpContiene, KPlanta, KPlanta}] = KElectrico
	Operaciones[OpKey{OpContiene, KPlanta, KFuego}] = KElectrico
}

// ResultadoOp devuelve el tipo del resultado de una operación, o false si la
// tabla 3.3 no la admite. Para los operadores unarios (- , no, tamaño) der es
// nil.
//
// Un operando posible solo se acepta en igual, diferente y sino: para
// cualquier otra cosa hay que comprobar antes que no sea fantasma.
func ResultadoOp(op string, izq, der *Type) (*Type, bool) {
	if izq == nil {
		return nil, false
	}
	if der == nil {
		return resultadoUnaria(op, izq)
	}
	switch op {
	case OpIgual, OpDiferente:
		if comparables(izq, der) {
			return Electrico(), true
		}
		return nil, false
	case OpRespaldo:
		if izq.Opcional && Asignable(der, Base(izq)) {
			return Base(izq), true
		}
		return nil, false
	}
	if izq.Opcional || der.Opcional || izq.Kind == KFantasma || der.Kind == KFantasma {
		return nil, false
	}
	switch op {
	case OpIndice:
		return resultadoIndice(izq, der)
	case OpContiene:
		var buscado *Type
		switch izq.Kind {
		case KEquipo:
			buscado = izq.Elem
		case KMochila:
			buscado = izq.Clave
		}
		if buscado != nil {
			if Asignable(der, buscado) {
				return Electrico(), true
			}
			return nil, false
		}
	}
	k, ok := Operaciones[OpKey{op, izq.Kind, der.Kind}]
	if !ok {
		return nil, false
	}
	return &Type{Kind: k}, true
}

func resultadoUnaria(op string, t *Type) (*Type, bool) {
	if t.Opcional {
		return nil, false
	}
	switch op {
	case OpResta:
		if t.Kind == KRoca || t.Kind == KAgua {
			return &Type{Kind: t.Kind}, true
		}
	case OpNo:
		if t.Kind == KElectrico {
			return Electrico(), true
		}
	case OpTamano:
		if t.Kind == KEquipo || t.Kind == KMochila || t.Kind == KPlanta {
			return Roca(), true
		}
	}
	return nil, false
}

// resultadoIndice resuelve equipo[roca], mochila[clave] y planta[roca].
func resultadoIndice(c, i *Type) (*Type, bool) {
	switch c.Kind {
	case KEquipo:
		if i.Kind == KRoca {
			return c.Elem, true
		}
	case KMochila:
		if Asignable(i, c.Clave) {
			return c.Elem, true
		}
	case KPlanta:
		if i.Kind == KRoca {
			return Fuego(), true
		}
	}
	return nil, false
}

// comparables dice si igual y diferente admiten los dos tipos: el mismo tipo,
// T contra posible T, o fantasma contra un posible (tabla 3.3 y decisión
// H18). roca contra agua no se compara con igual.
func comparables(a, b *Type) bool {
	if a.Kind == KFantasma {
		return b.Opcional || b.Kind == KFantasma
	}
	if b.Kind == KFantasma {
		return a.Opcional
	}
	return Iguales(Base(a), Base(b))
}
