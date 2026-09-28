package interprete

import (
	"strings"

	"github.com/keylorpineda/PokeScript/internal/ast"
)

// senal indica cómo terminó una instrucción. Los ciclos consumen huir y
// siguiente; la llamada a un movimiento consume entregar.
type senal int

const (
	normal senal = iota
	senalHuir
	senalSiguiente
	senalEntregar
)

// ejecutarBloque ejecuta las instrucciones en orden hasta terminar o hasta
// que una produzca una señal.
func (in *Interprete) ejecutarBloque(instrucciones []ast.Instr, env *Entorno) (senal, error) {
	for _, i := range instrucciones {
		s, err := in.ejecutar(i, env)
		if err != nil || s != normal {
			return s, err
		}
	}
	return normal, nil
}

func (in *Interprete) ejecutar(i ast.Instr, env *Entorno) (senal, error) {
	switch x := i.(type) {
	case *ast.DeclDato:
		return normal, in.declarar(x, x.Nombre, x.Tipo, x.Valor, x.Medalla, env)
	case *ast.Asignacion:
		return normal, in.ejecutarAsignacion(x, env)
	case *ast.Gritar:
		return normal, in.ejecutarGritar(x, env)
	case *ast.Capturar:
		return normal, in.ejecutarCapturar(x, env)
	case *ast.LlamadaInstr:
		_, _, err := in.llamar(x.Llamada, env) // el valor se descarta (advertencia 19)
		return normal, err
	case *ast.Sumar:
		return normal, in.ejecutarSumar(x, env)
	case *ast.Quitar:
		return normal, in.ejecutarQuitar(x, env)
	case *ast.Si:
		return in.ejecutarSi(x, env)
	case *ast.Segun:
		return in.ejecutarSegun(x, env)
	case *ast.Mientras:
		return in.ejecutarMientras(x, env)
	case *ast.RecorrerRango:
		return in.ejecutarRango(x, env)
	case *ast.RecorrerColeccion:
		return in.ejecutarRecorrido(x, env)
	case *ast.Huir:
		return senalHuir, nil
	case *ast.Siguiente:
		return senalSiguiente, nil
	case *ast.Entregar:
		return in.ejecutarEntregar(x, env)
	}
	return normal, in.interno(i, "instrucción de tipo %T sin implementar", i)
}

// ─── Datos ─────────────────────────────────────────────────────────────────

// declarar crea un dato en env. Sin valor inicial queda sin asignar hasta
// que reciba uno.
func (in *Interprete) declarar(nodo ast.Nodo, nombre *ast.Ident, tipo *ast.TipoExpr, valor ast.Expr, medalla bool, env *Entorno) error {
	v := &Variable{Tipo: tipo, Medalla: medalla}
	if valor != nil {
		val, err := in.evaluarCon(valor, tipo, env)
		if err != nil {
			return err
		}
		v.Valor, v.Asignada = Copiar(val), true
	}
	if !env.Declarar(nombre.Nombre, v) {
		return in.interno(nodo, "%s ya está declarado en este bloque", nombre.Nombre)
	}
	return nil
}

func (in *Interprete) ejecutarAsignacion(x *ast.Asignacion, env *Entorno) error {
	v, err := in.evaluarCon(x.Valor, in.tipoDe(x.Destino, env), env)
	if err != nil {
		return err
	}
	return in.asignar(x.Destino, v, env)
}

// asignar guarda una copia de v en un dato, un elemento o un campo.
func (in *Interprete) asignar(destino ast.Expr, v Value, env *Entorno) error {
	raiz := raizDe(destino)
	if raiz == nil {
		return in.interno(destino, "no se puede asignar a esta expresión")
	}
	variable := env.Buscar(raiz.Nombre)
	switch {
	case variable == nil:
		return in.interno(raiz, "%s no está declarado", raiz.Nombre)
	case variable.Medalla:
		return in.interno(destino, "%s es una medalla y no se puede modificar", raiz.Nombre)
	case variable.SoloLectura:
		return in.interno(destino, "%s es la variable de un recorrido y es de solo lectura", raiz.Nombre)
	}
	v = ajustar(v, in.tipoDe(destino, env))

	switch d := destino.(type) {
	case *ast.Ident:
		variable.Valor, variable.Asignada = Copiar(v), true
		return nil
	case *ast.Indice:
		c, err := in.evaluar(d.Coleccion, env)
		if err != nil {
			return err
		}
		i, err := in.evaluar(d.Indice, env)
		if err != nil {
			return err
		}
		switch col := c.(type) {
		case *Equipo:
			n, ok := i.(int64)
			if !ok {
				return in.interno(d, "el índice de un equipo debe ser roca")
			}
			if err := col.Poner(n, v); err != nil {
				return in.falloAcceso(d, d.Coleccion, err)
			}
			return nil
		case *Mochila:
			if !EsClave(i) {
				return in.interno(d, "%s no puede ser clave de una mochila (decisión H3)", literal(i))
			}
			col.Poner(i, v) // una clave nueva se agrega al final (decisión H2)
			return nil
		}
		return in.interno(d, "no se puede asignar con [ ] sobre %s", literal(c))
	case *ast.CampoAcceso:
		obj, err := in.evaluar(d.Objeto, env)
		if err != nil {
			return err
		}
		f, ok := obj.(*Ficha)
		if !ok || !f.TieneCampo(d.Nombre.Nombre) {
			return in.interno(d, "%s no es un campo de esa ficha", d.Nombre.Nombre)
		}
		f.PonerCampo(d.Nombre.Nombre, v)
		return nil
	}
	return in.interno(destino, "no se puede asignar a esta expresión")
}

// raizDe devuelve el dato del que cuelga un destino: e en e[1].vida.
func raizDe(e ast.Expr) *ast.Ident {
	for {
		switch x := e.(type) {
		case *ast.Ident:
			return x
		case *ast.Indice:
			e = x.Coleccion
		case *ast.CampoAcceso:
			e = x.Objeto
		default:
			return nil
		}
	}
}

// ─── Salida ────────────────────────────────────────────────────────────────

// ejecutarGritar escribe todos los valores seguidos, sin separador, y
// termina la línea (sección 5 y ejemplo de la sección 10).
func (in *Interprete) ejecutarGritar(x *ast.Gritar, env *Entorno) error {
	var b strings.Builder
	for _, a := range x.Args {
		v, err := in.evaluar(a, env)
		if err != nil {
			return err
		}
		b.WriteString(Texto(v))
	}
	in.ES.Escribir(b.String())
	return nil
}

// ─── Colecciones ───────────────────────────────────────────────────────────

func (in *Interprete) ejecutarSumar(x *ast.Sumar, env *Entorno) error {
	variable := env.Buscar(x.Coleccion.Nombre)
	if variable == nil || variable.Medalla || variable.SoloLectura {
		return in.interno(x, "no se puede sumar a %s", x.Coleccion.Nombre)
	}
	eq, ok := variable.Valor.(*Equipo)
	if !ok {
		return in.interno(x, "sumar solo se aplica a un equipo")
	}
	var elem *ast.TipoExpr
	if variable.Tipo != nil {
		elem = variable.Tipo.Elem
	}
	v, err := in.evaluarCon(x.Valor, elem, env)
	if err != nil {
		return err
	}
	eq.Sumar(v)
	return nil
}

func (in *Interprete) ejecutarQuitar(x *ast.Quitar, env *Entorno) error {
	variable := env.Buscar(x.Coleccion.Nombre)
	if variable == nil || variable.Medalla || variable.SoloLectura {
		return in.interno(x, "no se puede quitar de %s", x.Coleccion.Nombre)
	}
	i, err := in.evaluar(x.Indice, env)
	if err != nil {
		return err
	}
	switch col := variable.Valor.(type) {
	case *Equipo:
		n, ok := i.(int64)
		if !ok {
			return in.interno(x, "el índice de un equipo debe ser roca")
		}
		err = col.Quitar(n)
	case *Mochila:
		err = col.Quitar(i)
	default:
		return in.interno(x, "quitar solo se aplica a equipo o mochila")
	}
	if err != nil {
		return in.falloAcceso(x, x.Coleccion, err)
	}
	return nil
}

// ─── Control ───────────────────────────────────────────────────────────────

func (in *Interprete) ejecutarSi(x *ast.Si, env *Entorno) (senal, error) {
	for _, rama := range x.Ramas {
		ok, err := in.condicion(rama.Cond, env)
		if err != nil {
			return normal, err
		}
		if ok {
			return in.ejecutarBloque(rama.Cuerpo, NuevoEntorno(env))
		}
	}
	if x.TieneSino {
		return in.ejecutarBloque(x.Sino, NuevoEntorno(env))
	}
	return normal, nil
}

// ejecutarSegun busca la primera alternativa con un patrón igual al valor.
// El analizador garantiza que haya una o que exista otro.
func (in *Interprete) ejecutarSegun(x *ast.Segun, env *Entorno) (senal, error) {
	v, err := in.evaluar(x.Valor, env)
	if err != nil {
		return normal, err
	}
	for _, alt := range x.Alternativas {
		for _, p := range alt.Patrones {
			pv, err := in.evaluar(p, env)
			if err != nil {
				return normal, err
			}
			if Igual(v, pv) {
				return in.ejecutar(alt.Cuerpo, NuevoEntorno(env))
			}
		}
	}
	if x.Otro != nil {
		return in.ejecutar(x.Otro, NuevoEntorno(env))
	}
	return normal, nil
}

func (in *Interprete) ejecutarMientras(x *ast.Mientras, env *Entorno) (senal, error) {
	for {
		if err := in.detenido(); err != nil {
			return normal, err
		}
		ok, err := in.condicion(x.Cond, env)
		if err != nil || !ok {
			return normal, err
		}
		s, err := in.ejecutarBloque(x.Cuerpo, NuevoEntorno(env))
		if fin, s, err := finDeVuelta(s, err); fin {
			return s, err
		}
	}
}

// ejecutarRango recorre de A hasta B, ambos incluidos, con paso 1. A y B se
// evalúan una sola vez; si A > B el cuerpo no se ejecuta (sección 6).
func (in *Interprete) ejecutarRango(x *ast.RecorrerRango, env *Entorno) (senal, error) {
	a, err := in.evaluar(x.Desde, env)
	if err != nil {
		return normal, err
	}
	b, err := in.evaluar(x.Hasta, env)
	if err != nil {
		return normal, err
	}
	desde, ok1 := a.(int64)
	hasta, ok2 := b.(int64)
	if !ok1 || !ok2 {
		return normal, in.interno(x, "los extremos de recorrer deben ser roca")
	}
	for n := desde; n <= hasta; n++ {
		s, err := in.vuelta(x.Var, nil, n, nil, x.Cuerpo, env)
		if fin, s, err := finDeVuelta(s, err); fin {
			return s, err
		}
		if n == hasta {
			break // evita que n++ desborde cuando hasta es el máximo de roca
		}
	}
	return normal, nil
}

// ejecutarRecorrido visita un equipo, una mochila o un texto. Recorre una
// foto de la colección tomada al empezar.
func (in *Interprete) ejecutarRecorrido(x *ast.RecorrerColeccion, env *Entorno) (senal, error) {
	c, err := in.evaluar(x.Coleccion, env)
	if err != nil {
		return normal, err
	}
	var primeros, segundos []Value
	switch col := c.(type) {
	case *Equipo:
		primeros = col.Elementos()
	case *Mochila:
		// Una variable recibe la clave (decisión H6); dos, clave y valor.
		primeros = col.Claves()
		for _, k := range primeros {
			v, _ := col.Obtener(k)
			segundos = append(segundos, v)
		}
	case string:
		for _, r := range col {
			primeros = append(primeros, r) // un texto se recorre letra por letra (decisión H7)
		}
	default:
		return normal, in.interno(x, "recorrer no se aplica a %s", literal(c))
	}
	if x.Var2 != nil && segundos == nil {
		return normal, in.interno(x, "recorrer con dos variables solo se aplica a una mochila")
	}
	for i, v := range primeros {
		var segunda Value
		if x.Var2 != nil {
			segunda = segundos[i]
		}
		s, err := in.vuelta(x.Var, x.Var2, v, segunda, x.Cuerpo, env)
		if fin, s, err := finDeVuelta(s, err); fin {
			return s, err
		}
	}
	return normal, nil
}

// vuelta ejecuta una vuelta de recorrer con sus variables de solo lectura.
func (in *Interprete) vuelta(var1, var2 *ast.Ident, v1, v2 Value, cuerpo []ast.Instr, env *Entorno) (senal, error) {
	if err := in.detenido(); err != nil {
		return normal, err
	}
	local := NuevoEntorno(env)
	local.Declarar(var1.Nombre, &Variable{Valor: Copiar(v1), Asignada: true, SoloLectura: true})
	if var2 != nil {
		local.Declarar(var2.Nombre, &Variable{Valor: Copiar(v2), Asignada: true, SoloLectura: true})
	}
	return in.ejecutarBloque(cuerpo, local)
}

// finDeVuelta decide qué hace un ciclo después de una vuelta: huir lo
// termina, siguiente y normal continúan, entregar y los errores salen.
func finDeVuelta(s senal, err error) (bool, senal, error) {
	switch {
	case err != nil:
		return true, normal, err
	case s == senalHuir:
		return true, normal, nil
	case s == senalEntregar:
		return true, s, nil
	}
	return false, normal, nil
}

// ─── Movimientos ───────────────────────────────────────────────────────────

func (in *Interprete) ejecutarEntregar(x *ast.Entregar, env *Entorno) (senal, error) {
	mov := env.movimiento()
	if mov == nil {
		return normal, in.interno(x, "entregar solo puede usarse dentro de un movimiento")
	}
	in.entregado = nil
	if x.Valor != nil {
		v, err := in.evaluarCon(x.Valor, mov.Retorno, env)
		if err != nil {
			return normal, err
		}
		in.entregado = Copiar(v)
	}
	return senalEntregar, nil
}

// llamar ejecuta un movimiento. Los argumentos se copian (paso por valor) y
// el cuerpo corre en un alcance nuevo cuyo padre es el del archivo, no el
// de quien llama. Devuelve el valor entregado y si hubo entrega con valor.
func (in *Interprete) llamar(x *ast.Llamada, env *Entorno) (Value, bool, error) {
	mov, ok := in.movimientos[x.Nombre.Nombre]
	if !ok {
		return nil, false, in.interno(x, "el movimiento %s no existe", x.Nombre.Nombre)
	}
	if len(x.Args) != len(mov.Params) {
		return nil, false, in.interno(x, "%s recibe %d argumentos y se le dieron %d",
			x.Nombre.Nombre, len(mov.Params), len(x.Args))
	}
	if in.profundidad >= limiteLlamadasAnidadas {
		return nil, false, in.fallo(x, CodigoRecursion,
			"se superó el límite de 1000 llamadas anidadas al llamar a "+x.Nombre.Nombre+".",
			"un movimiento se llama a sí mismo, directa o indirectamente, sin llegar nunca a un caso que termine.",
			"revisa que cada llamada acerque el problema a una condición que deje de llamar.")
	}
	if err := in.detenido(); err != nil {
		return nil, false, err
	}

	archivoMov := in.origen[mov]
	local := NuevoEntorno(in.alcances[archivoMov])
	local.mov = mov
	for i, p := range mov.Params {
		v, err := in.evaluarCon(x.Args[i], p.Tipo, env)
		if err != nil {
			return nil, false, err
		}
		local.Declarar(p.Nombre.Nombre, &Variable{Valor: Copiar(v), Tipo: p.Tipo, Asignada: true})
	}

	// Mientras corre el cuerpo, los errores señalan el archivo del
	// movimiento, no el de quien llama.
	llamador := in.archivo
	in.archivo = archivoMov
	in.profundidad++
	s, err := in.ejecutarBloque(mov.Cuerpo, local)
	in.profundidad--
	in.archivo = llamador
	if err != nil {
		return nil, false, err
	}
	switch s {
	case senalEntregar:
		v := in.entregado
		in.entregado = nil
		return v, mov.Retorno != nil, nil
	case normal:
		if mov.Retorno != nil {
			return nil, false, in.interno(mov.FinPos, "%s llegó al fin sin entregar un valor", x.Nombre.Nombre)
		}
		return nil, false, nil
	}
	return nil, false, in.interno(x, "huir o siguiente salió de %s", x.Nombre.Nombre)
}
