package proyecto

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Operaciones del gestor de proyectos (sección 0: «FS: proyecto, archivos
// .pks»). Todas trabajan dentro de la carpeta del proyecto y validan cada
// nombre de archivo, así que no pueden escribir fuera de ella.

// PlantillaPrincipal es el contenido del archivo principal de un proyecto
// nuevo.
const PlantillaPrincipal = "// principal.pks · aquí empieza tu programa\ncombate\n    gritar \"¡Hola, mundo Pokémon!\"\nfin\n"

// Info describe un proyecto en disco para el gestor de la interfaz.
type Info struct {
	Nombre    string   `json:"nombre"`
	Principal string   `json:"principal"`
	Archivos  []string `json:"archivos"` // los .pks de la carpeta, en orden alfabético
}

// Crear arma un proyecto nuevo en carpeta: la crea si no existe (debe
// estar vacía si ya existe) y escribe proyecto.json y principal.pks.
func Crear(carpeta, nombre string) (Info, error) {
	if strings.TrimSpace(nombre) == "" {
		return Info{}, errors.New("el proyecto necesita un nombre")
	}
	if entradas, err := os.ReadDir(carpeta); err == nil && len(entradas) > 0 {
		return Info{}, fmt.Errorf("la carpeta %q no está vacía; elige una carpeta nueva para el proyecto", carpeta)
	}
	if err := os.MkdirAll(carpeta, 0o755); err != nil {
		return Info{}, fmt.Errorf("no se pudo crear la carpeta del proyecto: %w", err)
	}
	if err := escribirMetadatos(carpeta, metadatos{Nombre: nombre, Principal: PrincipalPorDefecto}); err != nil {
		return Info{}, err
	}
	if err := os.WriteFile(filepath.Join(carpeta, PrincipalPorDefecto), []byte(PlantillaPrincipal), 0o644); err != nil {
		return Info{}, err
	}
	return Leer(carpeta)
}

// Leer devuelve el nombre, el principal y los archivos de un proyecto.
func Leer(carpeta string) (Info, error) {
	meta, err := leerMetadatos(carpeta)
	if err != nil {
		return Info{}, err
	}
	archivos, err := filepath.Glob(filepath.Join(carpeta, "*.pks"))
	if err != nil {
		return Info{}, err
	}
	info := Info{Nombre: meta.Nombre, Principal: meta.Principal, Archivos: []string{}}
	for _, a := range archivos {
		info.Archivos = append(info.Archivos, filepath.Base(a))
	}
	sort.Strings(info.Archivos)
	return info, nil
}

// LeerArchivo devuelve el contenido de un .pks del proyecto.
func LeerArchivo(carpeta, archivo string) (string, error) {
	if err := validarNombre(archivo); err != nil {
		return "", err
	}
	datos, err := os.ReadFile(filepath.Join(carpeta, archivo))
	if err != nil {
		return "", fmt.Errorf("no se pudo leer «%s»: %w", archivo, err)
	}
	return string(datos), nil
}

// Guardar escribe el contenido de un .pks del proyecto; lo crea si no existe.
func Guardar(carpeta, archivo, contenido string) error {
	if err := validarNombre(archivo); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(carpeta, archivo), []byte(contenido), 0o644)
}

// NuevoArchivo crea un .pks vacío (con un comentario) en el proyecto.
func NuevoArchivo(carpeta, archivo string) error {
	if err := validarNombre(archivo); err != nil {
		return err
	}
	ruta := filepath.Join(carpeta, archivo)
	if _, err := os.Stat(ruta); err == nil {
		return fmt.Errorf("ya existe un archivo «%s» en el proyecto", archivo)
	}
	return os.WriteFile(ruta, []byte("// "+archivo+"\n"), 0o644)
}

// Renombrar cambia el nombre de un .pks. Si era el principal, actualiza
// proyecto.json. Las importaciones que lo usan no se cambian: el analizador
// las va a señalar como archivo inexistente.
func Renombrar(carpeta, de, a string) error {
	for _, n := range []string{de, a} {
		if err := validarNombre(n); err != nil {
			return err
		}
	}
	if _, err := os.Stat(filepath.Join(carpeta, a)); err == nil {
		return fmt.Errorf("ya existe un archivo «%s» en el proyecto", a)
	}
	if err := os.Rename(filepath.Join(carpeta, de), filepath.Join(carpeta, a)); err != nil {
		return fmt.Errorf("no se pudo renombrar «%s»: %w", de, err)
	}
	meta, err := leerMetadatos(carpeta)
	if err != nil || meta.Principal != de {
		return err
	}
	meta.Principal = a
	return escribirMetadatos(carpeta, meta)
}

// Borrar elimina un .pks del proyecto. El archivo principal no se puede
// borrar: primero hay que marcar otro como principal.
func Borrar(carpeta, archivo string) error {
	if err := validarNombre(archivo); err != nil {
		return err
	}
	meta, err := leerMetadatos(carpeta)
	if err != nil {
		return err
	}
	if meta.Principal == archivo {
		return fmt.Errorf("«%s» es el archivo principal y no se puede borrar; marca otro como principal primero", archivo)
	}
	return os.Remove(filepath.Join(carpeta, archivo))
}

// MarcarPrincipal cambia el archivo principal en proyecto.json.
func MarcarPrincipal(carpeta, archivo string) error {
	if err := validarNombre(archivo); err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(carpeta, archivo)); err != nil {
		return fmt.Errorf("no existe el archivo «%s» en el proyecto", archivo)
	}
	meta, err := leerMetadatos(carpeta)
	if err != nil {
		return err
	}
	meta.Principal = archivo
	return escribirMetadatos(carpeta, meta)
}

// validarNombre exige un nombre de archivo .pks sin carpetas (sección 8:
// sin subcarpetas en esta versión).
func validarNombre(archivo string) error {
	switch {
	case archivo == "" || strings.TrimSpace(archivo) != archivo:
		return fmt.Errorf("«%s» no es un nombre de archivo válido", archivo)
	case filepath.Base(archivo) != archivo || strings.ContainsAny(archivo, `/\`) || archivo == ".pks":
		return fmt.Errorf("«%s» no es válido: los archivos del proyecto van en su carpeta, sin subcarpetas", archivo)
	case filepath.Ext(archivo) != ".pks":
		return fmt.Errorf("«%s» no es válido: los archivos de PokeScript terminan en .pks", archivo)
	}
	return nil
}

// leerMetadatos lee proyecto.json; si no existe, usa los valores por
// defecto, igual que Cargar.
func leerMetadatos(carpeta string) (metadatos, error) {
	meta := metadatos{Nombre: filepath.Base(filepath.Clean(carpeta)), Principal: PrincipalPorDefecto}
	datos, err := os.ReadFile(filepath.Join(carpeta, ArchivoProyecto))
	if errors.Is(err, os.ErrNotExist) {
		return meta, nil
	}
	if err != nil {
		return meta, err
	}
	if err := json.Unmarshal(datos, &meta); err != nil {
		return meta, fmt.Errorf("%s no es válido: %w", ArchivoProyecto, err)
	}
	return meta, nil
}

func escribirMetadatos(carpeta string, meta metadatos) error {
	datos, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(carpeta, ArchivoProyecto), append(datos, '\n'), 0o644)
}

// Otros devuelve, en orden alfabético, los archivos de la carpeta que no son
// .pks ni proyecto.json. El IDE los muestra aparte porque no los edita.
func Otros(carpeta string) ([]string, error) {
	entradas, err := os.ReadDir(carpeta)
	if err != nil {
		return nil, err
	}
	otros := []string{}
	for _, e := range entradas {
		n := e.Name()
		if e.IsDir() || n == ArchivoProyecto || strings.HasSuffix(n, ".pks") {
			continue
		}
		otros = append(otros, n)
	}
	sort.Strings(otros)
	return otros, nil
}
