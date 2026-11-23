package sesion

import (
	"context"
	"errors"
	"testing"
	"time"

	memoria "uniapp/internal/adapters/repositories/memory/sesion"
	dominio "uniapp/internal/domain/sesion"
)

// reservasFalsas simula el flujo D: la cancelacion puede ocurrir en cualquier
// momento, incluso despues de verificar los minutos.
type reservasFalsas struct{ canceladas map[string]bool }

func (r *reservasFalsas) Cancelada(_ context.Context, reservaID string) (bool, error) {
	return r.canceladas[reservaID], nil
}

var inicioI17 = time.Date(2026, 11, 10, 15, 0, 0, 0, time.UTC)

func prepararServicioI17(t *testing.T) (ServicioMinutos, *reservasFalsas, time.Time) {
	t.Helper()
	ctx := context.Background()
	sesiones := memoria.NuevoRepositorioMemoria()
	s, err := dominio.NuevaSesion("sesion-a", "reserva-a", "tutor-a", "escolar-a", inicioI17, inicioI17.Add(dominio.DuracionSesion))
	if err != nil {
		t.Fatal(err)
	}
	if err := sesiones.CrearSesion(ctx, s); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"tutor-a", "escolar-a"} {
		if _, err := sesiones.DeclararAsistencia(ctx, "sesion-a", id, inicioI17); err != nil {
			t.Fatal(err)
		}
	}
	reservas := &reservasFalsas{canceladas: map[string]bool{}}
	servicio := ServicioMinutos{Sesiones: sesiones, Minutos: memoria.NuevoRepositorioMinutosMemoria(), Reservas: reservas}
	return servicio, reservas, inicioI17.Add(dominio.DuracionSesion + time.Minute)
}

func TestI17ServicioCancelacionPosteriorAnulaYSumaCero(t *testing.T) {
	ctx := context.Background()
	servicio, reservas, ahora := prepararServicioI17(t)
	if _, err := servicio.Proponer(ctx, "sesion-a", "tutor-a", 45, ahora); err != nil {
		t.Fatal(err)
	}
	if _, err := servicio.Aprobar(ctx, "sesion-a", "coord", ahora); err != nil {
		t.Fatal(err)
	}
	total, err := servicio.Total(ctx, "tutor-a", ahora)
	if err != nil || total.Verificados != 45 {
		t.Fatalf("antes de cancelar debe sumar 45: %d %v", total.Verificados, err)
	}

	reservas.canceladas["reserva-a"] = true

	total, err = servicio.Total(ctx, "tutor-a", ahora)
	if err != nil || total.Verificados != 0 {
		t.Fatalf("tras cancelar debe sumar 0: %d %v", total.Verificados, err)
	}
	// La anulacion quedo persistida y auditada, no solo calculada.
	registro, err := servicio.Minutos.MinutosPorSesion(ctx, "sesion-a")
	if err != nil {
		t.Fatal(err)
	}
	auditoria := registro.Auditoria()
	if registro.Estado() != dominio.MinutosAnulados || auditoria[len(auditoria)-1].Accion() != dominio.AccionAnular {
		t.Fatal("la anulacion debe persistirse con auditoria")
	}
	if _, err := servicio.Ajustar(ctx, "sesion-a", "coord", 60, "x", ahora); !errors.Is(err, ErrSesionCancelada) {
		t.Fatalf("no se debe poder ajustar una sesion cancelada: %v", err)
	}
}

func TestI17ServicioPendientesExcluyeCanceladas(t *testing.T) {
	ctx := context.Background()
	servicio, reservas, ahora := prepararServicioI17(t)
	if _, err := servicio.Proponer(ctx, "sesion-a", "tutor-a", 30, ahora); err != nil {
		t.Fatal(err)
	}
	pendientes, err := servicio.Pendientes(ctx, ahora)
	if err != nil || len(pendientes) != 1 {
		t.Fatalf("esperado 1 pendiente: %d %v", len(pendientes), err)
	}
	reservas.canceladas["reserva-a"] = true
	pendientes, err = servicio.Pendientes(ctx, ahora)
	if err != nil || len(pendientes) != 0 {
		t.Fatalf("una cancelada no debe quedar en la bandeja: %d %v", len(pendientes), err)
	}
}

func TestI17ServicioProponerRechazaCanceladaYErroresDelPuerto(t *testing.T) {
	ctx := context.Background()
	servicio, reservas, ahora := prepararServicioI17(t)
	reservas.canceladas["reserva-a"] = true
	if _, err := servicio.Proponer(ctx, "sesion-a", "tutor-a", 30, ahora); !errors.Is(err, ErrSesionCancelada) {
		t.Fatalf("esperado ErrSesionCancelada: %v", err)
	}
	if _, err := servicio.Minutos.MinutosPorSesion(ctx, "sesion-a"); !errors.Is(err, ErrMinutosNoEncontrados) {
		t.Fatal("no debe crearse registro para una reserva cancelada")
	}
}

func TestI17ServicioRechazoNoDejaCambios(t *testing.T) {
	ctx := context.Background()
	servicio, _, ahora := prepararServicioI17(t)
	if _, err := servicio.Proponer(ctx, "sesion-a", "tutor-a", 30, ahora); err != nil {
		t.Fatal(err)
	}
	if _, err := servicio.Aprobar(ctx, "sesion-a", "tutor-a", ahora); !errors.Is(err, ErrAutoaprobacion) {
		t.Fatalf("esperado ErrAutoaprobacion: %v", err)
	}
	registro, err := servicio.Minutos.MinutosPorSesion(ctx, "sesion-a")
	if err != nil || registro.Estado() != dominio.MinutosPropuestos || len(registro.Auditoria()) != 1 {
		t.Fatal("un rechazo no debe modificar el registro")
	}
}
