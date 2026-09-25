package parser

import (
	"fmt"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/lexer"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// Resultado es lo que produce el análisis sintáctico de un archivo.
type Resultado struct {
	// Programa siempre existe, aunque haya errores: tiene todo lo que se
	// pudo leer. Antes de ejecutarlo hay que revisar Diagnosticos.
	Programa *ast.Programa
	// Diagnosticos junta los errores léxicos y los sintácticos, en ese orden.
	Diagnosticos []diag.Diagnostic
	// Sangrias viene del lexer: columna del primer token de cada línea.
	Sangrias map[int]int
}

// TieneErrores informa si hay al menos un error (las advertencias no cuentan).
func (r Resultado) TieneErrores() bool {
	for _, d := range r.Diagnosticos {
		if d.Severity == diag.Error {
			return true
		}
	}
	return false
}

// Analizar lee un archivo .pks completo:
//
//	programa = { importacion } , { declaracion }
//
// Ante un error descarta tokens hasta el siguiente NEWLINE y sigue, hasta
// juntar MaxErrores diagnósticos sintácticos (sección 2.1).
func Analizar(archivo, fuente string) Resultado {
	lex := lexer.Analizar(archivo, fuente)
	p := nuevo(archivo, lex.Tokens)
	prog := p.programa()
	diags := append(append([]diag.Diagnostic{}, lex.Diagnosticos...), p.diags.Items()...)
	return Resultado{Programa: prog, Diagnosticos: diags, Sangrias: lex.Sangrias}
}

func (p *parser) programa() *ast.Programa {
	prog := &ast.Programa{Archivo: p.archivo}
	for !p.es(token.EOF) && !p.diags.Lleno() {
		p.intentar(func() {
			switch p.actual().Kind {
			case token.ENSENAR:
				if len(prog.Declaraciones) > 0 {
					t := p.actual()
					p.error(ast.DesdeToken(t), "importacion-fuera-de-lugar",
						"las importaciones van al inicio del archivo, antes de cualquier declaración.",
						"el orden de un archivo es: importaciones, declaraciones y combate.",
						"mueve esta línea «enseñar» arriba, con las demás importaciones.")
				}
				prog.Importaciones = append(prog.Importaciones, p.importacion())
			default:
				prog.Declaraciones = append(prog.Declaraciones, p.declaracion())
			}
			p.finDeLinea()
		})
	}
	return prog
}

// intentar ejecuta f; si f encuentra un error sintáctico, descarta el resto
// de la línea para seguir con la siguiente.
func (p *parser) intentar(f func()) (ok bool) {
	defer func() {
		if r := recover(); r != nil {
			if _, esSintaxis := r.(errSintaxis); !esSintaxis {
				panic(r)
			}
			p.sincronizar()
			ok = false
		}
	}()
	f()
	return true
}

// sincronizar descarta tokens hasta pasar el siguiente NEWLINE.
func (p *parser) sincronizar() {
	for !p.es(token.EOF) {
		if p.avanzar().Kind == token.NEWLINE {
			return
		}
	}
}

// finDeLinea exige que la instrucción termine aquí. Si sigue algo, lo más
// probable es que haya dos instrucciones en la misma línea (sección 1.5).
func (p *parser) finDeLinea() {
	if p.aceptar(token.NEWLINE) || p.es(token.EOF) {
		return
	}
	if iniciaDeclaracion(p.actual().Kind) {
		// Un bloque quedó sin fin y ya se reportó: no se suma otro error.
		return
	}
	t := p.actual()
	if iniciaInstruccion(t.Kind) {
		p.fallar(t, "dos-instrucciones",
			"hay dos instrucciones en la misma línea.",
			"en PokeScript cada instrucción va en su propia línea.",
			fmt.Sprintf("pasa %s y lo que sigue a una línea nueva.", describir(t)))
	}
	p.fallar(t, "token-inesperado",
		fmt.Sprintf("después de la instrucción sobra %s.", describir(t)),
		"la instrucción ya estaba completa en este punto.",
		"borra lo que sobra o revisa si falta un operador.")
}

// iniciaInstruccion informa si un token puede empezar una instrucción.
func iniciaInstruccion(k token.Kind) bool {
	switch k {
	case token.IDENT, token.GRITAR, token.CAPTURAR, token.SUMAR, token.QUITAR,
		token.SI, token.SEGUN, token.MIENTRAS, token.RECORRER, token.HUIR,
		token.SIGUIENTE, token.ENTREGAR, token.MEDALLA, token.POSIBLE,
		token.EQUIPO, token.MOCHILA:
		return true
	}
	return k.EsTipoSimple()
}

// iniciaDeclaracion informa si un token solo puede empezar una declaración
// de archivo. Si aparece dentro de un bloque, casi seguro falta un fin.
func iniciaDeclaracion(k token.Kind) bool {
	switch k {
	case token.MOVIMIENTO, token.COMBATE, token.ESPECIE, token.FICHA, token.ENSENAR:
		return true
	}
	return false
}

// ─── Importaciones y declaraciones ─────────────────────────────────────────

// importacion: "enseñar" identificador { "," identificador } "desde" literal_texto
func (p *parser) importacion() *ast.Importacion {
	ini := p.avanzar()
	im := &ast.Importacion{}
	for {
		im.Nombres = append(im.Nombres, p.identificador("el nombre de lo que quieres importar"))
		if !p.aceptar(token.COMMA) {
			break
		}
	}
	p.esperar(token.DESDE, "«desde» y el archivo de donde se importa")
	ruta := p.esperar(token.PLANTA_LIT, "el nombre del archivo entre comillas, como \"tipos.pks\"")
	im.Ruta, im.RutaPos = ruta.Lexeme, ast.DesdeToken(ruta)
	im.Pos = p.desde(ini)
	return im
}

func (p *parser) declaracion() ast.Decl {
	t := p.actual()
	switch t.Kind {
	case token.MEDALLA:
		return p.declMedalla()
	case token.ESPECIE:
		return p.declEspecie()
	case token.FICHA:
		return p.declFicha()
	case token.MOVIMIENTO:
		return p.declMovimiento()
	case token.COMBATE:
		return p.combate()
	case token.FIN:
		p.fallar(t, "fin-sobrante",
			"este «fin» no cierra ningún bloque.",
			"hay más «fin» que bloques abiertos.",
			"borra este «fin» o revisa si sobra en otra parte.")
	}
	if iniciaInstruccion(t.Kind) {
		p.fallar(t, "instruccion-fuera-de-bloque",
			fmt.Sprintf("%s está fuera de un combate o de un movimiento.", describir(t)),
			"las instrucciones solo pueden ir dentro de «combate … fin» o de un movimiento.",
			"mueve esta línea dentro del bloque combate.")
	}
	p.fallar(t, "declaracion-esperada",
		fmt.Sprintf("se esperaba una declaración y se encontró %s.", describir(t)),
		"fuera de los bloques solo van enseñar, medalla, especie, ficha, movimiento y combate.",
		"revisa el inicio de esta línea.")
	return nil
}

// declMedalla: "medalla" tipo identificador "=" expresion
func (p *parser) declMedalla() *ast.DeclMedalla {
	ini := p.avanzar()
	d := &ast.DeclMedalla{Tipo: p.tipo(), Nombre: p.identificador("el nombre de la medalla")}
	p.esperar(token.ASSIGN, "«=» y el valor de la medalla")
	d.Valor = p.expresion()
	d.Pos = p.desde(ini)
	return d
}

// declEspecie: "especie" identificador valores "fin". Los valores pueden
// estar en una o varias líneas, separados por coma.
func (p *parser) declEspecie() *ast.DeclEspecie {
	ini := p.avanzar()
	d := &ast.DeclEspecie{}
	p.cabecera(func() {
		d.Nombre = p.identificador("el nombre de la especie")
		d.Pos = p.desde(ini)
	})
	p.lineas(func() {
		for {
			d.Valores = append(d.Valores, p.identificador("un valor de la especie"))
			if !p.aceptar(token.COMMA) {
				return
			}
			p.aceptar(token.NEWLINE) // se permite cortar la lista después de una coma
		}
	})
	p.cerrar(ini)
	return d
}

// declFicha: "ficha" identificador { tipo identificador } "fin", un campo por línea.
func (p *parser) declFicha() *ast.DeclFicha {
	ini := p.avanzar()
	d := &ast.DeclFicha{}
	p.cabecera(func() {
		d.Nombre = p.identificador("el nombre de la ficha")
		d.Pos = p.desde(ini)
	})
	p.lineas(func() {
		c := &ast.Campo{Tipo: p.tipo()}
		c.Nombre = p.identificador("el nombre del campo")
		c.Pos = entre(c.Tipo.Pos, c.Nombre.Pos)
		d.Campos = append(d.Campos, c)
	})
	p.cerrar(ini)
	return d
}

// declMovimiento: "movimiento" [ tipo ] identificador "(" [ parametros ] ")" bloque "fin"
func (p *parser) declMovimiento() *ast.DeclMovimiento {
	ini := p.avanzar()
	d := &ast.DeclMovimiento{}
	p.cabecera(func() {
		if !p.es(token.IDENT) || p.ver(1).Kind != token.LPAREN {
			d.Retorno = p.tipo()
		}
		d.Nombre = p.identificador("el nombre del movimiento")
		p.esperar(token.LPAREN, "«(» con los parámetros del movimiento")
		if !p.aceptar(token.RPAREN) {
			for {
				pr := &ast.Param{Tipo: p.tipo()}
				pr.Nombre = p.identificador("el nombre del parámetro")
				pr.Pos = entre(pr.Tipo.Pos, pr.Nombre.Pos)
				d.Params = append(d.Params, pr)
				if !p.aceptar(token.COMMA) {
					break
				}
			}
			p.esperar(token.RPAREN, "«)» para cerrar los parámetros")
		}
		d.Pos = p.desde(ini)
	})
	d.Cuerpo = p.bloque()
	d.FinPos = p.cerrar(ini)
	return d
}

// combate: "combate" bloque "fin"
func (p *parser) combate() *ast.Combate {
	ini := p.avanzar()
	c := &ast.Combate{Pos: ast.DesdeToken(ini)}
	p.cabecera(func() {})
	c.Cuerpo = p.bloque()
	p.cerrar(ini)
	return c
}

// ─── Pila de bloques ───────────────────────────────────────────────────────
//
// La pila de bloques abiertos es la propia recursión del parser: cada bloque
// guarda el token que lo abrió y, al cerrarse, sabe su línea y columna.

// cabecera lee la primera línea de un bloque (si cond, mientras cond, …)
// hasta su NEWLINE. Si tiene un error, lo reporta y descarta el resto de la
// línea, pero el cuerpo del bloque se sigue leyendo: así un error en la
// condición no descuadra los fin que vienen después.
func (p *parser) cabecera(f func()) {
	p.intentar(func() {
		f()
		if !p.es(token.NEWLINE) && !p.es(token.EOF) {
			t := p.actual()
			p.fallar(t, "token-inesperado",
				fmt.Sprintf("después de la primera línea del bloque sobra %s.", describir(t)),
				"el cuerpo del bloque empieza en la línea siguiente.",
				"pasa lo que sobra a la línea de abajo.")
		}
		p.aceptar(token.NEWLINE)
	})
}

// cerrar consume el fin del bloque abierto por ini y lo saca de la pila.
// Si no hay fin, reporta el diagnóstico estrella: en qué línea se abrió el
// bloque que quedó sin cerrar (sección 2.1). Devuelve la posición del fin.
func (p *parser) cerrar(ini token.Token) ast.Pos {
	if p.es(token.FIN) {
		fin := p.avanzar()
		if fin.Col < ini.Col {
			// Un fin con menos sangría que su bloque probablemente era de un
			// bloque de afuera: el que cierra es el que no tiene el suyo.
			p.sospechosos = append(p.sospechosos, ini)
		}
		return ast.DesdeToken(fin)
	}
	p.bloqueSinCerrar(ini)
	return ast.Pos{}
}

func (p *parser) bloqueSinCerrar(ini token.Token) {
	causa := fmt.Sprintf("cada «%s» se cierra con su propio «fin».", ini.Kind)
	for _, s := range p.sospechosos {
		if s.Line > ini.Line {
			causa = fmt.Sprintf("por la sangría, parece que falta el «fin» del «%s» que abriste en la línea %d.", s.Kind, s.Line)
			break
		}
	}
	donde := "al final del archivo"
	if !p.es(token.EOF) {
		donde = fmt.Sprintf("a la línea %d", p.actual().Line)
	}
	p.error(ast.DesdeToken(ini), "bloque-sin-cerrar",
		fmt.Sprintf("el bloque «%s» que abriste en la línea %d no tiene su «fin»; se llegó %s sin cerrarlo.", ini.Kind, ini.Line, donde),
		causa,
		"agrega «fin» donde termina el bloque, con la misma sangría que la línea que lo abre.")
}

// lineas lee líneas de un bloque de declaración (especie, ficha) hasta su fin.
func (p *parser) lineas(f func()) {
	for !p.es(token.FIN) && !p.es(token.EOF) && !iniciaDeclaracion(p.actual().Kind) && !p.diags.Lleno() {
		p.intentar(func() {
			f()
			p.finDeLinea()
		})
	}
}
