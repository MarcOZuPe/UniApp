package acceso

import "context"

type Mensaje struct{ Para, Asunto, Cuerpo string }

type Mailer interface {
	Enviar(ctx context.Context, mensaje Mensaje) error
}
