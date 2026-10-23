package sesion

import (
	"context"
	"errors"
	"testing"
	"time"
	acceso "uniapp/internal/models/acceso"
)

type repoOutboxPrueba struct {
	entrega                     *Entrega
	codigo                      string
	errorReclamo, errorResolver error
	correo                      string
}

func (r *repoOutboxPrueba) Reclamar(context.Context, time.Time) (*Entrega, error) {
	return r.entrega, r.errorReclamo
}
func (r *repoOutboxPrueba) Resolver(_ context.Context, _ Entrega, _ time.Time, c string) error {
	r.codigo = c
	return r.errorResolver
}
func (r *repoOutboxPrueba) CorreoDe(context.Context, string) (string, error) { return r.correo, nil }

type mailerOutboxPrueba struct {
	llamadas int
	err      error
}

func (m *mailerOutboxPrueba) Enviar(context.Context, acceso.Mensaje) error {
	m.llamadas++
	return m.err
}
func TestProcesadorOutboxCanalesYFallos(t *testing.T) {
	for _, c := range []struct {
		canal, correo, codigo string
		falla                 bool
		llamadas              int
	}{{"interno", "", "", false, 0}, {"correo", "", "destinatario_no_disponible", false, 0}, {"correo", "a@example.test", "", false, 1}, {"correo", "a@example.test", "proveedor_no_disponible", true, 1}} {
		r := &repoOutboxPrueba{entrega: &Entrega{Canal: c.canal}, correo: c.correo}
		m := &mailerOutboxPrueba{}
		if c.falla {
			m.err = errors.New("secreto del proveedor")
		}
		hay, err := (ProcesadorOutbox{Repositorio: r, Mailer: m}).ProcesarUno(context.Background())
		if !hay || err != nil || m.llamadas != c.llamadas || r.codigo != c.codigo {
			t.Fatalf("resultado: %v %v %s", hay, err, r.codigo)
		}
	}
}
func TestProcesadorSinTrabajoYFalloRepositorio(t *testing.T) {
	r := &repoOutboxPrueba{}
	p := ProcesadorOutbox{Repositorio: r}
	if hay, err := p.ProcesarUno(context.Background()); hay || err != nil {
		t.Fatal("sin trabajo")
	}
	r.errorReclamo = errors.New("base caída")
	if _, err := p.ProcesarUno(context.Background()); err == nil {
		t.Fatal("ocultó fallo de reclamo")
	}
	r.errorReclamo = nil
	r.entrega = &Entrega{Canal: "interno"}
	r.errorResolver = errors.New("commit falló")
	if _, err := p.ProcesarUno(context.Background()); err == nil {
		t.Fatal("ocultó fallo de resultado")
	}
}
