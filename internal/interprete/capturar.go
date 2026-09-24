package interprete

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// ejecutarCapturar muestra el mensaje, lee una línea y la guarda en el
// destino. Si la línea no corresponde al tipo, explica por qué y vuelve a
// preguntar, sin límite de intentos (sección 6).
func (in *Interprete) ejecutarCapturar(x *ast.Capturar, env *Entorno) error {
	m, err := in.evaluar(x.Mensaje, env)
	if err != nil {
		return err
	}
	mensaje, ok := m.(string)
	if !ok {
		return in.interno(x.Mensaje, "el mensaje de capturar debe ser planta")
	}
	tipo := in.tipoDe(x.Destino, env)
	if tipo == nil || tipo.Forma == ast.TipoEquipo || tipo.Forma == ast.TipoMochila || in.esFicha(tipo) {
		return in.interno(x.Destino, "capturar solo guarda en datos de tipo simple o especie (decisión H14)")
	}
	for {
		if err := in.detenido(); err != nil {
			return err
		}
		texto, err := in.ES.Leer(mensaje)
		if err != nil {
			return err // se detuvo la ejecución mientras esperaba
		}
		v, problema := in.leerValor(texto, tipo)
		if problema == "" {
			return in.asignar(x.Destino, v, env)
		}
		in.ES.Escribir(problema)
	}
}

func (in *Interprete) esFicha(t *ast.TipoExpr) bool {
	_, ok := in.fichas[t.Nombre]
	return t.Forma == ast.TipoNombrado && ok
}

var (
	formaRoca = regexp.MustCompile(`^-?[0-9]+$`)
	formaAgua = regexp.MustCompile(`^-?[0-9]+(\.[0-9]+)?$`)
)

// leerValor interpreta un texto escrito por el usuario como un valor del
// tipo t (decisiones H4 y H5). Si no sirve, devuelve una explicación para
// mostrarle.
//
// Se aceptan las mismas formas que los literales, con dos excepciones
// hechas a propósito: los números pueden llevar un - delante (el literal
// roca no lo lleva porque en el código el - es un operador), y un agua
// acepta un entero porque roca → agua es automática.
//
// Los espacios de los extremos se ignoran en todo salvo planta y fuego, donde
// el espacio es parte del valor. En un posible, la línea vacía es fantasma,
// salvo en posible planta, donde es el texto vacío.
func (in *Interprete) leerValor(texto string, t *ast.TipoExpr) (Value, string) {
	esTexto := t.Forma == ast.TipoSimple && (t.Simple == token.PLANTA || t.Simple == token.FUEGO)
	limpio := texto
	if !esTexto {
		limpio = strings.TrimSpace(texto)
	}
	if t.Posible && texto == "" && (t.Forma != ast.TipoSimple || t.Simple != token.PLANTA) {
		return nil, ""
	}
	if t.Posible && !esTexto && limpio == "" {
		return nil, ""
	}

	if t.Forma == ast.TipoNombrado {
		return in.leerEspecie(limpio, t.Nombre)
	}
	switch t.Simple {
	case token.ROCA:
		if !formaRoca.MatchString(limpio) {
			return nil, fmt.Sprintf("«%s» no es un roca. Escribe un número entero, por ejemplo 42.", texto)
		}
		n, err := strconv.ParseInt(limpio, 10, 64)
		if err != nil {
			return nil, fmt.Sprintf("«%s» no cabe en un roca. %s", texto, descripcionRangoRoca)
		}
		return n, ""
	case token.AGUA:
		if !formaAgua.MatchString(limpio) {
			return nil, fmt.Sprintf("«%s» no es un agua. Escribe un número con punto decimal, por ejemplo 3.5.", texto)
		}
		f, err := strconv.ParseFloat(limpio, 64)
		if err != nil {
			return nil, fmt.Sprintf("«%s» es demasiado grande para un agua.", texto)
		}
		return f, ""
	case token.FUEGO:
		if utf8.RuneCountInString(texto) != 1 {
			return nil, fmt.Sprintf("«%s» no es un fuego. Escribe exactamente un carácter.", texto)
		}
		r, _ := utf8.DecodeRuneInString(texto)
		return r, ""
	case token.PLANTA:
		return texto, ""
	case token.ELECTRICO:
		switch limpio {
		case "verdadero":
			return true, ""
		case "falso":
			return false, ""
		}
		return nil, fmt.Sprintf("«%s» no es un electrico. Escribe verdadero o falso.", texto)
	}
	return nil, fmt.Sprintf("no se puede leer un valor de tipo %s.", nombreTipo(t))
}

// leerEspecie acepta el nombre exacto de uno de los valores de la especie.
func (in *Interprete) leerEspecie(texto, especie string) (Value, string) {
	decl, ok := in.especies[especie]
	if !ok {
		return nil, fmt.Sprintf("no se puede leer un valor de tipo %s.", especie)
	}
	nombres := make([]string, len(decl.Valores))
	for i, v := range decl.Valores {
		if v.Nombre == texto {
			return EspecieVal{Tipo: especie, Valor: texto}, ""
		}
		nombres[i] = v.Nombre
	}
	return nil, fmt.Sprintf("«%s» no es un valor de %s. Escribe uno de estos: %s.",
		texto, especie, strings.Join(nombres, ", "))
}
