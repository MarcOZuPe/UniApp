package sesion

import (
	"context"
	"sync"
	"testing"
	"time"

	memoriaSesion "uniapp/internal/adapters/repositories/memory/sesion"
)

// cancelacionesContadas implementa EstadoReservas y cuenta cuantas consultas
// recibe, para detectar el patron N+1 de F-17.
type cancelacionesContadas struct {
	mu         sync.Mutex
	canceladas map[string]bool
	llamadas   int
}

func (c *cancelacionesContadas) Cancelada(_ context.Context, reservaID string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.llamadas++
	return c.canceladas[reservaID], nil
}

func (c *cancelacionesContadas) totalLlamadas() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.llamadas
}

// minutosConContador cuenta las escrituras que el servicio dispara al leer.
type minutosConContador struct {
	*memoriaSesion.RepositorioMinutosMemoria
	mu              sync.Mutex
	actualizaciones int
}

func (r *minutosConContador) ActualizarMinutos(
	ctx context.Context, sesionID string, cambio func(*RegistroMinutos) error,
) (*RegistroMinutos, error) {
	r.mu.Lock()
	r.actualizaciones++
	r.mu.Unlock()
	return r.RepositorioMinutosMemoria.ActualizarMinutos(ctx, sesionID, cambio)
}

func (r *minutosConContador) escrituras() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.actualizaciones
}

func sesionConMinutos(t *testing.T, sesionID, tutorID, escolarID string, fin time.Time) *Sesion {
	t.Helper()
	sesion, err := NuevaSesion(sesionID, "res-"+sesionID, tutorID, escolarID, fin.Add(-time.Hour), fin)
	if err != nil {
		t.Fatal(err)
	}
	if err := sesion.RestaurarAsistencia(tutorID, fin.Add(-30*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := sesion.RestaurarAsistencia(escolarID, fin.Add(-30*time.Minute)); err != nil {
		t.Fatal(err)
	}
	return sesion
}

func crearRegistroMinutos(
	t *testing.T,
	repositorio RepositorioMinutos,
	sesionID, tutorID, escolarID string,
	fin time.Time,
) *RegistroMinutos {
	t.Helper()
	sesion := sesionConMinutos(t, sesionID, tutorID, escolarID, fin)
	registro, err := ProponerMinutos(sesion, tutorID, 60, fin.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if err := repositorio.CrearMinutos(context.Background(), registro); err != nil {
		t.Fatal(err)
	}
	return registro
}

// F-06: un GET de minutos (Consultar) no debe persistir una anulacion. La
// lectura debe ser idempotente; la anulacion pertenece a un comando/escritura.
func TestF06ConsultarNoDebeEscribirAlVerReservaCancelada(t *testing.T) {
	ahora := time.Date(2026, 3, 10, 15, 0, 0, 0, time.UTC)
	fin := ahora.Add(-2 * time.Hour)
	minutos := &minutosConContador{RepositorioMinutosMemoria: memoriaSesion.NuevoRepositorioMinutosMemoria()}
	crearRegistroMinutos(t, minutos, "s-f06", "t-f06", "e-f06", fin)

	reservas := &cancelacionesContadas{canceladas: map[string]bool{"res-s-f06": true}}
	servicio := ServicioMinutos{Minutos: minutos, Reservas: reservas}

	if _, err := servicio.Consultar(context.Background(), "s-f06", ahora); err != nil {
		t.Fatalf("la consulta no debia fallar: %v", err)
	}
	if escrituras := minutos.escrituras(); escrituras != 0 {
		t.Fatalf("F-06: Consultar (GET) disparo %d escritura(s) de anulacion; una lectura no debe modificar el registro", escrituras)
	}
}

// F-06: Total tambien es una lectura y no debe anular ni escribir registros.
func TestF06TotalNoDebeEscribirAlVerReservaCancelada(t *testing.T) {
	ahora := time.Date(2026, 3, 10, 15, 0, 0, 0, time.UTC)
	fin := ahora.Add(-2 * time.Hour)
	minutos := &minutosConContador{RepositorioMinutosMemoria: memoriaSesion.NuevoRepositorioMinutosMemoria()}
	crearRegistroMinutos(t, minutos, "s-f06b", "t-f06b", "e-f06b", fin)

	reservas := &cancelacionesContadas{canceladas: map[string]bool{"res-s-f06b": true}}
	servicio := ServicioMinutos{Minutos: minutos, Reservas: reservas}

	if _, err := servicio.Total(context.Background(), "t-f06b", ahora); err != nil {
		t.Fatalf("el total no debia fallar: %v", err)
	}
	if escrituras := minutos.escrituras(); escrituras != 0 {
		t.Fatalf("F-06: Total (lectura) disparo %d escritura(s) de anulacion; el GET no debe tener efectos secundarios", escrituras)
	}
}

// F-17: Total no debe consultar el estado de la reserva una vez por registro.
// La lista completa debe resolverse con una sola consulta/batch.
func TestF17TotalNoDebeConsultarCancelacionesPorFila(t *testing.T) {
	ahora := time.Date(2026, 3, 10, 15, 0, 0, 0, time.UTC)
	minutos := memoriaSesion.NuevoRepositorioMinutosMemoria()
	const registros = 8
	for i := 0; i < registros; i++ {
		fin := ahora.Add(-time.Duration(24+i) * time.Hour)
		crearRegistroMinutos(t, minutos, "s-f17-"+string(rune('a'+i)), "t-f17", "e-f17", fin)
	}

	reservas := &cancelacionesContadas{canceladas: map[string]bool{}}
	servicio := ServicioMinutos{Minutos: minutos, Reservas: reservas}

	if _, err := servicio.Total(context.Background(), "t-f17", ahora); err != nil {
		t.Fatalf("el total no debia fallar: %v", err)
	}
	if llamadas := reservas.totalLlamadas(); llamadas > 1 {
		t.Fatalf("F-17: Total consulto el estado de las reservas %d veces para %d registros (N+1); debe resolverse con una sola consulta", llamadas, registros)
	}
}
