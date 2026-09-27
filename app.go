package main

import (
	"context"
	"path/filepath"
	"sync/atomic"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/keylorpineda/PokeScript/internal/consulta"
	"github.com/keylorpineda/PokeScript/internal/explorador"
	"github.com/keylorpineda/PokeScript/internal/proyecto"
	"github.com/keylorpineda/PokeScript/internal/servicio"
)

// App es el puente entre la interfaz (Svelte) y el backend. Cada método es
// una o dos líneas: la lógica vive en internal/servicio, internal/proyecto e
// internal/consulta (sección 9 de la especificación).
type App struct {
	ctx     context.Context
	entorno *servicio.Entorno
}

// ProyectoAbierto es un proyecto con la carpeta donde vive, para que la
// interfaz sepa qué ruta usar en las demás llamadas.
type ProyectoAbierto struct {
	Ruta string `json:"ruta"`
	proyecto.Info
}

// NuevaApp crea el puente sin ninguna ejecución en curso.
func NuevaApp() *App {
	return &App{entorno: servicio.NuevoEntorno()}
}

// iniciar ajusta la ventana a la pantalla: la ventana no tiene marco de
// Windows, así que si fuera más grande que la pantalla, los botones de la
// barra propia quedarían afuera.
func (a *App) iniciar(ctx context.Context) {
	a.ctx = ctx
	pantallas, err := runtime.ScreenGetAll(ctx)
	if err != nil {
		return
	}
	for _, p := range pantallas {
		if !p.IsCurrent && !p.IsPrimary {
			continue
		}
		ancho := min(1400, p.Size.Width*9/10)
		alto := min(880, p.Size.Height*85/100)
		runtime.WindowSetSize(ctx, ancho, alto)
		runtime.WindowCenter(ctx)
		return
	}
}

func (a *App) cerrar(context.Context) { a.entorno.DetenerEjecucion() }

// ─── Compilar y ejecutar ────────────────────────────────────────────────────

// CompilarProyecto compila la carpeta del proyecto.
func (a *App) CompilarProyecto(ruta string) (servicio.ResultadoCompilacion, error) {
	return a.entorno.CompilarProyecto(ruta)
}

// EventoNumerado es un evento de la ejecución con su número de orden: Wails
// puede entregar los eventos desordenados y la interfaz los reacomoda con N.
type EventoNumerado struct {
	servicio.Evento
	N int64 `json:"n"`
}

// EjecutarProyecto compila y, si no hay errores, ejecuta. La salida llega por
// eventos: salida, pedir-entrada, error-ejecucion y fin-ejecucion, numerados
// desde 1 en cada ejecución.
func (a *App) EjecutarProyecto(ruta string) (servicio.ResultadoCompilacion, error) {
	var n atomic.Int64
	return a.entorno.EjecutarProyecto(ruta, func(e servicio.Evento) {
		runtime.EventsEmit(a.ctx, e.Tipo, EventoNumerado{Evento: e, N: n.Add(1)})
	})
}

// EnviarEntrada responde al capturar que está esperando.
func (a *App) EnviarEntrada(texto string) error { return a.entorno.EnviarEntrada(texto) }

// DetenerEjecucion corta el programa en curso.
func (a *App) DetenerEjecucion() { a.entorno.DetenerEjecucion() }

// ─── Menú de consulta ───────────────────────────────────────────────────────

// ObtenerTablaEfectividades devuelve la tabla 3.2 con sus encabezados.
func (a *App) ObtenerTablaEfectividades() [][]string { return consulta.TablaEfectividades() }

// ObtenerPalabrasReservadas devuelve las palabras reservadas explicadas.
func (a *App) ObtenerPalabrasReservadas() []consulta.PalabraDoc {
	return consulta.PalabrasReservadas()
}

// ─── Gestor de proyectos ───────────────────────────────────────────────────

// LugaresExplorador devuelve los accesos directos del explorador del IDE.
func (a *App) LugaresExplorador() []explorador.Lugar { return explorador.Lugares() }

// ListarCarpeta devuelve lo que hay en una carpeta para el explorador.
func (a *App) ListarCarpeta(ruta string) (explorador.Carpeta, error) {
	return explorador.Listar(ruta)
}

// ElegirCarpeta abre el diálogo de carpetas del sistema; vacío si se cancela.
func (a *App) ElegirCarpeta() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Elige una carpeta",
	})
}

// CrearProyecto crea el proyecto en una carpeta nueva dentro de padre.
func (a *App) CrearProyecto(padre, nombre string) (ProyectoAbierto, error) {
	ruta := filepath.Join(padre, nombre)
	info, err := proyecto.Crear(ruta, nombre)
	return ProyectoAbierto{Ruta: ruta, Info: info}, err
}

// LeerProyecto devuelve el nombre, el principal y los archivos.
func (a *App) LeerProyecto(ruta string) (proyecto.Info, error) { return proyecto.Leer(ruta) }

// OtrosArchivos lista lo que hay en la carpeta que no es .pks.
func (a *App) OtrosArchivos(ruta string) ([]string, error) { return proyecto.Otros(ruta) }

// PlantillaPrincipal es el contenido con que empieza un principal.pks nuevo.
func (a *App) PlantillaPrincipal() string { return proyecto.PlantillaPrincipal }

// LeerArchivo devuelve el contenido de un .pks.
func (a *App) LeerArchivo(ruta, archivo string) (string, error) {
	return proyecto.LeerArchivo(ruta, archivo)
}

// GuardarArchivo escribe un .pks.
func (a *App) GuardarArchivo(ruta, archivo, contenido string) error {
	return proyecto.Guardar(ruta, archivo, contenido)
}

// NuevoArchivo crea un .pks vacío.
func (a *App) NuevoArchivo(ruta, archivo string) error { return proyecto.NuevoArchivo(ruta, archivo) }

// RenombrarArchivo cambia el nombre de un .pks.
func (a *App) RenombrarArchivo(ruta, de, nuevo string) error {
	return proyecto.Renombrar(ruta, de, nuevo)
}

// BorrarArchivo borra un .pks que no sea el principal.
func (a *App) BorrarArchivo(ruta, archivo string) error { return proyecto.Borrar(ruta, archivo) }

// MarcarPrincipal cambia el archivo principal del proyecto.
func (a *App) MarcarPrincipal(ruta, archivo string) error {
	return proyecto.MarcarPrincipal(ruta, archivo)
}
