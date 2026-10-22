package oferta

import (
	"errors"
	"testing"
	"time"
)

func TestValidarSlotSemanal(t *testing.T) {
	inicio, err := MinutosDelDia(9, 30)
	if err != nil {
		t.Fatal(err)
	}
	slot := SlotSemanal{
		TutorID: "tutor-1", DiaSemana: time.Monday,
		Inicio: inicio, Fin: inicio + DuracionSlot,
	}
	if err := slot.Validar(); err != nil {
		t.Fatalf("slot válido: %v", err)
	}
	slot.Fin--
	if !errors.Is(slot.Validar(), ErrSlotInvalido) {
		t.Fatal("debe rechazar duraciones distintas de 60 minutos")
	}
}

func TestValidarSlotsIntervaloSemiabierto(t *testing.T) {
	primero := SlotSemanal{
		TutorID: "tutor-1", DiaSemana: time.Monday,
		Inicio: 9 * time.Hour, Fin: 10 * time.Hour,
	}
	adyacente := SlotSemanal{
		TutorID: "tutor-1", DiaSemana: time.Monday,
		Inicio: 10 * time.Hour, Fin: 11 * time.Hour,
	}
	if err := ValidarSlots([]SlotSemanal{primero, adyacente}); err != nil {
		t.Fatalf("los intervalos semiabiertos adyacentes no se solapan: %v", err)
	}
	adyacente.Inicio = 9*time.Hour + 30*time.Minute
	adyacente.Fin = 10*time.Hour + 30*time.Minute
	if !errors.Is(ValidarSlots([]SlotSemanal{primero, adyacente}), ErrSolapamiento) {
		t.Fatal("debe rechazar slots solapados del mismo tutor")
	}
}

func TestValidarCambioCupoSemanal(t *testing.T) {
	actual := CupoSemanal{TutorID: "tutor-1", Minutos: 360}
	nuevo := CupoSemanal{TutorID: "tutor-1", Minutos: 300}
	if !errors.Is(ValidarCambioCupoSemanal(actual, nuevo, 1), ErrReduccionConActivos) {
		t.Fatal("no debe permitir reducir el cupo con compromisos activos")
	}
	if err := ValidarCambioCupoSemanal(actual, nuevo, 0); err != nil {
		t.Fatalf("debe permitir reducirlo sin compromisos activos: %v", err)
	}
	nuevo.Minutos = 330
	if !errors.Is(nuevo.Validar(), ErrCupoInvalido) {
		t.Fatal("el cupo debe ser múltiplo de 60")
	}
}

func TestEnZonaHorariaLima(t *testing.T) {
	momento := time.Date(2026, 10, 19, 15, 0, 0, 0, time.UTC)
	local, err := EnZonaHorariaLima(momento)
	if err != nil {
		t.Fatal(err)
	}
	if local.Hour() != 10 || local.Minute() != 0 || local.Location().String() != ZonaHoraria {
		t.Fatalf("hora de Lima inesperada: %s", local)
	}
}

func TestMinutosDelDia(t *testing.T) {
	hora, err := MinutosDelDia(23, 59)
	if err != nil || hora != 23*time.Hour+59*time.Minute {
		t.Fatalf("conversión de hora incorrecta: %v, %v", hora, err)
	}
	if _, err := MinutosDelDia(24, 0); !errors.Is(err, ErrSlotInvalido) {
		t.Fatal("debe rechazar horas fuera del día")
	}
}
