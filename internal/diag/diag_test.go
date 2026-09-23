package diag

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStringFormatoDeLaEspecificacion(t *testing.T) {
	d := Diagnostic{
		Severity: Error,
		Category: Semantico,
		Code:     "tipo-incompatible",
		Heading:  EncabezadoTipos,
		File:     "principal.pks",
		Line:     12,
		Col:      13,
		Len:      5,
		Desc:     "no es posible sumar un valor de tipo roca con uno de tipo electrico.",
		Cause:    "la combinación de estos dos tipos no tiene efecto según la tabla de efectividades.",
		Suggest:  "consultar la tabla de efectividades desde el menú del entorno.",
	}
	want := strings.Join([]string{
		"No es muy efectivo…",
		"Tipo de error: semántico",
		"Archivo: principal.pks",
		"Línea: 12",
		"Columna: 13",
		"Descripción: no es posible sumar un valor de tipo roca con uno de tipo electrico.",
		"Posible causa: la combinación de estos dos tipos no tiene efecto según la tabla de efectividades.",
		"Sugerencia: consultar la tabla de efectividades desde el menú del entorno.",
	}, "\n")
	if got := d.String(); got != want {
		t.Errorf("String() =\n%s\n\nwant\n%s", got, want)
	}
}

func TestStringOmiteCamposVacios(t *testing.T) {
	d := Diagnostic{
		Severity: Advertencia,
		Category: Semantico,
		Heading:  EncabezadoAdvertencia,
		Line:     3,
		Col:      5,
		Desc:     "el dato vida se declaró y nunca se usa.",
	}
	got := d.String()
	for _, prohibido := range []string{"Archivo:", "Posible causa:", "Sugerencia:", "Tipo de error:"} {
		if strings.Contains(got, prohibido) {
			t.Errorf("String() no debería contener %q:\n%s", prohibido, got)
		}
	}
	if !strings.Contains(got, "Tipo de advertencia: semántico") {
		t.Errorf("String() debería indicar que es advertencia:\n%s", got)
	}
}

func TestNombreDeCategoria(t *testing.T) {
	casos := map[Categoria]string{
		Lexico:      "léxico",
		Sintactico:  "sintáctico",
		Semantico:   "semántico",
		Importacion: "importación",
		Ejecucion:   "ejecución",
	}
	for c, want := range casos {
		if got := c.Nombre(); got != want {
			t.Errorf("%q.Nombre() = %q, want %q", c, got, want)
		}
	}
}

func TestListaRespetaElLimiteDeErrores(t *testing.T) {
	l := Lista{Max: 2}
	err := Diagnostic{Severity: Error}
	adv := Diagnostic{Severity: Advertencia}

	if !l.Agregar(err) || !l.Agregar(adv) || !l.Agregar(err) {
		t.Fatal("los dos primeros errores y la advertencia deberían aceptarse")
	}
	if !l.Lleno() {
		t.Error("con 2 errores y Max 2 la lista debería estar llena")
	}
	if l.Agregar(err) {
		t.Error("el tercer error debería descartarse")
	}
	if !l.Agregar(adv) {
		t.Error("las advertencias no cuentan para el límite")
	}
	if n := len(l.Items()); n != 4 {
		t.Errorf("len(Items()) = %d, want 4", n)
	}
	if !l.TieneErrores() {
		t.Error("TieneErrores() debería ser true")
	}
}

func TestListaSinLimite(t *testing.T) {
	var l Lista
	for i := 0; i < 50; i++ {
		l.Agregar(Diagnostic{Severity: Error})
	}
	if l.Lleno() || len(l.Items()) != 50 {
		t.Errorf("sin Max la lista no tiene límite: Lleno=%v len=%d", l.Lleno(), len(l.Items()))
	}
}

func TestJSONParaElFrontend(t *testing.T) {
	d := Diagnostic{Severity: Error, Category: Sintactico, Code: "bloque-sin-cerrar", Line: 1, Col: 1, Len: 7}
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, campo := range []string{`"severity":"error"`, `"category":"sintactico"`, `"code":"bloque-sin-cerrar"`, `"len":7`} {
		if !strings.Contains(s, campo) {
			t.Errorf("JSON sin %s: %s", campo, s)
		}
	}
	for _, omitido := range []string{`"fix"`, `"cause"`, `"suggest"`} {
		if strings.Contains(s, omitido) {
			t.Errorf("JSON no debería incluir %s vacío: %s", omitido, s)
		}
	}
}
