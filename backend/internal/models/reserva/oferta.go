package reserva

import (
	"context"
	"fmt"
	"time"
)

// PuertoOferta es el contrato del flujo C (Mateo). La reserva consume los horarios por
// este puerto y nunca escribe en las tablas de oferta (ADR 0001, regla 2).
type PuertoOferta interface {
	Slots(ctx context.Context) ([]Slot, error)
	SlotPorID(ctx context.Context, id string) (Slot, error)
}

const (
	TutorSinteticoID     = "tutor-sintetico-i04"
	TutorSinteticoNombre = "Tutor Prueba (sintético)"
)

// OfertaSintetica implementa PuertoOferta con datos ficticios para I04. Los horarios se
// calculan a partir del arranque, así la demo siempre tiene opciones válidas, una dentro
// de las 48 horas y otra fuera del horizonte de 28 días.
type OfertaSintetica struct{ slots []Slot }

// NuevaOfertaSintetica usa el tutor ficticio de la demo. Para que la agenda real de un
// tutor reciba la reserva (I10) se usa NuevaOfertaSinteticaParaTutor con el UUID del
// tutor sembrado; mientras I08 no publique la búsqueda, el horario sigue siendo ficticio.
func NuevaOfertaSintetica(base time.Time) *OfertaSintetica {
	return NuevaOfertaSinteticaParaTutor(base, TutorSinteticoID, TutorSinteticoNombre)
}

func NuevaOfertaSinteticaParaTutor(base time.Time, tutorID, tutorNombre string) *OfertaSintetica {
	catalogo := []struct {
		curso string
		dias  int
		hora  int
	}{
		{"Álgebra", 3, 15},
		{"Aritmética", 4, 10},
		{"Geometría", 4, 16},
		{"Trigonometría", 6, 11},
	}
	slots := make([]Slot, 0, len(catalogo)+2)
	for indice, entrada := range catalogo {
		slots = append(slots, slotSintetico(
			fmt.Sprintf("slot-i04-%02d", indice+1), tutorID, tutorNombre, entrada.curso, horaEnLima(base, entrada.dias, entrada.hora)))
	}
	slots = append(slots,
		slotSintetico("slot-i04-cercano", tutorID, tutorNombre, "Álgebra", horaEnLima(base, 1, 9)),
		slotSintetico("slot-i04-lejano", tutorID, tutorNombre, "Geometría", horaEnLima(base, 30, 15)),
	)
	return &OfertaSintetica{slots: slots}
}

func slotSintetico(id, tutorID, tutorNombre, curso string, inicio time.Time) Slot {
	return Slot{
		ID:          id,
		TutorID:     tutorID,
		TutorNombre: tutorNombre,
		Curso:       curso,
		Inicio:      inicio.UTC(),
		Fin:         inicio.UTC().Add(DuracionSesion),
	}
}

// horaEnLima devuelve las hora:00 de America/Lima dentro de "dias" días.
func horaEnLima(base time.Time, dias, hora int) time.Time {
	local := base.In(ZonaLima)
	medianoche := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, ZonaLima)
	return medianoche.AddDate(0, 0, dias).Add(time.Duration(hora) * time.Hour)
}

func (o *OfertaSintetica) Slots(_ context.Context) ([]Slot, error) {
	copia := make([]Slot, len(o.slots))
	copy(copia, o.slots)
	return copia, nil
}

func (o *OfertaSintetica) SlotPorID(_ context.Context, id string) (Slot, error) {
	for _, slot := range o.slots {
		if slot.ID == id {
			return slot, nil
		}
	}
	return Slot{}, ErrSlotInexistente
}
