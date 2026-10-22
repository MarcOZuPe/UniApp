package reserva

import "context"

// Repositorio guarda solo las tablas del flujo D. La oferta y los usuarios se consultan
// por su API o por este modelo compartido, nunca escribiendo en sus tablas.
type Repositorio interface {
	// ReservarTransaccional crea la reserva y aplica D07 dentro de una transacción.
	// Devuelve creada=false cuando una clave de idempotencia repetida gana la carrera.
	ReservarTransaccional(ctx context.Context, reserva *Reserva, limites Limites) (*Reserva, bool, error)
	PorClaveIdempotencia(ctx context.Context, escolarID, clave string) (*Reserva, error)
	PorID(ctx context.Context, id string) (*Reserva, error)
	PorEscolar(ctx context.Context, escolarID string) ([]Reserva, error)
	SlotTomado(ctx context.Context, slotID string) (bool, error)

	ContarAdjuntos(ctx context.Context, reservaID string) (int, error)
	AgregarAdjunto(ctx context.Context, reservaID string, adjunto *Adjunto) (*Adjunto, error)
	AdjuntosDe(ctx context.Context, reservaID string) ([]Adjunto, error)
	AdjuntoPorID(ctx context.Context, reservaID, adjuntoID string) (*Adjunto, error)
	ActualizarAdjunto(ctx context.Context, adjuntoID string, estado EstadoAdjunto, motivo string) error
}
