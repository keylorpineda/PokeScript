// cspell:words gritra mientas recorer fni

package asistente

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// Entrada es lo que el asistente necesita saber del proyecto. No depende
// del analizador: quien lo llama (el servicio de compilación) le pasa los
// textos y los nombres.
type Entrada struct {
	// Fuentes tiene el texto de cada archivo, para leer la palabra que
	// señala cada diagnóstico.
	Fuentes map[string]string
	// Nombres devuelve los nombres que se pueden usar en un archivo:
	// declarados, importados y locales.
	Nombres func(archivo string) []string
	// Tipos devuelve las especies y fichas que se pueden usar en un archivo.
	Tipos func(archivo string) []string
	// Exportados devuelve lo que declara un archivo del proyecto, para
	// corregir el nombre de una importación.
	Exportados func(archivo string) []string
	// Archivos devuelve los .pks del proyecto.
	Archivos func() []string
}

// Asistir completa los diagnósticos con sugerencias de corrección: cuando
// un nombre, un tipo, un archivo o una palabra reservada parece mal
// escrito, agrega «¿Quisiste decir …?» al principio de Suggest y un Fix que
// el editor puede aplicar con un clic (sección 7.1). Devuelve una copia;
// los diagnósticos sin sugerencia quedan igual.
func Asistir(diags []diag.Diagnostic, e Entrada) []diag.Diagnostic {
	salida := make([]diag.Diagnostic, len(diags))
	for i, d := range diags {
		salida[i] = d
		if d.Fix != nil {
			continue
		}
		switch d.Code {
		case "nombre-no-declarado":
			sugerirPalabra(&salida[i], e, candidatos(e.Nombres, d.File), reservadas())
		case "tipo-desconocido":
			// «gritra vida» se lee como un dato de tipo «gritra»: si no se
			// parece a ningún tipo, puede ser una palabra reservada mal escrita.
			sugerirPalabra(&salida[i], e, candidatos(e.Tipos, d.File), reservadas())
		case "nombre-inexistente":
			sugerirPalabra(&salida[i], e, candidatos(e.Exportados, rutaImportada(e, d)), nil)
		case "archivo-inexistente":
			sugerirArchivo(&salida[i], e)
		case "asignacion-esperada", "instruccion-esperada", "instruccion-fuera-de-bloque", "declaracion-esperada":
			sugerirPalabraReservada(&salida[i], e)
		}
	}
	return salida
}

func candidatos(f func(string) []string, arg string) []string {
	if f == nil {
		return nil
	}
	return f(arg)
}

// sugerirPalabra corrige la palabra que señala el diagnóstico con el
// candidato más parecido; si ninguno se parece, prueba con los de respaldo
// (las palabras reservadas).
func sugerirPalabra(d *diag.Diagnostic, e Entrada, candidatos, respaldo []string) {
	palabra := textoEn(e.Fuentes[d.File], d.Line, d.Col, d.Len)
	if palabra == "" {
		return
	}
	if utf8.RuneCountInString(palabra) < minimoReservada {
		respaldo = nil
	}
	for _, lista := range [][]string{candidatos, respaldo} {
		if parecido, ok := Parecido(palabra, lista); ok {
			corregir(d, d.Line, d.Col, d.Len, parecido)
			return
		}
	}
}

// sugerirArchivo corrige la ruta de un «enseñar … desde "…"». La posición
// incluye las comillas, que se conservan en el reemplazo.
func sugerirArchivo(d *diag.Diagnostic, e Entrada) {
	texto := textoEn(e.Fuentes[d.File], d.Line, d.Col, d.Len)
	ruta := strings.Trim(texto, `"`)
	if ruta == "" || e.Archivos == nil {
		return
	}
	if parecido, ok := Parecido(ruta, e.Archivos()); ok {
		d.Fix = &diag.Fix{Line: d.Line, Col: d.Col, Len: d.Len, Replacement: `"` + parecido + `"`}
		d.Suggest = unirSugerencia(fmt.Sprintf("¿Quisiste decir «%s»?", parecido), d.Suggest)
	}
}

// rutaImportada lee el archivo de un «enseñar … desde "ruta"» en la línea
// del diagnóstico.
func rutaImportada(e Entrada, d diag.Diagnostic) string {
	linea := lineaDe(e.Fuentes[d.File], d.Line)
	if i := strings.Index(linea, `desde "`); i >= 0 {
		resto := linea[i+len(`desde "`):]
		if j := strings.Index(resto, `"`); j >= 0 {
			return resto[:j]
		}
	}
	return ""
}

// sugerirPalabraReservada revisa la primera palabra de la línea: si no es
// un nombre del programa pero se parece a una palabra reservada, casi
// seguro es esa palabra mal escrita («gritra», «mientas», «fni»).
func sugerirPalabraReservada(d *diag.Diagnostic, e Entrada) {
	linea := []rune(lineaDe(e.Fuentes[d.File], d.Line))
	inicio := 0
	for inicio < len(linea) && (linea[inicio] == ' ' || linea[inicio] == '\t') {
		inicio++
	}
	fin := inicio
	for fin < len(linea) && (unicode.IsLetter(linea[fin]) || unicode.IsDigit(linea[fin]) || linea[fin] == '_') {
		fin++
	}
	palabra := string(linea[inicio:fin])
	if palabra == "" || token.Buscar(palabra) != token.IDENT {
		return
	}
	for _, n := range candidatos(e.Nombres, d.File) {
		if n == palabra {
			return // es un nombre válido del programa: el error es otro
		}
	}
	if utf8.RuneCountInString(palabra) < minimoReservada {
		return
	}
	if parecido, ok := Parecido(palabra, reservadas()); ok {
		corregir(d, d.Line, inicio+1, fin-inicio, parecido)
	}
}

// minimoReservada es el largo mínimo de una palabra para sugerir una
// palabra reservada: «x» está a una letra de «a», «o» e «y», y proponerlas
// confundiría más de lo que ayuda.
const minimoReservada = 3

func reservadas() []string { return token.PalabrasReservadas() }

func corregir(d *diag.Diagnostic, linea, col, largo int, reemplazo string) {
	d.Fix = &diag.Fix{Line: linea, Col: col, Len: largo, Replacement: reemplazo}
	d.Suggest = unirSugerencia(fmt.Sprintf("¿Quisiste decir «%s»?", reemplazo), d.Suggest)
}

func unirSugerencia(pregunta, original string) string {
	if original == "" {
		return pregunta
	}
	return pregunta + " Si no, " + original
}

// lineaDe devuelve la línea n (base 1) del texto, sin el salto de línea.
func lineaDe(fuente string, n int) string {
	lineas := strings.Split(fuente, "\n")
	if n < 1 || n > len(lineas) {
		return ""
	}
	return strings.TrimRight(lineas[n-1], "\r")
}

// textoEn devuelve el texto de largo runas desde la columna col (base 1).
func textoEn(fuente string, linea, col, largo int) string {
	runas := []rune(lineaDe(fuente, linea))
	if col < 1 || largo < 1 || col-1+largo > len(runas) {
		return ""
	}
	return string(runas[col-1 : col-1+largo])
}
