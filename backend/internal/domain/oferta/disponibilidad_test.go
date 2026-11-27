package oferta

import (
	"testing"
	"time"
)

func TestSlotSemanalValidar(t *testing.T) {
	valido := SlotSemanal{TutorID: "t-1", DiaSemana: time.Monday, Inicio: 9 * time.Hour, Fin: 10 * time.Hour}
	if err := valido.Validar(); err != nil {
		t.Fatalf("slot válido rechazado: %v", err)
	}
	casos := []SlotSemanal{
		{},
		{TutorID: "t-1", DiaSemana: time.Monday, Inicio: 9 * time.Hour, Fin: 10*time.Hour + time.Minute},
		{TutorID: "t-1", DiaSemana: time.Monday, Inicio: 23*time.Hour + 30*time.Minute, Fin: 24*time.Hour + 30*time.Minute},
		{TutorID: "t-1", DiaSemana: time.Monday, Inicio: -time.Hour, Fin: time.Hour},
		{TutorID: "t-1", DiaSemana: time.Monday, Inicio: 9*time.Hour + 30*time.Minute + 15*time.Second, Fin: 10*time.Hour + 30*time.Minute + 15*time.Second},
	}
	for _, caso := range casos {
		if err := caso.Validar(); err == nil {
			t.Fatalf("slot inválido aceptado: %+v", caso)
		}
	}
}

func TestValidarSlotsDetectaSolapamiento(t *testing.T) {
	slots := []SlotSemanal{
		{TutorID: "t-1", DiaSemana: time.Monday, Inicio: 9 * time.Hour, Fin: 10 * time.Hour},
		{TutorID: "t-1", DiaSemana: time.Monday, Inicio: 9*time.Hour + 30*time.Minute, Fin: 10*time.Hour + 30*time.Minute},
	}
	if err := ValidarSlots(slots); err == nil {
		t.Fatal("se esperaba solapamiento")
	}
	slots[1].TutorID = "t-2"
	if err := ValidarSlots(slots); err != nil {
		t.Fatalf("dos tutores pueden solaparse: %v", err)
	}
}

func TestCupoSemanalYCambios(t *testing.T) {
	cupo := CupoSemanal{TutorID: "t-1", Minutos: 180}
	if err := cupo.Validar(); err != nil {
		t.Fatalf("cupo válido rechazado: %v", err)
	}
	if err := (CupoSemanal{TutorID: "t-1", Minutos: 250}).Validar(); err == nil {
		t.Fatal("250 no es múltiplo de 60")
	}
	if err := (CupoSemanal{TutorID: "t-1", Minutos: 120}).Validar(); err == nil {
		t.Fatal("120 está bajo el mínimo")
	}
	if err := ValidarCambioCupoSemanal(cupo, CupoSemanal{TutorID: "t-1", Minutos: 180}, 2); err != nil {
		t.Fatalf("mantener cupo con compromisos no debería fallar: %v", err)
	}
	if err := ValidarCambioCupoSemanal(cupo, CupoSemanal{TutorID: "t-1", Minutos: 120}, 1); err == nil {
		t.Fatal("reducir con compromisos activos debe fallar")
	}
	if err := ValidarCambioCupoSemanal(cupo, CupoSemanal{TutorID: "t-2", Minutos: 180}, 0); err == nil {
		t.Fatal("cambiar de tutor debe fallar")
	}
}

func TestHorariosYZona(t *testing.T) {
	if minutos, err := MinutosDelDia(9, 30); err != nil || minutos != 9*time.Hour+30*time.Minute {
		t.Fatalf("minutos inesperados: %v err=%v", minutos, err)
	}
	if _, err := MinutosDelDia(24, 0); err == nil {
		t.Fatal("hora inválida debe rechazarse")
	}
	if _, err := MinutosDelDia(10, 60); err == nil {
		t.Fatal("minuto inválido debe rechazarse")
	}
	if HorarioDesdeMinutos(90) != 90*time.Minute {
		t.Fatal("conversión de minutos incorrecta")
	}
	momento := time.Date(2026, 11, 26, 18, 0, 0, 0, time.UTC)
	lima, err := EnZonaHorariaLima(momento)
	if err != nil || lima.Hour() != 13 {
		t.Fatalf("zona Lima inesperada: %v err=%v", lima, err)
	}
	if _, err := AhoraEnLima(); err != nil {
		t.Fatalf("AhoraEnLima falló: %v", err)
	}
}
