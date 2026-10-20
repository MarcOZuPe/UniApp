package reserva

import (
	"context"
	"time"
)

// Servicio aplica las reglas del flujo D. El reloj es inyectable para probar las 48 horas
// sin esperar: el servidor es la autoridad temporal (RN-03).
type Servicio struct {
	repositorio Repositorio
	oferta      PuertoOferta
	ahora       func() time.Time
}

type Dependencias struct {
	Repositorio Repositorio
	Oferta      PuertoOferta
}

func NuevoServicio(dependencias Dependencias) *Servicio {
	return &Servicio{
		repositorio: dependencias.Repositorio,
		oferta:      dependencias.Oferta,
		ahora:       func() time.Time { return time.Now().UTC() },
	}
}

func (s *Servicio) FijarReloj(f func() time.Time) { s.ahora = f }

func (s *Servicio) Ahora() time.Time { return s.ahora().UTC() }

// Slots devuelve la oferta con el motivo por el que cada horario no puede reservarse.
// La vista solo pinta el resultado; no recalcula las 48 horas.
func (s *Servicio) Slots(ctx context.Context) ([]Disposicion, error) {
	slots, err := s.oferta.Slots(ctx)
	if err != nil {
		return nil, err
	}
	ahora := s.Ahora()
	disposiciones := make([]Disposicion, 0, len(slots))
	for _, slot := range slots {
		motivo := MotivoPorTiempo(slot, ahora)
		if motivo == "" {
			tomado, err := s.repositorio.SlotTomado(ctx, slot.ID)
			if err != nil {
				return nil, err
			}
			if tomado {
				motivo = ErrSlotOcupado.Error()
			}
		}
		disposiciones = append(disposiciones, Disposicion{Slot: slot, Motivo: motivo})
	}
	return disposiciones, nil
}

func (s *Servicio) Reservar(ctx context.Context, escolarID string, datos DatosReserva) (*Reserva, error) {
	slot, err := s.slotElegido(ctx, datos.SlotID)
	if err != nil {
		return nil, err
	}
	reserva, err := NuevaReserva(escolarID, datos, slot, s.Ahora())
	if err != nil {
		return nil, err
	}
	return s.repositorio.Crear(ctx, reserva)
}

func (s *Servicio) slotElegido(ctx context.Context, slotID string) (Slot, error) {
	if slotID == "" {
		return Slot{}, ErrValidacion{Campos: map[string]string{"slot_id": "elige un horario disponible"}}
	}
	return s.oferta.SlotPorID(ctx, slotID)
}

// ReservaDe solo devuelve la reserva a su dueño: otra persona recibe ErrNoAutorizado.
func (s *Servicio) ReservaDe(ctx context.Context, id, escolarID string) (*Reserva, error) {
	reserva, err := s.repositorio.PorID(ctx, id)
	if err != nil {
		return nil, err
	}
	if reserva.EscolarID != escolarID {
		return nil, ErrNoAutorizado
	}
	return reserva, nil
}

func (s *Servicio) ReservasDe(ctx context.Context, escolarID string) ([]Reserva, error) {
	return s.repositorio.PorEscolar(ctx, escolarID)
}
