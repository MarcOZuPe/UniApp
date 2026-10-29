package sesion

import (
	"context"
	"errors"
	"time"
)

const MaxIntentosOutbox = 5
const PlazoReclamoOutbox = 2 * time.Minute

type Entrega struct {
	ID                                                     int64
	ReservaID, DestinatarioID, Tipo, Canal, Asunto, Cuerpo string
	Intentos                                               int
	EntregadoEn                                            time.Time
}

type RepositorioOutbox interface {
	Reclamar(context.Context, time.Time) (*Entrega, error)
	Resolver(context.Context, Entrega, time.Time, string) error
	CorreoDe(context.Context, string) (string, error)
}

type Bandeja interface {
	NotificacionesDe(context.Context, string) ([]Entrega, error)
}

var ErrReclamoVencido = errors.New("reclamo de notificación vencido")

func EsperaOutbox(intento int) time.Duration {
	if intento < 1 {
		intento = 1
	}
	if intento > MaxIntentosOutbox {
		intento = MaxIntentosOutbox
	}
	return time.Minute * time.Duration(1<<(intento-1))
}
