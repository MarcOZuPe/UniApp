package observers

import (
	"context"
	"fmt"
	"log"
	"time"

	"uniapp/internal/adapters"
	"uniapp/internal/events"
)

type Auditoria struct{ log *log.Logger }

func NuevaAuditoria(registro *log.Logger) *Auditoria { return &Auditoria{log: registro} }

func (a *Auditoria) ID() string { return "auditoria" }

func (a *Auditoria) Actualizar(evento events.Event) {
	a.log.Printf("auditoria evento=%s usuario=%v rol=%v", evento.Nombre, evento.Payload["usuario_id"], evento.Payload["rol"])
}

type CorreoRecuperacion struct {
	mailer adapters.Mailer
	origen string
}

const tiempoLimiteEnvio = 20 * time.Second

func NuevoCorreoRecuperacion(mailer adapters.Mailer, origenFrontend string) *CorreoRecuperacion {
	return &CorreoRecuperacion{mailer: mailer, origen: origenFrontend}
}

func (c *CorreoRecuperacion) ID() string { return "correo-recuperacion" }

func (c *CorreoRecuperacion) Actualizar(evento events.Event) {
	contexto, cancelar := context.WithTimeout(context.Background(), tiempoLimiteEnvio)
	defer cancelar()
	enlace := fmt.Sprintf("%s/restablecer?token=%v", c.origen, evento.Payload["token"])
	err := c.mailer.Enviar(contexto, adapters.Mensaje{
		Para:   fmt.Sprint(evento.Payload["email"]),
		Asunto: "UniApp · restablecer contraseña",
		Cuerpo: "Para elegir una nueva contraseña abre este enlace (vence en 1 hora):\n" + enlace +
			"\nSi no lo pediste, ignora este mensaje.",
	})
	if err != nil {
		log.Printf("correo de recuperación no enviado: %v", err)
	}
}

type CorreoConfirmacion struct {
	mailer adapters.Mailer
	origen string
}

func NuevoCorreoConfirmacion(mailer adapters.Mailer, origenFrontend string) *CorreoConfirmacion {
	return &CorreoConfirmacion{mailer: mailer, origen: origenFrontend}
}

func (c *CorreoConfirmacion) ID() string { return "correo-confirmacion" }

func (c *CorreoConfirmacion) Actualizar(evento events.Event) {
	contexto, cancelar := context.WithTimeout(context.Background(), tiempoLimiteEnvio)
	defer cancelar()
	enlace := fmt.Sprintf("%s/confirmar?token=%v", c.origen, evento.Payload["token"])
	err := c.mailer.Enviar(contexto, adapters.Mensaje{
		Para:   fmt.Sprint(evento.Payload["email"]),
		Asunto: "UniApp · confirma tu correo",
		Cuerpo: "Para activar tu cuenta abre este enlace (vence en 30 minutos):\n" + enlace +
			"\nSi no creaste esta cuenta, ignora este mensaje.",
	})
	if err != nil {
		log.Printf("correo de confirmación no enviado: %v", err)
	}
}
