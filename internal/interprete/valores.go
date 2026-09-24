package interprete

import (
	"fmt"
	"unicode/utf8"
)

// Value es un valor de PokeScript durante la ejecución (sección 3.4 de la
// especificación). Cada tipo del lenguaje tiene una sola representación:
//
//	roca      → int64
//	agua      → float64
//	fuego     → rune
//	planta    → string
//	electrico → bool
//	especie   → EspecieVal
//	equipo    → *Equipo
//	mochila   → *Mochila
//	ficha     → *Ficha
//	fantasma  → nil
//
// Regla de propiedad: un contenedor (equipo, mochila, ficha) es dueño de lo
// que guarda. Todo lo que entra se copia en profundidad; lo que sale con
// Obtener o Campo NO se copia, para que el intérprete pueda asignar a
// destinos anidados como equipo[1][2]. Quien guarde un valor que salió de un
// contenedor debe copiarlo con Copiar.
type Value interface{}

// EspecieVal es un valor de una especie, por ejemplo Estado.SANO.
type EspecieVal struct {
	Tipo  string // nombre de la especie: "Estado"
	Valor string // nombre del valor: "SANO"
}

// ─── Equipo ────────────────────────────────────────────────────────────────

// Equipo es una lista con índices desde 1.
type Equipo struct {
	elems []Value
}

// NuevoEquipo crea un equipo con copias de los elementos dados.
func NuevoEquipo(elems ...Value) *Equipo {
	e := &Equipo{elems: make([]Value, 0, len(elems))}
	for _, v := range elems {
		e.Sumar(v)
	}
	return e
}

// Tamano devuelve la cantidad de elementos.
func (e *Equipo) Tamano() int { return len(e.elems) }

// Obtener devuelve el elemento en la posición i (desde 1), sin copiarlo.
func (e *Equipo) Obtener(i int64) (Value, error) {
	p, err := e.posicion(i)
	if err != nil {
		return nil, err
	}
	return e.elems[p], nil
}

// Poner reemplaza el elemento en la posición i (desde 1) por una copia de v.
func (e *Equipo) Poner(i int64, v Value) error {
	p, err := e.posicion(i)
	if err != nil {
		return err
	}
	e.elems[p] = Copiar(v)
	return nil
}

// Sumar agrega una copia de v al final (instrucción sumar).
func (e *Equipo) Sumar(v Value) {
	e.elems = append(e.elems, Copiar(v))
}

// Quitar elimina el elemento en la posición i (desde 1). Los elementos
// siguientes bajan una posición.
func (e *Equipo) Quitar(i int64) error {
	p, err := e.posicion(i)
	if err != nil {
		return err
	}
	e.elems = append(e.elems[:p], e.elems[p+1:]...)
	return nil
}

// Elementos devuelve los elementos en orden, sin copiarlos. El slice es
// nuevo, así que modificar el equipo después no altera el recorrido.
func (e *Equipo) Elementos() []Value {
	return append([]Value(nil), e.elems...)
}

// posicion convierte un índice de PokeScript (desde 1) en uno de Go.
func (e *Equipo) posicion(i int64) (int, error) {
	if i < 1 || i > int64(len(e.elems)) {
		return 0, &ErrFueraDeRango{Indice: i, Tamano: len(e.elems), Coleccion: "equipo"}
	}
	return int(i - 1), nil
}

// ─── Ficha ─────────────────────────────────────────────────────────────────

// Ficha es un registro con campos fijos. Guarda sus valores en un
// map[string]Value, como dice la sección 3.4; además conserva el orden en que
// se declararon los campos, que es el orden en que se muestran.
type Ficha struct {
	Tipo    string           // nombre de la ficha: "Pokemon"
	campos  []string         // orden de declaración; compartido y de solo lectura
	valores map[string]Value // valor de cada campo
}

// NuevaFicha crea una ficha del tipo dado. campos es el orden de la
// declaración y valores debe traer exactamente esos campos: el analizador lo
// garantiza (validación 15), así que si no se cumple es un error interno.
func NuevaFicha(tipo string, campos []string, valores map[string]Value) *Ficha {
	if len(valores) != len(campos) {
		panic(fmt.Sprintf("error interno: la ficha %s tiene %d campos y se recibieron %d valores",
			tipo, len(campos), len(valores)))
	}
	f := &Ficha{Tipo: tipo, campos: campos, valores: make(map[string]Value, len(campos))}
	for _, c := range campos {
		v, ok := valores[c]
		if !ok {
			panic(fmt.Sprintf("error interno: falta el campo %s de la ficha %s", c, tipo))
		}
		f.valores[c] = Copiar(v)
	}
	return f
}

// Campos devuelve los nombres de los campos en el orden de declaración.
func (f *Ficha) Campos() []string {
	return append([]string(nil), f.campos...)
}

// TieneCampo indica si la ficha tiene un campo con ese nombre.
func (f *Ficha) TieneCampo(nombre string) bool {
	_, ok := f.valores[nombre]
	return ok
}

// Campo devuelve el valor de un campo, sin copiarlo.
func (f *Ficha) Campo(nombre string) Value {
	v, ok := f.valores[nombre]
	if !ok {
		panic(fmt.Sprintf("error interno: la ficha %s no tiene el campo %s", f.Tipo, nombre))
	}
	return v
}

// PonerCampo reemplaza el valor de un campo por una copia de v.
func (f *Ficha) PonerCampo(nombre string, v Value) {
	if _, ok := f.valores[nombre]; !ok {
		panic(fmt.Sprintf("error interno: la ficha %s no tiene el campo %s", f.Tipo, nombre))
	}
	f.valores[nombre] = Copiar(v)
}

// ─── Planta como secuencia de letras ───────────────────────────────────────

// TamanoPlanta devuelve la cantidad de letras de un texto. Cuenta runas, no
// bytes: tamaño("Pokémon") es 7.
func TamanoPlanta(s string) int { return utf8.RuneCountInString(s) }

// LetraDe devuelve la letra en la posición i (desde 1) de un texto, como
// fuego. Cuenta runas: "Pokémon"[4] es 'é'.
func LetraDe(s string, i int64) (rune, error) {
	letras := []rune(s)
	if i < 1 || i > int64(len(letras)) {
		return 0, &ErrFueraDeRango{Indice: i, Tamano: len(letras), Coleccion: "texto"}
	}
	return letras[i-1], nil
}

// ─── Copia y comparación ───────────────────────────────────────────────────

// Copiar devuelve una copia profunda de v. Los valores simples se devuelven
// tal cual porque en Go ya se copian; equipos, mochilas y fichas se
// reconstruyen completos, incluido todo lo que llevan adentro. Así se cumple
// el paso por valor de la sección 3.4.
func Copiar(v Value) Value {
	switch x := v.(type) {
	case nil, int64, float64, rune, string, bool, EspecieVal:
		return x
	case *Equipo:
		c := &Equipo{elems: make([]Value, len(x.elems))}
		for i, e := range x.elems {
			c.elems[i] = Copiar(e)
		}
		return c
	case *Mochila:
		return x.copiar()
	case *Ficha:
		c := &Ficha{Tipo: x.Tipo, campos: x.campos, valores: make(map[string]Value, len(x.valores))}
		for k, e := range x.valores {
			c.valores[k] = Copiar(e)
		}
		return c
	default:
		panic(fmt.Sprintf("error interno: valor de tipo %T no pertenece a PokeScript", v))
	}
}

// Igual compara dos valores por contenido (operador igual).
//
// Como todo se pasa por valor, no existe la identidad de un objeto: dos
// equipos son iguales si tienen los mismos elementos en el mismo orden; dos
// mochilas, si tienen los mismos pares sin importar el orden; dos fichas, si
// son del mismo tipo y sus campos son iguales. fantasma solo es igual a
// fantasma, así que comparar T con posible T compara los valores.
//
// Dos valores de tipos distintos nunca son iguales; el analizador ya rechaza
// esas comparaciones, esto solo evita resultados raros si llegaran.
func Igual(a, b Value) bool {
	switch x := a.(type) {
	case nil:
		return b == nil
	case int64, float64, rune, string, bool, EspecieVal:
		return a == b
	case *Equipo:
		y, ok := b.(*Equipo)
		if !ok || len(x.elems) != len(y.elems) {
			return false
		}
		for i := range x.elems {
			if !Igual(x.elems[i], y.elems[i]) {
				return false
			}
		}
		return true
	case *Mochila:
		y, ok := b.(*Mochila)
		return ok && x.igual(y)
	case *Ficha:
		y, ok := b.(*Ficha)
		if !ok || x.Tipo != y.Tipo || len(x.valores) != len(y.valores) {
			return false
		}
		for k, v := range x.valores {
			w, existe := y.valores[k]
			if !existe || !Igual(v, w) {
				return false
			}
		}
		return true
	default:
		panic(fmt.Sprintf("error interno: valor de tipo %T no pertenece a PokeScript", a))
	}
}

// ─── Errores de acceso ─────────────────────────────────────────────────────

// ErrFueraDeRango indica un índice que no existe en un equipo o un texto. El
// intérprete lo convierte en un diagnóstico de ejecución con la posición y el
// nombre de la colección.
type ErrFueraDeRango struct {
	Indice    int64
	Tamano    int
	Coleccion string // "equipo" o "texto"
}

func (e *ErrFueraDeRango) Error() string {
	if e.Tamano == 0 {
		return fmt.Sprintf("el índice %d está fuera de rango: el %s está vacío", e.Indice, e.Coleccion)
	}
	unidad := "elementos"
	if e.Coleccion == "texto" {
		unidad = "letras"
	}
	return fmt.Sprintf("el índice %d está fuera de rango: el %s tiene %d %s, del 1 al %d",
		e.Indice, e.Coleccion, e.Tamano, unidad, e.Tamano)
}

// ErrClaveInexistente indica que se leyó o quitó una clave que la mochila no
// tiene.
type ErrClaveInexistente struct {
	Clave Value
}

func (e *ErrClaveInexistente) Error() string {
	return fmt.Sprintf("la clave %s no existe en la mochila", literal(e.Clave))
}
