package analizador

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/proyecto"
)

// Advertencias 20 a 23 de la sección 4. No impiden ejecutar el programa:
//
//	20  dato declarado y nunca utilizado
//	21  ciclo cuya condición no depende de nada modificado en su cuerpo
//	22  sangría inconsistente con la estructura de bloques
//	23  convención de nombres (mayúsculas y minúsculas)
func init() { registrar("advertencias", revisarAdvertencias) }

func revisarAdvertencias(c *Contexto) {
	for _, a := range c.Archivos() {
		for _, d := range a.Programa.Declaraciones {
			revisarSinUsar(c, a.Nombre, d)
			revisarCiclos(c, a.Nombre, d)
			revisarSangria(c, a, d)
		}
		revisarConvenciones(c, a)
	}
}

// ─── 20. Datos sin usar ────────────────────────────────────────────────────

type datoDeclarado struct {
	id    *ast.Ident
	usado bool
}

// usos lleva los datos locales por ámbito y marca cuáles se leen.
type usos struct {
	ambitos []map[string]*datoDeclarado
	todos   []*datoDeclarado
}

func (u *usos) abrir()  { u.ambitos = append(u.ambitos, map[string]*datoDeclarado{}) }
func (u *usos) cerrar() { u.ambitos = u.ambitos[:len(u.ambitos)-1] }

func (u *usos) declarar(id *ast.Ident) {
	if id == nil {
		return
	}
	d := &datoDeclarado{id: id}
	u.ambitos[len(u.ambitos)-1][id.Nombre] = d
	u.todos = append(u.todos, d)
}

func (u *usos) usar(nombre string) {
	for i := len(u.ambitos) - 1; i >= 0; i-- {
		if d, ok := u.ambitos[i][nombre]; ok {
			d.usado = true
			return
		}
	}
}

// leer marca como usados todos los nombres de una expresión.
func (u *usos) leer(e ast.Nodo) {
	if e == nil {
		return
	}
	ast.Inspeccionar(e, func(n ast.Nodo) bool {
		switch x := n.(type) {
		case *ast.Ident:
			u.usar(x.Nombre)
		case *ast.Llamada:
			// El nombre del movimiento no es un dato; sí sus argumentos.
			for _, a := range x.Args {
				u.leer(a)
			}
			return false
		case *ast.CampoAcceso:
			// El nombre del campo tampoco es un dato.
			u.leer(x.Objeto)
			return false
		}
		return true
	})
}

// destino marca lo que se lee al escribir en un destino: en x = … no se
// lee x; en x[i] = … o x.campo = … se usa x (se cambia parte de su valor)
// y se leen las expresiones de los índices.
func (u *usos) destino(e ast.Expr) {
	switch x := e.(type) {
	case *ast.Indice:
		u.usar(raiz(x.Coleccion))
		u.leer(x.Coleccion)
		u.leer(x.Indice)
	case *ast.CampoAcceso:
		u.usar(raiz(x.Objeto))
		u.leer(x.Objeto)
	}
}

func revisarSinUsar(c *Contexto, archivo string, d ast.Decl) {
	u := &usos{}
	u.abrir()
	switch x := d.(type) {
	case *ast.DeclMovimiento:
		u.bloque(x.Cuerpo)
	case *ast.Combate:
		u.bloque(x.Cuerpo)
	default:
		return
	}
	for _, dato := range u.todos {
		if !dato.usado {
			c.Advertencia(archivo, dato.id.Pos, "dato-sin-usar",
				fmt.Sprintf("«%s» se declara pero nunca se usa.", dato.id.Nombre),
				"ninguna instrucción lee su valor, así que no cambia lo que hace el programa.",
				fmt.Sprintf("usa «%s» o borra su declaración.", dato.id.Nombre))
		}
	}
}

func (u *usos) bloque(is []ast.Instr) {
	u.abrir()
	for _, i := range is {
		u.instruccion(i)
	}
	u.cerrar()
}

func (u *usos) instruccion(i ast.Instr) {
	switch x := i.(type) {
	case *ast.DeclDato:
		u.leer(x.Valor)
		u.declarar(x.Nombre)
	case *ast.Asignacion:
		u.destino(x.Destino)
		u.leer(x.Valor)
	case *ast.Capturar:
		u.destino(x.Destino)
		u.leer(x.Mensaje)
	case *ast.Sumar:
		u.leer(x.Valor)
		u.usar(x.Coleccion.Nombre)
	case *ast.Quitar:
		u.usar(x.Coleccion.Nombre)
		u.leer(x.Indice)
	case *ast.Gritar:
		for _, a := range x.Args {
			u.leer(a)
		}
	case *ast.LlamadaInstr:
		u.leer(x.Llamada)
	case *ast.Entregar:
		u.leer(x.Valor)
	case *ast.Si:
		for _, r := range x.Ramas {
			u.leer(r.Cond)
			u.bloque(r.Cuerpo)
		}
		u.bloque(x.Sino)
	case *ast.Segun:
		u.leer(x.Valor)
		for _, alt := range x.Alternativas {
			u.bloque([]ast.Instr{alt.Cuerpo})
		}
		if x.Otro != nil {
			u.bloque([]ast.Instr{x.Otro})
		}
	case *ast.Mientras:
		u.leer(x.Cond)
		u.bloque(x.Cuerpo)
	case *ast.RecorrerRango:
		u.leer(x.Desde)
		u.leer(x.Hasta)
		u.bloque(x.Cuerpo)
	case *ast.RecorrerColeccion:
		u.leer(x.Coleccion)
		u.bloque(x.Cuerpo)
	}
}

// ─── 21. Ciclos que no cambian ─────────────────────────────────────────────

func revisarCiclos(c *Contexto, archivo string, d ast.Decl) {
	ast.Inspeccionar(d, func(n ast.Nodo) bool {
		m, ok := n.(*ast.Mientras)
		if !ok || m.Cond == nil {
			return true
		}
		leidos, variable := nombresDe(m.Cond)
		if variable || puedeSalir(m.Cuerpo) {
			return true
		}
		modificados := modificadosEn(m.Cuerpo)
		for _, l := range leidos {
			if modificados[l] {
				return true
			}
		}
		c.Advertencia(archivo, m.Pos, "ciclo-sin-cambio",
			"la condición de este mientras no cambia dentro del ciclo.",
			"si la condición es verdadera al entrar, nunca deja de serlo: el ciclo no termina.",
			"cambia dentro del ciclo algún dato de la condición, o sal con «huir».")
		return true
	})
}

// nombresDe devuelve los nombres que lee una condición. variable es true si
// la condición puede cambiar sola, porque llama a un movimiento o a
// aleatorio.
func nombresDe(e ast.Expr) (nombres []string, variable bool) {
	ast.Inspeccionar(e, func(n ast.Nodo) bool {
		switch x := n.(type) {
		case *ast.Ident:
			nombres = append(nombres, x.Nombre)
		case *ast.Llamada, *ast.Aleatorio:
			variable = true
		case *ast.CampoAcceso:
			nombres = append(nombres, raiz(x.Objeto))
			return false
		}
		return true
	})
	return nombres, variable
}

// puedeSalir informa si el cuerpo tiene un huir o un entregar que saque de
// este ciclo (los huir de un ciclo interno no cuentan).
func puedeSalir(cuerpo []ast.Instr) bool {
	sale := false
	for _, i := range cuerpo {
		ast.Inspeccionar(i, func(n ast.Nodo) bool {
			switch n.(type) {
			case *ast.Mientras, *ast.RecorrerRango, *ast.RecorrerColeccion:
				// Un huir adentro sale de ese ciclo, pero un entregar sí
				// sale de todo.
				ast.Inspeccionar(n, func(m ast.Nodo) bool {
					if _, ok := m.(*ast.Entregar); ok {
						sale = true
					}
					return true
				})
				return false
			case *ast.Huir, *ast.Entregar:
				sale = true
			}
			return true
		})
	}
	return sale
}

// modificadosEn junta los datos que el cuerpo cambia.
func modificadosEn(cuerpo []ast.Instr) map[string]bool {
	m := map[string]bool{}
	for _, i := range cuerpo {
		ast.Inspeccionar(i, func(n ast.Nodo) bool {
			switch x := n.(type) {
			case *ast.Asignacion:
				m[raiz(x.Destino)] = true
			case *ast.Capturar:
				m[raiz(x.Destino)] = true
			case *ast.Sumar:
				m[x.Coleccion.Nombre] = true
			case *ast.Quitar:
				m[x.Coleccion.Nombre] = true
			}
			return true
		})
	}
	return m
}

// ─── 22. Sangría ───────────────────────────────────────────────────────────

// revisarSangria revisa que las líneas de cada bloque estén más adentro que
// la línea que lo abre y alineadas entre sí. Reporta una sola línea por
// bloque, la primera que no cuadra, para no llenar el panel de avisos.
func revisarSangria(c *Contexto, a *proyecto.Archivo, d ast.Decl) {
	col := func(linea int) int { return a.Sangrias[linea] }
	revisar := func(apertura int, lineas []int) {
		if len(lineas) == 0 || col(apertura) == 0 {
			return
		}
		primera := col(lineas[0])
		for _, l := range lineas {
			actual := col(l)
			if actual == 0 {
				continue
			}
			if actual <= col(apertura) || actual != primera {
				desc := "esta línea está a la misma altura que la que abre su bloque, o más afuera."
				if actual > col(apertura) {
					desc = "esta línea no está alineada con las demás de su bloque."
				}
				c.Advertencia(a.Nombre, ast.Pos{Line: l, Col: actual, Len: 1}, "sangria-inconsistente", desc,
					"la sangría no cambia lo que hace el programa, pero ayuda a ver qué fin cierra cada bloque; si no cuadra, confunde.",
					fmt.Sprintf("mueve la línea para que quede más adentro que la línea %d, igual que las demás de su bloque.", apertura))
				return
			}
		}
	}
	lineasDe := func(is []ast.Instr) []int {
		var ls []int
		for _, i := range is {
			if i != nil {
				ls = append(ls, i.Posicion().Line)
			}
		}
		return ls
	}

	ast.Inspeccionar(d, func(n ast.Nodo) bool {
		switch x := n.(type) {
		case *ast.Combate:
			revisar(x.Line, lineasDe(x.Cuerpo))
		case *ast.DeclMovimiento:
			revisar(x.Line, lineasDe(x.Cuerpo))
		case *ast.DeclFicha:
			var ls []int
			for _, campo := range x.Campos {
				ls = append(ls, campo.Line)
			}
			revisar(x.Line, ls)
		case *ast.RamaSi:
			revisar(x.Line, lineasDe(x.Cuerpo))
		case *ast.Si:
			if x.TieneSino {
				revisar(x.SinoPos.Line, lineasDe(x.Sino))
			}
		case *ast.Mientras:
			revisar(x.Line, lineasDe(x.Cuerpo))
		case *ast.RecorrerRango:
			revisar(x.Line, lineasDe(x.Cuerpo))
		case *ast.RecorrerColeccion:
			revisar(x.Line, lineasDe(x.Cuerpo))
		case *ast.Segun:
			var ls []int
			for _, alt := range x.Alternativas {
				ls = append(ls, alt.Line)
			}
			if x.Otro != nil {
				ls = append(ls, x.OtroPos.Line)
			}
			revisar(x.Line, ls)
		}
		return true
	})
}

// ─── 23. Convención de nombres ─────────────────────────────────────────────

// revisarConvenciones aplica la convención de estilo de la sección 1.3:
// especies y fichas con mayúscula inicial, valores de especie y medallas
// en mayúsculas, y datos en minúscula.
func revisarConvenciones(c *Contexto, a *proyecto.Archivo) {
	avisar := func(id *ast.Ident, que, regla, sugerido string) {
		if id == nil || sugerido == id.Nombre {
			return
		}
		c.Advertencia(a.Nombre, id.Pos, "convencion-de-nombres",
			fmt.Sprintf("«%s» es %s y no sigue la convención de nombres.", id.Nombre, que),
			regla,
			fmt.Sprintf("puedes llamarlo «%s».", sugerido))
	}
	const (
		reglaTipo  = "por convención, los nombres de especies y fichas empiezan con mayúscula."
		reglaConst = "por convención, las medallas y los valores de especie se escriben en MAYÚSCULAS."
		reglaDato  = "por convención, los datos empiezan con minúscula; las mayúsculas quedan para tipos y medallas."
	)
	ast.InspeccionarPrograma(a.Programa, func(n ast.Nodo) bool {
		switch x := n.(type) {
		case *ast.DeclEspecie:
			avisar(x.Nombre, "una especie", reglaTipo, conMayuscula(nombre(x.Nombre)))
			for _, v := range x.Valores {
				avisar(v, "un valor de especie", reglaConst, strings.ToUpper(v.Nombre))
			}
		case *ast.DeclFicha:
			avisar(x.Nombre, "una ficha", reglaTipo, conMayuscula(nombre(x.Nombre)))
		case *ast.DeclMedalla:
			avisar(x.Nombre, "una medalla", reglaConst, strings.ToUpper(nombre(x.Nombre)))
		case *ast.DeclDato:
			if x.Medalla {
				avisar(x.Nombre, "una medalla", reglaConst, strings.ToUpper(nombre(x.Nombre)))
			} else {
				avisar(x.Nombre, "un dato", reglaDato, conMinuscula(nombre(x.Nombre)))
			}
		case *ast.Param:
			avisar(x.Nombre, "un parámetro", reglaDato, conMinuscula(nombre(x.Nombre)))
		case *ast.RecorrerRango:
			avisar(x.Var, "la variable de un recorrido", reglaDato, conMinuscula(nombre(x.Var)))
		case *ast.RecorrerColeccion:
			avisar(x.Var, "la variable de un recorrido", reglaDato, conMinuscula(nombre(x.Var)))
			avisar(x.Var2, "la variable de un recorrido", reglaDato, conMinuscula(nombre(x.Var2)))
		}
		return true
	})
}

func conMayuscula(s string) string {
	r := []rune(s)
	if len(r) > 0 {
		r[0] = unicode.ToUpper(r[0])
	}
	return string(r)
}

// conMinuscula baja la primera letra; un dato escrito todo en mayúsculas
// (PODER) se pasa entero a minúsculas.
func conMinuscula(s string) string {
	if s == strings.ToUpper(s) {
		return strings.ToLower(s)
	}
	r := []rune(s)
	if len(r) > 0 {
		r[0] = unicode.ToLower(r[0])
	}
	return string(r)
}
