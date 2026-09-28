package analizador

import (
	"fmt"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/tipos"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// J1 · Tipos de expresiones (sección 4, pasada 2):
//
//	1  tipos de cada expresión contra la tabla de operaciones
//	2  condición de si y mientras es electrico (sin veracidad implícita)
//	3  cantidad y tipo de los argumentos en cada llamada
//	17 división o resto entre el literal 0
//	19 movimiento con tipo invocado como instrucción (advertencia)
//
// También reporta los nombres no declarados, porque sin ellos no se puede
// saber el tipo de nada.
//
// El recorrido lleva los datos locales en ámbitos. Otras verificaciones se
// enganchan a él desde su propio archivo (ver Ganchos, abajo) para no
// repetir el recorrido ni los ámbitos.
func init() { registrar("tipos", revisarTipos) }

func revisarTipos(c *Contexto) {
	for _, a := range c.Archivos() {
		v := &verificador{c: c, archivo: a.Nombre}
		for _, d := range a.Programa.Declaraciones {
			v.declaracion(d)
		}
	}
}

// ─── Ganchos ───────────────────────────────────────────────────────────────

// Puntos de extensión del recorrido. Cada uno tiene un comportamiento por
// defecto neutro; la verificación dueña lo reemplaza desde su init.
var (
	// tipoLiteral tipa un literal [ ] o { } con el tipo esperado del lugar
	// donde se guarda (J2, literales.go). Por defecto confía en el tipo
	// esperado.
	tipoLiteral = func(_ *verificador, _ ast.Expr, esperado *tipos.Type) *tipos.Type { return esperado }

	// revisarSegun revisa los patrones de un segun y si es exhaustivo cuyo
	// valor es de tipo t (J4, segun.go).
	revisarSegun = func(_ *verificador, _ *ast.Segun, _ *tipos.Type) {}

	// hechosDe devuelve los datos posible que quedan comprobados (distintos
	// de fantasma) si la condición es verdadera y si es falsa (J5,
	// posible.go).
	hechosDe = func(_ *verificador, _ ast.Expr) (siVerdadera, siFalsa []string) { return nil, nil }

	// posibleSinComprobar reporta un posible usado donde hace falta su valor
	// (J5, posible.go).
	posibleSinComprobar = func(_ *verificador, _ ast.Expr, _ *tipos.Type) {}
)

// ─── Verificador y ámbitos ─────────────────────────────────────────────────

// verificador recorre los cuerpos de un archivo con sus ámbitos.
type verificador struct {
	c       *Contexto
	archivo string
	amb     *alcance
	mov     *Movimiento // movimiento en curso; nil en combate
}

// alcance son los datos locales de un bloque.
type alcance struct {
	vars  map[string]*local
	padre *alcance
	// hechos: true = el dato posible está comprobado en este bloque;
	// false = se reasignó y dejó de estarlo.
	hechos map[string]bool
}

// local es un dato declarado dentro de un cuerpo: variable, parámetro o
// variable de recorrer.
type local struct {
	tipo *tipos.Type // nil si su tipo no se pudo resolver
}

func (v *verificador) abrir() {
	v.amb = &alcance{vars: map[string]*local{}, hechos: map[string]bool{}, padre: v.amb}
}
func (v *verificador) cerrar() { v.amb = v.amb.padre }

// enBloque ejecuta f en un ámbito nuevo con los hechos dados ya comprobados.
func (v *verificador) enBloque(hechos []string, f func()) {
	v.abrir()
	for _, h := range hechos {
		v.amb.hechos[h] = true
	}
	f()
	v.cerrar()
}

func (v *verificador) declarar(nombre string, t *tipos.Type) {
	if v.amb != nil {
		v.amb.vars[nombre] = &local{tipo: t}
	}
}

func (v *verificador) buscarLocal(nombre string) *local {
	for a := v.amb; a != nil; a = a.padre {
		if l, ok := a.vars[nombre]; ok {
			return l
		}
	}
	return nil
}

// comprobado dice si un dato posible se sabe distinto de fantasma aquí.
func (v *verificador) comprobado(nombre string) bool {
	for a := v.amb; a != nil; a = a.padre {
		if h, ok := a.hechos[nombre]; ok {
			return h
		}
		if _, declarado := a.vars[nombre]; declarado {
			return false // los hechos de afuera no alcanzan a un dato de este bloque
		}
	}
	return false
}

// olvidar marca que un dato se reasignó: deja de estar comprobado desde aquí
// hasta el final del bloque (sección 4.1).
func (v *verificador) olvidar(nombre string) {
	if v.amb != nil && v.comprobado(nombre) {
		v.amb.hechos[nombre] = false
	}
}

// ─── Reportes ──────────────────────────────────────────────────────────────

func (v *verificador) error(pos ast.Pos, codigo, encabezado, desc, causa, sugerencia string) {
	v.c.Error(v.archivo, pos, codigo, encabezado, desc, causa, sugerencia)
}

func (v *verificador) diagnostico(d *diag.Diagnostic) {
	if d != nil {
		v.c.diags = append(v.c.diags, *d)
	}
}

func (v *verificador) resolver(te *ast.TipoExpr) *tipos.Type {
	if te == nil {
		return nil
	}
	t, d := v.c.Tabla.ResolverTipo(v.archivo, te)
	v.diagnostico(d)
	return t
}

const (
	causaTabla      = "la combinación de estos dos tipos no tiene efecto según la tabla de efectividades."
	sugerenciaTabla = "consultar la tabla de efectividades desde el menú del entorno."
)

// asignable revisa que un valor de tipo t se pueda guardar en un lugar de
// tipo destino. que describe el lugar para el mensaje: "el dato «x»".
func (v *verificador) asignable(e ast.Expr, t, destino *tipos.Type, codigo, que string) {
	if t == nil || destino == nil || tipos.Asignable(t, destino) {
		return
	}
	if t.Opcional && tipos.Asignable(tipos.Base(t), destino) {
		posibleSinComprobar(v, e, t)
		return
	}
	causa := fmt.Sprintf("la tabla de efectividades no permite llevar un %s a un %s.", t, destino)
	sugerencia := "revisa el tipo del valor o el del lugar donde lo guardas."
	switch {
	case tipos.Convertible(t, destino) == tipos.RequiereConvertir:
		causa = fmt.Sprintf("la conversión de %s a %s existe, pero no es automática.", t, destino)
		sugerencia = fmt.Sprintf("usa convertir(valor) a %s.", destino)
	case t.Kind == tipos.KFantasma:
		causa = "fantasma solo se guarda en un dato posible."
		sugerencia = fmt.Sprintf("declara el lugar como «posible %s».", destino)
	}
	v.error(e.Posicion(), codigo, diag.EncabezadoTipos,
		fmt.Sprintf("no se puede usar un valor de tipo %s como %s, que es de tipo %s.", t, que, destino),
		causa, sugerencia)
}

// valor marca que una expresión de tipo t se usa por su valor: si es un
// posible sin comprobar, lo reporta y sigue con el tipo base para no
// producir errores en cascada.
func (v *verificador) valor(e ast.Expr, t *tipos.Type) *tipos.Type {
	if t != nil && t.Opcional {
		posibleSinComprobar(v, e, t)
		return tipos.Base(t)
	}
	return t
}

// ─── Declaraciones de archivo ──────────────────────────────────────────────

func (v *verificador) declaracion(d ast.Decl) {
	switch x := d.(type) {
	case *ast.DeclMedalla:
		m := v.c.Tabla.Medallas[nombre(x.Nombre)]
		if m == nil || m.Decl != x || x.Valor == nil {
			return
		}
		t := v.tipoCon(x.Valor, m.Tipo)
		v.asignable(x.Valor, t, m.Tipo, "asignacion-incompatible", "valor de la medalla «"+m.Nombre+"»")
	case *ast.DeclMovimiento:
		mov := v.c.Tabla.Movimientos[nombre(x.Nombre)]
		if mov == nil || mov.Decl != x {
			return
		}
		v.mov = mov
		v.abrir()
		for _, p := range mov.Params {
			v.declarar(p.Nombre, p.Tipo)
		}
		v.bloque(x.Cuerpo)
		v.cerrar()
		v.mov = nil
	case *ast.Combate:
		v.abrir()
		v.bloque(x.Cuerpo)
		v.cerrar()
	}
}

// ─── Instrucciones ─────────────────────────────────────────────────────────

func (v *verificador) bloque(is []ast.Instr) {
	for _, i := range is {
		v.instruccion(i)
	}
}

func (v *verificador) instruccion(i ast.Instr) { //nolint:gocyclo // un caso por instrucción
	switch x := i.(type) {
	case *ast.DeclDato:
		t := v.resolver(x.Tipo)
		if x.Valor != nil {
			vt := v.tipoCon(x.Valor, t)
			v.asignable(x.Valor, vt, t, "asignacion-incompatible", "valor del dato «"+nombre(x.Nombre)+"»")
		}
		v.declarar(nombre(x.Nombre), t)
	case *ast.Asignacion:
		t := v.tipoDestino(x.Destino)
		vt := v.tipoCon(x.Valor, t)
		v.asignable(x.Valor, vt, t, "asignacion-incompatible", "valor de "+describirDestino(x.Destino))
		if id, ok := x.Destino.(*ast.Ident); ok {
			v.olvidar(id.Nombre)
		}
	case *ast.Gritar:
		for _, a := range x.Args {
			v.tipo(a)
		}
	case *ast.Capturar:
		v.capturar(x)
	case *ast.LlamadaInstr:
		if mov := v.llamada(x.Llamada); mov != nil && mov.Retorno != nil {
			v.c.Advertencia(v.archivo, x.Pos, "valor-descartado",
				fmt.Sprintf("«%s» entrega un %s, pero ese valor no se usa.", mov.Nombre, mov.Retorno),
				"el movimiento se llamó como instrucción suelta, así que su resultado se pierde.",
				fmt.Sprintf("guárdalo en un dato: %s resultado = %s(…).", mov.Retorno, mov.Nombre))
		}
	case *ast.Sumar:
		v.sumar(x)
	case *ast.Quitar:
		v.quitar(x)
	case *ast.Si:
		v.si(x)
	case *ast.Segun:
		t := v.valor(x.Valor, v.tipo(x.Valor))
		revisarSegun(v, x, t)
		for _, alt := range x.Alternativas {
			v.enBloque(nil, func() { v.instruccion(alt.Cuerpo) })
		}
		if x.Otro != nil {
			v.enBloque(nil, func() { v.instruccion(x.Otro) })
		}
	case *ast.Mientras:
		v.condicion(x.Cond, "mientras")
		verdad, _ := hechosDe(v, x.Cond)
		v.enBloque(verdad, func() { v.bloque(x.Cuerpo) })
	case *ast.RecorrerRango:
		for _, e := range []ast.Expr{x.Desde, x.Hasta} {
			t := v.valor(e, v.tipo(e))
			v.asignable(e, t, tipos.Roca(), "tipo-incompatible", "extremo de recorrer")
		}
		v.enBloque(nil, func() {
			v.declarar(nombre(x.Var), tipos.Roca())
			v.bloque(x.Cuerpo)
		})
	case *ast.RecorrerColeccion:
		v.recorrer(x)
	case *ast.Entregar:
		if x.Valor == nil {
			return
		}
		var esperado *tipos.Type
		if v.mov != nil {
			esperado = v.mov.Retorno
		}
		t := v.tipoCon(x.Valor, esperado)
		if esperado != nil {
			v.asignable(x.Valor, t, esperado, "entrega-incompatible", "valor que entrega «"+v.mov.Nombre+"»")
		}
	}
}

// si revisa las condiciones y los cuerpos. Cada sino si y el sino se
// analizan sabiendo que las condiciones anteriores fueron falsas.
func (v *verificador) si(x *ast.Si) {
	var falsas []string
	for _, r := range x.Ramas {
		v.enBloque(falsas, func() { v.condicion(r.Cond, "si") })
		verdad, falso := hechosDe(v, r.Cond)
		v.enBloque(append(append([]string{}, falsas...), verdad...), func() { v.bloque(r.Cuerpo) })
		falsas = append(falsas, falso...)
	}
	if x.TieneSino {
		v.enBloque(falsas, func() { v.bloque(x.Sino) })
	}
}

// condicion exige un electrico: no hay veracidad implícita (validación 2).
func (v *verificador) condicion(e ast.Expr, que string) {
	if e == nil {
		return
	}
	t := v.valor(e, v.tipo(e))
	if t == nil || t.Kind == tipos.KElectrico {
		return
	}
	v.error(e.Posicion(), "condicion-no-electrico", diag.EncabezadoTipos,
		fmt.Sprintf("la condición de «%s» es de tipo %s, y debe ser electrico.", que, t),
		"en PokeScript no hay veracidad implícita: un número o un texto no cuentan como verdadero o falso.",
		"escribe una comparación, por ejemplo «vida > 0» o «nombre diferente \"\"».")
}

func (v *verificador) capturar(x *ast.Capturar) {
	t := v.tipoDestino(x.Destino)
	if t != nil {
		base := tipos.Base(t)
		if !base.Kind.Simple() && base.Kind != tipos.KEspecie {
			v.error(x.Destino.Posicion(), "capturar-invalido", diag.EncabezadoTipos,
				fmt.Sprintf("capturar no puede guardar en %s, que es de tipo %s.", describirDestino(x.Destino), t),
				"capturar lee una sola línea, así que solo llena datos de tipo simple o especie.",
				"captura cada valor por separado y agrégalo con sumar o con [ ].")
		}
	}
	if id, ok := x.Destino.(*ast.Ident); ok {
		v.olvidar(id.Nombre)
	}
	if x.Mensaje != nil {
		mt := v.tipo(x.Mensaje)
		v.asignable(x.Mensaje, mt, tipos.Planta(), "tipo-incompatible", "mensaje de capturar")
	}
}

func (v *verificador) sumar(x *ast.Sumar) {
	ct := v.valor(x.Coleccion, v.tipo(x.Coleccion))
	if ct == nil {
		v.tipo(x.Valor)
		return
	}
	if ct.Kind != tipos.KEquipo {
		v.tipo(x.Valor)
		v.error(x.Coleccion.Pos, "sumar-invalido", diag.EncabezadoTipos,
			fmt.Sprintf("sumar solo agrega a un equipo, y «%s» es de tipo %s.", x.Coleccion.Nombre, ct),
			"en una mochila cada valor necesita su clave.",
			fmt.Sprintf("para una mochila usa %s[clave] = valor.", x.Coleccion.Nombre))
		return
	}
	t := v.tipoCon(x.Valor, ct.Elem)
	v.asignable(x.Valor, t, ct.Elem, "tipo-incompatible", "elemento de «"+x.Coleccion.Nombre+"»")
}

func (v *verificador) quitar(x *ast.Quitar) {
	ct := v.valor(x.Coleccion, v.tipo(x.Coleccion))
	it := v.valor(x.Indice, v.tipo(x.Indice))
	if ct == nil {
		return
	}
	switch ct.Kind {
	case tipos.KEquipo:
		v.asignable(x.Indice, it, tipos.Roca(), "tipo-incompatible", "índice de «"+x.Coleccion.Nombre+"»")
	case tipos.KMochila:
		v.asignable(x.Indice, it, ct.Clave, "tipo-incompatible", "clave de «"+x.Coleccion.Nombre+"»")
	default:
		v.error(x.Coleccion.Pos, "quitar-invalido", diag.EncabezadoTipos,
			fmt.Sprintf("quitar solo se aplica a un equipo o una mochila, y «%s» es de tipo %s.", x.Coleccion.Nombre, ct),
			"solo las colecciones tienen elementos que quitar.",
			"revisa el nombre de la colección.")
	}
}

// recorrer declara la variable según la colección: el elemento de un
// equipo, la clave (y el valor) de una mochila, o cada letra de un texto.
func (v *verificador) recorrer(x *ast.RecorrerColeccion) {
	ct := v.valor(x.Coleccion, v.tipo(x.Coleccion))
	var t1, t2 *tipos.Type
	if ct != nil {
		switch ct.Kind {
		case tipos.KEquipo:
			t1 = ct.Elem
		case tipos.KMochila:
			t1, t2 = ct.Clave, ct.Elem
		case tipos.KPlanta:
			t1 = tipos.Fuego()
		default:
			v.error(x.Coleccion.Posicion(), "recorrer-invalido", diag.EncabezadoTipos,
				fmt.Sprintf("no se puede recorrer un valor de tipo %s.", ct),
				"recorrer … en visita un equipo, una mochila o las letras de un texto.",
				"para contar, usa recorrer n de 1 hasta 10.")
		}
		if x.Var2 != nil && ct.Kind != tipos.KMochila && t1 != nil {
			v.error(x.Var2.Pos, "recorrer-invalido", diag.EncabezadoTipos,
				fmt.Sprintf("recorrer con dos variables solo sirve para una mochila, y esto es un %s.", ct),
				"la segunda variable recibe el valor de cada clave.",
				"quita la segunda variable.")
		}
	}
	v.enBloque(nil, func() {
		v.declarar(nombre(x.Var), t1)
		if x.Var2 != nil {
			v.declarar(x.Var2.Nombre, t2)
		}
		v.bloque(x.Cuerpo)
	})
}

// ─── Expresiones ───────────────────────────────────────────────────────────

// tipoCon tipa una expresión que se guarda en un lugar de tipo esperado;
// los literales [ ] y { } lo necesitan (sección 2.1). Si el tipo del lugar
// no se pudo resolver (esperado nil, con el error ya reportado), un literal
// no se revisa, para no sumar un error en cascada.
func (v *verificador) tipoCon(e ast.Expr, esperado *tipos.Type) *tipos.Type {
	switch e.(type) {
	case *ast.LitEquipo, *ast.LitLlaves:
		if esperado == nil {
			return nil
		}
		return tipoLiteral(v, e, esperado)
	}
	return v.tipo(e)
}

// tipo devuelve el tipo de una expresión, o nil si no se puede saber (por
// un error ya reportado). Un nil no produce más errores.
func (v *verificador) tipo(e ast.Expr) *tipos.Type { //nolint:gocyclo // un caso por expresión
	switch x := e.(type) {
	case nil:
		return nil
	case *ast.LitRoca:
		return tipos.Roca()
	case *ast.LitAgua:
		return tipos.Agua()
	case *ast.LitFuego:
		return tipos.Fuego()
	case *ast.LitPlanta:
		return tipos.Planta()
	case *ast.LitElectrico:
		return tipos.Electrico()
	case *ast.LitFantasma:
		return tipos.Fantasma()
	case *ast.LitEquipo, *ast.LitLlaves:
		// Un literal suelto, sin lugar donde guardarse (validación 14).
		return tipoLiteral(v, e, nil)
	case *ast.Ident:
		return v.nombre(x)
	case *ast.Binaria:
		return v.binaria(x)
	case *ast.Unaria:
		t := v.valor(x.Operando, v.tipo(x.Operando))
		return v.operacion(x, x.Op.String(), t, nil, x.Operando, nil)
	case *ast.Respaldo:
		a, b := v.tipo(x.Valor), v.tipo(x.Reemplazo)
		if a == nil || b == nil {
			return nil
		}
		if r, ok := tipos.ResultadoOp(tipos.OpRespaldo, a, b); ok {
			return r
		}
		desc := fmt.Sprintf("«sino» da un valor de reemplazo para un posible, y aquí la izquierda es de tipo %s.", a)
		if a.Opcional {
			desc = fmt.Sprintf("el reemplazo es de tipo %s, pero el dato es %s.", b, a)
		}
		v.error(x.Pos, "tipo-incompatible", diag.EncabezadoTipos, desc,
			"el reemplazo tiene que servir en lugar del valor que falta.",
			"usa un reemplazo del mismo tipo que el dato posible.")
		return nil
	case *ast.Indice:
		c := v.valor(x.Coleccion, v.tipo(x.Coleccion))
		i := v.valor(x.Indice, v.tipo(x.Indice))
		return v.operacion(x, tipos.OpIndice, c, i, x.Coleccion, x.Indice)
	case *ast.CampoAcceso:
		return v.campo(x)
	case *ast.Llamada:
		mov := v.llamada(x)
		if mov == nil {
			return nil
		}
		if mov.Retorno == nil {
			v.error(x.Pos, "movimiento-sin-valor", diag.EncabezadoTipos,
				fmt.Sprintf("«%s» no entrega ningún valor, así que no se puede usar en una expresión.", mov.Nombre),
				"solo un movimiento declarado con tipo, como «movimiento roca …», produce un valor.",
				fmt.Sprintf("llama a «%s» como instrucción suelta, o dale un tipo y usa entregar.", mov.Nombre))
		}
		return mov.Retorno
	case *ast.Convertir:
		return v.convertir(x)
	case *ast.Tamano:
		t := v.valor(x.Valor, v.tipo(x.Valor))
		return v.operacion(x, tipos.OpTamano, t, nil, x.Valor, nil)
	case *ast.Aleatorio:
		for _, a := range []ast.Expr{x.Min, x.Max} {
			t := v.valor(a, v.tipo(a))
			v.asignable(a, t, tipos.Roca(), "tipo-incompatible", "extremo de aleatorio")
		}
		return tipos.Roca()
	case *ast.Redondear:
		t := v.valor(x.Valor, v.tipo(x.Valor))
		v.asignable(x.Valor, t, tipos.Agua(), "tipo-incompatible", "valor de redondear")
		return tipos.Roca()
	}
	return nil
}

// nombre tipa un identificador usado como valor: un dato local, una
// medalla o un valor de especie.
func (v *verificador) nombre(x *ast.Ident) *tipos.Type {
	if l := v.buscarLocal(x.Nombre); l != nil {
		if l.tipo != nil && l.tipo.Opcional && v.comprobado(x.Nombre) {
			return tipos.Base(l.tipo)
		}
		return l.tipo
	}
	s := v.c.Tabla.Buscar(v.archivo, x.Nombre)
	if s == nil {
		v.noDeclarado(x, "dato")
		return nil
	}
	switch s.Clase {
	case ClaseMedalla:
		return s.Medalla.Tipo
	case ClaseValorEspecie:
		return s.Especie.Tipo()
	}
	que := map[Clase]string{ClaseEspecie: "una especie", ClaseFicha: "una ficha", ClaseMovimiento: "un movimiento"}[s.Clase]
	v.error(x.Pos, "no-es-un-valor", diag.EncabezadoTipos,
		fmt.Sprintf("«%s» es %s, no un valor.", x.Nombre, que),
		"aquí se esperaba un dato, una medalla o un valor de especie.",
		"si es un movimiento, llámalo con paréntesis: "+x.Nombre+"(…).")
	return nil
}

// noDeclarado reporta un nombre que no existe desde este archivo. Si
// existe en otro archivo del proyecto, sugiere importarlo.
func (v *verificador) noDeclarado(x *ast.Ident, que string) {
	sugerencia := fmt.Sprintf("declara «%s» antes de usarlo, o revisa cómo está escrito.", x.Nombre)
	if origen := v.declaradoEn(x.Nombre); origen != "" && origen != v.archivo {
		sugerencia = fmt.Sprintf("«%s» está en «%s»: agrega «enseñar %s desde \"%s\"» al inicio del archivo.", x.Nombre, origen, x.Nombre, origen)
	}
	v.error(x.Pos, "nombre-no-declarado", diag.EncabezadoSinValor,
		fmt.Sprintf("«%s» no está declarado.", x.Nombre),
		fmt.Sprintf("no hay ningún %s, medalla ni valor de especie con ese nombre en este archivo.", que),
		sugerencia)
}

// declaradoEn devuelve el archivo donde se declara un nombre de alcance de
// archivo, aunque este archivo no lo vea.
func (v *verificador) declaradoEn(n string) string {
	t := v.c.Tabla
	switch {
	case t.Movimientos[n] != nil:
		return t.Movimientos[n].Archivo
	case t.Medallas[n] != nil:
		return t.Medallas[n].Archivo
	case t.Especies[n] != nil:
		return t.Especies[n].Archivo
	case t.Fichas[n] != nil:
		return t.Fichas[n].Archivo
	case t.Valores[n] != nil:
		return t.Valores[n].Archivo
	}
	return ""
}

func (v *verificador) binaria(x *ast.Binaria) *tipos.Type {
	op := x.Op.String()
	if x.Op == token.Y {
		// Lo que la izquierda comprueba vale al analizar la derecha,
		// igual que el cortocircuito (sección 4.1).
		a := v.valor(x.Izq, v.tipo(x.Izq))
		verdad, _ := hechosDe(v, x.Izq)
		var b *tipos.Type
		v.enBloque(verdad, func() { b = v.valor(x.Der, v.tipo(x.Der)) })
		return v.operacion(x, op, a, b, x.Izq, x.Der)
	}
	a, b := v.tipo(x.Izq), v.tipo(x.Der)
	if x.Op != token.IGUAL && x.Op != token.DIFERENTE {
		a, b = v.valor(x.Izq, a), v.valor(x.Der, b)
	}
	if (x.Op == token.SLASH || x.Op == token.RESTO) && esCeroLiteral(x.Der) {
		accion := "dividir"
		if x.Op == token.RESTO {
			accion = "calcular el resto"
		}
		v.error(x.Der.Posicion(), "division-entre-cero", diag.EncabezadoTipos,
			fmt.Sprintf("no se puede %s entre 0.", accion),
			"dividir entre cero no tiene resultado; este programa fallaría siempre al llegar aquí.",
			"revisa el divisor.")
	}
	return v.operacion(x, op, a, b, x.Izq, x.Der)
}

// operacion consulta la tabla 3.3 y, si la combinación no existe, lo
// explica. der es nil en los operadores de un solo operando.
func (v *verificador) operacion(n ast.Nodo, op string, a, b *tipos.Type, ea, eb ast.Expr) *tipos.Type {
	if a == nil || (eb != nil && b == nil) {
		return nil
	}
	if r, ok := tipos.ResultadoOp(op, a, b); ok {
		return r
	}
	if (a.Opcional || (b != nil && b.Opcional)) && op != tipos.OpIgual && op != tipos.OpDiferente {
		return nil // ya lo reportó posibleSinComprobar
	}
	v.error(n.Posicion(), "tipo-incompatible", diag.EncabezadoTipos, descripcionOp(op, a, b), causaTabla, sugerenciaTabla)
	return nil
}

func descripcionOp(op string, a, b *tipos.Type) string {
	verbos := map[string]string{
		tipos.OpSuma: "sumar", tipos.OpResta: "restar", tipos.OpProducto: "multiplicar",
		tipos.OpDivision: "dividir", tipos.OpResto: "calcular el resto de",
		tipos.OpMayor: "comparar", tipos.OpMenor: "comparar", tipos.OpMayorIgual: "comparar",
		tipos.OpMenorIgual: "comparar", tipos.OpIgual: "comparar con «igual»",
		tipos.OpDiferente: "comparar con «diferente»",
	}
	switch {
	case b == nil && op == tipos.OpResta:
		return fmt.Sprintf("no es posible cambiar de signo un valor de tipo %s.", a)
	case b == nil && op == tipos.OpNo:
		return fmt.Sprintf("«no» solo se aplica a un electrico, y este valor es de tipo %s.", a)
	case b == nil && op == tipos.OpTamano:
		return fmt.Sprintf("tamaño se aplica a un equipo, una mochila o un planta, y este valor es de tipo %s.", a)
	case op == tipos.OpY || op == tipos.OpO:
		return fmt.Sprintf("«%s» combina dos electrico, y aquí hay un %s y un %s.", op, a, b)
	case op == tipos.OpContiene:
		return fmt.Sprintf("no es posible buscar un valor de tipo %s dentro de uno de tipo %s.", b, a)
	case op == tipos.OpIndice:
		return fmt.Sprintf("no es posible usar [ ] sobre un valor de tipo %s con un índice de tipo %s.", a, b)
	case verbos[op] != "":
		return fmt.Sprintf("no es posible %s un valor de tipo %s con uno de tipo %s.", verbos[op], a, b)
	}
	return fmt.Sprintf("la operación «%s» no admite un %s y un %s.", op, a, b)
}

func (v *verificador) campo(x *ast.CampoAcceso) *tipos.Type {
	o := v.valor(x.Objeto, v.tipo(x.Objeto))
	if o == nil || x.Nombre == nil {
		return nil
	}
	if o.Kind != tipos.KFicha {
		v.error(x.Pos, "no-es-ficha", diag.EncabezadoTipos,
			fmt.Sprintf("solo una ficha tiene campos, y este valor es de tipo %s.", o),
			"el punto accede a un campo de una ficha.",
			"revisa el nombre del dato.")
		return nil
	}
	f := v.c.Tabla.Fichas[o.Nombre]
	if f == nil {
		return nil
	}
	c := f.Campo(x.Nombre.Nombre)
	if c == nil {
		v.error(x.Nombre.Pos, "campo-inexistente", diag.EncabezadoSinValor,
			fmt.Sprintf("la ficha %s no tiene un campo «%s».", f.Nombre, x.Nombre.Nombre),
			fmt.Sprintf("los campos de %s son: %s.", f.Nombre, camposDe(f)),
			"revisa cómo está escrito el campo.")
		return nil
	}
	return c.Tipo
}

func camposDe(f *Ficha) string {
	s := ""
	for i, c := range f.Campos {
		if i > 0 {
			s += ", "
		}
		s += c.Nombre
	}
	return s
}

// llamada revisa que el movimiento exista y sus argumentos (validación 3).
// Devuelve el movimiento, o nil si no se pudo identificar.
func (v *verificador) llamada(x *ast.Llamada) *Movimiento {
	if x == nil || x.Nombre == nil {
		return nil
	}
	s := v.c.Tabla.Buscar(v.archivo, x.Nombre.Nombre)
	if s == nil || s.Clase != ClaseMovimiento {
		for _, a := range x.Args {
			v.tipo(a)
		}
		if s == nil && v.buscarLocal(x.Nombre.Nombre) == nil {
			v.noDeclarado(x.Nombre, "movimiento")
		} else {
			v.error(x.Nombre.Pos, "no-es-movimiento", diag.EncabezadoTipos,
				fmt.Sprintf("«%s» no es un movimiento, así que no se puede llamar.", x.Nombre.Nombre),
				"los paréntesis después de un nombre llaman a un movimiento.",
				"quita los paréntesis si querías usar su valor.")
		}
		return nil
	}
	mov := s.Movimiento
	if len(x.Args) != len(mov.Params) {
		for _, a := range x.Args {
			v.tipo(a)
		}
		v.error(x.Pos, "cantidad-de-argumentos", diag.EncabezadoTipos,
			fmt.Sprintf("«%s» recibe %s y se le dieron %d.", mov.Nombre, contarParams(mov), len(x.Args)),
			fmt.Sprintf("la firma es %s.", firma(mov)),
			"pasa un argumento por cada parámetro, en el mismo orden.")
		return mov
	}
	for i, a := range x.Args {
		p := mov.Params[i]
		t := v.tipoCon(a, p.Tipo)
		v.asignable(a, t, p.Tipo, "argumento-incompatible",
			fmt.Sprintf("argumento %d de «%s» (el parámetro «%s»%s)", i+1, mov.Nombre, p.Nombre, deOtroArchivo(mov, v.archivo)))
	}
	return mov
}

func deOtroArchivo(mov *Movimiento, archivo string) string {
	if mov.Archivo != archivo {
		return ", declarado en " + mov.Archivo
	}
	return ""
}

func contarParams(mov *Movimiento) string {
	if len(mov.Params) == 1 {
		return "1 argumento"
	}
	return fmt.Sprintf("%d argumentos", len(mov.Params))
}

// firma escribe la cabecera de un movimiento: curar(roca vida, planta nombre).
func firma(mov *Movimiento) string {
	s := mov.Nombre + "("
	for i, p := range mov.Params {
		if i > 0 {
			s += ", "
		}
		s += fmt.Sprintf("%s %s", p.Tipo, p.Nombre)
	}
	return s + ")"
}

func (v *verificador) convertir(x *ast.Convertir) *tipos.Type {
	o := v.valor(x.Valor, v.tipo(x.Valor))
	d := v.resolver(x.Destino)
	if o == nil || d == nil {
		return d
	}
	if tipos.Convertible(o, d) == tipos.SinEfecto {
		v.error(x.Pos, "conversion-imposible", diag.EncabezadoTipos,
			fmt.Sprintf("no se puede convertir un %s a %s.", o, d),
			"según la tabla de efectividades esa conversión no tiene efecto, ni siquiera con convertir.",
			sugerenciaTabla)
	}
	return d
}

// tipoDestino es el tipo del lugar donde se guarda: el tipo declarado del
// dato (aunque sea posible), o el del elemento o campo.
func (v *verificador) tipoDestino(e ast.Expr) *tipos.Type {
	if id, ok := e.(*ast.Ident); ok {
		if l := v.buscarLocal(id.Nombre); l != nil {
			return l.tipo
		}
		// Asignar a una medalla o a algo que no es un dato lo revisa K3;
		// aquí solo interesa el tipo.
		return v.tipo(id)
	}
	return v.tipo(e)
}

func describirDestino(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		return "el dato «" + x.Nombre + "»"
	case *ast.CampoAcceso:
		if x.Nombre != nil {
			return "el campo «" + x.Nombre.Nombre + "»"
		}
	case *ast.Indice:
		return "un elemento"
	}
	return "el destino"
}

// esCeroLiteral reconoce 0, 0.0 y sus formas con - (validación 17).
func esCeroLiteral(e ast.Expr) bool {
	switch x := e.(type) {
	case *ast.LitRoca:
		return x.Valor == 0
	case *ast.LitAgua:
		return x.Valor == 0
	case *ast.Unaria:
		return x.Op == token.MINUS && esCeroLiteral(x.Operando)
	}
	return false
}
