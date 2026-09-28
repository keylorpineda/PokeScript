package consulta

import (
	"sort"
	"strings"
	"testing"

	"github.com/keylorpineda/PokeScript/internal/lexer"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// TestNoFaltaNingunaPalabra compara el menú con la tabla del lexer: si
// alguien agrega una palabra reservada, esta prueba obliga a documentarla.
func TestNoFaltaNingunaPalabra(t *testing.T) {
	var documentadas []string
	vistas := map[string]bool{}
	for _, p := range PalabrasReservadas() {
		if vistas[p.Palabra] {
			t.Errorf("«%s» está documentada dos veces", p.Palabra)
		}
		vistas[p.Palabra] = true
		if p.Palabra != "sino si" {
			documentadas = append(documentadas, p.Palabra)
		}
	}
	reservadas := token.PalabrasReservadas()
	sort.Strings(documentadas)
	sort.Strings(reservadas)
	if strings.Join(documentadas, " ") != strings.Join(reservadas, " ") {
		t.Errorf("documentadas:\n%v\nreservadas:\n%v", documentadas, reservadas)
	}
	if !vistas["sino si"] {
		t.Error("falta «sino si»")
	}
}

func TestCadaPalabraTieneDescripcionYEjemplo(t *testing.T) {
	grupos := map[string]bool{
		GrupoDeclaracion: true, GrupoTipos: true, GrupoConectores: true, GrupoLiterales: true,
		GrupoEstructura: true, GrupoControl: true, GrupoDatos: true, GrupoOperadores: true,
	}
	for _, p := range PalabrasReservadas() {
		if !grupos[p.Grupo] || p.Descripcion == "" || !strings.HasSuffix(p.Descripcion, ".") {
			t.Errorf("«%s»: grupo o descripción inválidos: %+v", p.Palabra, p)
		}
		// El ejemplo usa la palabra y no tiene errores léxicos.
		r := lexer.Analizar("ejemplo.pks", p.Ejemplo)
		if len(r.Diagnosticos) > 0 {
			t.Errorf("«%s»: el ejemplo tiene errores léxicos: %s", p.Palabra, r.Diagnosticos[0].Desc)
		}
		usa := false
		for _, tk := range r.Tokens {
			if tk.Kind.String() == p.Palabra {
				usa = true
			}
		}
		if !usa {
			t.Errorf("«%s»: el ejemplo %q no usa la palabra", p.Palabra, p.Ejemplo)
		}
	}
}

func TestPalabrasReservadasDevuelveUnaCopia(t *testing.T) {
	PalabrasReservadas()[0].Palabra = "cambiada"
	if PalabrasReservadas()[0].Palabra != "especie" {
		t.Error("modificar el resultado no debe cambiar el menú")
	}
}

func TestTablaEfectividades(t *testing.T) {
	tabla := TablaEfectividades()
	if len(tabla) != 7 || len(tabla[0]) != 7 {
		t.Fatalf("la tabla debe ser de 7×7 con encabezados: %v", tabla)
	}
	if tabla[1][0] != "roca" || tabla[1][2] != "EF" {
		t.Errorf("roca → agua debe ser EF: %v", tabla[1])
	}
}
