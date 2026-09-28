package parser

import (
	"fmt"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/token"
)

// tipo: tipo_opcional | tipo_equipo | tipo_mochila
//
//	tipo_opcional  = [ "posible" ] ( tipo_simple | tipo_nombrado )
//	tipo_equipo    = "equipo" "de" tipo_contenido
//	tipo_mochila   = "mochila" "de" tipo_clave "a" tipo_contenido
func (p *parser) tipo() *ast.TipoExpr {
	ini := p.actual()
	if p.aceptar(token.POSIBLE) {
		t := p.actual()
		if t.Kind == token.EQUIPO || t.Kind == token.MOCHILA {
			p.fallar(t, "posible-no-permitido",
				fmt.Sprintf("«posible» no se puede aplicar a un %s.", t.Kind),
				"solo los tipos simples y las especies pueden ser posible.",
				"quita «posible» o usa una colección vacía para indicar que no hay elementos.")
		}
		base := p.tipoBase("un tipo después de «posible»")
		base.Posible = true
		base.Pos = p.desde(ini)
		return base
	}
	switch ini.Kind {
	case token.EQUIPO, token.MOCHILA:
		return p.tipoColeccion()
	}
	return p.tipoBase("un tipo")
}

// tipoBase: tipo_simple | tipo_nombrado
func (p *parser) tipoBase(que string) *ast.TipoExpr {
	t := p.actual()
	switch {
	case t.Kind.EsTipoSimple():
		p.avanzar()
		return &ast.TipoExpr{Pos: ast.DesdeToken(t), Forma: ast.TipoSimple, Simple: t.Kind}
	case t.Kind == token.IDENT:
		p.avanzar()
		return &ast.TipoExpr{Pos: ast.DesdeToken(t), Forma: ast.TipoNombrado, Nombre: t.Lexeme}
	}
	p.fallar(t, "tipo-esperado",
		fmt.Sprintf("se esperaba %s y se encontró %s.", que, describir(t)),
		"los tipos son roca, agua, fuego, planta, electrico, equipo, mochila o el nombre de una especie o ficha.",
		"escribe uno de esos tipos.")
	return nil
}

// tipoContenido: tipo_simple | tipo_nombrado | tipo_equipo | tipo_mochila
func (p *parser) tipoContenido() *ast.TipoExpr {
	t := p.actual()
	switch t.Kind {
	case token.EQUIPO, token.MOCHILA:
		return p.tipoColeccion()
	case token.POSIBLE:
		p.fallar(t, "posible-no-permitido",
			"«posible» no se puede usar para los elementos de una colección.",
			"una colección guarda valores que siempre existen.",
			"quita «posible»; si un elemento no existe, simplemente no se agrega.")
	}
	return p.tipoBase("el tipo de los elementos")
}

// tipoColeccion: "equipo" "de" tipo_contenido | "mochila" "de" tipo_clave "a" tipo_contenido
func (p *parser) tipoColeccion() *ast.TipoExpr {
	ini := p.avanzar()
	p.esperar(token.DE, fmt.Sprintf("«de» después de «%s»", ini.Kind))
	if ini.Kind == token.EQUIPO {
		elem := p.tipoContenido()
		return &ast.TipoExpr{Pos: p.desde(ini), Forma: ast.TipoEquipo, Elem: elem}
	}
	clave := p.tipoClave()
	p.esperar(token.A, "«a» entre el tipo de la clave y el del valor")
	valor := p.tipoContenido()
	return &ast.TipoExpr{Pos: p.desde(ini), Forma: ast.TipoMochila, Clave: clave, Elem: valor}
}

// tipoClave: "roca" | "fuego" | "planta" | tipo_nombrado
func (p *parser) tipoClave() *ast.TipoExpr {
	t := p.actual()
	switch t.Kind {
	case token.ROCA, token.FUEGO, token.PLANTA:
		p.avanzar()
		return &ast.TipoExpr{Pos: ast.DesdeToken(t), Forma: ast.TipoSimple, Simple: t.Kind}
	case token.IDENT:
		p.avanzar()
		return &ast.TipoExpr{Pos: ast.DesdeToken(t), Forma: ast.TipoNombrado, Nombre: t.Lexeme}
	}
	p.fallar(t, "clave-de-mochila-invalida",
		fmt.Sprintf("%s no puede ser la clave de una mochila.", describir(t)),
		"las claves de una mochila son roca, fuego, planta o una especie.",
		"usa uno de esos tipos para la clave.")
	return nil
}
