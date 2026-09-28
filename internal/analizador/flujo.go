package analizador

import (
	"fmt"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/diag"
)

// J3 · Flujo (sección 4, pasada 2, y sección 4.2):
//
//	4 un movimiento con tipo entrega un valor por todos los caminos
//	5 entregar con valor en un movimiento sin tipo, o sin valor en uno con
//	  tipo; entregar fuera de un movimiento (decisión H12)
//	6 asignación definida: ningún camino lee un dato antes de darle valor
//
// Recorre cada cuerpo siguiendo los caminos posibles:
//
//   - tras un si con sino, un dato tiene valor solo si todas las ramas se
//     lo dan; sin sino, la rama que no se ejecuta cuenta como un camino más;
//   - un segun sin otro que no es exhaustivo también tiene ese camino extra;
//   - el cuerpo de un mientras o de un recorrer puede no ejecutarse, así que
//     lo que asigna no cuenta después del ciclo;
//   - después de entregar, huir o siguiente el camino termina y no cuenta
//     al unir las ramas.
func init() { registrar("flujo", revisarFlujo) }

func revisarFlujo(c *Contexto) {
	for _, a := range c.Archivos() {
		for _, d := range a.Programa.Declaraciones {
			switch x := d.(type) {
			case *ast.DeclMovimiento:
				f := &flujo{c: c, archivo: a.Nombre, mov: x}
				f.abrir()
				for _, p := range x.Params {
					f.declarar(nombre(p.Nombre), true)
				}
				fin := f.bloque(x.Cuerpo, caminoNuevo())
				if x.Retorno != nil && !fin.termino {
					f.faltaEntregar(x)
				}
			case *ast.Combate:
				f := &flujo{c: c, archivo: a.Nombre}
				f.abrir()
				f.bloque(x.Cuerpo, caminoNuevo())
			}
		}
	}
}

// camino es lo que se sabe en un punto del cuerpo.
type camino struct {
	// conValor son los datos declarados sin valor que ya recibieron uno en
	// todos los caminos que llegan aquí.
	conValor map[*datoFlujo]bool
	// termino indica que ningún camino llega aquí: todos pasaron por
	// entregar, huir o siguiente.
	termino bool
}

func caminoNuevo() camino { return camino{conValor: map[*datoFlujo]bool{}} }

func (c camino) copia() camino {
	n := camino{conValor: make(map[*datoFlujo]bool, len(c.conValor)), termino: c.termino}
	for d := range c.conValor {
		n.conValor[d] = true
	}
	return n
}

// unir junta los caminos que salen de las ramas de un si o un segun. Los
// que ya terminaron no cuentan; de los demás, un dato tiene valor solo si
// lo tiene en todos.
func unir(caminos ...camino) camino {
	var vivos []camino
	for _, c := range caminos {
		if !c.termino {
			vivos = append(vivos, c)
		}
	}
	if len(vivos) == 0 {
		return camino{conValor: map[*datoFlujo]bool{}, termino: true}
	}
	r := vivos[0].copia()
	for _, c := range vivos[1:] {
		for d := range r.conValor {
			if !c.conValor[d] {
				delete(r.conValor, d)
			}
		}
	}
	return r
}

// datoFlujo es un dato local. Solo interesan los declarados sin valor.
type datoFlujo struct {
	nombre      string
	sinValor    bool
	reportado   bool
	declaradoEn ast.Pos
}

type flujo struct {
	c       *Contexto
	archivo string
	mov     *ast.DeclMovimiento // nil en combate
	alcance []map[string]*datoFlujo
}

func (f *flujo) abrir()  { f.alcance = append(f.alcance, map[string]*datoFlujo{}) }
func (f *flujo) cerrar() { f.alcance = f.alcance[:len(f.alcance)-1] }

func (f *flujo) declarar(n string, conValor bool) *datoFlujo {
	d := &datoFlujo{nombre: n, sinValor: !conValor}
	f.alcance[len(f.alcance)-1][n] = d
	return d
}

func (f *flujo) buscar(n string) *datoFlujo {
	for i := len(f.alcance) - 1; i >= 0; i-- {
		if d, ok := f.alcance[i][n]; ok {
			return d
		}
	}
	return nil
}

// bloque recorre las instrucciones en un alcance nuevo. Lo que viene
// después de un entregar, huir o siguiente no se alcanza y no se revisa.
func (f *flujo) bloque(is []ast.Instr, c camino) camino {
	f.abrir()
	defer f.cerrar()
	for _, i := range is {
		if c.termino {
			break
		}
		c = f.instruccion(i, c)
	}
	return c
}

func (f *flujo) instruccion(i ast.Instr, c camino) camino { //nolint:gocyclo // un caso por instrucción
	switch x := i.(type) {
	case *ast.DeclDato:
		f.leer(x.Valor, c)
		d := f.declarar(nombre(x.Nombre), x.Valor != nil)
		if x.Nombre != nil {
			d.declaradoEn = x.Nombre.Pos
		}
	case *ast.Asignacion:
		f.leer(x.Valor, c)
		f.asignar(x.Destino, c)
	case *ast.Capturar:
		f.leer(x.Mensaje, c)
		f.asignar(x.Destino, c)
	case *ast.Gritar:
		for _, a := range x.Args {
			f.leer(a, c)
		}
	case *ast.LlamadaInstr:
		f.leer(x.Llamada, c)
	case *ast.Sumar:
		f.leer(x.Valor, c)
		f.leer(x.Coleccion, c)
	case *ast.Quitar:
		f.leer(x.Coleccion, c)
		f.leer(x.Indice, c)
	case *ast.Si:
		var salidas []camino
		for _, r := range x.Ramas {
			f.leer(r.Cond, c)
			salidas = append(salidas, f.bloque(r.Cuerpo, c.copia()))
		}
		if x.TieneSino {
			salidas = append(salidas, f.bloque(x.Sino, c.copia()))
		} else {
			salidas = append(salidas, c) // ninguna condición se cumplió
		}
		return unir(salidas...)
	case *ast.Segun:
		f.leer(x.Valor, c)
		var salidas []camino
		for _, alt := range x.Alternativas {
			for _, p := range alt.Patrones {
				f.leer(p, c)
			}
			salidas = append(salidas, f.bloque([]ast.Instr{alt.Cuerpo}, c.copia()))
		}
		switch {
		case x.Otro != nil:
			salidas = append(salidas, f.bloque([]ast.Instr{x.Otro}, c.copia()))
		case !f.segunExhaustivo(x):
			salidas = append(salidas, c) // ninguna rama coincidió
		}
		return unir(salidas...)
	case *ast.Mientras:
		f.leer(x.Cond, c)
		f.bloque(x.Cuerpo, c.copia())
	case *ast.RecorrerRango:
		f.leer(x.Desde, c)
		f.leer(x.Hasta, c)
		f.cicloCon(x.Cuerpo, c, x.Var, nil)
	case *ast.RecorrerColeccion:
		f.leer(x.Coleccion, c)
		f.cicloCon(x.Cuerpo, c, x.Var, x.Var2)
	case *ast.Huir, *ast.Siguiente:
		c.termino = true
	case *ast.Entregar:
		f.leer(x.Valor, c)
		f.revisarEntregar(x)
		c.termino = true
	}
	return c
}

// cicloCon recorre el cuerpo de un recorrer con sus variables, que siempre
// tienen valor. Lo que el cuerpo asigna no cuenta después del ciclo.
func (f *flujo) cicloCon(cuerpo []ast.Instr, c camino, v1, v2 *ast.Ident) {
	f.abrir()
	f.declarar(nombre(v1), true)
	if v2 != nil {
		f.declarar(v2.Nombre, true)
	}
	f.bloque(cuerpo, c.copia())
	f.cerrar()
}

// asignar marca que un dato recibe valor. Asignar a un elemento o a un
// campo no le da valor al dato: lo lee.
func (f *flujo) asignar(destino ast.Expr, c camino) {
	if id, ok := destino.(*ast.Ident); ok {
		if d := f.buscar(id.Nombre); d != nil && d.sinValor {
			c.conValor[d] = true
		}
		return
	}
	f.leer(destino, c)
}

// leer revisa que cada dato que aparece en e ya tenga valor en este camino
// (validación 6). No cuentan como lecturas el nombre de un campo en p.vida,
// el de un movimiento en curar(…) ni las claves de un literal { }, que
// pueden ser nombres de campo.
func (f *flujo) leer(e ast.Nodo, c camino) {
	ast.Inspeccionar(e, func(n ast.Nodo) bool {
		switch x := n.(type) {
		case *ast.Ident:
			f.revisarLectura(x, c)
		case *ast.CampoAcceso:
			f.leer(x.Objeto, c)
			return false
		case *ast.Llamada:
			for _, a := range x.Args {
				f.leer(a, c)
			}
			return false
		case *ast.LitLlaves:
			for _, p := range x.Pares {
				if _, esNombre := p.Clave.(*ast.Ident); !esNombre {
					f.leer(p.Clave, c)
				}
				f.leer(p.Valor, c)
			}
			return false
		}
		return true
	})
}

func (f *flujo) revisarLectura(x *ast.Ident, c camino) {
	d := f.buscar(x.Nombre)
	if d == nil || !d.sinValor || c.conValor[d] || d.reportado {
		return
	}
	d.reportado = true // un aviso por dato basta
	f.c.Error(f.archivo, x.Pos, "dato-sin-valor", diag.EncabezadoSinValor,
		fmt.Sprintf("«%s» se lee antes de tener un valor.", x.Nombre),
		fmt.Sprintf("«%s» se declaró sin valor en la línea %d, y hay un camino que llega aquí sin asignárselo: "+
			"un si sin sino, un segun sin todas sus ramas o un ciclo pueden saltarse la asignación.", x.Nombre, d.declaradoEn.Line),
		"dale un valor al declararlo, o asigna uno en todos los caminos antes de usarlo.")
}

// segunExhaustivo dice si las ramas cubren todos los valores posibles: los
// de una especie o verdadero y falso. Se deduce de los patrones, sin
// conocer el tipo del valor.
func (f *flujo) segunExhaustivo(s *ast.Segun) bool {
	cubiertos := map[string]bool{}
	var especie *Especie
	hayVerdadero, hayFalso := false, false
	for _, alt := range s.Alternativas {
		for _, p := range alt.Patrones {
			switch x := p.(type) {
			case *ast.LitElectrico:
				hayVerdadero = hayVerdadero || x.Valor
				hayFalso = hayFalso || !x.Valor
			case *ast.Ident:
				if e := f.c.Tabla.Valores[x.Nombre]; e != nil {
					especie = e
					cubiertos[x.Nombre] = true
				}
			}
		}
	}
	if hayVerdadero && hayFalso {
		return true
	}
	if especie == nil {
		return false
	}
	for _, v := range especie.Valores {
		if !cubiertos[v] {
			return false
		}
	}
	return true
}

// revisarEntregar aplica la validación 5 y la decisión H12.
func (f *flujo) revisarEntregar(x *ast.Entregar) {
	switch {
	case f.mov == nil:
		f.c.Error(f.archivo, x.Pos, "entregar-fuera-de-movimiento", diag.EncabezadoTipos,
			"entregar solo existe dentro de un movimiento.",
			"combate es el programa principal: no tiene a quién entregarle un valor.",
			"para terminar un ciclo antes usa huir.")
	case f.mov.Retorno == nil && x.Valor != nil:
		f.c.Error(f.archivo, x.Valor.Posicion(), "entregar-con-valor", diag.EncabezadoTipos,
			fmt.Sprintf("«%s» no tiene tipo, así que su entregar no puede llevar un valor.", nombre(f.mov.Nombre)),
			"solo un movimiento declarado con tipo, como «movimiento roca …», entrega un valor.",
			fmt.Sprintf("quita el valor, o declara el tipo: movimiento <tipo> %s(…).", nombre(f.mov.Nombre)))
	case f.mov.Retorno != nil && x.Valor == nil:
		f.c.Error(f.archivo, x.Pos, "entregar-sin-valor", diag.EncabezadoTipos,
			fmt.Sprintf("«%s» es un movimiento con tipo, y este entregar no lleva valor.", nombre(f.mov.Nombre)),
			"un movimiento con tipo siempre entrega un valor de ese tipo.",
			"escribe el valor después de entregar.")
	}
}

// faltaEntregar reporta un movimiento con tipo que puede llegar al fin sin
// entregar (validación 4).
func (f *flujo) faltaEntregar(x *ast.DeclMovimiento) {
	pos := x.FinPos
	if pos.Line == 0 && x.Nombre != nil {
		pos = x.Nombre.Pos
	}
	f.c.Error(f.archivo, pos, "falta-entregar", diag.EncabezadoTipos,
		fmt.Sprintf("«%s» debe entregar un valor, pero hay un camino que llega al fin sin entregar.", nombre(x.Nombre)),
		"un movimiento con tipo entrega un valor por todos los caminos; un si sin sino o un ciclo pueden no ejecutarse.",
		"agrega un entregar antes del fin del movimiento, o un sino que también entregue.")
}
