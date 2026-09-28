package servicio

import (
	"context"
	"sync"

	"github.com/keylorpineda/PokeScript/internal/interprete"
)

// Entorno reúne lo que el IDE necesita del backend, con los métodos de la
// sección 9. Guarda la ejecución en curso, así que app.go de Wails solo
// tiene que envolverlo y reenviar los eventos:
//
//	func (a *App) EjecutarProyecto(ruta string) (servicio.ResultadoCompilacion, error) {
//	    return a.entorno.EjecutarProyecto(ruta, func(e servicio.Evento) {
//	        runtime.EventsEmit(a.ctx, e.Tipo, e)
//	    })
//	}
//
// El gestor de proyectos (proyecto.Crear, Guardar, …) y el menú de consulta
// (consulta.PalabrasReservadas, TablaEfectividades) no guardan estado y se
// llaman directo.
type Entorno struct {
	mu     sync.Mutex
	sesion *Sesion
	// NuevoInterprete crea el intérprete de cada ejecución; las pruebas lo
	// reemplazan para fijar la semilla de aleatorio.
	NuevoInterprete func() *interprete.Interprete
}

// NuevoEntorno crea un entorno sin ninguna ejecución en curso.
func NuevoEntorno() *Entorno {
	return &Entorno{NuevoInterprete: func() *interprete.Interprete { return interprete.Nuevo(nil) }}
}

// CompilarProyecto compila la carpeta o el archivo de ruta y devuelve el
// resultado para la interfaz.
func (e *Entorno) CompilarProyecto(ruta string) (ResultadoCompilacion, error) {
	c, err := Compilar(ruta)
	if err != nil {
		return ResultadoCompilacion{}, err
	}
	return c.Resultado(), nil
}

// EjecutarProyecto compila y, si no hay errores, empieza a ejecutar: la
// salida llega por emitir, en streaming. Si ya había una ejecución en
// curso, la detiene primero. Devuelve el resultado de la compilación; si
// tiene errores, no ejecuta nada.
func (e *Entorno) EjecutarProyecto(ruta string, emitir func(Evento)) (ResultadoCompilacion, error) {
	c, err := Compilar(ruta)
	if err != nil {
		return ResultadoCompilacion{}, err
	}
	r := c.Resultado()
	if !r.Exito {
		return r, nil
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	if e.sesion != nil {
		e.sesion.Detener()
		_ = e.sesion.Esperar()
	}
	s, err := Iniciar(context.Background(), c, e.NuevoInterprete(), emitir)
	if err != nil {
		return r, err
	}
	e.sesion = s
	return r, nil
}

// EnviarEntrada responde al capturar que está esperando.
func (e *Entorno) EnviarEntrada(texto string) error {
	s := e.actual()
	if s == nil {
		return ErrNoEsperaEntrada
	}
	return s.EnviarEntrada(texto)
}

// DetenerEjecucion corta la ejecución en curso, si hay una.
func (e *Entorno) DetenerEjecucion() {
	if s := e.actual(); s != nil {
		s.Detener()
	}
}

// Ejecutando informa si hay un programa corriendo.
func (e *Entorno) Ejecutando() bool {
	s := e.actual()
	return s != nil && !s.Terminada()
}

// Esperar bloquea hasta que termine la ejecución en curso, si hay una.
func (e *Entorno) Esperar() error {
	if s := e.actual(); s != nil {
		return s.Esperar()
	}
	return nil
}

func (e *Entorno) actual() *Sesion {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.sesion
}
