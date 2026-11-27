package reserva

import (
	"context"
	"errors"
	"testing"
	"time"

	memoriaReserva "uniapp/internal/adapters/repositories/memory/reserva"
	modeloAcceso "uniapp/internal/domain/acceso"
)

type autorizacionVigente struct{}

func (autorizacionVigente) AutorizacionVigente(context.Context, string) (bool, error) {
	return true, nil
}

func servicioParaCancelar(t *testing.T) (*Servicio, *memoriaReserva.RepositorioMemoria) {
	t.Helper()
	base := memoriaReserva.NuevoRepositorioMemoria()
	servicio := NuevoServicio(Dependencias{
		Repositorio:  base,
		Oferta:       NuevaOfertaSintetica(time.Now().UTC()),
		Autorizacion: autorizacionVigente{},
	})
	return servicio, base
}

// I16: cancelar deja la reserva en estado cancelada y guarda el evento en el historial.
func TestI16CancelarReservaConHistorial(t *testing.T) {
	servicio, _ := servicioParaCancelar(t)
	ctx := context.Background()

	reserva, _, err := servicio.Reservar(ctx, "escolar-i16", datosReservaValidos("slot-i04-01"))
	if err != nil {
		t.Fatalf("crear reserva: %v", err)
	}

	cancelada, err := servicio.Cancelar(ctx, "escolar-i16", modeloAcceso.RolEscolar, reserva.ID, "  enfermedad  ")
	if err != nil {
		t.Fatalf("cancelar: %v", err)
	}
	if cancelada.Estado != EstadoCancelada {
		t.Fatalf("estado = %q, se esperaba cancelada", cancelada.Estado)
	}
	if cancelada.CanceladaEn == nil {
		t.Fatal("cancelada_en quedó vacío")
	}
	if cancelada.MotivoCancelacion != "enfermedad" {
		t.Fatalf("motivo = %q, se esperaba enfermedad", cancelada.MotivoCancelacion)
	}

	historial, err := servicio.Historial(ctx, "escolar-i16", modeloAcceso.RolEscolar, reserva.ID)
	if err != nil {
		t.Fatalf("historial: %v", err)
	}
	if len(historial) != 1 {
		t.Fatalf("historial con %d eventos, se esperaba 1", len(historial))
	}
	if historial[0].EstadoNuevo != EstadoCancelada || historial[0].EstadoAnterior != EstadoConfirmada {
		t.Fatalf("evento inesperado: %+v", historial[0])
	}
	if historial[0].ActorID != "escolar-i16" {
		t.Fatalf("actor = %q, se esperaba escolar-i16", historial[0].ActorID)
	}
}

// I16: una reserva ya cancelada no se puede volver a cancelar.
func TestI16NoSePuedeCancelarDosVeces(t *testing.T) {
	servicio, _ := servicioParaCancelar(t)
	ctx := context.Background()

	reserva, _, err := servicio.Reservar(ctx, "escolar-i16-b", datosReservaValidos("slot-i04-01"))
	if err != nil {
		t.Fatalf("crear reserva: %v", err)
	}
	if _, err := servicio.Cancelar(ctx, "escolar-i16-b", modeloAcceso.RolEscolar, reserva.ID, ""); err != nil {
		t.Fatalf("primera cancelación: %v", err)
	}
	if _, err := servicio.Cancelar(ctx, "escolar-i16-b", modeloAcceso.RolEscolar, reserva.ID, ""); !errors.Is(err, ErrReservaNoCancelable) {
		t.Fatalf("segunda cancelación: se esperaba ErrReservaNoCancelable, se obtuvo %v", err)
	}
}

// I16: una sesión que ya empezó no se cancela.
func TestI16NoSeCancelaUnaSesionEmpezada(t *testing.T) {
	reserva := &Reserva{
		ID:     "r-i16-c",
		Inicio: time.Now().UTC().Add(-time.Hour),
		Estado: EstadoConfirmada,
	}
	if err := reserva.Cancelar("", time.Now().UTC()); !errors.Is(err, ErrCancelacionTardia) {
		t.Fatalf("se esperaba ErrCancelacionTardia, se obtuvo %v", err)
	}
}
