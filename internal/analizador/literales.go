package analizador

import (
	"fmt"
	"strings"

	"github.com/keylorpineda/PokeScript/internal/ast"
	"github.com/keylorpineda/PokeScript/internal/diag"
	"github.com/keylorpineda/PokeScript/internal/tipos"
)

// J2 · Literales de colección (sección 2.1 y pasada 2):
//
//	14 literal de colección sin tipo esperado
//	15 literal de ficha: campos completos, sin repetir, sin ajenos
//
// El parser no puede saber si { } es una mochila o una ficha: aquí se
// decide con el tipo esperado y se anota en LitLlaves.Resuelto, que el
// intérprete respeta.
//
// No recorre el programa por su cuenta: se engancha al recorrido de
// tipos_expr.go, que es el que conoce el tipo esperado de cada lugar.
func init() { tipoLiteral = revisarLiteral }

func revisarLiteral(v *verificador, e ast.Expr, esperado *tipos.Type) *tipos.Type {
	switch x := e.(type) {
	case *ast.LitEquipo:
		return v.literalEquipo(x, esperado)
	case *ast.LitLlaves:
		return v.literalLlaves(x, esperado)
	}
	return nil
}

func (v *verificador) literalEquipo(x *ast.LitEquipo, esperado *tipos.Type) *tipos.Type {
	if esperado == nil || esperado.Kind != tipos.KEquipo {
		for _, el := range x.Elems {
			v.tipo(el)
		}
		v.literalFueraDeLugar(x.Pos, "de equipo", esperado)
		return nil
	}
	for _, el := range x.Elems {
		t := v.tipoCon(el, esperado.Elem)
		v.asignable(el, t, esperado.Elem, "tipo-incompatible", "elemento de un "+esperado.String())
	}
	return esperado
}

func (v *verificador) literalLlaves(x *ast.LitLlaves, esperado *tipos.Type) *tipos.Type {
	switch {
	case esperado != nil && esperado.Kind == tipos.KMochila:
		x.Resuelto = ast.LlavesMochila
		for _, p := range x.Pares {
			kt := v.tipoCon(p.Clave, esperado.Clave)
			v.asignable(p.Clave, kt, esperado.Clave, "tipo-incompatible", "clave de una "+esperado.String())
			vt := v.tipoCon(p.Valor, esperado.Elem)
			v.asignable(p.Valor, vt, esperado.Elem, "tipo-incompatible", "valor de una "+esperado.String())
		}
		return esperado
	case esperado != nil && esperado.Kind == tipos.KFicha && v.c.Tabla.Fichas[esperado.Nombre] != nil:
		x.Resuelto = ast.LlavesFicha
		v.literalFicha(x, v.c.Tabla.Fichas[esperado.Nombre])
		return esperado
	}
	// Sin saber qué es, las claves pueden ser nombres de campo: solo se
	// revisan los valores.
	for _, p := range x.Pares {
		v.tipo(p.Valor)
	}
	v.literalFueraDeLugar(x.Pos, "{ }", esperado)
	return nil
}

// literalFicha revisa que el literal tenga cada campo de la ficha una sola
// vez y ninguno ajeno (validación 15).
func (v *verificador) literalFicha(x *ast.LitLlaves, f *Ficha) {
	vistos := map[string]bool{}
	for _, p := range x.Pares {
		id, ok := p.Clave.(*ast.Ident)
		if !ok {
			v.tipo(p.Valor)
			v.error(p.Clave.Posicion(), "clave-de-ficha-invalida", diag.EncabezadoTipos,
				fmt.Sprintf("en un literal de %s, cada clave es el nombre de un campo.", f.Nombre),
				"una ficha tiene campos con nombre, no claves calculadas.",
				fmt.Sprintf("escribe el campo sin comillas, por ejemplo {%s: …}.", primerCampo(f)))
			continue
		}
		c := f.Campo(id.Nombre)
		switch {
		case c == nil:
			v.tipo(p.Valor)
			v.error(id.Pos, "campo-ajeno", diag.EncabezadoTipos,
				fmt.Sprintf("la ficha %s no tiene un campo «%s».", f.Nombre, id.Nombre),
				fmt.Sprintf("los campos de %s son: %s.", f.Nombre, camposDe(f)),
				"quita ese campo o revisa cómo está escrito.")
			continue
		case vistos[id.Nombre]:
			v.error(id.Pos, "campo-repetido-en-literal", diag.EncabezadoTipos,
				fmt.Sprintf("el campo «%s» aparece dos veces en este literal de %s.", id.Nombre, f.Nombre),
				"cada campo recibe un solo valor.",
				"deja una sola vez ese campo.")
		}
		vistos[id.Nombre] = true
		t := v.tipoCon(p.Valor, c.Tipo)
		v.asignable(p.Valor, t, c.Tipo, "tipo-incompatible", "campo «"+c.Nombre+"» de "+f.Nombre)
	}
	var faltan []string
	for _, c := range f.Campos {
		if !vistos[c.Nombre] {
			faltan = append(faltan, c.Nombre)
		}
	}
	if len(faltan) > 0 {
		que := "falta el campo"
		if len(faltan) > 1 {
			que = "faltan los campos"
		}
		v.error(x.Pos, "ficha-incompleta", diag.EncabezadoSinValor,
			fmt.Sprintf("a este literal de %s le %s %s.", f.Nombre, que, strings.Join(faltan, ", ")),
			"una ficha nace con todos sus campos; no hay campos vacíos.",
			fmt.Sprintf("agrega %s con su valor.", strings.Join(faltan, ", ")))
	}
}

// literalFueraDeLugar explica un literal sin tipo esperado (validación 14)
// o que no corresponde al tipo del lugar.
func (v *verificador) literalFueraDeLugar(pos ast.Pos, que string, esperado *tipos.Type) {
	if esperado == nil {
		v.error(pos, "literal-sin-tipo", diag.EncabezadoTipos,
			fmt.Sprintf("no se sabe de qué tipo es este literal %s.", que),
			"un literal de colección toma su tipo del lugar donde se guarda, y aquí no se guarda en ninguno.",
			"guárdalo primero en un dato con tipo, por ejemplo: equipo de roca e = [1, 2].")
		return
	}
	v.error(pos, "literal-incompatible", diag.EncabezadoTipos,
		fmt.Sprintf("aquí se espera un valor de tipo %s, y esto es un literal %s.", esperado, que),
		"el literal no corresponde al tipo del lugar donde se guarda.",
		"revisa el tipo declarado o el literal.")
}

func primerCampo(f *Ficha) string {
	if len(f.Campos) == 0 {
		return "campo"
	}
	return f.Campos[0].Nombre
}
