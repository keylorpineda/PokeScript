package astprueba

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/keylorpineda/PokeScript/internal/ast"
)

var (
	tipoPos = reflect.TypeOf(ast.Pos{})
	tipoSi  = reflect.TypeOf(ast.Si{})
)

// Diferencia compara dos árboles ignorando todas las posiciones (los campos
// de tipo ast.Pos). Devuelve "" si son iguales, o la ruta del primer lugar
// donde difieren, por ejemplo "Declaraciones[1].Cuerpo[3].Valor.Op: + ≠ -".
//
// En general un slice nil y uno vacío cuentan como iguales. La excepción es
// Si.Sino: el contrato del AST dice que nunca es nil, así que ahí sí se
// distinguen y la prueba del parser lo hace cumplir.
//
// Cuando exista el parser, la prueba de contrato queda así:
//
//	for archivo, want := range astprueba.Seccion10() {
//	    got := parser.Analizar(archivo, fuente)
//	    if d := astprueba.Diferencia(got.Programa, want); d != "" {
//	        t.Errorf("%s: %s", archivo, d)
//	    }
//	}
func Diferencia(a, b any) string {
	return diferencia("", reflect.ValueOf(a), reflect.ValueOf(b))
}

func diferencia(ruta string, a, b reflect.Value) string {
	if a.IsValid() != b.IsValid() {
		return ruta + ": uno de los dos falta"
	}
	if !a.IsValid() {
		return ""
	}
	if a.Type() != b.Type() {
		return fmt.Sprintf("%s: %s ≠ %s", nombreRuta(ruta), a.Type(), b.Type())
	}
	switch a.Kind() {
	case reflect.Pointer, reflect.Interface:
		if a.IsNil() || b.IsNil() {
			if a.IsNil() != b.IsNil() {
				return fmt.Sprintf("%s: %s ≠ %s", nombreRuta(ruta), describirNil(a), describirNil(b))
			}
			return ""
		}
		return diferencia(ruta, a.Elem(), b.Elem())
	case reflect.Struct:
		if a.Type() == tipoPos {
			return ""
		}
		for i := 0; i < a.NumField(); i++ {
			campo := a.Type().Field(i)
			if !campo.IsExported() {
				continue
			}
			sub := campo.Name
			if campo.Anonymous && campo.Type == tipoPos {
				continue
			}
			if ruta != "" {
				sub = ruta + "." + sub
			}
			if a.Type() == tipoSi && campo.Name == "Sino" && a.Field(i).IsNil() != b.Field(i).IsNil() {
				return fmt.Sprintf("%s: nil ≠ vacío (el contrato pide un slice vacío)", sub)
			}
			if d := diferencia(sub, a.Field(i), b.Field(i)); d != "" {
				return d
			}
		}
		return ""
	case reflect.Slice:
		// Un slice nil y uno vacío son iguales: ninguno tiene elementos
		// (salvo Si.Sino, que se revisa arriba).
		if a.Len() != b.Len() {
			return fmt.Sprintf("%s: %d elementos ≠ %d", nombreRuta(ruta), a.Len(), b.Len())
		}
		for i := 0; i < a.Len(); i++ {
			if d := diferencia(ruta+"["+strconv.Itoa(i)+"]", a.Index(i), b.Index(i)); d != "" {
				return d
			}
		}
		return ""
	}
	if !reflect.DeepEqual(a.Interface(), b.Interface()) {
		return fmt.Sprintf("%s: %v ≠ %v", nombreRuta(ruta), a.Interface(), b.Interface())
	}
	return ""
}

func nombreRuta(r string) string {
	if r == "" {
		return "raíz"
	}
	return r
}

func describirNil(v reflect.Value) string {
	if v.IsNil() {
		return "nil"
	}
	return v.Elem().Type().String()
}

// Hojas devuelve, en el orden en que aparecen en el código, los nombres y
// literales de un árbol: identificadores, nombres de tipo, rutas de
// importación y valores de los literales. Debe coincidir con los tokens
// IDENT y *_LIT que el lexer produce del mismo archivo; así se comprueba
// que un fixture armado a mano corresponde a su .pks.
func Hojas(n any) []string {
	var hojas []string
	var visitar func(v reflect.Value)
	visitar = func(v reflect.Value) {
		switch v.Kind() {
		case reflect.Pointer, reflect.Interface:
			if !v.IsNil() {
				visitar(v.Elem())
			}
			return
		case reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				visitar(v.Index(i))
			}
			return
		case reflect.Struct:
		default:
			return
		}
		switch x := v.Addr().Interface().(type) {
		case *ast.Ident:
			hojas = append(hojas, x.Nombre)
			return
		case *ast.LitRoca:
			hojas = append(hojas, strconv.FormatInt(x.Valor, 10))
			return
		case *ast.LitAgua:
			hojas = append(hojas, strconv.FormatFloat(x.Valor, 'f', -1, 64))
			return
		case *ast.LitFuego:
			hojas = append(hojas, string(x.Valor))
			return
		case *ast.LitPlanta:
			hojas = append(hojas, x.Valor)
			return
		case *ast.TipoExpr:
			if x.Forma == ast.TipoNombrado {
				hojas = append(hojas, x.Nombre)
				return
			}
		case *ast.Importacion:
			for _, id := range x.Nombres {
				hojas = append(hojas, id.Nombre)
			}
			hojas = append(hojas, x.Ruta)
			return
		case *ast.Programa:
			for _, im := range x.Importaciones {
				visitar(reflect.ValueOf(im))
			}
			for _, d := range x.Declaraciones {
				visitar(reflect.ValueOf(d))
			}
			return
		}
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).IsExported() {
				visitar(v.Field(i))
			}
		}
	}
	visitar(reflect.ValueOf(n))
	return hojas
}
