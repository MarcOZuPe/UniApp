package reserva

import (
	"context"
	"testing"
	"time"
)

func TestPorTutorSoloDevuelveReservasAsignadas(t *testing.T) {
	ctx := context.Background()
	repo := NuevoRepositorioMemoria()

	ahora := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)

	casos := []struct {
		escolarID string
		tutorID   string
		slotID    string
		inicio    time.Time
	}{
		{"escolar-1", "tutor-1", "slot-1", ahora.Add(72 * time.Hour)},
		{"escolar-2", "tutor-1", "slot-2", ahora.Add(96 * time.Hour)},
		{"escolar-3", "tutor-2", "slot-3", ahora.Add(120 * time.Hour)},
	}

	for _, caso := range casos {
		reserva := &Reserva{
			EscolarID: caso.escolarID,
			TutorID:   caso.tutorID,
			SlotID:    caso.slotID,
			Curso:     "Matematicas",
			Inicio:    caso.inicio,
			Fin:       caso.inicio.Add(time.Hour),
			Estado:    EstadoConfirmada,
			CreadaEn:  ahora,
		}

		_, _, err := repo.ReservarTransaccional(
			ctx, reserva, LimitesPorDefecto,
		)
		if err != nil {
			t.Fatalf("No se pudo crear reserva: %v", err)
		}
	}

	reservas, err := repo.PorTutor(ctx, "tutor-1")
	if err != nil {
		t.Fatalf("PorTutor fallo: %v", err)
	}

	if len(reservas) != 2 {
		t.Fatalf("Se esperaban 2 reservas, se obtuvieron %d", len(reservas))
	}

	for _, reserva := range reservas {
		if reserva.TutorID != "tutor-1" {
			t.Fatalf("Se devolvio una reserva de otro tutor: %s", reserva.TutorID)
		}
	}

	if reservas[0].Inicio.After(reservas[1].Inicio) {
		t.Fatal("Las reservas no estan ordenadas cronologicamente")
	}

	vacias, err := repo.PorTutor(ctx, "tutor-inexistente")
	if err != nil {
		t.Fatalf("Error consultando tutor sin reservas: %v", err)
	}
	if len(vacias) != 0 {
		t.Fatalf("Un tutor sin reservas obtuvo %d resultados", len(vacias))
	}
}
