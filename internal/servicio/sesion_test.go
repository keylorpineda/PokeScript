package servicio

import (
	"context"
	"errors"
	"math/rand/v2"
	"path/filepath"
	"testing"
	"time"

	"github.com/keylorpineda/PokeScript/internal/interprete"
)

// conductor ejecuta una sesión y entrega sus eventos por un canal, como
// los recibiría la interfaz.
type conductor struct {
	t       *testing.T
	s       *Sesion
	eventos chan Evento
}

func iniciar(t *testing.T, archivos map[string]string) *conductor {
	t.Helper()
	c := compilarMapa(t, archivos)
	d := &conductor{t: t, eventos: make(chan Evento, 1000)}
	in := interprete.Nuevo(nil)
	in.Azar = rand.New(rand.NewPCG(1, 2))
	s, err := Iniciar(context.Background(), c, in, func(e Evento) { d.eventos <- e })
	if err != nil {
		t.Fatal(err)
	}
	d.s = s
	return d
}

// siguiente espera el próximo evento.
func (d *conductor) siguiente() Evento {
	d.t.Helper()
	select {
	case e := <-d.eventos:
		return e
	case <-time.After(5 * time.Second):
		d.t.Fatal("la sesión no emitió ningún evento en 5 segundos")
		return Evento{}
	}
}

func (d *conductor) esperar(tipo, texto string) Evento {
	d.t.Helper()
	e := d.siguiente()
	if e.Tipo != tipo || (texto != "" && e.Texto != texto) {
		d.t.Fatalf("evento = %+v, want %s %q", e, tipo, texto)
	}
	return e
}

func TestSesionSimple(t *testing.T) {
	d := iniciar(t, map[string]string{"principal.pks": "combate\n    gritar \"¡Hola!\"\n    gritar 1 + 1\nfin\n"})
	d.esperar(EventoSalida, "¡Hola!")
	d.esperar(EventoSalida, "2")
	d.esperar(EventoFin, FinTerminado)
	if err := d.s.Esperar(); err != nil || !d.s.Terminada() {
		t.Errorf("Esperar = %v, Terminada = %v", err, d.s.Terminada())
	}
}

func TestSesionConCapturar(t *testing.T) {
	c, err := Compilar(filepath.Join("..", "..", "ejemplos", "centro.pks"))
	if err != nil {
		t.Fatal(err)
	}
	d := &conductor{t: t, eventos: make(chan Evento, 100)}
	s, err := Iniciar(context.Background(), c, nil, func(e Evento) { d.eventos <- e })
	if err != nil {
		t.Fatal(err)
	}
	d.s = s

	d.esperar(EventoPedirEntrada, "¿Qué Pokémon atrapaste? ")
	if !s.EsperandoEntrada() {
		t.Error("tras pedir-entrada, la sesión debe estar esperando")
	}
	if err := s.EnviarEntrada("Eevee"); err != nil {
		t.Fatal(err)
	}
	d.esperar(EventoPedirEntrada, "¿De qué nivel? ")
	// Una entrada que no es un roca: el intérprete explica y vuelve a preguntar.
	if err := s.EnviarEntrada("doce"); err != nil {
		t.Fatal(err)
	}
	d.esperar(EventoSalida, "")
	d.esperar(EventoPedirEntrada, "¿De qué nivel? ")
	if err := s.EnviarEntrada("12"); err != nil {
		t.Fatal(err)
	}
	d.esperar(EventoSalida, "Eevee va al 12% del camino")
	d.esperar(EventoFin, FinTerminado)
	if err := s.EnviarEntrada("tarde"); !errors.Is(err, ErrNoEsperaEntrada) {
		t.Errorf("después del fin, EnviarEntrada = %v", err)
	}
}

func TestSesionDetenidaMientrasEspera(t *testing.T) {
	d := iniciar(t, map[string]string{"principal.pks": "combate\n    planta x\n    capturar(x, \"? \")\n    gritar x\nfin\n"})
	d.esperar(EventoPedirEntrada, "? ")
	d.s.Detener()
	d.esperar(EventoFin, FinDetenido)
	if err := d.s.Esperar(); err != nil {
		t.Errorf("detener no es un error: %v", err)
	}
}

func TestSesionDetieneUnCicloInfinito(t *testing.T) {
	d := iniciar(t, map[string]string{"principal.pks": "combate\n    roca n = 0\n    mientras verdadero\n        n = n + 1\n        gritar n\n    fin\nfin\n"})
	d.esperar(EventoSalida, "1")
	d.s.Detener()
	for {
		if e := d.siguiente(); e.Tipo == EventoFin {
			if e.Texto != FinDetenido {
				t.Errorf("fin = %q", e.Texto)
			}
			break
		}
	}
}

func TestSesionConErrorDeEjecucion(t *testing.T) {
	d := iniciar(t, map[string]string{"principal.pks": "combate\n    roca cero = 0\n    gritar 10 / cero\nfin\n"})
	e := d.esperar(EventoErrorEjecucion, "")
	if e.Diagnostico == nil || e.Diagnostico.Code != interprete.CodigoDivisionCero || e.Diagnostico.Line != 3 {
		t.Errorf("diagnóstico = %+v", e.Diagnostico)
	}
	d.esperar(EventoFin, FinConError)
	var errEjecucion *interprete.ErrorEjecucion
	if err := d.s.Esperar(); !errors.As(err, &errEjecucion) {
		t.Errorf("Esperar = %v", err)
	}
}

func TestSesionNoIniciaConErrores(t *testing.T) {
	c := compilarMapa(t, map[string]string{"principal.pks": "combate\n    gritar x\nfin\n"})
	if _, err := Iniciar(context.Background(), c, nil, func(Evento) {}); !errors.Is(err, ErrHayErrores) {
		t.Errorf("err = %v", err)
	}
}
