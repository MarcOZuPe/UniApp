package acceso

import "context"

type ConfirmacionListener struct {
	servicio *ConfirmacionService
}

func NuevoConfirmacionListener(servicio *ConfirmacionService) *ConfirmacionListener {
	return &ConfirmacionListener{servicio: servicio}
}

func (listener *ConfirmacionListener) ID() string { return "confirmacion-correo-listener" }

func (listener *ConfirmacionListener) Actualizar(evento Event) {
	if evento.Nombre != UsuarioRegistrado {
		return
	}

	id, ok := evento.Payload["usuario_id"].(string)
	if !ok || id == "" {
		return
	}

	usuario, err := listener.servicio.RepositorioUsuarios.UsuarioPorID(context.Background(), id)
	if err != nil {
		return
	}
	_ = listener.servicio.SolicitarConfirmacion(context.Background(), usuario)
}
