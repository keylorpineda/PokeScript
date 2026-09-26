package proyecto

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/parser"
)

// ArchivoProyecto es el nombre del archivo de metadatos (sección 8).
const ArchivoProyecto = "proyecto.json"

// PrincipalPorDefecto se usa cuando no hay proyecto.json.
const PrincipalPorDefecto = "principal.pks"

// Proyecto es un conjunto de archivos .pks en una carpeta, sin subcarpetas.
type Proyecto struct {
	Nombre    string
	Principal string
	// Archivos tiene todos los .pks de la carpeta, por nombre.
	Archivos map[string]*Archivo
	// Orden son los archivos alcanzables desde el principal siguiendo las
	// importaciones, con cada archivo después de los que importa y el
	// principal al final. Es el orden en que se analizan y se ejecutan.
	Orden []string
}

// Archivo es un .pks leído y analizado sintácticamente.
type Archivo struct {
	Nombre   string
	Fuente   string // el texto del archivo, para leer lo que señala cada diagnóstico
	Programa *ast.Programa
	Sangrias map[int]int
	// Dependencias son los archivos que este importa y que existen, sin
	// repetir, en el orden en que aparecen.
	Dependencias []string
}

// Resultado es lo que devuelve la carga de un proyecto.
type Resultado struct {
	Proyecto *Proyecto
	// Diagnosticos junta los errores léxicos y sintácticos de cada archivo
	// y los de importación.
	Diagnosticos []diag.Diagnostic
}

// TieneErrores informa si hay al menos un error.
func (r Resultado) TieneErrores() bool {
	for _, d := range r.Diagnosticos {
		if d.Severity == diag.Error {
			return true
		}
	}
	return false
}

type metadatos struct {
	Nombre    string `json:"nombre"`
	Principal string `json:"principal"`
}

// Cargar lee la carpeta de un proyecto: proyecto.json, todos sus .pks y
// sus importaciones. Devuelve error solo si no se puede armar el proyecto
// (carpeta inexistente, proyecto.json inválido o sin archivo principal);
// los problemas del código van en Resultado.Diagnosticos.
func Cargar(carpeta string) (Resultado, error) {
	info, err := os.Stat(carpeta)
	if err != nil || !info.IsDir() {
		return Resultado{}, fmt.Errorf("no se encontró la carpeta del proyecto %q", carpeta)
	}
	return CargarFS(os.DirFS(carpeta), filepath.Base(filepath.Clean(carpeta)))
}

// CargarArchivo arma un proyecto de un solo archivo, que hace de principal.
// Sus importaciones se buscan en la misma carpeta.
func CargarArchivo(ruta string) (Resultado, error) {
	if _, err := os.Stat(ruta); err != nil {
		return Resultado{}, fmt.Errorf("no se encontró el archivo %q", ruta)
	}
	carpeta, nombre := filepath.Split(ruta)
	if carpeta == "" {
		carpeta = "."
	}
	return cargar(os.DirFS(carpeta), metadatos{Nombre: nombre, Principal: nombre}, false)
}

// CargarFS es Cargar sobre cualquier sistema de archivos; las pruebas lo
// usan con fstest.MapFS. nombre es el nombre del proyecto si proyecto.json
// no trae uno.
func CargarFS(fsys fs.FS, nombre string) (Resultado, error) {
	meta := metadatos{Nombre: nombre, Principal: PrincipalPorDefecto}
	datos, err := fs.ReadFile(fsys, ArchivoProyecto)
	switch {
	case err == nil:
		var delArchivo metadatos
		if err := json.Unmarshal(datos, &delArchivo); err != nil {
			return Resultado{}, fmt.Errorf("%s no es válido: %w", ArchivoProyecto, err)
		}
		if delArchivo.Principal == "" {
			return Resultado{}, fmt.Errorf(`%s no dice cuál es el archivo principal ("principal": "…")`, ArchivoProyecto)
		}
		meta.Principal = delArchivo.Principal
		if delArchivo.Nombre != "" {
			meta.Nombre = delArchivo.Nombre
		}
	case !errors.Is(err, fs.ErrNotExist):
		return Resultado{}, fmt.Errorf("no se pudo leer %s: %w", ArchivoProyecto, err)
	}
	return cargar(fsys, meta, true)
}

// cargador lee los archivos a medida que hacen falta.
type cargador struct {
	fsys       fs.FS
	p          *Proyecto
	diags      []diag.Diagnostic
	pendientes []string // archivos leídos cuyas importaciones falta resolver
}

// cargar arma el proyecto. Con todos, lee todos los .pks de la carpeta
// (para ver los errores de cada uno en el IDE); sin todos, solo el principal
// y lo que importa (para correr un archivo suelto).
func cargar(fsys fs.FS, meta metadatos, todos bool) (Resultado, error) {
	c := &cargador{fsys: fsys, p: &Proyecto{Nombre: meta.Nombre, Principal: meta.Principal, Archivos: map[string]*Archivo{}}}
	if todos {
		nombres, err := fs.Glob(fsys, "*.pks")
		if err != nil {
			return Resultado{}, err
		}
		sort.Strings(nombres)
		for _, n := range nombres {
			if _, err := c.leer(n); err != nil {
				return Resultado{}, err
			}
		}
	}
	a, err := c.leer(meta.Principal)
	if err != nil {
		return Resultado{}, err
	}
	if a == nil {
		return Resultado{}, fmt.Errorf("no se encontró el archivo principal %q en el proyecto", meta.Principal)
	}

	for len(c.pendientes) > 0 {
		n := c.pendientes[0]
		c.pendientes = c.pendientes[1:]
		if err := c.resolverImportaciones(c.p.Archivos[n]); err != nil {
			return Resultado{}, err
		}
	}
	orden, ciclos := c.p.ordenar()
	c.p.Orden = orden
	c.diags = append(c.diags, ciclos...)
	return Resultado{Proyecto: c.p, Diagnosticos: c.diags}, nil
}

// leer devuelve el archivo n, leyéndolo y analizándolo la primera vez.
// Devuelve nil si el archivo no existe.
func (c *cargador) leer(n string) (*Archivo, error) {
	if a, ok := c.p.Archivos[n]; ok {
		return a, nil
	}
	fuente, err := fs.ReadFile(c.fsys, n)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer %q: %w", n, err)
	}
	r := parser.Analizar(n, string(fuente))
	a := &Archivo{Nombre: n, Fuente: string(fuente), Programa: r.Programa, Sangrias: r.Sangrias}
	c.p.Archivos[n] = a
	c.diags = append(c.diags, r.Diagnosticos...)
	c.pendientes = append(c.pendientes, n)
	return a, nil
}

// ─── Importaciones ─────────────────────────────────────────────────────────

// Declarados devuelve lo que un archivo declara y otro puede importar:
// movimientos, especies, fichas y medallas de alcance de archivo.
func Declarados(prog *ast.Programa) map[string]ast.Decl {
	m := map[string]ast.Decl{}
	for _, d := range prog.Declaraciones {
		if n := nombreDe(d); n != nil {
			if _, repetido := m[n.Nombre]; !repetido {
				m[n.Nombre] = d
			}
		}
	}
	return m
}

func nombreDe(d ast.Decl) *ast.Ident {
	switch x := d.(type) {
	case *ast.DeclMedalla:
		return x.Nombre
	case *ast.DeclEspecie:
		return x.Nombre
	case *ast.DeclFicha:
		return x.Nombre
	case *ast.DeclMovimiento:
		return x.Nombre
	}
	return nil
}

// resolverImportaciones valida cada «enseñar … desde» de un archivo y
// llena sus dependencias.
func (c *cargador) resolverImportaciones(a *Archivo) error {
	diags := &c.diags
	locales := Declarados(a.Programa)
	importados := map[string]bool{}
	vistas := map[string]bool{}

	for _, im := range a.Programa.Importaciones {
		if im.Ruta == "" { // la línea tenía un error de sintaxis
			continue
		}
		if path.Base(im.Ruta) != im.Ruta || filepath.Base(im.Ruta) != im.Ruta {
			*diags = append(*diags, errorImportacion(a.Nombre, im.RutaPos, "ruta-invalida",
				fmt.Sprintf("«%s» apunta a otra carpeta.", im.Ruta),
				"los archivos de un proyecto están todos en la misma carpeta, sin subcarpetas.",
				"escribe solo el nombre del archivo, por ejemplo \"tipos.pks\"."))
			continue
		}
		destino, err := c.leer(im.Ruta)
		if err != nil {
			return err
		}
		if destino == nil {
			*diags = append(*diags, errorImportacion(a.Nombre, im.RutaPos, "archivo-inexistente",
				fmt.Sprintf("el archivo «%s» no existe en el proyecto.", im.Ruta),
				"la ruta de «enseñar … desde» debe ser un archivo .pks de la carpeta del proyecto.",
				c.sugerirArchivo()))
			continue
		}
		if !vistas[im.Ruta] {
			vistas[im.Ruta] = true
			a.Dependencias = append(a.Dependencias, im.Ruta)
		}
		if im.Ruta == a.Nombre {
			continue // se reporta como ciclo; revisar los nombres solo repetiría el error
		}

		exportados := Declarados(destino.Programa)
		for _, n := range im.Nombres {
			switch {
			case exportados[n.Nombre] == nil:
				*diags = append(*diags, errorImportacion(a.Nombre, n.Pos, "nombre-inexistente",
					fmt.Sprintf("«%s» no está declarado en «%s».", n.Nombre, im.Ruta),
					"solo se pueden importar movimientos, especies, fichas y medallas que el archivo declara.",
					sugerirNombre(n.Nombre, destino.Programa)))
			case locales[n.Nombre] != nil:
				*diags = append(*diags, errorImportacion(a.Nombre, n.Pos, "colision-con-importacion",
					fmt.Sprintf("«%s» se importa desde «%s», pero este archivo también declara un «%s».", n.Nombre, im.Ruta, n.Nombre),
					"un mismo nombre no puede referirse a dos cosas distintas.",
					"cambia el nombre de la declaración local o quita la importación."))
			case importados[n.Nombre]:
				*diags = append(*diags, errorImportacion(a.Nombre, n.Pos, "importacion-repetida",
					fmt.Sprintf("«%s» ya se importó antes en este archivo.", n.Nombre),
					"cada nombre se importa una sola vez.",
					"borra esta importación repetida."))
			default:
				importados[n.Nombre] = true
			}
		}
	}
	return nil
}

func (c *cargador) sugerirArchivo() string {
	disponibles, _ := fs.Glob(c.fsys, "*.pks")
	sort.Strings(disponibles)
	var nombres []string
	for _, n := range disponibles {
		nombres = append(nombres, "«"+n+"»")
	}
	if len(nombres) == 0 {
		return "crea el archivo en la carpeta del proyecto."
	}
	return "revisa el nombre; los archivos del proyecto son: " + unir(nombres) + "."
}

// sugerirNombre explica qué se puede importar del archivo; si el nombre
// es un valor de especie, sugiere importar la especie.
func sugerirNombre(nombre string, prog *ast.Programa) string {
	for _, d := range prog.Declaraciones {
		if e, ok := d.(*ast.DeclEspecie); ok {
			for _, v := range e.Valores {
				if v.Nombre == nombre {
					return fmt.Sprintf("«%s» es un valor de la especie «%s»: importa «%s» y sus valores quedan disponibles.", nombre, e.Nombre.Nombre, e.Nombre.Nombre)
				}
			}
		}
	}
	var nombres []string
	for n := range Declarados(prog) {
		nombres = append(nombres, "«"+n+"»")
	}
	sort.Strings(nombres)
	if len(nombres) == 0 {
		return "ese archivo no declara nada que se pueda importar."
	}
	return "ese archivo declara: " + unir(nombres) + "."
}

func unir(xs []string) string {
	s := ""
	for i, x := range xs {
		switch {
		case i == 0:
		case i == len(xs)-1:
			s += " y "
		default:
			s += ", "
		}
		s += x
	}
	return s
}

func errorImportacion(archivo string, pos ast.Pos, codigo, desc, causa, sugerencia string) diag.Diagnostic {
	if pos.Len < 1 {
		pos.Len = 1
	}
	return diag.Diagnostic{
		Severity: diag.Error,
		Category: diag.Importacion,
		Code:     codigo,
		Heading:  diag.EncabezadoImportacion,
		File:     archivo,
		Line:     pos.Line,
		Col:      pos.Col,
		Len:      pos.Len,
		Desc:     desc,
		Cause:    causa,
		Suggest:  sugerencia,
	}
}
