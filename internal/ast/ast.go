package ast

import "github.com/keylorpineda/PokeScript/internal/token"

// Pos es la posición de un nodo en el archivo fuente. Todos los nodos la
// tienen: sin ella no hay subrayado en el editor ni diagnósticos útiles.
type Pos struct {
	Line int // base 1
	Col  int // base 1, en runas
	Len  int // en runas
}

// Posicion devuelve la posición del nodo. Al estar en Pos, cualquier nodo
// que la incluya cumple la interfaz Nodo.
func (p Pos) Posicion() Pos { return p }

// DesdeToken arma una Pos con la ubicación de un token.
func DesdeToken(t token.Token) Pos { return Pos{Line: t.Line, Col: t.Col, Len: t.Len} }

// Nodo es cualquier elemento del árbol.
type Nodo interface{ Posicion() Pos }

// Decl es una declaración de alcance de archivo.
type Decl interface {
	Nodo
	decl()
}

// Instr es una instrucción dentro de un bloque.
type Instr interface {
	Nodo
	instr()
}

// Expr es una expresión que produce un valor.
type Expr interface {
	Nodo
	expr()
}

// ─── Programa e importaciones ──────────────────────────────────────────────

// Programa es un archivo .pks: importaciones → declaraciones → combate.
type Programa struct {
	Archivo       string
	Importaciones []*Importacion
	Declaraciones []Decl
}

// Importacion: enseñar calcular_dano, Estado desde "operaciones.pks"
type Importacion struct {
	Pos
	Nombres []*Ident
	Ruta    string
	RutaPos Pos // para subrayar solo la ruta cuando el archivo no existe
}

// ─── Tipos escritos en el código ───────────────────────────────────────────

// FormaTipo distingue las variantes de TipoExpr.
type FormaTipo int

const (
	TipoSimple   FormaTipo = iota // roca, agua, fuego, planta, electrico
	TipoNombrado                  // Estado, Pokemon (especie o ficha, lo resuelve el analizador)
	TipoEquipo                    // equipo de T
	TipoMochila                   // mochila de C a V
)

// TipoExpr es un tipo tal como aparece en el código. El analizador lo
// convierte en un tipos.Type.
type TipoExpr struct {
	Pos
	Forma   FormaTipo
	Posible bool       // posible T (solo simples y nombrados)
	Simple  token.Kind // en TipoSimple: ROCA, AGUA, FUEGO, PLANTA o ELECTRICO
	Nombre  string     // en TipoNombrado
	Clave   *TipoExpr  // en TipoMochila
	Elem    *TipoExpr  // en TipoEquipo y TipoMochila
}

// ─── Declaraciones de archivo ──────────────────────────────────────────────

// DeclMedalla: medalla roca VIDA_MAXIMA = 100 (alcance de archivo).
type DeclMedalla struct {
	Pos
	Tipo   *TipoExpr
	Nombre *Ident
	Valor  Expr
}

// DeclEspecie: especie Estado SANO, DORMIDO fin
type DeclEspecie struct {
	Pos
	Nombre  *Ident
	Valores []*Ident
}

// DeclFicha: ficha Pokemon planta nombre roca vida fin
type DeclFicha struct {
	Pos
	Nombre *Ident
	Campos []*Campo
}

// Campo de una ficha.
type Campo struct {
	Pos
	Tipo   *TipoExpr
	Nombre *Ident
}

// DeclMovimiento: movimiento [tipo] nombre(parámetros) … fin
type DeclMovimiento struct {
	Pos
	Retorno *TipoExpr // nil si el movimiento no entrega valor
	Nombre  *Ident
	Params  []*Param
	Cuerpo  []Instr
	FinPos  Pos // posición del fin, para diagnósticos de retorno faltante
}

// Param es un parámetro de movimiento.
type Param struct {
	Pos
	Tipo   *TipoExpr
	Nombre *Ident
}

// Combate es el bloque principal del programa.
type Combate struct {
	Pos
	Cuerpo []Instr
}

// ─── Instrucciones ─────────────────────────────────────────────────────────

// DeclDato: [medalla] tipo nombre [= valor]
type DeclDato struct {
	Pos
	Medalla bool
	Tipo    *TipoExpr
	Nombre  *Ident
	Valor   Expr // nil si se declara sin valor
}

// Asignacion: destino = valor. Destino es *Ident, *Indice o *CampoAcceso.
type Asignacion struct {
	Pos
	Destino Expr
	Valor   Expr
}

// Gritar: gritar expr, expr, …
type Gritar struct {
	Pos
	Args []Expr
}

// Capturar: capturar(destino, mensaje). Destino es *Ident, *Indice o *CampoAcceso.
type Capturar struct {
	Pos
	Destino Expr
	Mensaje Expr
}

// LlamadaInstr es una llamada a movimiento usada como instrucción.
type LlamadaInstr struct {
	Pos
	Llamada *Llamada
}

// Sumar: sumar valor a coleccion
type Sumar struct {
	Pos
	Valor     Expr
	Coleccion *Ident
}

// Quitar: quitar coleccion[indice o clave]
type Quitar struct {
	Pos
	Coleccion *Ident
	Indice    Expr
}

// Si: si cond … {sino si cond …} [sino …] fin
type Si struct {
	Pos
	Ramas []*RamaSi // la primera es el si; las demás, los sino si
	// Sino son las instrucciones del sino. El parser lo deja como slice
	// vacío, nunca nil, cuando no hay sino; así se puede recorrer sin
	// comprobar nada.
	Sino      []Instr
	TieneSino bool // distingue "sin sino" de "sino vacío" (asignación definida)
	SinoPos   Pos  // cero si no hay sino
}

// RamaSi es el si inicial o un sino si.
type RamaSi struct {
	Pos
	Cond   Expr
	Cuerpo []Instr
}

// Segun: segun valor  patrón, patrón entonces instr … [otro entonces instr] fin
type Segun struct {
	Pos
	Valor        Expr
	Alternativas []*Alternativa
	Otro         Instr // nil si no hay otro
	OtroPos      Pos
}

// Alternativa de un segun. Cada patrón es un literal o un *Ident.
type Alternativa struct {
	Pos
	Patrones []Expr
	Cuerpo   Instr
}

// Mientras: mientras cond … fin
type Mientras struct {
	Pos
	Cond   Expr
	Cuerpo []Instr
}

// RecorrerColeccion: recorrer x [, y] en coleccion … fin
type RecorrerColeccion struct {
	Pos
	Var       *Ident
	Var2      *Ident // valor, al recorrer una mochila con dos variables; nil si no hay
	Coleccion Expr
	Cuerpo    []Instr
}

// RecorrerRango: recorrer n de desde hasta hasta … fin
type RecorrerRango struct {
	Pos
	Var    *Ident
	Desde  Expr
	Hasta  Expr
	Cuerpo []Instr
}

// Huir sale del ciclo más interno.
type Huir struct{ Pos }

// Siguiente pasa a la vuelta siguiente del ciclo más interno.
type Siguiente struct{ Pos }

// Entregar: entregar [valor]
type Entregar struct {
	Pos
	Valor Expr // nil en movimientos sin tipo
}

// ─── Expresiones ───────────────────────────────────────────────────────────

// Ident es un nombre: dato, movimiento, especie, ficha, valor de especie o campo.
type Ident struct {
	Pos
	Nombre string
}

// LitRoca es un literal entero.
type LitRoca struct {
	Pos
	Valor int64
}

// LitAgua es un literal decimal.
type LitAgua struct {
	Pos
	Valor float64
}

// LitFuego es un literal de un carácter.
type LitFuego struct {
	Pos
	Valor rune
}

// LitPlanta es un literal de texto, ya sin comillas y con escapes resueltos.
type LitPlanta struct {
	Pos
	Valor string
}

// LitElectrico es verdadero o falso.
type LitElectrico struct {
	Pos
	Valor bool
}

// LitFantasma es el valor nulo de los tipos posible.
type LitFantasma struct{ Pos }

// LitEquipo: [a, b, c]
type LitEquipo struct {
	Pos
	Elems []Expr
}

// LitLlaves: { … }. El parser no puede saber si es mochila o ficha y no lo
// intenta: deja Resuelto en LlavesSinResolver y el analizador lo llena según
// el tipo esperado (sección 2.1).
type LitLlaves struct {
	Pos
	Pares    []*Par
	Resuelto FormaLlaves // lo llena el analizador
}

// FormaLlaves dice qué es un literal { }. Se define aquí y no se usa
// tipos.Kind para que ast no dependa de tipos.
type FormaLlaves int

const (
	LlavesSinResolver FormaLlaves = iota // el parser siempre deja este valor
	LlavesMochila
	LlavesFicha
)

// Par clave: valor dentro de LitLlaves. En una ficha, Clave es un *Ident.
type Par struct {
	Pos
	Clave Expr
	Valor Expr
}

// Binaria: izq op der. Op es PLUS, MINUS, STAR, SLASH, RESTO, GT, LT, GE,
// LE, IGUAL, DIFERENTE, CONTIENE, Y u O.
type Binaria struct {
	Pos
	Op    token.Kind
	OpPos Pos
	Izq   Expr
	Der   Expr
}

// Unaria: op operando. Op es MINUS o NO.
type Unaria struct {
	Pos
	Op       token.Kind
	Operando Expr
}

// Respaldo: valor sino reemplazo (valor de respaldo para un posible).
type Respaldo struct {
	Pos
	Valor     Expr
	Reemplazo Expr
}

// Indice: coleccion[indice]
type Indice struct {
	Pos
	Coleccion Expr
	Indice    Expr
}

// CampoAcceso: objeto.nombre (acceso a un campo de ficha)
type CampoAcceso struct {
	Pos
	Objeto Expr
	Nombre *Ident
}

// Llamada: nombre(args)
type Llamada struct {
	Pos
	Nombre *Ident
	Args   []Expr
}

// Convertir: convertir(valor) a tipo
type Convertir struct {
	Pos
	Valor   Expr
	Destino *TipoExpr
}

// Tamano: tamaño(valor)
type Tamano struct {
	Pos
	Valor Expr
}

// Aleatorio: aleatorio(min, max)
type Aleatorio struct {
	Pos
	Min Expr
	Max Expr
}

// Redondear: redondear(valor)
type Redondear struct {
	Pos
	Valor Expr
}

// ─── Marcas de interfaz ────────────────────────────────────────────────────

func (*DeclMedalla) decl()    {}
func (*DeclEspecie) decl()    {}
func (*DeclFicha) decl()      {}
func (*DeclMovimiento) decl() {}
func (*Combate) decl()        {}

func (*DeclDato) instr()          {}
func (*Asignacion) instr()        {}
func (*Gritar) instr()            {}
func (*Capturar) instr()          {}
func (*LlamadaInstr) instr()      {}
func (*Sumar) instr()             {}
func (*Quitar) instr()            {}
func (*Si) instr()                {}
func (*Segun) instr()             {}
func (*Mientras) instr()          {}
func (*RecorrerColeccion) instr() {}
func (*RecorrerRango) instr()     {}
func (*Huir) instr()              {}
func (*Siguiente) instr()         {}
func (*Entregar) instr()          {}

func (*Ident) expr()        {}
func (*LitRoca) expr()      {}
func (*LitAgua) expr()      {}
func (*LitFuego) expr()     {}
func (*LitPlanta) expr()    {}
func (*LitElectrico) expr() {}
func (*LitFantasma) expr()  {}
func (*LitEquipo) expr()    {}
func (*LitLlaves) expr()    {}
func (*Binaria) expr()      {}
func (*Unaria) expr()       {}
func (*Respaldo) expr()     {}
func (*Indice) expr()       {}
func (*CampoAcceso) expr()  {}
func (*Llamada) expr()      {}
func (*Convertir) expr()    {}
func (*Tamano) expr()       {}
func (*Aleatorio) expr()    {}
func (*Redondear) expr()    {}
