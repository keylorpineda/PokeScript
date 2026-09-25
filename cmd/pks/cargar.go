package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/parser"
)

// proyectoJSON es el contenido de proyecto.json (sección 8).
type proyectoJSON struct {
	Nombre    string `json:"nombre"`
	Principal string `json:"principal"`
}

// cargar lee un archivo .pks o una carpeta de proyecto y devuelve un único
// programa listo para ejecutar, junto con los diagnósticos de todos los
// archivos.
//
// Mientras no exista internal/proyecto (hito 9), una carpeta se carga
// juntando las declaraciones de todos sus .pks, con el archivo principal al
// final; las líneas «enseñar» se leen pero todavía no se validan.
func cargar(ruta string) (*ast.Programa, []diag.Diagnostic, error) {
	info, err := os.Stat(ruta)
	if err != nil {
		return nil, nil, fmt.Errorf("no se encontró %q", ruta)
	}
	if !info.IsDir() {
		return cargarArchivos(filepath.Dir(ruta), []string{filepath.Base(ruta)}, filepath.Base(ruta))
	}

	principal, err := archivoPrincipal(ruta)
	if err != nil {
		return nil, nil, err
	}
	rutas, err := filepath.Glob(filepath.Join(ruta, "*.pks"))
	if err != nil {
		return nil, nil, err
	}
	var nombres []string
	for _, r := range rutas {
		if n := filepath.Base(r); n != principal {
			nombres = append(nombres, n)
		}
	}
	sort.Strings(nombres)
	nombres = append(nombres, principal)
	return cargarArchivos(ruta, nombres, principal)
}

// archivoPrincipal lee el principal de proyecto.json; si no existe el
// archivo, se usa principal.pks.
func archivoPrincipal(carpeta string) (string, error) {
	datos, err := os.ReadFile(filepath.Join(carpeta, "proyecto.json"))
	if errors.Is(err, os.ErrNotExist) {
		return "principal.pks", nil
	}
	if err != nil {
		return "", err
	}
	var p proyectoJSON
	if err := json.Unmarshal(datos, &p); err != nil {
		return "", fmt.Errorf("proyecto.json no es válido: %w", err)
	}
	if p.Principal == "" {
		return "", errors.New(`proyecto.json no dice cuál es el archivo principal ("principal": "…")`)
	}
	return p.Principal, nil
}

func cargarArchivos(carpeta string, nombres []string, principal string) (*ast.Programa, []diag.Diagnostic, error) {
	prog := &ast.Programa{Archivo: principal}
	var diags []diag.Diagnostic
	for _, n := range nombres {
		fuente, err := os.ReadFile(filepath.Join(carpeta, n))
		if err != nil {
			return nil, nil, fmt.Errorf("no se pudo leer %q", n)
		}
		r := parser.Analizar(n, string(fuente))
		diags = append(diags, r.Diagnosticos...)
		prog.Declaraciones = append(prog.Declaraciones, r.Programa.Declaraciones...)
	}
	return prog, diags, nil
}
