package main

import (
	"os"

	"github.com/keylorpineda/PokeScript/internal/analizador"
	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/proyecto"
)

// cargar lee un archivo .pks o una carpeta de proyecto, con sus
// importaciones, y devuelve un único programa listo para ejecutar junto con
// los diagnósticos de todos los archivos.
//
// Mientras el intérprete ejecute un solo programa (tarea J6), las
// declaraciones se juntan en el orden de internal/proyecto: cada archivo
// después de los que importa y el principal al final.
func cargar(ruta string) (*ast.Programa, []diag.Diagnostic, error) {
	info, err := os.Stat(ruta)
	var r proyecto.Resultado
	switch {
	case err == nil && info.IsDir():
		r, err = proyecto.Cargar(ruta)
	default:
		r, err = proyecto.CargarArchivo(ruta)
	}
	if err != nil {
		return nil, nil, err
	}

	p := r.Proyecto
	diags := r.Diagnosticos
	if !r.TieneErrores() {
		// La pasada 1 del analizador solo tiene sentido sobre un proyecto
		// que se pudo leer completo.
		diags = append(diags, analizador.Recolectar(p).Diagnosticos...)
	}

	prog := &ast.Programa{Archivo: p.Principal}
	for _, n := range p.Orden {
		prog.Declaraciones = append(prog.Declaraciones, p.Archivos[n].Programa.Declaraciones...)
	}
	return prog, diags, nil
}
