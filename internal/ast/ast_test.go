package ast

import (
	"testing"

	"github.com/keylorpineda/PokeScript/internal/token"
)

// Si un nodo deja de cumplir su interfaz, esto no compila.
var (
	_ Decl = (*DeclMedalla)(nil)
	_ Decl = (*DeclEspecie)(nil)
	_ Decl = (*DeclFicha)(nil)
	_ Decl = (*DeclMovimiento)(nil)
	_ Decl = (*Combate)(nil)

	_ Instr = (*DeclDato)(nil)
	_ Instr = (*Asignacion)(nil)
	_ Instr = (*Gritar)(nil)
	_ Instr = (*Capturar)(nil)
	_ Instr = (*LlamadaInstr)(nil)
	_ Instr = (*Sumar)(nil)
	_ Instr = (*Quitar)(nil)
	_ Instr = (*Si)(nil)
	_ Instr = (*Segun)(nil)
	_ Instr = (*Mientras)(nil)
	_ Instr = (*RecorrerColeccion)(nil)
	_ Instr = (*RecorrerRango)(nil)
	_ Instr = (*Huir)(nil)
	_ Instr = (*Siguiente)(nil)
	_ Instr = (*Entregar)(nil)

	_ Expr = (*Ident)(nil)
	_ Expr = (*LitRoca)(nil)
	_ Expr = (*LitAgua)(nil)
	_ Expr = (*LitFuego)(nil)
	_ Expr = (*LitPlanta)(nil)
	_ Expr = (*LitElectrico)(nil)
	_ Expr = (*LitFantasma)(nil)
	_ Expr = (*LitEquipo)(nil)
	_ Expr = (*LitLlaves)(nil)
	_ Expr = (*Binaria)(nil)
	_ Expr = (*Unaria)(nil)
	_ Expr = (*Respaldo)(nil)
	_ Expr = (*Indice)(nil)
	_ Expr = (*CampoAcceso)(nil)
	_ Expr = (*Llamada)(nil)
	_ Expr = (*Convertir)(nil)
	_ Expr = (*Tamano)(nil)
	_ Expr = (*Aleatorio)(nil)
	_ Expr = (*Redondear)(nil)

	_ Nodo = (*Importacion)(nil)
	_ Nodo = (*TipoExpr)(nil)
	_ Nodo = (*Campo)(nil)
	_ Nodo = (*Param)(nil)
	_ Nodo = (*RamaSi)(nil)
	_ Nodo = (*Alternativa)(nil)
	_ Nodo = (*Par)(nil)
)

func TestPosicionDesdeToken(t *testing.T) {
	tk := token.Token{Kind: token.IDENT, Lexeme: "año", Line: 3, Col: 5, Len: 3}
	var n Expr = &Ident{Pos: DesdeToken(tk), Nombre: tk.Lexeme}
	if got, want := n.Posicion(), (Pos{Line: 3, Col: 5, Len: 3}); got != want {
		t.Errorf("Posicion() = %+v, want %+v", got, want)
	}
}
