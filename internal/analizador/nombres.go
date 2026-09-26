package analizador

import (
	"fmt"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/diag"
)

// Verificaciones de nombres y ámbitos (sección 4, pasada 2):
//
//	 7  reasignar una medalla o modificar su contenido
//	 8  ocultar un nombre que ya existe (incluye variables de recorrer anidados)
//	12  huir o siguiente fuera de un ciclo
//	13  modificar una colección mientras se la recorre
//	16  modificar la variable de un recorrido (es de solo lectura)
//
// Lleva su propio recorrido con ámbitos: además del nombre, necesita saber
// qué clase de dato es cada uno, cuántos ciclos hay abiertos y qué
// colecciones se están recorriendo.
func init() { registrar("nombres", revisarNombres) }

func revisarNombres(c *Contexto) {
	for _, a := range c.Archivos() {
		n := &nombrador{c: c, archivo: a.Nombre}
		for _, d := range a.Programa.Declaraciones {
			n.declaracion(d)
		}
	}
}

// claseLocal dice qué es un dato declarado dentro de un cuerpo.
type claseLocal int

const (
	localDato claseLocal = iota
	localMedalla
	localParametro
	localRecorrido
)

type localNombrado struct {
	clase claseLocal
	linea int
}

type nombrador struct {
	c       *Contexto
	archivo string
	ambitos []map[string]localNombrado
	ciclos  int      // ciclos abiertos en el cuerpo actual
	recorre []string // colecciones que se están recorriendo, de afuera hacia adentro
}

func (n *nombrador) abrir()  { n.ambitos = append(n.ambitos, map[string]localNombrado{}) }
func (n *nombrador) cerrar() { n.ambitos = n.ambitos[:len(n.ambitos)-1] }

func (n *nombrador) buscar(nombre string) (localNombrado, bool) {
	for i := len(n.ambitos) - 1; i >= 0; i-- {
		if l, ok := n.ambitos[i][nombre]; ok {
			return l, true
		}
	}
	return localNombrado{}, false
}

func (n *nombrador) error(pos ast.Pos, codigo, desc, causa, sugerencia string) {
	n.c.Error(n.archivo, pos, codigo, diag.EncabezadoTipos, desc, causa, sugerencia)
}

// ─── Declaraciones ─────────────────────────────────────────────────────────

func (n *nombrador) declaracion(d ast.Decl) {
	switch x := d.(type) {
	case *ast.DeclMovimiento:
		n.abrir()
		for _, p := range x.Params {
			// Los parámetros repetidos ya los reporta la pasada 1; aquí solo
			// se revisa que no oculten un nombre del archivo.
			if p.Nombre != nil && n.ocultaGlobal(p.Nombre) {
				continue
			}
			if p.Nombre != nil {
				n.ambitos[0][p.Nombre.Nombre] = localNombrado{clase: localParametro, linea: p.Nombre.Line}
			}
		}
		n.cuerpo(x.Cuerpo)
		n.cerrar()
	case *ast.Combate:
		n.cuerpo(x.Cuerpo)
	}
}

// cuerpo recorre el cuerpo de un movimiento o del combate: los ciclos de
// afuera no cuentan adentro, porque un movimiento no sabe desde dónde lo
// llaman.
func (n *nombrador) cuerpo(is []ast.Instr) {
	ciclos, recorre := n.ciclos, n.recorre
	n.ciclos, n.recorre = 0, nil
	n.bloque(is)
	n.ciclos, n.recorre = ciclos, recorre
}

// declarar registra un dato local, salvo que oculte otro nombre (8).
func (n *nombrador) declarar(id *ast.Ident, clase claseLocal) {
	if id == nil {
		return
	}
	if previo, existe := n.buscar(id.Nombre); existe {
		n.error(id.Pos, "nombre-ocultado",
			fmt.Sprintf("«%s» ya existe: es %s declarado en la línea %d.", id.Nombre, describirLocal(previo.clase), previo.linea),
			"un bloque de adentro no puede declarar un nombre que ya existe afuera, ni siquiera en otro recorrido: quedaría escondido y no se sabría a cuál se refiere cada uso.",
			fmt.Sprintf("usa otro nombre, por ejemplo «%s2» o uno que diga qué guarda.", id.Nombre))
		return
	}
	if n.ocultaGlobal(id) {
		return
	}
	n.ambitos[len(n.ambitos)-1][id.Nombre] = localNombrado{clase: clase, linea: id.Line}
}

// ocultaGlobal reporta si id repite un nombre de alcance de archivo visible.
func (n *nombrador) ocultaGlobal(id *ast.Ident) bool {
	s := n.c.Tabla.Buscar(n.archivo, id.Nombre)
	if s == nil {
		return false
	}
	n.error(id.Pos, "nombre-ocultado",
		fmt.Sprintf("«%s» ya es el nombre de %s.", id.Nombre, describirSimbolo(s)),
		"un dato no puede llamarse igual que algo declarado fuera de los bloques: quedaría escondido y no se sabría a cuál se refiere cada uso.",
		"usa otro nombre para el dato.")
	return true
}

func describirLocal(c claseLocal) string {
	switch c {
	case localMedalla:
		return "una medalla"
	case localParametro:
		return "un parámetro"
	case localRecorrido:
		return "la variable de un recorrido"
	}
	return "un dato"
}

func describirSimbolo(s *Simbolo) string {
	switch s.Clase {
	case ClaseEspecie:
		return "la especie «" + s.Especie.Nombre + "»"
	case ClaseValorEspecie:
		return "un valor de la especie «" + s.Especie.Nombre + "»"
	case ClaseFicha:
		return "la ficha «" + s.Ficha.Nombre + "»"
	case ClaseMedalla:
		return "una medalla del archivo"
	}
	return "un movimiento"
}

// ─── Instrucciones ─────────────────────────────────────────────────────────

func (n *nombrador) bloque(is []ast.Instr) {
	n.abrir()
	for _, i := range is {
		n.instruccion(i)
	}
	n.cerrar()
}

func (n *nombrador) instruccion(i ast.Instr) {
	switch x := i.(type) {
	case *ast.DeclDato:
		clase := localDato
		if x.Medalla {
			clase = localMedalla
		}
		n.declarar(x.Nombre, clase)
	case *ast.Asignacion:
		n.modifica(x.Destino, x.Pos, esNombre(x.Destino))
	case *ast.Capturar:
		n.modifica(x.Destino, x.Pos, esNombre(x.Destino))
	case *ast.Sumar:
		n.modifica(x.Coleccion, x.Pos, false)
	case *ast.Quitar:
		n.modifica(x.Coleccion, x.Pos, false)
	case *ast.Huir, *ast.Siguiente:
		if n.ciclos == 0 {
			palabra := "huir"
			if _, ok := x.(*ast.Siguiente); ok {
				palabra = "siguiente"
			}
			n.error(x.Posicion(), "corte-fuera-de-ciclo",
				fmt.Sprintf("«%s» solo se puede usar dentro de un ciclo.", palabra),
				"huir sale del ciclo más cercano y siguiente pasa a su próxima vuelta; aquí no hay ningún mientras ni recorrer abierto.",
				"si querías terminar un movimiento, usa «entregar»; si no, mueve esta línea dentro del ciclo.")
		}
	case *ast.Si:
		for _, r := range x.Ramas {
			n.bloque(r.Cuerpo)
		}
		n.bloque(x.Sino)
	case *ast.Segun:
		for _, alt := range x.Alternativas {
			n.bloque([]ast.Instr{alt.Cuerpo})
		}
		if x.Otro != nil {
			n.bloque([]ast.Instr{x.Otro})
		}
	case *ast.Mientras:
		n.ciclos++
		n.bloque(x.Cuerpo)
		n.ciclos--
	case *ast.RecorrerRango:
		n.enCiclo(x.Cuerpo, "", x.Var)
	case *ast.RecorrerColeccion:
		n.enCiclo(x.Cuerpo, raiz(x.Coleccion), x.Var, x.Var2)
	}
}

// enCiclo recorre el cuerpo de un recorrer con sus variables declaradas y,
// si recorre una colección con nombre, la marca como intocable.
func (n *nombrador) enCiclo(cuerpo []ast.Instr, coleccion string, vars ...*ast.Ident) {
	n.abrir()
	for _, v := range vars {
		n.declarar(v, localRecorrido)
	}
	n.ciclos++
	if coleccion != "" {
		n.recorre = append(n.recorre, coleccion)
	}
	n.bloque(cuerpo)
	if coleccion != "" {
		n.recorre = n.recorre[:len(n.recorre)-1]
	}
	n.ciclos--
	n.cerrar()
}

// modifica revisa una instrucción que cambia el dato destino (asignación,
// capturar, sumar o quitar): 7, 13 y 16. entero dice si la instrucción
// reemplaza el valor completo (x = …) o solo parte de él (x[1] = …, sumar).
func (n *nombrador) modifica(destino ast.Expr, pos ast.Pos, entero bool) {
	nombre := raiz(destino)
	if nombre == "" {
		return
	}

	if l, local := n.buscar(nombre); local {
		switch l.clase {
		case localMedalla:
			n.medallaModificada(nombre, pos, entero, l.linea)
			return
		case localRecorrido:
			n.error(pos, "variable-de-recorrido",
				fmt.Sprintf("«%s» es la variable del recorrido: es de solo lectura.", nombre),
				"en cada vuelta, recorrer le da el valor del elemento que toca; cambiarlo no cambiaría la colección y confundiría el recorrido.",
				fmt.Sprintf("si necesitas un valor distinto, copia «%s» a otro dato y cambia ese.", nombre))
			return
		}
	} else if s := n.c.Tabla.Buscar(n.archivo, nombre); s != nil && s.Clase == ClaseMedalla {
		n.medallaModificada(nombre, pos, entero, 0)
		return
	}

	for _, c := range n.recorre {
		if c == nombre {
			n.error(pos, "coleccion-en-recorrido",
				fmt.Sprintf("«%s» se está recorriendo en este momento y no se puede modificar.", nombre),
				"si la colección cambia mientras se la recorre, no se sabe qué elementos quedan por visitar.",
				"guarda los cambios en otra colección y aplica esos cambios cuando termine el recorrido.")
			return
		}
	}
}

func (n *nombrador) medallaModificada(nombre string, pos ast.Pos, entero bool, linea int) {
	desc := fmt.Sprintf("«%s» es una medalla y su contenido no se puede modificar.", nombre)
	codigo := "medalla-modificada"
	if entero {
		desc = fmt.Sprintf("«%s» es una medalla y no se le puede asignar otro valor.", nombre)
		codigo = "medalla-reasignada"
	}
	causa := "una medalla guarda un valor que no cambia durante todo el programa."
	if linea > 0 {
		causa = fmt.Sprintf("«%s» se declaró como medalla en la línea %d: su valor no cambia.", nombre, linea)
	}
	n.error(pos, codigo, desc, causa,
		"si el valor tiene que cambiar, declara un dato común en lugar de una medalla.")
}

func esNombre(e ast.Expr) bool {
	_, ok := e.(*ast.Ident)
	return ok
}

// raiz devuelve el nombre del dato que un destino modifica: e[1].x → e.
func raiz(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		if x == nil {
			return ""
		}
		return x.Nombre
	case *ast.Indice:
		return raiz(x.Coleccion)
	case *ast.CampoAcceso:
		return raiz(x.Objeto)
	}
	return ""
}
