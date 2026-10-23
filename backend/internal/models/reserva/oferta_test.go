package reserva

import (
	"context"
	"testing"
	"time"
)

func TestOfertaSinteticaParaTutor(t *testing.T) {
	base := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	const tutorID = "0f8fad5b-d9cb-469f-a165-70867728950e"
	oferta := NuevaOfertaSinteticaParaTutor(base, tutorID, "Tutor Real")
	slots, err := oferta.Slots(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(slots) != 6 {
		t.Fatalf("se esperaban 6 horarios: %d", len(slots))
	}
	for _, slot := range slots {
		if slot.TutorID != tutorID || slot.TutorNombre != "Tutor Real" {
			t.Fatalf("el horario debe conservar el tutor real: %+v", slot)
		}
		if slot.Fin.Sub(slot.Inicio) != DuracionSesion {
			t.Fatalf("cada sesión dura 60 minutos: %+v", slot)
		}
	}
	primero, err := oferta.SlotPorID(context.Background(), "slot-i04-01")
	if err != nil || primero.TutorID != tutorID {
		t.Fatalf("SlotPorID debe devolver el tutor real: %v %+v", err, primero)
	}

	defecto, err := NuevaOfertaSintetica(base).Slots(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if defecto[0].TutorID != TutorSinteticoID {
		t.Fatalf("la oferta por defecto conserva el tutor ficticio: %+v", defecto[0])
	}
}
