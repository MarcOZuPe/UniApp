package sesion

import (
	"context"
	"errors"
	"time"
	acceso "uniapp/internal/models/acceso"
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

type ProcesadorOutbox struct {
	Repositorio RepositorioOutbox
	Mailer      acceso.Mailer
	Ahora       func() time.Time
}

// Un intento SMTP por reclamo. Los reintentos viven en PostgreSQL, no en memoria.
func (p ProcesadorOutbox) ProcesarUno(ctx context.Context) (bool, error) {
	ahora := time.Now
	if p.Ahora != nil {
		ahora = p.Ahora
	}
	entrega, err := p.Repositorio.Reclamar(ctx, ahora().UTC())
	if err != nil || entrega == nil {
		return false, err
	}
	codigo := ""
	if entrega.Canal == "correo" {
		correo, err := p.Repositorio.CorreoDe(ctx, entrega.DestinatarioID)
		if err != nil || correo == "" {
			codigo = "destinatario_no_disponible"
		} else {
			envio, cancelar := context.WithTimeout(ctx, 30*time.Second)
			err = p.Mailer.Enviar(envio, acceso.Mensaje{Para: correo, Asunto: entrega.Asunto, Cuerpo: entrega.Cuerpo})
			cancelar()
			if err != nil {
				codigo = "proveedor_no_disponible"
			}
		}
	}
	// Un fallo al guardar deja el lease para recuperación; nunca modifica la reserva.
	return true, p.Repositorio.Resolver(ctx, *entrega, ahora().UTC(), codigo)
}

func EsperaOutbox(intento int) time.Duration {
	if intento < 1 {
		intento = 1
	}
	if intento > MaxIntentosOutbox {
		intento = MaxIntentosOutbox
	}
	return time.Minute * time.Duration(1<<(intento-1))
}

var ErrReclamoVencido = errors.New("reclamo de notificación vencido")
