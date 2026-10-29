package sesion

import (
	"context"
	"time"
	acceso "uniapp/internal/domain/acceso"
)

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
