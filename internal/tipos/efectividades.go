package tipos

// Efecto es una casilla de la tabla de efectividades (sección 3.2).
type Efecto int

const (
	SinEfecto         Efecto = iota // SE: error, ni siquiera con convertir
	MismoTipo                       // MT
	Efectivo                        // EF: conversión automática
	RequiereConvertir               // RC: solo con convertir
)

// Abreviatura devuelve la sigla de la tabla: MT, EF, RC o SE.
func (e Efecto) Abreviatura() string {
	switch e {
	case MismoTipo:
		return "MT"
	case Efectivo:
		return "EF"
	case RequiereConvertir:
		return "RC"
	}
	return "SE"
}

// Conversiones es la tabla de efectividades, indexada por [origen][destino].
// Una casilla que no aparece es SinEfecto. Las filas y columnas de especie
// valen para dos especies distintas; una especie consigo misma es MismoTipo.
var Conversiones = map[Kind]map[Kind]Efecto{
	KRoca:      {KRoca: MismoTipo, KAgua: Efectivo, KPlanta: RequiereConvertir},
	KAgua:      {KRoca: RequiereConvertir, KAgua: MismoTipo, KPlanta: RequiereConvertir},
	KFuego:     {KFuego: MismoTipo, KPlanta: RequiereConvertir},
	KPlanta:    {KRoca: RequiereConvertir, KAgua: RequiereConvertir, KPlanta: MismoTipo},
	KElectrico: {KPlanta: RequiereConvertir, KElectrico: MismoTipo},
	KEspecie:   {KPlanta: RequiereConvertir, KEspecie: MismoTipo},
}

// ordenTabla es el orden de filas y columnas de la tabla 3.2.
var ordenTabla = []Kind{KRoca, KAgua, KFuego, KPlanta, KElectrico, KEspecie}

// Convertible dice qué hace falta para llevar un valor de tipo origen a un
// lugar de tipo destino.
//
//   - Tipos simples y especies: la tabla de efectividades.
//   - T → posible T es automática; posible T → T no (necesita comprobar
//     antes que no sea fantasma, o usar sino).
//   - fantasma solo va a un posible.
//   - equipo, mochila y ficha: solo el tipo exactamente igual.
func Convertible(origen, destino *Type) Efecto {
	if origen == nil || destino == nil {
		return SinEfecto
	}
	if Iguales(origen, destino) {
		return MismoTipo
	}
	if origen.Kind == KFantasma {
		if destino.Opcional {
			return Efectivo
		}
		return SinEfecto
	}
	if origen.Opcional {
		return SinEfecto
	}
	if destino.Opcional {
		switch Convertible(origen, Base(destino)) {
		case MismoTipo, Efectivo:
			return Efectivo
		}
		return SinEfecto
	}
	fila, ok := Conversiones[origen.Kind]
	if !ok {
		return SinEfecto // equipo, mochila y ficha que no son iguales
	}
	e := fila[destino.Kind]
	if e == MismoTipo && origen.Kind == KEspecie && origen.Nombre != destino.Nombre {
		return SinEfecto // dos especies distintas
	}
	return e
}

// Asignable dice si un valor de tipo valor se puede guardar sin convertir en
// un lugar de tipo destino: mismo tipo, roca → agua, T → posible T o
// fantasma → posible T.
func Asignable(valor, destino *Type) bool {
	switch Convertible(valor, destino) {
	case MismoTipo, Efectivo:
		return true
	}
	return false
}

// TablaEfectividades devuelve la tabla 3.2 como texto, con una fila de
// encabezado, para el menú de consulta de la interfaz. Se genera desde
// Conversiones, así que la tabla que se consulta y la que valida nunca se
// contradicen.
func TablaEfectividades() [][]string {
	encabezado := []string{"Origen \\ Destino"}
	for _, k := range ordenTabla {
		encabezado = append(encabezado, k.String())
	}
	tabla := [][]string{encabezado}
	for _, o := range ordenTabla {
		fila := []string{o.String()}
		for _, d := range ordenTabla {
			fila = append(fila, Conversiones[o][d].Abreviatura())
		}
		tabla = append(tabla, fila)
	}
	return tabla
}
