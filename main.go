// PokeScript: el IDE de escritorio. Wails abre la interfaz de frontend/ y le
// expone los métodos de App (app.go).
package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var interfaz embed.FS

func main() {
	app := NuevaApp()
	err := wails.Run(&options.App{
		Title:            "PokeScript",
		Width:            1400,
		Height:           880,
		MinWidth:         760,
		MinHeight:        560,
		BackgroundColour: &options.RGBA{R: 16, G: 16, B: 24, A: 255},
		// Sin el marco de Windows: la interfaz dibuja su propia barra con
		// minimizar, maximizar, pantalla completa y cerrar.
		Frameless:   true,
		AssetServer: &assetserver.Options{Assets: interfaz},
		OnStartup:   app.iniciar,
		OnShutdown:  app.cerrar,
		Bind:        []interface{}{app},
	})
	if err != nil {
		log.Fatal(err)
	}
}
