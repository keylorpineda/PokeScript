// Command pks compila y ejecuta un programa PokeScript desde la terminal,
// sin abrir el IDE:
//
//	go run ./cmd/pks ejemplos/combate
//	go run ./cmd/pks programa.pks
//
// Sirve para probar el compilador sin Wails y como plan B de la demo.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/interprete"
	"github.com/keylorpineda/PokeScript/internal/servicio"
)

// Códigos de salida.
const (
	salidaOK         = 0
	salidaCompilar   = 1 // errores léxicos o sintácticos
	salidaEjecucion  = 2 // error de ejecución
	salidaUso        = 3 // argumentos o archivos inválidos
	salidaInterrumpe = 130
)

func main() {
	ctx, cancelar := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelar()
	os.Exit(correr(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr, nil))
}

// correr hace todo el trabajo de main y devuelve el código de salida. Se
// separa para poder probarlo; motor permite fijar la semilla de aleatorio.
func correr(ctx context.Context, args []string, entrada io.Reader, salida, errores io.Writer, motor *interprete.Interprete) int {
	if len(args) != 1 {
		fmt.Fprintln(errores, "Uso: pks <archivo.pks | carpeta del proyecto>")
		return salidaUso
	}
	c, err := servicio.Compilar(args[0])
	if err != nil {
		fmt.Fprintln(errores, "✖", err)
		return salidaUso
	}
	imprimir(errores, c.Diagnosticos) // errores, o advertencias si las hay
	if c.TieneErrores() {
		return salidaCompilar
	}

	if motor == nil {
		motor = interprete.Nuevo(nil)
	}
	motor.ES = nuevaConsola(entrada, salida)
	err = servicio.Ejecutar(ctx, c, motor)

	var errEjecucion *interprete.ErrorEjecucion
	switch {
	case err == nil:
		return salidaOK
	case errors.As(err, &errEjecucion):
		fmt.Fprintln(errores)
		imprimir(errores, []diag.Diagnostic{errEjecucion.Diag})
		return salidaEjecucion
	case errors.Is(err, context.Canceled), errors.Is(err, interprete.ErrEntradaCerrada):
		fmt.Fprintln(errores, "\nEjecución detenida.")
		return salidaInterrumpe
	default:
		fmt.Fprintln(errores, "✖", err)
		return salidaEjecucion
	}
}

func imprimir(w io.Writer, diags []diag.Diagnostic) {
	for i, d := range diags {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, d.String())
	}
}
