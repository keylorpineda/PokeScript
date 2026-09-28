package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/keylorpineda/PokeScript/internal/interprete"
)

// consola implementa interprete.ES sobre la terminal: gritar escribe una
// línea y capturar muestra el mensaje y lee lo que el usuario escribe.
type consola struct {
	salida  io.Writer
	entrada *bufio.Reader
}

func nuevaConsola(entrada io.Reader, salida io.Writer) *consola {
	return &consola{salida: salida, entrada: bufio.NewReader(entrada)}
}

func (c *consola) Escribir(linea string) { fmt.Fprintln(c.salida, linea) }

func (c *consola) Leer(prompt string) (string, error) {
	fmt.Fprint(c.salida, prompt)
	linea, err := c.entrada.ReadString('\n')
	if err != nil && linea == "" {
		return "", interprete.ErrEntradaCerrada
	}
	return strings.TrimRight(linea, "\r\n"), nil
}
