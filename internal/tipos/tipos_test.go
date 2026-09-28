package tipos

import (
	"reflect"
	"testing"
)

// tablaEspecificacion es la tabla 3.2 copiada tal cual de la especificación.
// Si alguien cambia Conversiones sin cambiar la especificación, esta prueba
// lo detecta celda por celda.
var tablaEspecificacion = [][]string{
	{"Origen \\ Destino", "roca", "agua", "fuego", "planta", "electrico", "especie"},
	{"roca", "MT", "EF", "SE", "RC", "SE", "SE"},
	{"agua", "RC", "MT", "SE", "RC", "SE", "SE"},
	{"fuego", "SE", "SE", "MT", "RC", "SE", "SE"},
	{"planta", "RC", "RC", "SE", "MT", "SE", "SE"},
	{"electrico", "SE", "SE", "SE", "RC", "MT", "SE"},
	{"especie", "SE", "SE", "SE", "RC", "SE", "MT"},
}

func TestTablaEfectividadesCeldaPorCelda(t *testing.T) {
	got := TablaEfectividades()
	if len(got) != len(tablaEspecificacion) {
		t.Fatalf("la tabla tiene %d filas, want %d", len(got), len(tablaEspecificacion))
	}
	for i := range tablaEspecificacion {
		for j := range tablaEspecificacion[i] {
			if got[i][j] != tablaEspecificacion[i][j] {
				t.Errorf("fila %s, columna %s: %s, want %s",
					tablaEspecificacion[i][0], tablaEspecificacion[0][j], got[i][j], tablaEspecificacion[i][j])
			}
		}
	}
}

// Convertible debe coincidir con la tabla para cada par de tipos simples.
func TestConvertibleCoincideConLaTabla(t *testing.T) {
	tipos := []*Type{Roca(), Agua(), Fuego(), Planta(), Electrico(), Especie("Estado")}
	for i, o := range tipos {
		for j, d := range tipos {
			want := tablaEspecificacion[i+1][j+1]
			if got := Convertible(o, d).Abreviatura(); got != want {
				t.Errorf("Convertible(%s, %s) = %s, want %s", o, d, got, want)
			}
		}
	}
}

func TestConvertibleCasosEspeciales(t *testing.T) {
	estado := Especie("Estado")
	casos := []struct {
		nombre          string
		origen, destino *Type
		want            Efecto
	}{
		{"dos especies distintas", estado, Especie("Clima"), SinEfecto},
		{"T a posible T", Planta(), Posible(Planta()), Efectivo},
		{"roca a posible agua", Roca(), Posible(Agua()), Efectivo},
		{"roca a posible planta necesita convertir", Roca(), Posible(Planta()), SinEfecto},
		{"posible T a T", Posible(Planta()), Planta(), SinEfecto},
		{"posible T a posible T", Posible(Planta()), Posible(Planta()), MismoTipo},
		{"posible roca a posible agua", Posible(Roca()), Posible(Agua()), SinEfecto},
		{"fantasma a posible", Fantasma(), Posible(estado), Efectivo},
		{"fantasma a no posible", Fantasma(), Roca(), SinEfecto},
		{"equipo igual", Equipo(Roca()), Equipo(Roca()), MismoTipo},
		{"equipo de roca a equipo de agua", Equipo(Roca()), Equipo(Agua()), SinEfecto},
		{"ficha igual", Ficha("Pokemon"), Ficha("Pokemon"), MismoTipo},
		{"ficha a planta", Ficha("Pokemon"), Planta(), SinEfecto},
		{"mochila a planta", Mochila(Planta(), Roca()), Planta(), SinEfecto},
	}
	for _, c := range casos {
		if got := Convertible(c.origen, c.destino); got != c.want {
			t.Errorf("%s: %s → %s = %s, want %s", c.nombre, c.origen, c.destino, got.Abreviatura(), c.want.Abreviatura())
		}
	}
}

func TestAsignable(t *testing.T) {
	casos := []struct {
		valor, destino *Type
		want           bool
	}{
		{Roca(), Roca(), true},
		{Roca(), Agua(), true}, // ensanchamiento
		{Agua(), Roca(), false},
		{Roca(), Planta(), false}, // necesita convertir
		{Planta(), Posible(Planta()), true},
		{Fantasma(), Posible(Roca()), true},
		{Posible(Roca()), Roca(), false},
		{Equipo(Equipo(Roca())), Equipo(Equipo(Roca())), true},
		{Equipo(Equipo(Roca())), Equipo(Equipo(Agua())), false},
		{Mochila(Planta(), Roca()), Mochila(Planta(), Roca()), true},
		{Mochila(Planta(), Roca()), Mochila(Fuego(), Roca()), false},
	}
	for _, c := range casos {
		if got := Asignable(c.valor, c.destino); got != c.want {
			t.Errorf("Asignable(%s, %s) = %v, want %v", c.valor, c.destino, got, c.want)
		}
	}
}

func TestIguales(t *testing.T) {
	casos := []struct {
		a, b *Type
		want bool
	}{
		{Roca(), Roca(), true},
		{Roca(), Agua(), false},
		{Roca(), Posible(Roca()), false},
		{Especie("Estado"), Especie("Estado"), true},
		{Especie("Estado"), Especie("Clima"), false},
		{Especie("Estado"), Ficha("Estado"), false},
		{Equipo(Equipo(Roca())), Equipo(Equipo(Roca())), true},
		{Equipo(Equipo(Roca())), Equipo(Roca()), false},
		{Mochila(Planta(), Equipo(Roca())), Mochila(Planta(), Equipo(Roca())), true},
		{Mochila(Planta(), Roca()), Mochila(Roca(), Roca()), false},
		{nil, nil, true},
		{Roca(), nil, false},
	}
	for _, c := range casos {
		if got := Iguales(c.a, c.b); got != c.want {
			t.Errorf("Iguales(%s, %s) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestResultadoOp(t *testing.T) {
	estado := Especie("Estado")
	equipo := Equipo(Planta())
	mochila := Mochila(Planta(), Agua())
	casos := []struct {
		op       string
		izq, der *Type
		want     *Type // nil = la operación no existe
	}{
		// Aritmética y ensanchamiento.
		{OpSuma, Roca(), Roca(), Roca()},
		{OpSuma, Roca(), Agua(), Agua()},
		{OpResta, Agua(), Roca(), Agua()},
		{OpProducto, Agua(), Agua(), Agua()},
		{OpDivision, Roca(), Roca(), Agua()},
		{OpResto, Roca(), Roca(), Roca()},
		{OpResto, Agua(), Roca(), nil},
		{OpSuma, Planta(), Planta(), Planta()},
		{OpSuma, Planta(), Roca(), nil},
		{OpResta, Planta(), Planta(), nil},
		{OpSuma, Roca(), Electrico(), nil},
		// Comparaciones.
		{OpMayor, Roca(), Agua(), Electrico()},
		{OpMenor, Fuego(), Fuego(), Electrico()},
		{OpMayor, Planta(), Planta(), nil},
		{OpMayor, estado, estado, nil},
		// Igualdad.
		{OpIgual, estado, estado, Electrico()},
		{OpIgual, Planta(), Posible(Planta()), Electrico()},
		{OpDiferente, Posible(Planta()), Fantasma(), Electrico()},
		{OpIgual, Fantasma(), Posible(Roca()), Electrico()},
		{OpIgual, Roca(), Fantasma(), nil},
		{OpIgual, Roca(), Agua(), nil},
		{OpIgual, equipo, Equipo(Planta()), Electrico()},
		{OpIgual, equipo, Equipo(Roca()), nil},
		// Lógicos.
		{OpY, Electrico(), Electrico(), Electrico()},
		{OpO, Electrico(), Roca(), nil},
		// contiene (H16 incluido).
		{OpContiene, equipo, Planta(), Electrico()},
		{OpContiene, equipo, Roca(), nil},
		{OpContiene, mochila, Planta(), Electrico()},
		{OpContiene, mochila, Agua(), nil},
		{OpContiene, Planta(), Planta(), Electrico()},
		{OpContiene, Planta(), Fuego(), Electrico()},
		{OpContiene, Planta(), Roca(), nil},
		{OpContiene, Fuego(), Fuego(), nil},
		// [ ]
		{OpIndice, equipo, Roca(), Planta()},
		{OpIndice, equipo, Agua(), nil},
		{OpIndice, mochila, Planta(), Agua()},
		{OpIndice, mochila, Roca(), nil},
		{OpIndice, Planta(), Roca(), Fuego()},
		// sino de respaldo.
		{OpRespaldo, Posible(Planta()), Planta(), Planta()},
		{OpRespaldo, Posible(Agua()), Roca(), Agua()},
		{OpRespaldo, Planta(), Planta(), nil},
		// Un posible sin comprobar no opera.
		{OpSuma, Posible(Roca()), Roca(), nil},
		{OpContiene, Posible(Planta()), Planta(), nil},
		// Unarios: der == nil.
		{OpResta, Roca(), nil, Roca()},
		{OpResta, Planta(), nil, nil},
		{OpNo, Electrico(), nil, Electrico()},
		{OpNo, Posible(Electrico()), nil, nil},
		{OpTamano, equipo, nil, Roca()},
		{OpTamano, Planta(), nil, Roca()},
		{OpTamano, Roca(), nil, nil},
	}
	for _, c := range casos {
		got, ok := ResultadoOp(c.op, c.izq, c.der)
		if c.want == nil {
			if ok {
				t.Errorf("%s %s %s debería no existir, dio %s", c.izq, c.op, c.der, got)
			}
			continue
		}
		if !ok || !Iguales(got, c.want) {
			t.Errorf("%s %s %s = %s, %v; want %s", c.izq, c.op, c.der, got, ok, c.want)
		}
	}
}

func TestValido(t *testing.T) {
	casos := []struct {
		tipo *Type
		want bool
	}{
		{Posible(Roca()), true},
		{Posible(Especie("Estado")), true},
		{Posible(Ficha("Pokemon")), false},
		{Posible(Equipo(Roca())), false},
		{Equipo(Posible(Roca())), false},
		{Equipo(Equipo(Roca())), true},
		{Mochila(Planta(), Roca()), true},
		{Mochila(Especie("Estado"), Roca()), true},
		{Mochila(Ficha("Pokemon"), Roca()), false}, // decisión H3
		{Mochila(Agua(), Roca()), false},
		{Mochila(Posible(Planta()), Roca()), false},
		{Mochila(Planta(), Posible(Roca())), false},
		{Fantasma(), false},
	}
	for _, c := range casos {
		if got, motivo := Valido(c.tipo); got != c.want {
			t.Errorf("Valido(%s) = %v (%s), want %v", c.tipo, got, motivo, c.want)
		}
	}
}

func TestString(t *testing.T) {
	casos := map[string]*Type{
		"roca":                     Roca(),
		"posible planta":           Posible(Planta()),
		"Estado":                   Especie("Estado"),
		"posible Estado":           Posible(Especie("Estado")),
		"equipo de equipo de roca": Equipo(Equipo(Roca())),
		"mochila de planta a agua": Mochila(Planta(), Agua()),
		"fantasma":                 Fantasma(),
	}
	for want, tipo := range casos {
		if got := tipo.String(); got != want {
			t.Errorf("String() = %q, want %q", got, want)
		}
	}
}

func TestLosConstructoresNoCompartenTipos(t *testing.T) {
	a, b := Roca(), Roca()
	a.Opcional = true
	if b.Opcional || !reflect.DeepEqual(Roca(), b) {
		t.Error("cambiar un tipo afectó a otro")
	}
	p := Posible(Planta())
	if Base(p).Opcional || !p.Opcional {
		t.Error("Base debe devolver una copia sin posible y no tocar el original")
	}
}
