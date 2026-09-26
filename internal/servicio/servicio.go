// Package servicio une todas las fases en dos operaciones: Compilar (leer
// el proyecto, analizarlo y completar los diagnósticos con el asistente) y
// Ejecutar. Es lo único que necesitan cmd/pks y el futuro app.go de Wails:
// el puente con la interfaz queda en unas pocas líneas.
package servicio

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"sort"

	"github.com/keylorpineda/PokeScript/internal/analizador"
	"github.com/keylorpineda/PokeScript/internal/asistente"
	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/interprete"
	"github.com/keylorpineda/PokeScript/internal/proyecto"
)

// Compilacion es el resultado de compilar un proyecto.
type Compilacion struct {
	Proyecto *proyecto.Proyecto
	// Tabla es nil si el proyecto tiene errores de sintaxis o de
	// importación: en ese caso el analizador no llega a correr.
	Tabla *analizador.Tabla
	// Diagnosticos junta, en orden, los errores léxicos y sintácticos, los
	// de importación y los del analizador, ya completados por el asistente.
	Diagnosticos []diag.Diagnostic
}

// TieneErrores informa si hay al menos un error (las advertencias no impiden
// ejecutar).
func (c Compilacion) TieneErrores() bool {
	for _, d := range c.Diagnosticos {
		if d.Severity == diag.Error {
			return true
		}
	}
	return false
}

// Compilar lee y analiza un proyecto. ruta puede ser una carpeta de
// proyecto o un archivo .pks suelto. Devuelve error solo si no se puede
// armar el proyecto; los problemas del código van en Diagnosticos.
func Compilar(ruta string) (Compilacion, error) {
	info, err := os.Stat(ruta)
	var r proyecto.Resultado
	if err == nil && info.IsDir() {
		r, err = proyecto.Cargar(ruta)
	} else {
		r, err = proyecto.CargarArchivo(ruta)
	}
	if err != nil {
		return Compilacion{}, err
	}
	return compilar(r), nil
}

// CompilarFS compila un proyecto de cualquier sistema de archivos; lo usan
// las pruebas y el IDE para compilar lo que hay en el editor sin guardarlo.
func CompilarFS(fsys fs.FS, nombre string) (Compilacion, error) {
	r, err := proyecto.CargarFS(fsys, nombre)
	if err != nil {
		return Compilacion{}, err
	}
	return compilar(r), nil
}

func compilar(r proyecto.Resultado) Compilacion {
	c := Compilacion{Proyecto: r.Proyecto, Diagnosticos: r.Diagnosticos}
	if !r.TieneErrores() {
		// El analizador solo tiene sentido sobre un proyecto que se pudo
		// leer completo.
		a := analizador.Analizar(r.Proyecto)
		c.Tabla = a.Tabla
		c.Diagnosticos = append(c.Diagnosticos, a.Diagnosticos...)
	}
	c.Diagnosticos = asistente.Asistir(c.Diagnosticos, c.entradaAsistente())
	return c
}

// Ejecutar corre el programa ya compilado con el intérprete dado, que
// escribe y lee por su interfaz ES. No ejecuta nada si la compilación tiene
// errores. Devuelve lo mismo que interprete.EjecutarProyecto.
func Ejecutar(ctx context.Context, c Compilacion, in *interprete.Interprete) error {
	if c.TieneErrores() {
		return ErrHayErrores
	}
	var archivos []*ast.Programa
	for _, n := range c.Proyecto.Orden {
		archivos = append(archivos, c.Proyecto.Archivos[n].Programa)
	}
	return in.EjecutarProyecto(ctx, archivos)
}

// ErrHayErrores indica que se intentó ejecutar un programa con errores.
var ErrHayErrores = errors.New("el programa tiene errores y no se puede ejecutar")

// ─── Datos para el asistente ───────────────────────────────────────────────

func (c Compilacion) entradaAsistente() asistente.Entrada {
	p := c.Proyecto
	fuentes := map[string]string{}
	for n, a := range p.Archivos {
		fuentes[n] = a.Fuente
	}
	return asistente.Entrada{
		Fuentes: fuentes,
		Nombres: func(archivo string) []string {
			return append(c.visibles(archivo), locales(p.Archivos[archivo])...)
		},
		Tipos: func(archivo string) []string {
			var tipos []string
			for _, n := range c.visibles(archivo) {
				if s := c.Tabla.Buscar(archivo, n); s != nil && (s.Clase == analizador.ClaseEspecie || s.Clase == analizador.ClaseFicha) {
					tipos = append(tipos, n)
				}
			}
			return tipos
		},
		Exportados: func(archivo string) []string {
			a := p.Archivos[archivo]
			if a == nil {
				return nil
			}
			var nombres []string
			for n := range proyecto.Declarados(a.Programa) {
				nombres = append(nombres, n)
			}
			return nombres
		},
		Archivos: func() []string {
			var nombres []string
			for n := range p.Archivos {
				nombres = append(nombres, n)
			}
			sort.Strings(nombres)
			return nombres
		},
	}
}

// visibles son los nombres de alcance de archivo que un archivo puede usar;
// vacío si el analizador no corrió.
func (c Compilacion) visibles(archivo string) []string {
	if c.Tabla == nil {
		return nil
	}
	return c.Tabla.Visibles(archivo)
}

// locales junta los datos, parámetros y variables de recorrido declarados
// en un archivo. No distingue bloques: para sugerir un nombre mal escrito
// basta con saber que existe en el archivo.
func locales(a *proyecto.Archivo) []string {
	if a == nil {
		return nil
	}
	var nombres []string
	agregar := func(id *ast.Ident) {
		if id != nil {
			nombres = append(nombres, id.Nombre)
		}
	}
	ast.InspeccionarPrograma(a.Programa, func(n ast.Nodo) bool {
		switch x := n.(type) {
		case *ast.DeclDato:
			agregar(x.Nombre)
		case *ast.Param:
			agregar(x.Nombre)
		case *ast.RecorrerColeccion:
			agregar(x.Var)
			agregar(x.Var2)
		case *ast.RecorrerRango:
			agregar(x.Var)
		}
		return true
	})
	return nombres
}
