package interprete

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// evaluar calcula el valor de una expresión.
//
// Lo que devuelve NO es una copia: un dato que guarda un equipo devuelve ese
// mismo equipo. Así asignar a e[1][2] modifica el original. Quien guarde el
// resultado en otro lugar debe copiarlo (lo hacen declarar, asignar y los
// contenedores).
func (in *Interprete) evaluar(e ast.Expr, env *Entorno) (Value, error) {
	switch x := e.(type) {
	case *ast.LitRoca:
		return x.Valor, nil
	case *ast.LitAgua:
		return x.Valor, nil
	case *ast.LitFuego:
		return x.Valor, nil
	case *ast.LitPlanta:
		return x.Valor, nil
	case *ast.LitElectrico:
		return x.Valor, nil
	case *ast.LitFantasma:
		return nil, nil
	case *ast.Ident:
		return in.leerNombre(x, env)
	case *ast.LitEquipo, *ast.LitLlaves:
		return in.evaluarCon(e, nil, env)
	case *ast.Binaria:
		return in.evaluarBinaria(x, env)
	case *ast.Unaria:
		return in.evaluarUnaria(x, env)
	case *ast.Respaldo:
		v, err := in.evaluar(x.Valor, env)
		if err != nil || v != nil {
			return v, err
		}
		return in.evaluar(x.Reemplazo, env)
	case *ast.Indice:
		return in.evaluarIndice(x, env)
	case *ast.CampoAcceso:
		obj, err := in.evaluar(x.Objeto, env)
		if err != nil {
			return nil, err
		}
		f, ok := obj.(*Ficha)
		if !ok {
			return nil, in.interno(x, "se usó . sobre un valor que no es ficha: %s", literal(obj))
		}
		if !f.TieneCampo(x.Nombre.Nombre) {
			return nil, in.interno(x, "la ficha %s no tiene el campo %s", f.Tipo, x.Nombre.Nombre)
		}
		return f.Campo(x.Nombre.Nombre), nil
	case *ast.Llamada:
		v, entrego, err := in.llamar(x, env)
		if err == nil && !entrego {
			return nil, in.interno(x, "%s no entrega ningún valor y se usó en una expresión", x.Nombre.Nombre)
		}
		return v, err
	case *ast.Convertir:
		return in.evaluarConvertir(x, env)
	case *ast.Tamano:
		return in.evaluarTamano(x, env)
	case *ast.Aleatorio:
		return in.evaluarAleatorio(x, env)
	case *ast.Redondear:
		return in.evaluarRedondear(x, env)
	}
	return nil, in.interno(e, "expresión de tipo %T sin implementar", e)
}

// evaluarCon evalúa una expresión sabiendo el tipo del lugar donde se va a
// guardar. Hace falta para los literales: { } es mochila o ficha según ese
// tipo (sección 2.1), y los elementos de un literal se ensanchan al tipo de
// elemento (decisión H11). El resultado ya viene ajustado a esperado.
func (in *Interprete) evaluarCon(e ast.Expr, esperado *ast.TipoExpr, env *Entorno) (Value, error) {
	switch x := e.(type) {
	case *ast.LitEquipo:
		var elem *ast.TipoExpr
		if esperado != nil && esperado.Forma == ast.TipoEquipo {
			elem = esperado.Elem
		}
		eq := NuevoEquipo()
		for _, el := range x.Elems {
			v, err := in.evaluarCon(el, elem, env)
			if err != nil {
				return nil, err
			}
			eq.Sumar(v)
		}
		return eq, nil
	case *ast.LitLlaves:
		return in.evaluarLlaves(x, esperado, env)
	}
	v, err := in.evaluar(e, env)
	if err != nil {
		return nil, err
	}
	return ajustar(v, esperado), nil
}

// evaluarLlaves construye una mochila o una ficha.
func (in *Interprete) evaluarLlaves(x *ast.LitLlaves, esperado *ast.TipoExpr, env *Entorno) (Value, error) {
	if esperado == nil {
		return nil, in.interno(x, "un literal { } necesita un tipo esperado para saber si es mochila o ficha")
	}
	// Si el analizador ya resolvió el literal, su decisión tiene que
	// coincidir con el tipo esperado; si no lo resolvió (todavía no existe),
	// manda el tipo esperado.
	esMochila := esperado.Forma == ast.TipoMochila
	if (x.Resuelto == ast.LlavesMochila && !esMochila) || (x.Resuelto == ast.LlavesFicha && esMochila) {
		return nil, in.interno(x, "el analizador resolvió este literal { } distinto del tipo %s", nombreTipo(esperado))
	}
	if esMochila {
		m := NuevaMochila()
		for _, p := range x.Pares {
			k, err := in.evaluarCon(p.Clave, esperado.Clave, env)
			if err != nil {
				return nil, err
			}
			if !EsClave(k) {
				return nil, in.interno(p, "%s no puede ser clave de una mochila (decisión H3)", literal(k))
			}
			v, err := in.evaluarCon(p.Valor, esperado.Elem, env)
			if err != nil {
				return nil, err
			}
			m.Poner(k, v)
		}
		return m, nil
	}
	decl, ok := in.fichas[esperado.Nombre]
	if esperado.Forma != ast.TipoNombrado || !ok {
		return nil, in.interno(x, "un literal { } no corresponde al tipo %s", nombreTipo(esperado))
	}
	tipos := map[string]*ast.TipoExpr{}
	campos := make([]string, len(decl.Campos))
	for i, c := range decl.Campos {
		campos[i] = c.Nombre.Nombre
		tipos[c.Nombre.Nombre] = c.Tipo
	}
	valores := map[string]Value{}
	for _, p := range x.Pares {
		id, ok := p.Clave.(*ast.Ident)
		if !ok {
			return nil, in.interno(p, "un campo de ficha debe ser un nombre")
		}
		t, existe := tipos[id.Nombre]
		if !existe {
			return nil, in.interno(p, "la ficha %s no tiene el campo %s", decl.Nombre.Nombre, id.Nombre)
		}
		v, err := in.evaluarCon(p.Valor, t, env)
		if err != nil {
			return nil, err
		}
		valores[id.Nombre] = v
	}
	if len(valores) != len(campos) {
		return nil, in.interno(x, "el literal de %s no tiene todos sus campos", decl.Nombre.Nombre)
	}
	return NuevaFicha(decl.Nombre.Nombre, campos, valores), nil
}

// leerNombre resuelve un identificador: un dato, o un valor de especie.
func (in *Interprete) leerNombre(x *ast.Ident, env *Entorno) (Value, error) {
	if v := env.Buscar(x.Nombre); v != nil {
		if !v.Asignada {
			return nil, in.interno(x, "se leyó %s antes de que tuviera un valor", x.Nombre)
		}
		return v.Valor, nil
	}
	if ev, ok := in.valores[x.Nombre]; ok {
		return ev, nil
	}
	return nil, in.interno(x, "%s no está declarado", x.Nombre)
}

// ─── Operadores ────────────────────────────────────────────────────────────

func (in *Interprete) evaluarBinaria(x *ast.Binaria, env *Entorno) (Value, error) {
	if x.Op == token.Y || x.Op == token.O {
		return in.evaluarLogica(x, env)
	}
	a, err := in.evaluar(x.Izq, env)
	if err != nil {
		return nil, err
	}
	b, err := in.evaluar(x.Der, env)
	if err != nil {
		return nil, err
	}

	switch x.Op {
	case token.PLUS, token.MINUS, token.STAR, token.SLASH, token.RESTO:
		v, err := operarAritmetica(x.Op, a, b)
		return v, in.falloOperacion(x, a, b, err)
	case token.GT, token.LT, token.GE, token.LE:
		v, err := comparar(x.Op, a, b)
		if err != nil {
			return nil, in.falloOperacion(x, a, b, err)
		}
		return v, nil
	case token.IGUAL:
		return Igual(a, b), nil
	case token.DIFERENTE:
		return !Igual(a, b), nil
	case token.CONTIENE:
		return in.contiene(x, a, b)
	}
	return nil, in.interno(x, "operador %s sin implementar", x.Op)
}

// evaluarLogica aplica y / o en cortocircuito: si el lado izquierdo decide,
// el derecho no se evalúa (sección 6).
func (in *Interprete) evaluarLogica(x *ast.Binaria, env *Entorno) (Value, error) {
	a, err := in.condicion(x.Izq, env)
	if err != nil {
		return nil, err
	}
	if (x.Op == token.Y && !a) || (x.Op == token.O && a) {
		return a, nil
	}
	b, err := in.condicion(x.Der, env)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// condicion evalúa una expresión que debe ser electrico: no hay veracidad
// implícita (validación 2).
func (in *Interprete) condicion(e ast.Expr, env *Entorno) (bool, error) {
	v, err := in.evaluar(e, env)
	if err != nil {
		return false, err
	}
	b, ok := v.(bool)
	if !ok {
		return false, in.interno(e, "se esperaba un electrico y llegó %s", literal(v))
	}
	return b, nil
}

func (in *Interprete) evaluarUnaria(x *ast.Unaria, env *Entorno) (Value, error) {
	if x.Op == token.NO {
		b, err := in.condicion(x.Operando, env)
		return !b, err
	}
	a, err := in.evaluar(x.Operando, env)
	if err != nil {
		return nil, err
	}
	v, err := negar(a)
	if errors.Is(err, errDesbordamiento) {
		return nil, in.falloDesbordamiento(x, "-"+describir(x.Operando, a))
	}
	if err != nil {
		return nil, in.interno(x, "- no se aplica a %s", literal(a))
	}
	return v, nil
}

// falloOperacion traduce el error de una operación a su diagnóstico.
func (in *Interprete) falloOperacion(x *ast.Binaria, a, b Value, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errDivisionCero):
		accion := "dividir"
		if x.Op == token.RESTO {
			accion = "calcular el resto de"
		}
		return in.fallo(x, CodigoDivisionCero,
			fmt.Sprintf("no se puede %s %s entre %s.", accion, describir(x.Izq, a), describir(x.Der, b)),
			"dividir entre cero no tiene resultado.", sugerenciaDivisionCero)
	case errors.Is(err, errDesbordamiento):
		return in.falloDesbordamiento(x,
			fmt.Sprintf("%s %s %s", describir(x.Izq, a), x.Op, describir(x.Der, b)))
	}
	return in.interno(x, "%s no se aplica a %s y %s", x.Op, literal(a), literal(b))
}

// contiene busca un elemento en un equipo, una clave en una mochila (tabla
// 3.3) o un texto o una letra dentro de un texto (decisión H16). El texto
// vacío siempre está contenido.
func (in *Interprete) contiene(x *ast.Binaria, a, b Value) (Value, error) {
	switch c := a.(type) {
	case string:
		switch buscado := b.(type) {
		case string:
			return strings.Contains(c, buscado), nil
		case rune:
			return strings.ContainsRune(c, buscado), nil
		}
	case *Equipo:
		for _, e := range c.elems {
			if Igual(e, ensancharComo(b, e)) {
				return true, nil
			}
		}
		return false, nil
	case *Mochila:
		return c.Contiene(b), nil
	}
	return nil, in.interno(x, "contiene no se aplica a %s", literal(a))
}

// ─── Acceso con [ ] ────────────────────────────────────────────────────────

func (in *Interprete) evaluarIndice(x *ast.Indice, env *Entorno) (Value, error) {
	c, err := in.evaluar(x.Coleccion, env)
	if err != nil {
		return nil, err
	}
	i, err := in.evaluar(x.Indice, env)
	if err != nil {
		return nil, err
	}
	var v Value
	switch col := c.(type) {
	case *Equipo:
		n, ok := i.(int64)
		if !ok {
			return nil, in.interno(x, "el índice de un equipo debe ser roca")
		}
		v, err = col.Obtener(n)
	case *Mochila:
		v, err = col.Obtener(i)
	case string:
		n, ok := i.(int64)
		if !ok {
			return nil, in.interno(x, "el índice de un texto debe ser roca")
		}
		v, err = LetraDe(col, n)
	default:
		return nil, in.interno(x, "[ ] no se aplica a %s", literal(c))
	}
	if err != nil {
		return nil, in.falloAcceso(x, x.Coleccion, err)
	}
	return v, nil
}

// ─── Funciones incorporadas ────────────────────────────────────────────────

func (in *Interprete) evaluarTamano(x *ast.Tamano, env *Entorno) (Value, error) {
	v, err := in.evaluar(x.Valor, env)
	if err != nil {
		return nil, err
	}
	switch c := v.(type) {
	case *Equipo:
		return int64(c.Tamano()), nil
	case *Mochila:
		return int64(c.Tamano()), nil
	case string:
		return int64(TamanoPlanta(c)), nil
	}
	return nil, in.interno(x, "tamaño no se aplica a %s", literal(v))
}

// evaluarAleatorio devuelve un roca al azar entre los extremos, ambos
// incluidos (sección 5).
func (in *Interprete) evaluarAleatorio(x *ast.Aleatorio, env *Entorno) (Value, error) {
	a, err := in.evaluar(x.Min, env)
	if err != nil {
		return nil, err
	}
	b, err := in.evaluar(x.Max, env)
	if err != nil {
		return nil, err
	}
	minimo, ok1 := a.(int64)
	maximo, ok2 := b.(int64)
	if !ok1 || !ok2 {
		return nil, in.interno(x, "aleatorio necesita dos roca")
	}
	if minimo > maximo {
		return nil, in.fallo(x, CodigoAleatorioRango,
			fmt.Sprintf("aleatorio(%s, %s) no tiene números posibles.", describir(x.Min, a), describir(x.Max, b)),
			"el primer extremo es mayor que el segundo.",
			"escribe primero el menor: aleatorio(menor, mayor).")
	}
	// La resta en uint64 no desborda aunque el rango sea todo int64.
	ancho := uint64(maximo) - uint64(minimo)
	var salto uint64
	if ancho == math.MaxUint64 {
		salto = in.Azar.Uint64()
	} else {
		salto = in.Azar.Uint64N(ancho + 1)
	}
	return int64(uint64(minimo) + salto), nil
}

// evaluarRedondear redondea al entero más cercano; la mitad exacta se aleja
// de cero (sección 5), que es lo que hace math.Round.
func (in *Interprete) evaluarRedondear(x *ast.Redondear, env *Entorno) (Value, error) {
	v, err := in.evaluar(x.Valor, env)
	if err != nil {
		return nil, err
	}
	switch n := v.(type) {
	case int64:
		return n, nil // roca → agua es automática y redondear no la cambia
	case float64:
		r, err := aRoca(math.Round(n))
		if err != nil {
			return nil, in.falloDesbordamiento(x, "redondear("+describir(x.Valor, v)+")")
		}
		return r, nil
	}
	return nil, in.interno(x, "redondear no se aplica a %s", literal(v))
}

// evaluarConvertir aplica las casillas RC y MT de la tabla 3.2.
func (in *Interprete) evaluarConvertir(x *ast.Convertir, env *Entorno) (Value, error) {
	v, err := in.evaluar(x.Valor, env)
	if err != nil {
		return nil, err
	}
	d := x.Destino
	if d.Forma == ast.TipoSimple {
		switch d.Simple {
		case token.PLANTA:
			switch v.(type) {
			case int64, float64, rune, string, bool, EspecieVal:
				return Texto(v), nil
			}
		case token.ROCA:
			switch n := v.(type) {
			case int64:
				return n, nil
			case float64:
				r, err := aRoca(math.Trunc(n)) // convertir trunca, no redondea
				if err != nil {
					return nil, in.falloDesbordamiento(x, "convertir("+describir(x.Valor, v)+") a roca")
				}
				return r, nil
			case string:
				return in.convertirTexto(x, n, d)
			}
		case token.AGUA:
			switch n := v.(type) {
			case int64:
				return float64(n), nil
			case float64:
				return n, nil
			case string:
				return in.convertirTexto(x, n, d)
			}
		case token.FUEGO, token.ELECTRICO:
			if (d.Simple == token.FUEGO && esRune(v)) || (d.Simple == token.ELECTRICO && esBool(v)) {
				return v, nil
			}
		}
	} else if ev, ok := v.(EspecieVal); ok && d.Forma == ast.TipoNombrado && ev.Tipo == d.Nombre {
		return ev, nil
	}
	return nil, in.interno(x, "la tabla de efectividades no permite convertir %s a %s", literal(v), nombreTipo(d))
}

// convertirTexto lee un número desde un planta con las mismas reglas que
// capturar (decisión H5).
func (in *Interprete) convertirTexto(x *ast.Convertir, texto string, d *ast.TipoExpr) (Value, error) {
	v, problema := in.leerValor(texto, d)
	if problema != "" {
		return nil, in.fallo(x, CodigoConversion,
			fmt.Sprintf("no se puede convertir %s a %s.", describir(x.Valor, texto), nombreTipo(d)),
			problema, "comprueba el texto antes de convertirlo, o pídelo con capturar, que vuelve a preguntar.")
	}
	return v, nil
}

func esRune(v Value) bool { _, ok := v.(rune); return ok }
func esBool(v Value) bool { _, ok := v.(bool); return ok }

// ensancharComo aplica roca → agua a v si el valor con el que se va a
// comparar es un agua. El analizador acepta "equipo de agua contiene 1"
// porque la conversión es automática (tipos.Asignable); sin esto, 1 y 1.0
// nunca serían iguales.
func ensancharComo(v, modelo Value) Value {
	if n, ok := v.(int64); ok {
		if _, esAgua := modelo.(float64); esAgua {
			return float64(n)
		}
	}
	return v
}
