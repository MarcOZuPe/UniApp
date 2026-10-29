package acceso

import (
	"net/http"

	vista "uniapp/internal/adapters/presentation/acceso"
)

type ConfirmacionControlador struct{ *controladorBase }

type confirmacionCorreoJSON struct {
	Token string `json:"token"`
}

type reenvioConfirmacionJSON struct {
	Email string `json:"email"`
}

func (controlador *ConfirmacionControlador) Montar(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/confirmar", controlador.confirmar)
	mux.HandleFunc("POST /api/auth/reenviar-confirmacion", controlador.reenviarConfirmacion)
}

func (controlador *ConfirmacionControlador) confirmar(writer http.ResponseWriter, request *http.Request) {
	var payload confirmacionCorreoJSON
	if !controlador.decodificarCuerpo(writer, request, &payload) {
		return
	}

	if err := controlador.servicio.Confirmacion.ConfirmarCorreo(request.Context(), payload.Token); err != nil {
		controlador.responderError(writer, err)
		return
	}

	controlador.renderizarRespuestaConfirmacion(writer)
}

func (controlador *ConfirmacionControlador) reenviarConfirmacion(writer http.ResponseWriter, request *http.Request) {
	var payload reenvioConfirmacionJSON
	if !controlador.decodificarCuerpo(writer, request, &payload) {
		return
	}

	if err := controlador.servicio.Confirmacion.ReenviarConfirmacion(request.Context(), payload.Email); err != nil {
		controlador.responderError(writer, err)
		return
	}

	controlador.renderizarRespuestaReenvio(writer)
}

func (controlador *ConfirmacionControlador) renderizarRespuestaConfirmacion(writer http.ResponseWriter) {
	vista.EscribirJSON(writer, http.StatusOK, map[string]string{
		"mensaje": "Correo confirmado. Ya puedes ingresar.",
	})
}

func (controlador *ConfirmacionControlador) renderizarRespuestaReenvio(writer http.ResponseWriter) {
	vista.EscribirJSON(writer, http.StatusAccepted, map[string]string{
		"mensaje": "Si el correo está registrado y sin confirmar, recibirás un nuevo enlace.",
	})
}
