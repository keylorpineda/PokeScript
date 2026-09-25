package analizador

import (
	"sort"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/proyecto"
)

// Analizar corre las dos pasadas de la sección 4 sobre un proyecto ya
// cargado: primero Recolectar arma la Tabla, después cada verificación
// registrada revisa los cuerpos. No se detiene en el primer error.
//
// Los diagnósticos de la pasada 1 van primero; los de la pasada 2 se
// ordenan por archivo (en el orden del proyecto), línea y columna.
func Analizar(p *proyecto.Proyecto) Resultado {
	r := Recolectar(p)
	c := &Contexto{Proyecto: p, Tabla: r.Tabla}
	for _, v := range verificacionesOrdenadas() {
		v.revisar(c)
	}
	c.ordenar()
	return Resultado{Tabla: r.Tabla, Diagnosticos: append(r.Diagnosticos, c.diags...)}
}

// Contexto es lo que recibe cada verificación de la pasada 2.
type Contexto struct {
	Proyecto *proyecto.Proyecto
	Tabla    *Tabla
	diags    []diag.Diagnostic
}

// Archivos devuelve los archivos a revisar, en el orden del proyecto.
func (c *Contexto) Archivos() []*proyecto.Archivo {
	var as []*proyecto.Archivo
	for _, n := range c.Proyecto.Orden {
		as = append(as, c.Proyecto.Archivos[n])
	}
	return as
}

// Error reporta un error semántico en la posición de un nodo. Para los
// errores de tipos se usa diag.EncabezadoTipos; para nombres no declarados
// o datos sin valor, diag.EncabezadoSinValor (sección 7).
func (c *Contexto) Error(archivo string, pos ast.Pos, codigo, encabezado, desc, causa, sugerencia string) {
	c.diags = append(c.diags, *errorSemantico(archivo, pos, codigo, encabezado, desc, causa, sugerencia))
}

// Advertencia reporta algo que no impide ejecutar el programa (sección 4,
// validaciones 18 a 23). Siempre usa diag.EncabezadoAdvertencia.
func (c *Contexto) Advertencia(archivo string, pos ast.Pos, codigo, desc, causa, sugerencia string) {
	d := errorSemantico(archivo, pos, codigo, diag.EncabezadoAdvertencia, desc, causa, sugerencia)
	d.Severity = diag.Advertencia
	c.diags = append(c.diags, *d)
}

func (c *Contexto) ordenar() {
	orden := map[string]int{}
	for i, n := range c.Proyecto.Orden {
		orden[n] = i
	}
	sort.SliceStable(c.diags, func(i, j int) bool {
		a, b := c.diags[i], c.diags[j]
		if a.File != b.File {
			return orden[a.File] < orden[b.File]
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.Col < b.Col
	})
}

// ─── Registro de verificaciones ────────────────────────────────────────────

type verificacion struct {
	nombre  string
	revisar func(*Contexto)
}

var verificaciones []verificacion

// registrar agrega una verificación de la pasada 2. Cada archivo del
// paquete llama a registrar desde su propio init, así nadie tiene que
// editar una lista compartida:
//
//	func init() { registrar("tipos", revisarTipos) }
//
//	func revisarTipos(c *Contexto) {
//	    for _, a := range c.Archivos() {
//	        … recorrer a.Programa y llamar a c.Error(…) …
//	    }
//	}
func registrar(nombre string, revisar func(*Contexto)) {
	verificaciones = append(verificaciones, verificacion{nombre: nombre, revisar: revisar})
}

// verificacionesOrdenadas las devuelve por nombre, para que el resultado no
// dependa del orden en que Go ejecuta los init.
func verificacionesOrdenadas() []verificacion {
	vs := append([]verificacion{}, verificaciones...)
	sort.Slice(vs, func(i, j int) bool { return vs[i].nombre < vs[j].nombre })
	return vs
}
