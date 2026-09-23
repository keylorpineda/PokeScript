package token

import "testing"

func TestHay48PalabrasReservadas(t *testing.T) {
	if n := len(PalabrasReservadas()); n != 48 {
		t.Fatalf("PalabrasReservadas() tiene %d palabras, la especificación define 48", n)
	}
}

func TestBuscar(t *testing.T) {
	casos := []struct {
		entrada string
		want    Kind
	}{
		{"combate", COMBATE},
		{"enseñar", ENSENAR},
		{"tamaño", TAMANO},
		{"electrico", ELECTRICO},
		{"sino", SINO},
		{"si", SI},
		{"a", A},
		{"y", Y},
		{"vida", IDENT},
		{"Roca", IDENT},      // distingue mayúsculas
		{"eléctrico", IDENT}, // las palabras reservadas no llevan tilde
		{"ensenar", IDENT},   // la ñ se conserva
		{"sino si", IDENT},   // lo arma el lexer, no la tabla
	}
	for _, c := range casos {
		if got := Buscar(c.entrada); got != c.want {
			t.Errorf("Buscar(%q) = %v, want %v", c.entrada, got, c.want)
		}
	}
}

func TestCadaPalabraVuelveASuKind(t *testing.T) {
	for _, p := range PalabrasReservadas() {
		k := Buscar(p)
		if !k.EsPalabraReservada() {
			t.Errorf("Buscar(%q) = %v, no es palabra reservada", p, k)
		}
		if k.String() != p {
			t.Errorf("Buscar(%q).String() = %q", p, k.String())
		}
	}
}

func TestTodosLosKindTienenNombre(t *testing.T) {
	marcadores := map[Kind]bool{
		literalStart: true, literalEnd: true,
		simboloStart: true, simboloEnd: true,
		palabraStart: true, palabraEnd: true,
	}
	vistos := map[string]Kind{}
	for k := ILLEGAL; k < palabraEnd; k++ {
		if marcadores[k] {
			continue
		}
		n := k.String()
		if n == "Kind(?)" {
			t.Errorf("Kind %d no tiene nombre", int(k))
			continue
		}
		if otro, ok := vistos[n]; ok {
			t.Errorf("Kind %d y %d comparten el nombre %q", int(otro), int(k), n)
		}
		vistos[n] = k
	}
}

func TestCategoria(t *testing.T) {
	casos := []struct {
		k    Kind
		want Categoria
	}{
		{COMBATE, CatReservada},
		{SINO_SI, CatReservada},
		{GRITAR, CatReservada},
		{ROCA, CatTipo},
		{MOCHILA, CatTipo},
		{POSIBLE, CatTipo},
		{VERDADERO, CatLiteral},
		{FANTASMA, CatLiteral},
		{ROCA_LIT, CatLiteral},
		{PLANTA_LIT, CatLiteral},
		{IDENT, CatIdentificador},
		{RESTO, CatOperador},
		{IGUAL, CatOperador},
		{CONTIENE, CatOperador},
		{PLUS, CatOperador},
		{GE, CatOperador},
		{LPAREN, CatSimbolo},
		{COMMA, CatSimbolo},
		{NEWLINE, CatOtro},
	}
	for _, c := range casos {
		if got := c.k.Categoria(); got != c.want {
			t.Errorf("%v.Categoria() = %q, want %q", c.k, got, c.want)
		}
	}
}

func TestEsTipoSimple(t *testing.T) {
	for _, k := range []Kind{ROCA, AGUA, FUEGO, PLANTA, ELECTRICO} {
		if !k.EsTipoSimple() {
			t.Errorf("%v debería ser tipo simple", k)
		}
	}
	for _, k := range []Kind{EQUIPO, MOCHILA, POSIBLE, IDENT} {
		if k.EsTipoSimple() {
			t.Errorf("%v no debería ser tipo simple", k)
		}
	}
}
