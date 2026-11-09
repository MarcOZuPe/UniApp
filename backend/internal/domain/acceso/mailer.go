package acceso

import "context"

type Mensaje struct {
	Para   string
	Asunto string
	Cuerpo string
}

type Mailer interface {
	Enviar(contexto context.Context, mensaje Mensaje) error
}
