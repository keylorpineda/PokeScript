// Package diag define la estructura única de diagnóstico que Go envía a
// Svelte (sección 7 de la especificación).
package diag

import (
	"fmt"
	"strings"
)

// Severidad de un diagnóstico.
type Severidad string

const (
	Error       Severidad = "error"
	Advertencia Severidad = "advertencia"
)

// Categoria indica en qué fase se detectó el problema.
type Categoria string

const (
	Lexico      Categoria = "lexico"
	Sintactico  Categoria = "sintactico"
	Semantico   Categoria = "semantico"
	Importacion Categoria = "importacion"
	Ejecucion   Categoria = "ejecucion"
)

// Nombre devuelve la categoría escrita para mostrarla al usuario.
func (c Categoria) Nombre() string {
	switch c {
	case Lexico:
		return "léxico"
	case Sintactico:
		return "sintáctico"
	case Semantico:
		return "semántico"
	case Importacion:
		return "importación"
	case Ejecucion:
		return "ejecución"
	default:
		return string(c)
	}
}

// Encabezados temáticos (sección 7).
const (
	EncabezadoTipos       = "No es muy efectivo…"
	EncabezadoSintaxis    = "¡Se escapó!"
	EncabezadoSinValor    = "¡No pasó nada!"
	EncabezadoEjecucion   = "¡Falló el ataque!"
	EncabezadoImportacion = "No se encontró la ruta"
	EncabezadoAdvertencia = "¿Seguro que quieres hacer eso?"
	EncabezadoExito       = "¡Es superefectivo!"
)

// Fix es una corrección que el editor puede aplicar con un clic.
type Fix struct {
	Line        int    `json:"line"`
	Col         int    `json:"col"`
	Len         int    `json:"len"`
	Replacement string `json:"replacement"`
}

// Diagnostic describe un error o una advertencia con su posición exacta.
//
// Code no está en la especificación: es un subcódigo estable (por ejemplo
// "bloque-sin-cerrar") que el asistente usa junto con Category para buscar
// su plantilla de explicación. Ver docs/PLAN.md, registro de decisiones.
type Diagnostic struct {
	Severity Severidad `json:"severity"`
	Category Categoria `json:"category"`
	Code     string    `json:"code"`
	Heading  string    `json:"heading"`
	File     string    `json:"file"`
	Line     int       `json:"line"` // base 1
	Col      int       `json:"col"`  // base 1, en runas
	Len      int       `json:"len"`  // en runas
	Desc     string    `json:"desc"`
	Cause    string    `json:"cause,omitempty"`
	Suggest  string    `json:"suggest,omitempty"`
	Fix      *Fix      `json:"fix,omitempty"`
}

// String arma el mensaje completo en el formato de la sección 7.
func (d Diagnostic) String() string {
	var b strings.Builder
	b.WriteString(d.Heading)
	if d.Severity == Advertencia {
		fmt.Fprintf(&b, "\nTipo de advertencia: %s", d.Category.Nombre())
	} else {
		fmt.Fprintf(&b, "\nTipo de error: %s", d.Category.Nombre())
	}
	if d.File != "" {
		fmt.Fprintf(&b, "\nArchivo: %s", d.File)
	}
	fmt.Fprintf(&b, "\nLínea: %d\nColumna: %d\nDescripción: %s", d.Line, d.Col, d.Desc)
	if d.Cause != "" {
		fmt.Fprintf(&b, "\nPosible causa: %s", d.Cause)
	}
	if d.Suggest != "" {
		fmt.Fprintf(&b, "\nSugerencia: %s", d.Suggest)
	}
	return b.String()
}

// Lista acumula diagnósticos. Si Max es mayor que cero, deja de aceptar
// errores al llegar a ese número (el parser usa 20, sección 2.1); las
// advertencias no cuentan para el límite.
type Lista struct {
	Max     int
	items   []Diagnostic
	errores int
}

// Agregar suma d a la lista. Devuelve false si se alcanzó el límite de
// errores y d se descartó.
func (l *Lista) Agregar(d Diagnostic) bool {
	if d.Severity == Error {
		if l.Max > 0 && l.errores >= l.Max {
			return false
		}
		l.errores++
	}
	l.items = append(l.items, d)
	return true
}

// Lleno informa si ya se alcanzó el límite de errores.
func (l *Lista) Lleno() bool { return l.Max > 0 && l.errores >= l.Max }

// TieneErrores informa si hay al menos un diagnóstico de severidad Error.
func (l *Lista) TieneErrores() bool { return l.errores > 0 }

// Items devuelve los diagnósticos en el orden en que se agregaron.
func (l *Lista) Items() []Diagnostic { return l.items }
