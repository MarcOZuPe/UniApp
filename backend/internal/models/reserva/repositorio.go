package reserva

import "context"

// Repositorio guarda solo las tablas del flujo D. La oferta y los usuarios se consultan
// por su API o por este modelo compartido, nunca escribiendo en sus tablas.
type Repositorio interface {
	Crear(ctx context.Context, reserva *Reserva) (*Reserva, error)
	PorID(ctx context.Context, id string) (*Reserva, error)
	PorEscolar(ctx context.Context, escolarID string) ([]Reserva, error)
	SlotTomado(ctx context.Context, slotID string) (bool, error)
}
