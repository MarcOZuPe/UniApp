package sesion

import (
	"net/http"

	acceso "uniapp/internal/adapters/http/acceso"
	comun "uniapp/internal/adapters/presentation/acceso"
	vista "uniapp/internal/adapters/presentation/sesion"
	modeloAcceso "uniapp/internal/application/acceso"
	modelo "uniapp/internal/application/sesion"
)

type ControladorNotificaciones struct {
	bandeja modelo.Bandeja
}

func RutasNotificaciones(bandeja modelo.Bandeja, autenticar func(http.Handler) http.Handler) http.Handler {
	controlador := &ControladorNotificaciones{bandeja: bandeja}
	multiplexor := http.NewServeMux()
	multiplexor.Handle(
		"GET /api/notificaciones",
		autenticar(http.HandlerFunc(controlador.listar)),
	)
	return multiplexor
}

func (controlador *ControladorNotificaciones) listar(response http.ResponseWriter, request *http.Request) {
	usuario := acceso.UsuarioDe(request.Context())
	if usuario == nil {
		responderError(response, detalleErrorHTTP{
			estado: http.StatusUnauthorized, codigo: "no_autenticado", mensaje: "Debes iniciar sesión.",
		})
		return
	}
	if usuario.Rol != modeloAcceso.RolTutor && usuario.Rol != modeloAcceso.RolEscolar {
		responderError(response, detalleErrorHTTP{
			estado: http.StatusForbidden, codigo: "rol_no_permitido", mensaje: "Tu rol no tiene acceso a notificaciones.",
		})
		return
	}

	notificaciones, err := controlador.bandeja.NotificacionesDe(request.Context(), usuario.ID)
	if err != nil {
		responderError(response, detalleErrorHTTP{
			estado:  http.StatusInternalServerError,
			codigo:  "error_notificaciones",
			mensaje: "No se pudieron consultar las notificaciones.",
		})
		return
	}

	comun.EscribirJSON(response, http.StatusOK, map[string]any{
		"notificaciones": vista.PresentarNotificaciones(notificaciones),
	})
}
