// Package explorador recorre las carpetas del disco para el explorador del
// IDE (el «PC de Bill»): los lugares de siempre y lo que hay dentro de cada
// carpeta, marcando cuáles son proyectos de PokeScript.
package explorador

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// MaxEntradas limita cuántas cosas se muestran de una carpeta enorme.
const MaxEntradas = 400

// Lugar es un acceso directo: la carpeta personal, el Escritorio, una unidad…
type Lugar struct {
	Nombre string `json:"nombre"`
	Ruta   string `json:"ruta"`
	Tipo   string `json:"tipo"` // "personal", "carpeta" o "unidad"
}

// Subcarpeta es una carpeta dentro de la que se está viendo.
type Subcarpeta struct {
	Nombre   string `json:"nombre"`
	Ruta     string `json:"ruta"`
	Proyecto bool   `json:"proyecto"` // tiene proyecto.json o algún .pks
	Pks      int    `json:"pks"`      // cuántos .pks tiene
}

// Archivo es un archivo de la carpeta que se está viendo.
type Archivo struct {
	Nombre string `json:"nombre"`
	Pks    bool   `json:"pks"`
}

// Carpeta es lo que muestra el explorador de una carpeta.
type Carpeta struct {
	Ruta      string       `json:"ruta"`
	Nombre    string       `json:"nombre"`
	Padre     string       `json:"padre"` // vacío en la raíz de una unidad
	Proyecto  bool         `json:"proyecto"`
	Carpetas  []Subcarpeta `json:"carpetas"`
	Archivos  []Archivo    `json:"archivos"`
	Recortada bool         `json:"recortada"` // había más de MaxEntradas
}

// Lugares devuelve la carpeta personal, las carpetas de siempre que existan
// y, en Windows, las unidades.
func Lugares() []Lugar {
	var lugares []Lugar
	if casa, err := os.UserHomeDir(); err == nil {
		lugares = append(lugares, Lugar{"Carpeta personal", casa, "personal"})
		vistos := map[string]bool{}
		for _, n := range []string{"Desktop", "Escritorio", "Documents", "Documentos", filepath.Join("OneDrive", "Desktop"), filepath.Join("OneDrive", "Escritorio"), filepath.Join("OneDrive", "Documents"), filepath.Join("OneDrive", "Documentos")} {
			ruta := filepath.Join(casa, n)
			if !esCarpeta(ruta) {
				continue
			}
			nombre := filepath.Base(n)
			if nombre == "Desktop" {
				nombre = "Escritorio"
			} else if nombre == "Documents" {
				nombre = "Documentos"
			}
			if strings.HasPrefix(n, "OneDrive") {
				nombre += " (OneDrive)"
			}
			if !vistos[nombre] {
				vistos[nombre] = true
				lugares = append(lugares, Lugar{nombre, ruta, "carpeta"})
			}
		}
	}
	if runtime.GOOS == "windows" {
		for l := 'A'; l <= 'Z'; l++ {
			ruta := string(l) + `:\`
			if esCarpeta(ruta) {
				lugares = append(lugares, Lugar{"Unidad " + string(l) + ":", ruta, "unidad"})
			}
		}
	} else {
		lugares = append(lugares, Lugar{"Raíz", "/", "unidad"})
	}
	return lugares
}

// Listar devuelve las subcarpetas y los archivos de una carpeta, primero las
// carpetas y cada grupo en orden alfabético. Se saltan los ocultos.
func Listar(ruta string) (Carpeta, error) {
	ruta = filepath.Clean(ruta)
	entradas, err := os.ReadDir(ruta)
	if err != nil {
		return Carpeta{}, err
	}
	c := Carpeta{
		Ruta:     ruta,
		Nombre:   nombreDe(ruta),
		Carpetas: []Subcarpeta{},
		Archivos: []Archivo{},
	}
	if padre := filepath.Dir(ruta); padre != ruta {
		c.Padre = padre
	}
	for _, e := range entradas {
		n := e.Name()
		if oculto(n) {
			continue
		}
		if len(c.Carpetas)+len(c.Archivos) >= MaxEntradas {
			c.Recortada = true
			break
		}
		if e.IsDir() {
			sub := filepath.Join(ruta, n)
			pks, conJSON := contar(sub)
			c.Carpetas = append(c.Carpetas, Subcarpeta{n, sub, conJSON || pks > 0, pks})
			continue
		}
		esPks := strings.EqualFold(filepath.Ext(n), ".pks")
		c.Archivos = append(c.Archivos, Archivo{n, esPks})
		if esPks || n == "proyecto.json" {
			c.Proyecto = true
		}
	}
	sort.Slice(c.Carpetas, func(i, j int) bool {
		return strings.ToLower(c.Carpetas[i].Nombre) < strings.ToLower(c.Carpetas[j].Nombre)
	})
	sort.Slice(c.Archivos, func(i, j int) bool {
		return strings.ToLower(c.Archivos[i].Nombre) < strings.ToLower(c.Archivos[j].Nombre)
	})
	return c, nil
}

// contar dice cuántos .pks tiene una carpeta y si tiene proyecto.json. Una
// carpeta que no se puede leer cuenta como vacía.
func contar(carpeta string) (pks int, conJSON bool) {
	entradas, err := os.ReadDir(carpeta)
	if err != nil {
		return 0, false
	}
	for _, e := range entradas {
		switch {
		case e.IsDir():
		case e.Name() == "proyecto.json":
			conJSON = true
		case strings.EqualFold(filepath.Ext(e.Name()), ".pks"):
			pks++
		}
	}
	return pks, conJSON
}

func oculto(nombre string) bool {
	switch nombre {
	case "$Recycle.Bin", "System Volume Information", "$WinREAgent", "node_modules":
		return true
	}
	return strings.HasPrefix(nombre, ".") || strings.HasPrefix(nombre, "$")
}

func nombreDe(ruta string) string {
	if n := filepath.Base(ruta); n != "." && n != string(filepath.Separator) && !strings.HasSuffix(n, ":") && n != ruta {
		return n
	}
	return ruta
}

func esCarpeta(ruta string) bool {
	info, err := os.Stat(ruta)
	return err == nil && info.IsDir()
}
