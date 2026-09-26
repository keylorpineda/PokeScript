package main

import (
	"os"

	"github.com/keylorpineda/PokeScript/internal/analizador"
	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/proyecto"
)

// cargar lee un archivo .pks o una carpeta de proyecto, con sus
// importaciones, y devuelve sus archivos listos para ejecutar junto con los
// diagnósticos de todos ellos.
//
// Los archivos van en el orden de internal/proyecto: cada archivo después
// de los que importa y el principal al final. El intérprete los ejecuta con
// un alcance por archivo (tarea J6).
func cargar(ruta string) ([]*ast.Programa, []diag.Diagnostic, error) {
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
		// El analizador solo tiene sentido sobre un proyecto que se pudo
		// leer completo.
		diags = append(diags, analizador.Analizar(p).Diagnosticos...)
	}

	var archivos []*ast.Programa
	for _, n := range p.Orden {
		archivos = append(archivos, p.Archivos[n].Programa)
	}
	return archivos, diags, nil
}
