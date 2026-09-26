package interprete

import (
	"context"
	"math/rand/v2"
	"time"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// Interprete ejecuta un programa recorriendo su AST.
//
// Ejecuta un archivo suelto o un proyecto de varios archivos con sus
// importaciones (EjecutarProyecto). Supone un AST ya validado: lo que el
// analizador debería rechazar se reporta como error interno.
type Interprete struct {
	// ES recibe lo que escribe gritar y responde a capturar.
	ES ES
	// Azar genera los números de aleatorio. Las pruebas usan una semilla
	// fija para que la salida sea siempre la misma.
	Azar *rand.Rand

	// archivo es el que se está ejecutando, para los errores de ejecución.
	archivo string
	ctx     context.Context
	// alcances tiene, por archivo, sus medallas y las que importa.
	alcances map[string]*Entorno
	// medallasDe dice en qué archivo se declara cada medalla.
	medallasDe map[string]string
	// origen dice en qué archivo se declara cada movimiento.
	origen      map[*ast.DeclMovimiento]string
	movimientos map[string]*ast.DeclMovimiento // por nombre
	fichas      map[string]*ast.DeclFicha      // por nombre del tipo
	especies    map[string]*ast.DeclEspecie    // por nombre del tipo
	valores     map[string]EspecieVal          // SANO → Estado.SANO
	profundidad int                            // llamadas anidadas en curso
	entregado   Value                          // valor del último entregar
}

// Nuevo crea un intérprete que escribe y lee con es.
func Nuevo(es ES) *Interprete {
	semilla := uint64(time.Now().UnixNano())
	return &Interprete{ES: es, Azar: rand.New(rand.NewPCG(semilla, semilla>>1))}
}

// Ejecutar corre un programa de un solo archivo. Es EjecutarProyecto con
// una lista de un archivo.
func (in *Interprete) Ejecutar(ctx context.Context, prog *ast.Programa) error {
	return in.EjecutarProyecto(ctx, []*ast.Programa{prog})
}

// EjecutarProyecto corre un proyecto de varios archivos (sección 6):
// registra especies, fichas y movimientos de todos los archivos, evalúa las
// medallas archivo por archivo y ejecuta combate.
//
// archivos va en el orden de internal/proyecto: cada archivo después de los
// que importa y el principal al final. Ese es también el orden en que se
// evalúan las medallas (decisión H8), así una medalla importada ya tiene
// valor cuando otro archivo la usa.
//
// Cada archivo tiene su propio alcance con sus medallas y las que importa;
// un movimiento se ejecuta en el alcance de su archivo, y un error de
// ejecución señala el archivo donde ocurrió.
//
// Devuelve nil si el programa terminó, un *ErrorEjecucion si falló, o el
// error de ctx o ErrEntradaCerrada si se detuvo desde afuera.
func (in *Interprete) EjecutarProyecto(ctx context.Context, archivos []*ast.Programa) error {
	in.ctx = ctx
	in.alcances = map[string]*Entorno{}
	in.medallasDe = map[string]string{}
	in.origen = map[*ast.DeclMovimiento]string{}
	in.movimientos = map[string]*ast.DeclMovimiento{}
	in.fichas = map[string]*ast.DeclFicha{}
	in.especies = map[string]*ast.DeclEspecie{}
	in.valores = map[string]EspecieVal{}
	in.profundidad = 0

	var combate *ast.Combate
	var archivoCombate string
	for _, prog := range archivos {
		in.archivo = prog.Archivo
		for _, d := range prog.Declaraciones {
			switch x := d.(type) {
			case *ast.DeclEspecie:
				in.especies[x.Nombre.Nombre] = x
				for _, v := range x.Valores {
					in.valores[v.Nombre] = EspecieVal{Tipo: x.Nombre.Nombre, Valor: v.Nombre}
				}
			case *ast.DeclFicha:
				in.fichas[x.Nombre.Nombre] = x
			case *ast.DeclMovimiento:
				in.movimientos[x.Nombre.Nombre] = x
				in.origen[x] = prog.Archivo
			case *ast.DeclMedalla:
				in.medallasDe[x.Nombre.Nombre] = prog.Archivo
			case *ast.Combate:
				if combate != nil {
					return in.interno(x, "hay más de un bloque combate")
				}
				combate, archivoCombate = x, prog.Archivo
			}
		}
	}

	// Las medallas se evalúan después de registrar todo lo demás, porque su
	// valor puede usar valores de especie.
	for _, prog := range archivos {
		if err := in.prepararArchivo(prog); err != nil {
			return err
		}
	}

	if combate == nil {
		in.archivo = ""
		if len(archivos) > 0 {
			in.archivo = archivos[len(archivos)-1].Archivo
		}
		return in.interno(ast.Pos{Line: 1, Col: 1, Len: 1}, "el programa no tiene bloque combate")
	}
	in.archivo = archivoCombate
	s, err := in.ejecutarBloque(combate.Cuerpo, NuevoEntorno(in.alcances[archivoCombate]))
	if err != nil {
		return err
	}
	if s != normal {
		return in.interno(combate, "huir, siguiente o entregar quedó fuera de su lugar")
	}
	return nil
}

// prepararArchivo arma el alcance de un archivo: primero las medallas que
// importa (la misma variable que en su archivo de origen), después las
// propias en orden de aparición.
func (in *Interprete) prepararArchivo(prog *ast.Programa) error {
	in.archivo = prog.Archivo
	alcance := NuevoEntorno(nil)
	in.alcances[prog.Archivo] = alcance
	for _, im := range prog.Importaciones {
		origen := in.alcances[im.Ruta]
		if origen == nil {
			continue
		}
		for _, n := range im.Nombres {
			if in.medallasDe[n.Nombre] != im.Ruta {
				continue // no es una medalla de ese archivo
			}
			if v := origen.vars[n.Nombre]; v != nil {
				alcance.Declarar(n.Nombre, v)
			}
		}
	}
	for _, d := range prog.Declaraciones {
		if m, ok := d.(*ast.DeclMedalla); ok {
			if err := in.declarar(m, m.Nombre, m.Tipo, m.Valor, true, alcance); err != nil {
				return err
			}
		}
	}
	return nil
}

// detenido devuelve el error de ctx si alguien pidió detener la ejecución.
func (in *Interprete) detenido() error {
	if in.ctx == nil {
		return nil
	}
	return in.ctx.Err()
}

// ─── Tipos declarados ──────────────────────────────────────────────────────

// ajustar aplica la conversión automática de la tabla 3.2 al guardar un
// valor en un lugar de tipo t: roca → agua. T → posible T no cambia el
// valor.
func ajustar(v Value, t *ast.TipoExpr) Value {
	if t != nil && t.Forma == ast.TipoSimple && t.Simple == token.AGUA {
		if x, ok := v.(int64); ok {
			return float64(x)
		}
	}
	return v
}

var tipoFuego = &ast.TipoExpr{Forma: ast.TipoSimple, Simple: token.FUEGO}

// tipoDe devuelve el tipo declarado de un destino (dato, elemento o campo),
// o nil si no se conoce (variables de recorrer).
func (in *Interprete) tipoDe(e ast.Expr, env *Entorno) *ast.TipoExpr {
	switch x := e.(type) {
	case *ast.Ident:
		if v := env.Buscar(x.Nombre); v != nil {
			return v.Tipo
		}
	case *ast.Indice:
		base := in.tipoDe(x.Coleccion, env)
		if base == nil {
			return nil
		}
		switch {
		case base.Forma == ast.TipoEquipo || base.Forma == ast.TipoMochila:
			return base.Elem
		case base.Forma == ast.TipoSimple && base.Simple == token.PLANTA:
			return tipoFuego
		}
	case *ast.CampoAcceso:
		base := in.tipoDe(x.Objeto, env)
		if base == nil || base.Forma != ast.TipoNombrado {
			return nil
		}
		if f, ok := in.fichas[base.Nombre]; ok {
			for _, c := range f.Campos {
				if c.Nombre.Nombre == x.Nombre.Nombre {
					return c.Tipo
				}
			}
		}
	}
	return nil
}

// nombreTipo escribe un tipo como en el código: "roca", "equipo de planta".
func nombreTipo(t *ast.TipoExpr) string {
	if t == nil {
		return "desconocido"
	}
	s := ""
	if t.Posible {
		s = "posible "
	}
	switch t.Forma {
	case ast.TipoSimple:
		return s + t.Simple.String()
	case ast.TipoNombrado:
		return s + t.Nombre
	case ast.TipoEquipo:
		return "equipo de " + nombreTipo(t.Elem)
	case ast.TipoMochila:
		return "mochila de " + nombreTipo(t.Clave) + " a " + nombreTipo(t.Elem)
	}
	return "desconocido"
}
