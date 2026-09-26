package servicio

import (
	"context"
	"errors"
	"sync"

	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/interprete"
)

// Tipos de evento que emite una sesión (sección 9: la ejecución va en
// streaming porque capturar necesita ir y volver). app.go los reenvía tal
// cual con runtime.EventsEmit.
const (
	EventoSalida         = "salida"          // Texto: una línea de gritar
	EventoPedirEntrada   = "pedir-entrada"   // Texto: el mensaje de capturar
	EventoErrorEjecucion = "error-ejecucion" // Diagnostico: el error que detuvo el programa
	EventoFin            = "fin-ejecucion"   // Texto: "terminado" o "detenido"
)

// Evento es un mensaje de la ejecución hacia la interfaz.
type Evento struct {
	Tipo        string           `json:"tipo"`
	Texto       string           `json:"texto,omitempty"`
	Diagnostico *diag.Diagnostic `json:"diagnostico,omitempty"`
}

// Motivos de fin, en Evento.Texto de EventoFin.
const (
	FinTerminado = "terminado"
	FinDetenido  = "detenido"
	FinConError  = "error"
)

// ErrNoEsperaEntrada indica que se envió una entrada cuando el programa no
// estaba en un capturar.
var ErrNoEsperaEntrada = errors.New("el programa no está esperando una entrada")

// Sesion es una ejecución en curso. Se crea con Iniciar; el programa corre
// en su propia goroutine y avisa todo lo que pasa con la función emitir.
type Sesion struct {
	cancelar context.CancelFunc
	entradas chan string
	hecho    chan struct{}

	mu        sync.Mutex
	esperando bool
	err       error
}

// Iniciar empieza a ejecutar un programa compilado. emitir recibe cada
// evento en orden, desde la goroutine de la ejecución; el último siempre es
// EventoFin. emitir no debe llamar a EnviarEntrada directamente (se
// bloquearía esperando a su propia goroutine): la entrada llega después,
// desde la interfaz, como pasa con los eventos de Wails. in es el intérprete a usar (nil crea uno nuevo); su ES se
// reemplaza por la de la sesión.
func Iniciar(ctx context.Context, c Compilacion, in *interprete.Interprete, emitir func(Evento)) (*Sesion, error) {
	if c.TieneErrores() {
		return nil, ErrHayErrores
	}
	ctx, cancelar := context.WithCancel(ctx)
	s := &Sesion{cancelar: cancelar, entradas: make(chan string), hecho: make(chan struct{})}
	if in == nil {
		in = interprete.Nuevo(nil)
	}
	in.ES = &esSesion{s: s, ctx: ctx, emitir: emitir}

	go func() {
		defer close(s.hecho)
		defer cancelar()
		err := Ejecutar(ctx, c, in)

		var errEjecucion *interprete.ErrorEjecucion
		motivo := FinTerminado
		switch {
		case err == nil:
		case errors.As(err, &errEjecucion):
			d := errEjecucion.Diag
			emitir(Evento{Tipo: EventoErrorEjecucion, Diagnostico: &d})
			motivo = FinConError
		case errors.Is(err, context.Canceled), errors.Is(err, interprete.ErrEntradaCerrada):
			motivo = FinDetenido
			err = nil
		default:
			motivo = FinConError
		}
		s.mu.Lock()
		s.err = err
		s.mu.Unlock()
		emitir(Evento{Tipo: EventoFin, Texto: motivo})
	}()
	return s, nil
}

// EnviarEntrada responde al capturar que está esperando. Devuelve
// ErrNoEsperaEntrada si el programa no está en un capturar.
func (s *Sesion) EnviarEntrada(texto string) error {
	if !s.EsperandoEntrada() {
		return ErrNoEsperaEntrada
	}
	select {
	case s.entradas <- texto:
		return nil
	case <-s.hecho:
		return ErrNoEsperaEntrada
	}
}

// EsperandoEntrada informa si el programa está detenido en un capturar.
func (s *Sesion) EsperandoEntrada() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.esperando
}

// Detener pide terminar la ejecución: el programa se corta en la próxima
// instrucción o en el capturar que está esperando. No espera a que termine;
// para eso está Esperar.
func (s *Sesion) Detener() { s.cancelar() }

// Esperar bloquea hasta que el programa termina y devuelve su error: nil si
// terminó bien o se detuvo, un *interprete.ErrorEjecucion si falló.
func (s *Sesion) Esperar() error {
	<-s.hecho
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

// Terminada informa si la ejecución ya acabó.
func (s *Sesion) Terminada() bool {
	select {
	case <-s.hecho:
		return true
	default:
		return false
	}
}

// esSesion es la interfaz ES del intérprete dentro de una sesión.
type esSesion struct {
	s      *Sesion
	ctx    context.Context
	emitir func(Evento)
}

func (e *esSesion) Escribir(linea string) {
	e.emitir(Evento{Tipo: EventoSalida, Texto: linea})
}

func (e *esSesion) Leer(prompt string) (string, error) {
	e.s.mu.Lock()
	e.s.esperando = true
	e.s.mu.Unlock()
	defer func() {
		e.s.mu.Lock()
		e.s.esperando = false
		e.s.mu.Unlock()
	}()

	e.emitir(Evento{Tipo: EventoPedirEntrada, Texto: prompt})
	select {
	case texto := <-e.s.entradas:
		return texto, nil
	case <-e.ctx.Done():
		return "", interprete.ErrEntradaCerrada
	}
}
