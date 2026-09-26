package servicio

import (
	"fmt"
	"sort"
	"strings"

	"github.com/keylorpineda/PokeScript/internal/analizador"
	"github.com/keylorpineda/PokeScript/internal/asistente"
	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/tipos"
)

// ResultadoCompilacion es lo que CompilarProyecto le devuelve a la interfaz
// (sección 9), listo para serializar a JSON.
type ResultadoCompilacion struct {
	// Exito es true si no hay errores; puede haber advertencias.
	Exito bool `json:"exito"`
	// Encabezado resume el resultado: «¡Es superefectivo!» si compiló.
	Encabezado   string        `json:"encabezado"`
	Principal    string        `json:"principal"`
	Archivos     []string      `json:"archivos"`
	Diagnosticos []Diagnostico `json:"diagnosticos"`
	// Simbolos tiene, por archivo, los nombres para el autocompletado.
	Simbolos map[string][]Simbolo `json:"simbolos"`
}

// Diagnostico es un diag.Diagnostic con la lección del asistente para el
// panel pedagógico.
type Diagnostico struct {
	diag.Diagnostic
	Leccion asistente.Leccion `json:"leccion"`
}

// Clases de símbolo para el autocompletado.
const (
	SimboloMovimiento = "movimiento"
	SimboloEspecie    = "especie"
	SimboloValor      = "valor"
	SimboloFicha      = "ficha"
	SimboloMedalla    = "medalla"
	SimboloDato       = "dato"
	SimboloParametro  = "parametro"
)

// Simbolo es un nombre que el editor puede sugerir.
type Simbolo struct {
	Nombre string `json:"nombre"`
	Clase  string `json:"clase"`
	// Tipo es el tipo del dato, medalla o valor, o lo que entrega un
	// movimiento; vacío si no se conoce o no aplica.
	Tipo string `json:"tipo,omitempty"`
	// Detalle es la firma de un movimiento, los campos de una ficha o los
	// valores de una especie, para mostrar junto a la sugerencia.
	Detalle string `json:"detalle,omitempty"`
}

// Resultado arma lo que se envía a la interfaz.
func (c Compilacion) Resultado() ResultadoCompilacion {
	r := ResultadoCompilacion{
		Exito:        !c.TieneErrores(),
		Principal:    c.Proyecto.Principal,
		Diagnosticos: []Diagnostico{},
		Simbolos:     map[string][]Simbolo{},
	}
	if r.Exito {
		r.Encabezado = diag.EncabezadoExito
	} else {
		r.Encabezado = c.Diagnosticos[primerError(c.Diagnosticos)].Heading
	}
	for n := range c.Proyecto.Archivos {
		r.Archivos = append(r.Archivos, n)
		r.Simbolos[n] = c.Simbolos(n)
	}
	sort.Strings(r.Archivos)
	for _, d := range c.Diagnosticos {
		r.Diagnosticos = append(r.Diagnosticos, Diagnostico{Diagnostic: d, Leccion: asistente.Explicar(d)})
	}
	return r
}

func primerError(ds []diag.Diagnostic) int {
	for i, d := range ds {
		if d.Severity == diag.Error {
			return i
		}
	}
	return 0
}

// Simbolos devuelve los nombres que se pueden usar en un archivo, ordenados
// por nombre: lo que declara e importa, con sus tipos, y los datos,
// parámetros y medallas locales de sus cuerpos. Sin tabla (el proyecto tiene
// errores de sintaxis o de importación) devuelve solo los locales.
func (c Compilacion) Simbolos(archivo string) []Simbolo {
	var ss []Simbolo
	vistos := map[string]bool{}
	agregar := func(s Simbolo) {
		if !vistos[s.Nombre+"/"+s.Clase] {
			vistos[s.Nombre+"/"+s.Clase] = true
			ss = append(ss, s)
		}
	}
	if c.Tabla != nil {
		for _, n := range c.Tabla.Visibles(archivo) {
			if s := c.Tabla.Buscar(archivo, n); s != nil {
				agregar(simboloDe(n, s))
			}
		}
	}
	for _, s := range c.locales(archivo) {
		agregar(s)
	}
	sort.Slice(ss, func(i, j int) bool {
		if ss[i].Nombre != ss[j].Nombre {
			return ss[i].Nombre < ss[j].Nombre
		}
		return ss[i].Clase < ss[j].Clase
	})
	return ss
}

func simboloDe(nombre string, s *analizador.Simbolo) Simbolo {
	switch s.Clase {
	case analizador.ClaseMovimiento:
		m := s.Movimiento
		var params []string
		for _, p := range m.Params {
			params = append(params, texto(p.Tipo)+" "+p.Nombre)
		}
		firma := fmt.Sprintf("%s(%s)", m.Nombre, strings.Join(params, ", "))
		return Simbolo{Nombre: nombre, Clase: SimboloMovimiento, Tipo: texto(m.Retorno), Detalle: firma}
	case analizador.ClaseEspecie:
		return Simbolo{Nombre: nombre, Clase: SimboloEspecie, Detalle: strings.Join(s.Especie.Valores, ", ")}
	case analizador.ClaseValorEspecie:
		return Simbolo{Nombre: nombre, Clase: SimboloValor, Tipo: s.Especie.Nombre}
	case analizador.ClaseFicha:
		var campos []string
		for _, c := range s.Ficha.Campos {
			campos = append(campos, texto(c.Tipo)+" "+c.Nombre)
		}
		return Simbolo{Nombre: nombre, Clase: SimboloFicha, Detalle: strings.Join(campos, ", ")}
	}
	return Simbolo{Nombre: nombre, Clase: SimboloMedalla, Tipo: texto(s.Medalla.Tipo)}
}

// locales junta los datos, parámetros y variables de recorrido de un
// archivo, con su tipo cuando está escrito en la declaración.
func (c Compilacion) locales(archivo string) []Simbolo {
	a := c.Proyecto.Archivos[archivo]
	if a == nil {
		return nil
	}
	tipo := func(te *ast.TipoExpr) string {
		if c.Tabla == nil || te == nil {
			return ""
		}
		t, d := c.Tabla.ResolverTipo(archivo, te)
		if d != nil {
			return ""
		}
		return texto(t)
	}
	var ss []Simbolo
	ast.InspeccionarPrograma(a.Programa, func(n ast.Nodo) bool {
		switch x := n.(type) {
		case *ast.DeclDato:
			clase := SimboloDato
			if x.Medalla {
				clase = SimboloMedalla
			}
			if x.Nombre != nil {
				ss = append(ss, Simbolo{Nombre: x.Nombre.Nombre, Clase: clase, Tipo: tipo(x.Tipo)})
			}
		case *ast.Param:
			if x.Nombre != nil {
				ss = append(ss, Simbolo{Nombre: x.Nombre.Nombre, Clase: SimboloParametro, Tipo: tipo(x.Tipo)})
			}
		case *ast.RecorrerRango:
			if x.Var != nil {
				ss = append(ss, Simbolo{Nombre: x.Var.Nombre, Clase: SimboloDato, Tipo: "roca"})
			}
		case *ast.RecorrerColeccion:
			for _, v := range []*ast.Ident{x.Var, x.Var2} {
				if v != nil {
					ss = append(ss, Simbolo{Nombre: v.Nombre, Clase: SimboloDato})
				}
			}
		}
		return true
	})
	return ss
}

func texto(t *tipos.Type) string {
	if t == nil {
		return ""
	}
	return t.String()
}
