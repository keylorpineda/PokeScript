package servicio

import (
	"errors"
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/keylorpineda/PokeScript/internal/interprete"
)

func entornoDePrueba() *Entorno {
	e := NuevoEntorno()
	e.NuevoInterprete = func() *interprete.Interprete {
		in := interprete.Nuevo(nil)
		in.Azar = rand.New(rand.NewPCG(1, 2))
		return in
	}
	return e
}

func TestEntornoFlujoCompleto(t *testing.T) {
	e := entornoDePrueba()
	ruta := filepath.Join("..", "..", "ejemplos", "combate")

	r, err := e.CompilarProyecto(ruta)
	if err != nil || !r.Exito {
		t.Fatalf("CompilarProyecto: %v %+v", err, r)
	}
	if e.Ejecutando() || e.EnviarEntrada("x") == nil {
		t.Error("sin ejecución no puede haber entrada")
	}

	eventos := make(chan Evento, 1000)
	if _, err := e.EjecutarProyecto(ruta, func(ev Evento) { eventos <- ev }); err != nil {
		t.Fatal(err)
	}
	espera := func() Evento {
		select {
		case ev := <-eventos:
			return ev
		case <-time.After(5 * time.Second):
			t.Fatal("sin eventos")
			return Evento{}
		}
	}
	if ev := espera(); ev.Tipo != EventoPedirEntrada {
		t.Fatalf("primer evento = %+v", ev)
	}
	if !e.Ejecutando() {
		t.Error("debe estar ejecutando mientras espera la entrada")
	}
	if err := e.EnviarEntrada("Pikachu"); err != nil {
		t.Fatal(err)
	}
	for ev := espera(); ev.Tipo != EventoFin; ev = espera() {
		if ev.Tipo == EventoErrorEjecucion {
			t.Fatalf("error de ejecución: %+v", ev.Diagnostico)
		}
	}
	if err := e.Esperar(); err != nil || e.Ejecutando() {
		t.Errorf("Esperar = %v, Ejecutando = %v", err, e.Ejecutando())
	}
}

func TestEntornoNoEjecutaConErrores(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "principal.pks"), []byte("combate\n    gritar x\nfin\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	e := entornoDePrueba()
	r, err := e.EjecutarProyecto(dir, func(Evento) { t.Error("no debe emitir nada") })
	if err != nil || r.Exito || e.Ejecutando() {
		t.Errorf("r = %+v, err = %v", r, err)
	}
	if _, err := e.CompilarProyecto(filepath.Join(dir, "no-existe")); err == nil {
		t.Error("una ruta inexistente debe dar error")
	}
	if _, err := e.EjecutarProyecto(filepath.Join(dir, "no-existe"), nil); err == nil {
		t.Error("una ruta inexistente debe dar error")
	}
}

func TestEntornoReemplazaLaEjecucionAnterior(t *testing.T) {
	dir := t.TempDir()
	programa := "combate\n    planta x\n    capturar(x, \"? \")\n    gritar x\nfin\n"
	if err := os.WriteFile(filepath.Join(dir, "principal.pks"), []byte(programa), 0o600); err != nil {
		t.Fatal(err)
	}
	e := entornoDePrueba()
	primera := make(chan Evento, 10)
	if _, err := e.EjecutarProyecto(dir, func(ev Evento) { primera <- ev }); err != nil {
		t.Fatal(err)
	}
	<-primera // pedir-entrada

	segunda := make(chan Evento, 10)
	if _, err := e.EjecutarProyecto(dir, func(ev Evento) { segunda <- ev }); err != nil {
		t.Fatal(err)
	}
	if ev := <-primera; ev.Tipo != EventoFin || ev.Texto != FinDetenido {
		t.Errorf("la primera ejecución debe terminar detenida: %+v", ev)
	}
	<-segunda // pedir-entrada de la segunda
	e.DetenerEjecucion()
	if err := e.Esperar(); err != nil {
		t.Error(err)
	}
	if err := e.EnviarEntrada("tarde"); !errors.Is(err, ErrNoEsperaEntrada) {
		t.Errorf("EnviarEntrada tras detener = %v", err)
	}
}
