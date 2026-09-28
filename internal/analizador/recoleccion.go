package analizador

import (
	"fmt"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/proyecto"
	"github.com/keylorpineda/PokeScript/internal/tipos"
)

// Resultado es lo que produce una pasada del analizador.
type Resultado struct {
	Tabla        *Tabla
	Diagnosticos []diag.Diagnostic
}

// Recolectar es la pasada 1 (sección 4): registra especies, fichas,
// medallas y cabeceras de movimientos de todos los archivos alcanzables del
// proyecto, y verifica que haya exactamente un combate, en el principal.
// Las importaciones ya las validó internal/proyecto.
func Recolectar(p *proyecto.Proyecto) Resultado {
	r := &recolector{
		p: p,
		t: &Tabla{
			Especies:    map[string]*Especie{},
			Fichas:      map[string]*Ficha{},
			Medallas:    map[string]*Medalla{},
			Movimientos: map[string]*Movimiento{},
			Valores:     map[string]*Especie{},
			visibles:    map[string]map[string]bool{},
		},
		origen: map[string]origen{},
	}
	for _, n := range p.Orden {
		r.registrar(p.Archivos[n])
	}
	r.revisarCombate()
	for _, n := range p.Orden {
		r.calcularVisibles(p.Archivos[n])
	}
	for _, n := range p.Orden {
		r.resolverTipos(n)
	}
	return Resultado{Tabla: r.t, Diagnosticos: r.diags}
}

type recolector struct {
	p     *proyecto.Proyecto
	t     *Tabla
	diags []diag.Diagnostic
	// origen dice dónde se declaró cada nombre de alcance de archivo,
	// incluidos los valores de especie, para reportar duplicados.
	origen   map[string]origen
	combates []combateEn
}

type origen struct {
	archivo string
	pos     ast.Pos
	que     string // "el movimiento", "la especie", "el valor de especie"…
}

type combateEn struct {
	archivo string
	pos     ast.Pos
}

func (r *recolector) error(d *diag.Diagnostic) { r.diags = append(r.diags, *d) }

// ─── Paso 1: registrar nombres ─────────────────────────────────────────────

func (r *recolector) registrar(a *proyecto.Archivo) {
	for _, d := range a.Programa.Declaraciones {
		switch x := d.(type) {
		case *ast.DeclEspecie:
			if x.Nombre == nil || !r.nuevoNombre(a.Nombre, x.Nombre, "la especie") {
				continue
			}
			e := &Especie{Nombre: x.Nombre.Nombre, Archivo: a.Nombre, Decl: x}
			r.t.Especies[e.Nombre] = e
			for _, v := range x.Valores {
				if r.nuevoNombre(a.Nombre, v, "el valor de especie") {
					e.Valores = append(e.Valores, v.Nombre)
					r.t.Valores[v.Nombre] = e
				}
			}
		case *ast.DeclFicha:
			if x.Nombre != nil && r.nuevoNombre(a.Nombre, x.Nombre, "la ficha") {
				r.t.Fichas[x.Nombre.Nombre] = &Ficha{Nombre: x.Nombre.Nombre, Archivo: a.Nombre, Decl: x}
			}
		case *ast.DeclMedalla:
			if x.Nombre != nil && r.nuevoNombre(a.Nombre, x.Nombre, "la medalla") {
				r.t.Medallas[x.Nombre.Nombre] = &Medalla{Nombre: x.Nombre.Nombre, Archivo: a.Nombre, Decl: x}
			}
		case *ast.DeclMovimiento:
			if x.Nombre != nil && r.nuevoNombre(a.Nombre, x.Nombre, "el movimiento") {
				r.t.Movimientos[x.Nombre.Nombre] = &Movimiento{Nombre: x.Nombre.Nombre, Archivo: a.Nombre, Decl: x}
			}
		case *ast.Combate:
			r.combates = append(r.combates, combateEn{archivo: a.Nombre, pos: x.Pos})
		}
	}
}

// nuevoNombre registra un nombre de alcance de archivo. Si ya existía en el
// proyecto, reporta el duplicado señalando dónde está el primero y devuelve
// false.
func (r *recolector) nuevoNombre(archivo string, id *ast.Ident, que string) bool {
	previo, existe := r.origen[id.Nombre]
	if !existe {
		r.origen[id.Nombre] = origen{archivo: archivo, pos: id.Pos, que: que}
		return true
	}
	codigo := "nombre-duplicado"
	switch {
	case que == "el movimiento" && previo.que == "el movimiento":
		codigo = "movimiento-duplicado"
	case que == "el valor de especie" && previo.que == "el valor de especie":
		codigo = "valor-de-especie-repetido"
	}
	causa := "cada nombre declarado fuera de un bloque es único en todo el proyecto."
	if codigo == "movimiento-duplicado" {
		causa = "en PokeScript no hay sobrecarga: dos movimientos no pueden llamarse igual, aunque reciban parámetros distintos."
	}
	r.error(errorSemantico(archivo, id.Pos, codigo, diag.EncabezadoTipos,
		fmt.Sprintf("«%s» ya existe: %s «%s» está declarado en %s, línea %d.", id.Nombre, previo.que, id.Nombre, previo.archivo, previo.pos.Line),
		causa,
		"cambia uno de los dos nombres."))
	return false
}

// revisarCombate exige exactamente un combate, y que esté en el principal.
func (r *recolector) revisarCombate() {
	enPrincipal := 0
	for _, c := range r.combates {
		switch {
		case c.archivo != r.p.Principal:
			r.error(errorSemantico(c.archivo, c.pos, "combate-fuera-del-principal", diag.EncabezadoTipos,
				fmt.Sprintf("este archivo tiene un bloque combate, pero el archivo principal del proyecto es «%s».", r.p.Principal),
				"el programa empieza en el combate del archivo principal; los demás archivos solo declaran lo que el principal importa.",
				"mueve este combate al archivo principal o quita ese bloque."))
		case enPrincipal > 0:
			r.error(errorSemantico(c.archivo, c.pos, "combate-repetido", diag.EncabezadoTipos,
				"el archivo principal tiene más de un bloque combate.",
				"el programa tiene un único punto de inicio.",
				"junta los dos bloques en uno."))
		default:
			enPrincipal++
		}
	}
	if enPrincipal == 0 {
		r.error(errorSemantico(r.p.Principal, ast.Pos{Line: 1, Col: 1, Len: 1}, "sin-combate", diag.EncabezadoTipos,
			fmt.Sprintf("el archivo principal «%s» no tiene un bloque combate.", r.p.Principal),
			"el programa empieza a ejecutarse en combate.",
			"agrega al final del archivo:\ncombate\n    gritar \"¡Hola!\"\nfin"))
	}
}

// ─── Paso 2: visibilidad ───────────────────────────────────────────────────

// calcularVisibles junta lo que el archivo declara y lo que importa.
func (r *recolector) calcularVisibles(a *proyecto.Archivo) {
	ve := map[string]bool{}
	for nombre, o := range r.origen {
		if o.archivo == a.Nombre && o.que != "el valor de especie" {
			ve[nombre] = true
		}
	}
	for _, im := range a.Programa.Importaciones {
		for _, n := range im.Nombres {
			// Solo si el archivo de origen de verdad lo declara: los demás
			// casos ya los reportó internal/proyecto.
			if o, ok := r.origen[n.Nombre]; ok && o.archivo == im.Ruta && o.que != "el valor de especie" {
				ve[n.Nombre] = true
			}
		}
	}
	r.t.visibles[a.Nombre] = ve
}

// ─── Paso 3: tipos de las cabeceras ────────────────────────────────────────

func (r *recolector) resolverTipos(archivo string) {
	for _, d := range r.p.Archivos[archivo].Programa.Declaraciones {
		switch x := d.(type) {
		case *ast.DeclFicha:
			if f := r.t.Fichas[nombre(x.Nombre)]; f != nil && f.Decl == x {
				r.resolverCampos(f)
			}
		case *ast.DeclMedalla:
			if m := r.t.Medallas[nombre(x.Nombre)]; m != nil && m.Decl == x {
				m.Tipo = r.tipo(archivo, x.Tipo)
			}
		case *ast.DeclMovimiento:
			if mv := r.t.Movimientos[nombre(x.Nombre)]; mv != nil && mv.Decl == x {
				r.resolverFirma(mv)
			}
		}
	}
}

func (r *recolector) tipo(archivo string, te *ast.TipoExpr) *tipos.Type {
	t, d := r.t.ResolverTipo(archivo, te)
	if d != nil {
		r.error(d)
	}
	return t
}

func (r *recolector) resolverCampos(f *Ficha) {
	vistos := map[string]ast.Pos{}
	for _, c := range f.Decl.Campos {
		if c.Nombre == nil {
			continue
		}
		if previo, repetido := vistos[c.Nombre.Nombre]; repetido {
			r.error(errorSemantico(f.Archivo, c.Nombre.Pos, "campo-repetido", diag.EncabezadoTipos,
				fmt.Sprintf("la ficha «%s» ya tiene un campo «%s» (línea %d).", f.Nombre, c.Nombre.Nombre, previo.Line),
				"cada campo de una ficha tiene un nombre distinto.",
				"cambia el nombre de uno de los dos campos."))
			continue
		}
		vistos[c.Nombre.Nombre] = c.Nombre.Pos
		f.Campos = append(f.Campos, Campo{Nombre: c.Nombre.Nombre, Tipo: r.tipo(f.Archivo, c.Tipo), Pos: c.Nombre.Pos})
	}
}

func (r *recolector) resolverFirma(mv *Movimiento) {
	vistos := map[string]bool{}
	for _, p := range mv.Decl.Params {
		if p.Nombre == nil {
			continue
		}
		if vistos[p.Nombre.Nombre] {
			r.error(errorSemantico(mv.Archivo, p.Nombre.Pos, "parametro-repetido", diag.EncabezadoTipos,
				fmt.Sprintf("el movimiento «%s» ya tiene un parámetro «%s».", mv.Nombre, p.Nombre.Nombre),
				"cada parámetro de un movimiento tiene un nombre distinto.",
				"cambia el nombre de uno de los dos parámetros."))
			continue
		}
		vistos[p.Nombre.Nombre] = true
		mv.Params = append(mv.Params, Param{Nombre: p.Nombre.Nombre, Tipo: r.tipo(mv.Archivo, p.Tipo), Pos: p.Nombre.Pos})
	}
	if mv.Decl.Retorno != nil {
		mv.Retorno = r.tipo(mv.Archivo, mv.Decl.Retorno)
	}
}

func nombre(id *ast.Ident) string {
	if id == nil {
		return ""
	}
	return id.Nombre
}
