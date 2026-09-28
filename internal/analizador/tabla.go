package analizador

import (
	"fmt"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/tipos"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// Tabla es la tabla de símbolos del proyecto: todo lo que se declara con
// alcance de archivo, con sus tipos ya resueltos. La arma la pasada 1
// (Recolectar) y la usan las validaciones de la pasada 2.
//
// Todos los nombres de alcance de archivo son únicos en el proyecto (ver
// docs/PLAN.md, registro de decisiones), así que cada mapa se indexa solo
// por nombre. Lo que cada archivo puede ver se consulta con Buscar.
type Tabla struct {
	Especies    map[string]*Especie
	Fichas      map[string]*Ficha
	Medallas    map[string]*Medalla
	Movimientos map[string]*Movimiento
	// Valores lleva cada valor de especie a su especie: "SANO" → Estado.
	Valores map[string]*Especie

	// visibles guarda, por archivo, los nombres de alcance de archivo que
	// ese archivo puede usar: los propios y los importados.
	visibles map[string]map[string]bool
}

// Especie es una enumeración: especie Estado SANO, DORMIDO fin.
type Especie struct {
	Nombre  string
	Archivo string
	Decl    *ast.DeclEspecie
	Valores []string // en el orden en que se declararon
}

// Tipo devuelve el tipo de los valores de la especie.
func (e *Especie) Tipo() *tipos.Type { return tipos.Especie(e.Nombre) }

// Ficha es un registro: ficha Pokemon planta nombre … fin.
type Ficha struct {
	Nombre  string
	Archivo string
	Decl    *ast.DeclFicha
	Campos  []Campo // en el orden en que se declararon
}

// Tipo devuelve el tipo de los valores de la ficha.
func (f *Ficha) Tipo() *tipos.Type { return tipos.Ficha(f.Nombre) }

// Campo busca un campo por nombre; devuelve nil si no existe.
func (f *Ficha) Campo(nombre string) *Campo {
	for i := range f.Campos {
		if f.Campos[i].Nombre == nombre {
			return &f.Campos[i]
		}
	}
	return nil
}

// Campo es un campo de una ficha.
type Campo struct {
	Nombre string
	Tipo   *tipos.Type // nil si su tipo no se pudo resolver
	Pos    ast.Pos
}

// Medalla es una constante de alcance de archivo.
type Medalla struct {
	Nombre  string
	Archivo string
	Decl    *ast.DeclMedalla
	Tipo    *tipos.Type // nil si su tipo no se pudo resolver
}

// Movimiento es la cabecera de un movimiento: su firma, sin el cuerpo.
type Movimiento struct {
	Nombre  string
	Archivo string
	Decl    *ast.DeclMovimiento
	Params  []Param
	Retorno *tipos.Type // nil si el movimiento no entrega valor
}

// Param es un parámetro de movimiento.
type Param struct {
	Nombre string
	Tipo   *tipos.Type // nil si su tipo no se pudo resolver
	Pos    ast.Pos
}

// ─── Consultas ─────────────────────────────────────────────────────────────

// Clase dice qué es un símbolo.
type Clase int

const (
	ClaseEspecie Clase = iota
	ClaseValorEspecie
	ClaseFicha
	ClaseMedalla
	ClaseMovimiento
)

// Simbolo es el resultado de Buscar. Solo uno de los punteros tiene valor,
// según Clase; en ClaseValorEspecie, Especie es la especie del valor.
type Simbolo struct {
	Clase      Clase
	Especie    *Especie
	Ficha      *Ficha
	Medalla    *Medalla
	Movimiento *Movimiento
}

// Buscar devuelve el símbolo de alcance de archivo llamado nombre, si el
// archivo lo puede ver (lo declara o lo importa). Un valor de especie se ve
// cuando se ve su especie. Devuelve nil si no existe o no es visible.
func (t *Tabla) Buscar(archivo, nombre string) *Simbolo {
	ve := t.visibles[archivo]
	if e, ok := t.Valores[nombre]; ok && ve[e.Nombre] {
		return &Simbolo{Clase: ClaseValorEspecie, Especie: e}
	}
	if !ve[nombre] {
		return nil
	}
	switch {
	case t.Especies[nombre] != nil:
		return &Simbolo{Clase: ClaseEspecie, Especie: t.Especies[nombre]}
	case t.Fichas[nombre] != nil:
		return &Simbolo{Clase: ClaseFicha, Ficha: t.Fichas[nombre]}
	case t.Medallas[nombre] != nil:
		return &Simbolo{Clase: ClaseMedalla, Medalla: t.Medallas[nombre]}
	case t.Movimientos[nombre] != nil:
		return &Simbolo{Clase: ClaseMovimiento, Movimiento: t.Movimientos[nombre]}
	}
	return nil
}

// Visibles devuelve los nombres que un archivo puede usar, incluidos los
// valores de sus especies. El asistente los usa para sugerir correcciones.
func (t *Tabla) Visibles(archivo string) []string {
	var nombres []string
	for n := range t.visibles[archivo] {
		nombres = append(nombres, n)
		if e := t.Especies[n]; e != nil {
			nombres = append(nombres, e.Valores...)
		}
	}
	return nombres
}

// ResolverTipo convierte un tipo escrito en el código en un tipos.Type,
// buscando los nombres de especie y ficha desde el archivo dado. Si el tipo
// no existe o no es válido, devuelve nil y el diagnóstico que lo explica.
func (t *Tabla) ResolverTipo(archivo string, te *ast.TipoExpr) (*tipos.Type, *diag.Diagnostic) {
	tipo, d := t.resolver(archivo, te)
	if d != nil {
		return nil, d
	}
	if ok, porque := tipos.Valido(tipo); !ok {
		return nil, errorSemantico(archivo, te.Pos, "tipo-invalido", diag.EncabezadoTipos,
			fmt.Sprintf("el tipo «%s» no se puede declarar: %s.", tipo, porque),
			"las reglas de la sección 3.1 limitan dónde se puede usar posible y qué puede ser clave de una mochila.",
			"cambia el tipo; por ejemplo, quita «posible» de una colección o de una ficha.")
	}
	return tipo, nil
}

func (t *Tabla) resolver(archivo string, te *ast.TipoExpr) (*tipos.Type, *diag.Diagnostic) {
	if te == nil {
		return nil, nil
	}
	var tipo *tipos.Type
	switch te.Forma {
	case ast.TipoSimple:
		tipo = tipoSimple(te.Simple)
	case ast.TipoNombrado:
		s := t.Buscar(archivo, te.Nombre)
		switch {
		case s != nil && s.Clase == ClaseEspecie:
			tipo = s.Especie.Tipo()
		case s != nil && s.Clase == ClaseFicha:
			tipo = s.Ficha.Tipo()
		default:
			return nil, t.tipoDesconocido(archivo, te)
		}
	case ast.TipoEquipo:
		elem, d := t.resolver(archivo, te.Elem)
		if d != nil {
			return nil, d
		}
		tipo = tipos.Equipo(elem)
	case ast.TipoMochila:
		clave, d := t.resolver(archivo, te.Clave)
		if d != nil {
			return nil, d
		}
		valor, d := t.resolver(archivo, te.Elem)
		if d != nil {
			return nil, d
		}
		tipo = tipos.Mochila(clave, valor)
	}
	if te.Posible {
		tipo = tipos.Posible(tipo)
	}
	return tipo, nil
}

func (t *Tabla) tipoDesconocido(archivo string, te *ast.TipoExpr) *diag.Diagnostic {
	sugerencia := "declara la especie o la ficha, o importa su nombre con «enseñar … desde»."
	if t.Especies[te.Nombre] != nil || t.Fichas[te.Nombre] != nil {
		origen := ""
		if e := t.Especies[te.Nombre]; e != nil {
			origen = e.Archivo
		} else {
			origen = t.Fichas[te.Nombre].Archivo
		}
		sugerencia = fmt.Sprintf("«%s» está en «%s»: agrega «enseñar %s desde \"%s\"» al inicio del archivo.", te.Nombre, origen, te.Nombre, origen)
	}
	return errorSemantico(archivo, te.Pos, "tipo-desconocido", diag.EncabezadoSinValor,
		fmt.Sprintf("el tipo «%s» no existe en este archivo.", te.Nombre),
		"un tipo con nombre debe ser una especie o una ficha declarada aquí o importada.",
		sugerencia)
}

func tipoSimple(k token.Kind) *tipos.Type {
	switch k {
	case token.ROCA:
		return tipos.Roca()
	case token.AGUA:
		return tipos.Agua()
	case token.FUEGO:
		return tipos.Fuego()
	case token.PLANTA:
		return tipos.Planta()
	}
	return tipos.Electrico()
}

func errorSemantico(archivo string, pos ast.Pos, codigo, encabezado, desc, causa, sugerencia string) *diag.Diagnostic {
	if pos.Len < 1 {
		pos.Len = 1
	}
	return &diag.Diagnostic{
		Severity: diag.Error,
		Category: diag.Semantico,
		Code:     codigo,
		Heading:  encabezado,
		File:     archivo,
		Line:     pos.Line,
		Col:      pos.Col,
		Len:      pos.Len,
		Desc:     desc,
		Cause:    causa,
		Suggest:  sugerencia,
	}
}
