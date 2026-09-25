package parser

import (
	"fmt"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// bloque lee instrucciones, una por línea, hasta un fin, el final del
// archivo, una palabra de parar o el inicio de una declaración de archivo
// (señal de que a este bloque le faltó su fin). No consume lo que lo detiene.
func (p *parser) bloque(parar ...token.Kind) []ast.Instr {
	var cuerpo []ast.Instr
	for !p.es(token.FIN) && !p.es(token.EOF) && !iniciaDeclaracion(p.actual().Kind) && !p.diags.Lleno() {
		if contiene(parar, p.actual().Kind) {
			break
		}
		p.intentar(func() {
			i := p.instruccion()
			cuerpo = append(cuerpo, i)
			p.finDeLinea()
		})
	}
	return cuerpo
}

func contiene(ks []token.Kind, k token.Kind) bool {
	for _, x := range ks {
		if x == k {
			return true
		}
	}
	return false
}

// instruccion lee una instrucción sin su NEWLINE final.
func (p *parser) instruccion() ast.Instr {
	t := p.actual()
	switch t.Kind {
	case token.GRITAR:
		return p.gritar()
	case token.CAPTURAR:
		return p.capturar()
	case token.SUMAR:
		return p.sumar()
	case token.QUITAR:
		return p.quitar()
	case token.SI:
		return p.si()
	case token.SEGUN:
		return p.segun()
	case token.MIENTRAS:
		return p.mientras()
	case token.RECORRER:
		return p.recorrer()
	case token.HUIR:
		p.avanzar()
		return &ast.Huir{Pos: ast.DesdeToken(t)}
	case token.SIGUIENTE:
		p.avanzar()
		return &ast.Siguiente{Pos: ast.DesdeToken(t)}
	case token.ENTREGAR:
		p.avanzar()
		e := &ast.Entregar{}
		if !p.es(token.NEWLINE) && !p.es(token.EOF) {
			e.Valor = p.expresion()
		}
		e.Pos = p.desde(t)
		return e
	case token.MEDALLA:
		p.avanzar()
		d := p.declDato(t)
		d.Medalla = true
		return d
	case token.EQUIPO, token.MOCHILA, token.POSIBLE:
		return p.declDato(t)
	case token.IDENT:
		return p.instruccionConNombre()
	case token.SINO, token.SINO_SI:
		p.fallar(t, "sino-sin-si",
			fmt.Sprintf("este «%s» no pertenece a ningún «si».", t.Kind),
			"un «sino» solo puede ir dentro de un bloque «si … fin».",
			"revisa si falta el «si» o si un «fin» cerró el bloque antes de tiempo.")
	case token.OTRO:
		p.fallar(t, "otro-sin-segun",
			"este «otro» no pertenece a ningún «segun».",
			"«otro entonces …» es la última rama de un «segun».",
			"revisa si falta el «segun» o si un «fin» lo cerró antes de tiempo.")
	}
	if t.Kind.EsTipoSimple() {
		return p.declDato(t)
	}
	p.fallar(t, "instruccion-esperada",
		fmt.Sprintf("se esperaba una instrucción y se encontró %s.", describir(t)),
		"una instrucción empieza con una declaración, una asignación, gritar, capturar, si, segun, mientras, recorrer, huir, siguiente o entregar.",
		"revisa el inicio de esta línea.")
	return nil
}

// instruccionConNombre resuelve la anticipación de dos tokens (sección 2.1):
//
//	nombre (         → llamada
//	nombre nombre    → declaración con tipo nombrado (Estado x)
//	nombre = [ .     → asignación
func (p *parser) instruccionConNombre() ast.Instr {
	t := p.actual()
	switch p.ver(1).Kind {
	case token.LPAREN:
		l := p.llamada()
		return &ast.LlamadaInstr{Pos: l.Pos, Llamada: l}
	case token.IDENT:
		return p.declDato(t)
	}
	destino := p.destino()
	if !p.es(token.ASSIGN) {
		sig := p.actual()
		p.fallar(sig, "asignacion-esperada",
			fmt.Sprintf("se esperaba «=» después de «%s» y se encontró %s.", t.Lexeme, describir(sig)),
			"una línea que empieza con un nombre es una asignación, una declaración o una llamada; un valor suelto no hace nada.",
			fmt.Sprintf("si querías cambiar el valor, escribe «%s = …».", t.Lexeme))
	}
	p.avanzar()
	valor := p.expresion()
	return &ast.Asignacion{Pos: p.desde(t), Destino: destino, Valor: valor}
}

// destino: identificador { "[" expresion "]" | "." identificador }
func (p *parser) destino() ast.Expr {
	var d ast.Expr = p.identificador("el nombre de un dato")
	for {
		switch {
		case p.aceptar(token.LBRACKET):
			indice := p.expresion()
			p.esperarCierre(token.RBRACKET, "]", d.Posicion())
			d = &ast.Indice{Pos: entre(d.Posicion(), ast.DesdeToken(p.anterior())), Coleccion: d, Indice: indice}
		case p.aceptar(token.DOT):
			nombre := p.identificador("el nombre de un campo")
			d = &ast.CampoAcceso{Pos: entre(d.Posicion(), nombre.Pos), Objeto: d, Nombre: nombre}
		default:
			return d
		}
	}
}

// declDato: [ "medalla" ] tipo identificador [ "=" expresion ]. ini es el
// primer token de la línea (la medalla, si la hay).
func (p *parser) declDato(ini token.Token) *ast.DeclDato {
	d := &ast.DeclDato{Tipo: p.tipo()}
	d.Nombre = p.identificador("el nombre del dato")
	if p.aceptar(token.ASSIGN) {
		d.Valor = p.expresion()
	}
	d.Pos = p.desde(ini)
	return d
}

// gritar: "gritar" expresion { "," expresion }
func (p *parser) gritar() *ast.Gritar {
	ini := p.avanzar()
	g := &ast.Gritar{}
	for {
		g.Args = append(g.Args, p.expresion())
		if !p.aceptar(token.COMMA) {
			break
		}
	}
	g.Pos = p.desde(ini)
	return g
}

// capturar: "capturar" "(" destino "," expresion ")"
func (p *parser) capturar() *ast.Capturar {
	ini := p.avanzar()
	abre := p.esperar(token.LPAREN, "«(» después de «capturar»")
	c := &ast.Capturar{Destino: p.destino()}
	p.esperar(token.COMMA, "«,» y el mensaje que se le muestra al usuario")
	c.Mensaje = p.expresion()
	p.esperarCierre(token.RPAREN, ")", ast.DesdeToken(abre))
	c.Pos = p.desde(ini)
	return c
}

// sumar: "sumar" expresion "a" identificador
func (p *parser) sumar() *ast.Sumar {
	ini := p.avanzar()
	s := &ast.Sumar{Valor: p.expresion()}
	p.esperar(token.A, "«a» y el equipo al que se suma el valor")
	s.Coleccion = p.identificador("el nombre del equipo")
	s.Pos = p.desde(ini)
	return s
}

// quitar: "quitar" identificador "[" expresion "]"
func (p *parser) quitar() *ast.Quitar {
	ini := p.avanzar()
	q := &ast.Quitar{Coleccion: p.identificador("el nombre de la colección")}
	abre := p.esperar(token.LBRACKET, "«[» con la posición o la clave que se quita")
	q.Indice = p.expresion()
	p.esperarCierre(token.RBRACKET, "]", ast.DesdeToken(abre))
	q.Pos = p.desde(ini)
	return q
}

// si: "si" expresion bloque { "sino si" expresion bloque } [ "sino" bloque ] "fin"
func (p *parser) si() *ast.Si {
	ini := p.actual()
	s := &ast.Si{Sino: []ast.Instr{}}
	s.Ramas = append(s.Ramas, p.rama())
	s.Pos = s.Ramas[0].Pos
	for p.es(token.SINO_SI) {
		s.Ramas = append(s.Ramas, p.rama())
	}
	if p.es(token.SINO) {
		sino := p.avanzar()
		s.TieneSino, s.SinoPos = true, ast.DesdeToken(sino)
		p.cabecera(func() {})
		if cuerpo := p.bloque(); cuerpo != nil {
			s.Sino = cuerpo
		}
	}
	p.cerrar(ini)
	return s
}

// rama lee "si cond" o "sino si cond" y su cuerpo.
func (p *parser) rama() *ast.RamaSi {
	ini := p.avanzar()
	r := &ast.RamaSi{Pos: ast.DesdeToken(ini)}
	p.cabecera(func() {
		r.Cond = p.expresion()
		r.Pos = p.desde(ini)
	})
	r.Cuerpo = p.bloque(token.SINO, token.SINO_SI)
	return r
}

// mientras: "mientras" expresion bloque "fin"
func (p *parser) mientras() *ast.Mientras {
	ini := p.avanzar()
	m := &ast.Mientras{Pos: ast.DesdeToken(ini)}
	p.cabecera(func() {
		m.Cond = p.expresion()
		m.Pos = p.desde(ini)
	})
	m.Cuerpo = p.bloque()
	p.cerrar(ini)
	return m
}

// recorrer tiene dos formas (sección 2):
//
//	"recorrer" identificador [ "," identificador ] "en" expresion bloque "fin"
//	"recorrer" identificador "de" expresion "hasta" expresion bloque "fin"
func (p *parser) recorrer() ast.Instr {
	ini := p.avanzar()
	var (
		rango     *ast.RecorrerRango
		recorrido *ast.RecorrerColeccion
	)
	p.cabecera(func() {
		v := p.identificador("el nombre de la variable del recorrido")
		if p.aceptar(token.DE) {
			rango = &ast.RecorrerRango{Var: v, Desde: p.expresion()}
			if !p.es(token.HASTA) {
				t := p.actual()
				if t.Kind == token.A {
					p.fallar(t, "rango-sin-hasta",
						"en un recorrido por rango se usa «hasta», no «a».",
						"la forma es: recorrer n de inicio hasta fin.",
						"cambia «a» por «hasta».")
				}
				p.esperar(token.HASTA, "«hasta» y el final del rango")
			}
			p.avanzar()
			rango.Hasta = p.expresion()
			rango.Pos = p.desde(ini)
			return
		}
		recorrido = &ast.RecorrerColeccion{Var: v}
		if p.aceptar(token.COMMA) {
			recorrido.Var2 = p.identificador("el nombre de la segunda variable")
		}
		p.esperar(token.EN, "«en» y la colección, o «de» y un rango")
		recorrido.Coleccion = p.expresion()
		recorrido.Pos = p.desde(ini)
	})
	cuerpo := p.bloque()
	p.cerrar(ini)
	switch {
	case rango != nil:
		rango.Cuerpo = cuerpo
		return rango
	case recorrido != nil:
		recorrido.Cuerpo = cuerpo
		return recorrido
	}
	// La cabecera falló antes de saber la forma: igual se leyó el cuerpo
	// para no descuadrar los fin, y se devuelve un recorrido vacío.
	return &ast.RecorrerColeccion{Pos: ast.DesdeToken(ini), Cuerpo: cuerpo}
}

// segun: "segun" expresion { alternativa } [ "otro" "entonces" instruccion ] "fin"
//
//	alternativa = patron { "," patron } "entonces" instruccion
func (p *parser) segun() *ast.Segun {
	ini := p.avanzar()
	s := &ast.Segun{Pos: ast.DesdeToken(ini)}
	p.cabecera(func() {
		s.Valor = p.expresion()
		s.Pos = p.desde(ini)
	})
	for !p.es(token.FIN) && !p.es(token.EOF) && !iniciaDeclaracion(p.actual().Kind) && !p.diags.Lleno() {
		p.intentar(func() {
			if p.es(token.OTRO) {
				otro := p.avanzar()
				if s.Otro != nil {
					p.fallar(otro, "otro-repetido",
						"este «segun» ya tiene una rama «otro».",
						"solo puede haber una rama «otro», y va al final.",
						"junta las dos ramas «otro» en una.")
				}
				p.esperar(token.ENTONCES, "«entonces» después de «otro»")
				s.OtroPos = ast.DesdeToken(otro)
				s.Otro = p.instruccion()
			} else {
				if s.Otro != nil {
					t := p.actual()
					p.fallar(t, "rama-despues-de-otro",
						"hay una rama después de «otro».",
						"«otro» atrapa todos los casos que faltan, así que debe ser la última rama.",
						"mueve la rama «otro» al final del «segun».")
				}
				s.Alternativas = append(s.Alternativas, p.alternativa())
			}
			p.finDeLinea()
		})
	}
	p.cerrar(ini)
	return s
}

func (p *parser) alternativa() *ast.Alternativa {
	ini := p.actual()
	a := &ast.Alternativa{}
	for {
		a.Patrones = append(a.Patrones, p.patron())
		if !p.aceptar(token.COMMA) {
			break
		}
	}
	p.esperar(token.ENTONCES, "«entonces» y la instrucción de esta rama")
	a.Pos = p.desde(ini)
	a.Cuerpo = p.instruccion()
	return a
}

// patron: literal | identificador
func (p *parser) patron() ast.Expr {
	t := p.actual()
	switch t.Kind {
	case token.ROCA_LIT, token.AGUA_LIT, token.FUEGO_LIT, token.PLANTA_LIT,
		token.VERDADERO, token.FALSO:
		return p.primaria()
	case token.IDENT:
		p.avanzar()
		return &ast.Ident{Pos: ast.DesdeToken(t), Nombre: t.Lexeme}
	}
	p.fallar(t, "patron-invalido",
		fmt.Sprintf("%s no puede ser un caso de «segun».", describir(t)),
		"cada caso es un valor fijo: un número, un carácter, un texto, verdadero, falso o un valor de especie.",
		"escribe el valor tal cual; para comparar con cálculos usa «si».")
	return nil
}
