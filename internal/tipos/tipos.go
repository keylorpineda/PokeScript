package tipos

// Kind es la forma de un tipo (sección 3.1).
type Kind int

const (
	KRoca Kind = iota
	KAgua
	KFuego
	KPlanta
	KElectrico
	KEspecie
	KFicha
	KEquipo
	KMochila
	// KFantasma es el tipo del literal fantasma. No se puede declarar; solo
	// sirve para que el analizador sepa que fantasma se asigna únicamente a
	// un posible (sección 1.4).
	KFantasma
)

var nombresKind = [...]string{
	KRoca:      "roca",
	KAgua:      "agua",
	KFuego:     "fuego",
	KPlanta:    "planta",
	KElectrico: "electrico",
	KEspecie:   "especie",
	KFicha:     "ficha",
	KEquipo:    "equipo",
	KMochila:   "mochila",
	KFantasma:  "fantasma",
}

func (k Kind) String() string {
	if k >= 0 && int(k) < len(nombresKind) {
		return nombresKind[k]
	}
	return "desconocido"
}

// Simple indica si k es uno de los cinco tipos simples.
func (k Kind) Simple() bool { return k >= KRoca && k <= KElectrico }

// Type es un tipo de PokeScript (sección 3.1).
type Type struct {
	Kind     Kind
	Opcional bool   // posible
	Nombre   string // especie / ficha
	Clave    *Type  // mochila
	Elem     *Type  // equipo / mochila
}

// Constructores. Cada llamada devuelve un tipo nuevo, así nadie modifica un
// tipo compartido por accidente.

func Roca() *Type                 { return &Type{Kind: KRoca} }
func Agua() *Type                 { return &Type{Kind: KAgua} }
func Fuego() *Type                { return &Type{Kind: KFuego} }
func Planta() *Type               { return &Type{Kind: KPlanta} }
func Electrico() *Type            { return &Type{Kind: KElectrico} }
func Fantasma() *Type             { return &Type{Kind: KFantasma} }
func Especie(nombre string) *Type { return &Type{Kind: KEspecie, Nombre: nombre} }
func Ficha(nombre string) *Type   { return &Type{Kind: KFicha, Nombre: nombre} }
func Equipo(elem *Type) *Type     { return &Type{Kind: KEquipo, Elem: elem} }
func Mochila(clave, elem *Type) *Type {
	return &Type{Kind: KMochila, Clave: clave, Elem: elem}
}

// Posible devuelve una copia de t marcada como posible.
func Posible(t *Type) *Type {
	c := *t
	c.Opcional = true
	return &c
}

// Base devuelve t sin la marca posible.
func Base(t *Type) *Type {
	if t == nil || !t.Opcional {
		return t
	}
	c := *t
	c.Opcional = false
	return &c
}

// String escribe el tipo como en el código: "posible roca", "equipo de
// planta", "mochila de planta a roca", "Estado".
func (t *Type) String() string {
	if t == nil {
		return "desconocido"
	}
	s := ""
	if t.Opcional {
		s = "posible "
	}
	switch t.Kind {
	case KEspecie, KFicha:
		return s + t.Nombre
	case KEquipo:
		return s + "equipo de " + t.Elem.String()
	case KMochila:
		return s + "mochila de " + t.Clave.String() + " a " + t.Elem.String()
	}
	return s + t.Kind.String()
}

// Valido revisa las reglas de forma de la sección 3.1 y de la decisión H3.
// Devuelve una explicación si el tipo no se puede declarar.
func Valido(t *Type) (bool, string) {
	if t == nil {
		return false, "falta el tipo"
	}
	if t.Opcional && !t.Kind.Simple() && t.Kind != KEspecie {
		return false, "posible solo se aplica a tipos simples y especies"
	}
	switch t.Kind {
	case KFantasma:
		return false, "fantasma es un valor, no un tipo"
	case KEquipo:
		if t.Elem != nil && t.Elem.Opcional {
			return false, "un equipo no puede guardar valores posible"
		}
		return Valido(t.Elem)
	case KMochila:
		switch {
		case t.Clave == nil:
			return false, "falta el tipo de la clave"
		case t.Clave.Opcional:
			return false, "la clave de una mochila no puede ser posible"
		case t.Clave.Kind != KRoca && t.Clave.Kind != KFuego && t.Clave.Kind != KPlanta && t.Clave.Kind != KEspecie:
			return false, "la clave de una mochila debe ser roca, fuego, planta o una especie"
		}
		if t.Elem != nil && t.Elem.Opcional {
			return false, "una mochila no puede guardar valores posible"
		}
		return Valido(t.Elem)
	}
	return true, ""
}

// Iguales compara dos tipos por estructura exacta: equipo de equipo de roca
// solo es igual a equipo de equipo de roca, y posible roca no es igual a
// roca.
func Iguales(a, b *Type) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Kind != b.Kind || a.Opcional != b.Opcional || a.Nombre != b.Nombre {
		return false
	}
	return Iguales(a.Clave, b.Clave) && Iguales(a.Elem, b.Elem)
}
