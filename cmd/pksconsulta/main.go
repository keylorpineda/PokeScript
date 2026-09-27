// Comando pksconsulta escribe en JSON el menú de consulta (palabras
// reservadas y tabla de efectividades) para que la interfaz lo use sin Wails.
// Uso: go run ./cmd/pksconsulta > frontend/src/lib/consulta.json
package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/keylorpineda/PokeScript/internal/consulta"
)

type menu struct {
	Palabras      []consulta.PalabraDoc `json:"palabras"`
	Efectividades [][]string            `json:"efectividades"`
}

func main() {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(menu{consulta.PalabrasReservadas(), consulta.TablaEfectividades()}); err != nil {
		log.Fatal(err)
	}
}
