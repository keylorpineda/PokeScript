package interprete

import (
	"errors"
	"sync"
)

// ES es la entrada y salida del programa en ejecución. El intérprete solo
// conoce esta interfaz, nunca a Wails, así que se puede probar sin interfaz
// gráfica.
//
// En la aplicación, app.go la implementa con eventos de Wails:
//
//	Escribir → runtime.EventsEmit(ctx, "salida", linea)
//	Leer     → runtime.EventsEmit(ctx, "pedir-entrada", prompt), y espera en
//	           un canal que llena EnviarEntrada(texto) desde Svelte.
//
// Al terminar, app.go emite "fin-ejecucion"; ante un error de ejecución,
// "error-ejecucion" con el diag.Diagnostic.
type ES interface {
	// Escribir muestra una línea completa (gritar ya agrega el salto de línea).
	Escribir(linea string)
	// Leer muestra el prompt y bloquea hasta recibir una línea del usuario.
	// Devuelve ErrEntradaCerrada si la ejecución se detuvo mientras esperaba.
	Leer(prompt string) (string, error)
}

// ErrEntradaCerrada indica que ya no llegará más entrada: se detuvo la
// ejecución o se acabaron las respuestas preparadas en una prueba.
var ErrEntradaCerrada = errors.New("la entrada se cerró antes de recibir una respuesta")

// ESMemoria implementa ES en memoria, para pruebas. Las respuestas a
// capturar se toman en orden de Entradas; lo escrito y los prompts quedan
// guardados para compararlos con la salida esperada.
type ESMemoria struct {
	mu       sync.Mutex
	Entradas []string
	Salida   []string
	Prompts  []string
}

// NuevaESMemoria crea una ESMemoria que responderá con entradas, en orden.
func NuevaESMemoria(entradas ...string) *ESMemoria {
	return &ESMemoria{Entradas: entradas}
}

// Escribir guarda la línea.
func (e *ESMemoria) Escribir(linea string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Salida = append(e.Salida, linea)
}

// Leer guarda el prompt y devuelve la siguiente entrada preparada.
func (e *ESMemoria) Leer(prompt string) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Prompts = append(e.Prompts, prompt)
	if len(e.Entradas) == 0 {
		return "", ErrEntradaCerrada
	}
	r := e.Entradas[0]
	e.Entradas = e.Entradas[1:]
	return r, nil
}
