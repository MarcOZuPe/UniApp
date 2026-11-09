package acceso

import (
	"net/http"

	vista "uniapp/internal/adapters/presentation/acceso"
	modelo "uniapp/internal/application/acceso"
)

type RecuperacionControlador struct{ *controladorBase }

type recuperacionClaveJSON struct {
	Email string `json:"email"`
}

type restablecimientoClaveJSON struct {
	Token string `json:"token"`
	Clave string `json:"clave"`
}

func (controlador *RecuperacionControlador) Montar(multiplexor *http.ServeMux) {
	multiplexor.HandleFunc("POST /api/auth/recuperar", controlador.recuperar)
	multiplexor.HandleFunc("POST /api/auth/restablecer", controlador.restablecer)
}

func (controlador *RecuperacionControlador) recuperar(response http.ResponseWriter, request *http.Request) {
	var payload recuperacionClaveJSON
	if !controlador.decodificarCuerpo(response, request, &payload) {
		return
	}
	if err := controlador.servicio.Recuperacion.SolicitarRecuperacion(request.Context(), payload.Email); err != nil {
		controlador.responderError(response, err)
		return
	}
	controlador.renderizarRespuestaRecuperacion(response)
}

func (controlador *RecuperacionControlador) restablecer(response http.ResponseWriter, request *http.Request) {
	var payload restablecimientoClaveJSON
	if !controlador.decodificarCuerpo(response, request, &payload) {
		return
	}
	solicitud := modelo.SolicitudRestablecimiento{
		Token:      payload.Token,
		NuevaClave: payload.Clave,
	}
	if err := controlador.servicio.Recuperacion.RestablecerClave(request.Context(), solicitud); err != nil {
		controlador.responderError(response, err)
		return
	}
	controlador.renderizarRespuestaRestablecimiento(response)
}

func (controlador *RecuperacionControlador) renderizarRespuestaRecuperacion(response http.ResponseWriter) {
	vista.EscribirJSON(response, http.StatusAccepted, map[string]string{
		"mensaje": "Si el correo está registrado, recibirás un enlace para restablecer tu contraseña.",
	})
}

func (controlador *RecuperacionControlador) renderizarRespuestaRestablecimiento(response http.ResponseWriter) {
	vista.EscribirJSON(response, http.StatusOK, map[string]string{
		"mensaje": "Contraseña actualizada. Inicia sesión con la nueva.",
	})
}
