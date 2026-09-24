package interprete

import "fmt"

// Mochila es un diccionario que conserva el orden de inserción, porque
// recorrer la visita en ese orden (sección 3.4). Un map de Go no garantiza
// orden, así que se guardan las claves y los valores en slices paralelos y
// el map solo sirve para encontrar la posición de una clave.
//
// Las claves posibles son roca (int64), fuego (rune), planta (string) y
// especie (EspecieVal). Todas son comparables en Go, así que sirven como
// clave de map, y como int64 y rune son tipos distintos, la roca 65 y el
// fuego 'A' no chocan.
type Mochila struct {
	claves  []Value
	valores []Value
	indice  map[Value]int
}

// NuevaMochila crea una mochila vacía.
func NuevaMochila() *Mochila {
	return &Mochila{indice: map[Value]int{}}
}

// Tamano devuelve la cantidad de pares.
func (m *Mochila) Tamano() int { return len(m.claves) }

// Contiene indica si la mochila tiene la clave (operador contiene).
func (m *Mochila) Contiene(clave Value) bool {
	_, ok := m.indice[clave]
	return ok
}

// Obtener devuelve el valor de una clave, sin copiarlo. Si la clave no
// existe devuelve ErrClaveInexistente.
func (m *Mochila) Obtener(clave Value) (Value, error) {
	p, ok := m.indice[clave]
	if !ok {
		return nil, &ErrClaveInexistente{Clave: clave}
	}
	return m.valores[p], nil
}

// Poner guarda una copia de valor en la clave. Si la clave ya existe, se
// reemplaza su valor y conserva su posición; si no existe, se agrega al final
// (decisión H2 del equipo).
func (m *Mochila) Poner(clave, valor Value) {
	validarClave(clave)
	if p, ok := m.indice[clave]; ok {
		m.valores[p] = Copiar(valor)
		return
	}
	m.indice[clave] = len(m.claves)
	m.claves = append(m.claves, clave)
	m.valores = append(m.valores, Copiar(valor))
}

// Quitar elimina una clave y su valor. Los pares siguientes conservan su
// orden. Si la clave no existe devuelve ErrClaveInexistente.
func (m *Mochila) Quitar(clave Value) error {
	p, ok := m.indice[clave]
	if !ok {
		return &ErrClaveInexistente{Clave: clave}
	}
	m.claves = append(m.claves[:p], m.claves[p+1:]...)
	m.valores = append(m.valores[:p], m.valores[p+1:]...)
	delete(m.indice, clave)
	for i := p; i < len(m.claves); i++ {
		m.indice[m.claves[i]] = i
	}
	return nil
}

// Claves devuelve las claves en orden de inserción. El slice es nuevo, así
// que modificar la mochila después no altera el recorrido.
func (m *Mochila) Claves() []Value {
	return append([]Value(nil), m.claves...)
}

// copiar devuelve una copia profunda. Las claves son valores simples y no
// necesitan copia; los valores sí.
func (m *Mochila) copiar() *Mochila {
	c := &Mochila{
		claves:  append([]Value(nil), m.claves...),
		valores: make([]Value, len(m.valores)),
		indice:  make(map[Value]int, len(m.indice)),
	}
	for i, v := range m.valores {
		c.valores[i] = Copiar(v)
	}
	for k, p := range m.indice {
		c.indice[k] = p
	}
	return c
}

// igual compara por contenido sin importar el orden (decisión H13).
func (m *Mochila) igual(o *Mochila) bool {
	if len(m.claves) != len(o.claves) {
		return false
	}
	for i, k := range m.claves {
		p, ok := o.indice[k]
		if !ok || !Igual(m.valores[i], o.valores[p]) {
			return false
		}
	}
	return true
}

// EsClave indica si v puede ser clave de una mochila.
func EsClave(v Value) bool {
	switch v.(type) {
	case int64, rune, string, EspecieVal:
		return true
	}
	return false
}

// validarClave detiene la ejecución si la clave no es de un tipo permitido.
// El analizador lo impide (tipo_clave en la gramática y decisión H3), así
// que llegar aquí es un error interno.
func validarClave(clave Value) {
	if !EsClave(clave) {
		panic(fmt.Sprintf("error interno: %T no puede ser clave de una mochila", clave))
	}
}
