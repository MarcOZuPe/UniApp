package acceso

import (
	"net/http"

	vista "uniapp/internal/adapters/presentation/acceso"
	modelo "uniapp/internal/application/acceso"
)

type VinculacionControlador struct{ *controladorBase }

type solicitudVinculacionJSON struct {
	EmailEscolar string `json:"email_escolar"`
}

type tokenVinculacionJSON struct {
	Token string `json:"token"`
}

func (controlador *VinculacionControlador) Montar(multiplexor *http.ServeMux) {
	multiplexor.Handle("POST /api/apoderado/vinculaciones", Encadenar(
		http.HandlerFunc(controlador.solicitarVinculacion), controlador.Autenticar(),
	))
	multiplexor.Handle("GET /api/apoderado/vinculaciones", Encadenar(
		http.HandlerFunc(controlador.listarVinculaciones), controlador.Autenticar(),
	))
	multiplexor.Handle("POST /api/escolar/vinculaciones/consultar", Encadenar(
		http.HandlerFunc(controlador.consultarVinculacion), controlador.Autenticar(),
	))
	multiplexor.Handle("POST /api/escolar/vinculaciones/aceptar", Encadenar(
		http.HandlerFunc(controlador.aceptarVinculacion), controlador.Autenticar(),
	))
	multiplexor.Handle("GET /api/escolar/apoderado", Encadenar(
		http.HandlerFunc(controlador.miApoderado), controlador.Autenticar(),
	))
}

func (controlador *VinculacionControlador) solicitarVinculacion(response http.ResponseWriter, request *http.Request) {
	var payload solicitudVinculacionJSON
	if !controlador.decodificarCuerpo(response, request, &payload) {
		return
	}

	vinculacion, err := controlador.servicio.Vinculacion.SolicitarVinculacion(
		request.Context(), UsuarioDe(request.Context()), payload.EmailEscolar,
	)
	if err != nil {
		controlador.responderError(response, err)
		return
	}

	controlador.renderizarSolicitudVinculacion(response, vinculacion)
}

func (controlador *VinculacionControlador) listarVinculaciones(response http.ResponseWriter, request *http.Request) {
	vinculaciones, err := controlador.servicio.Vinculacion.VinculacionesDeApoderado(
		request.Context(), UsuarioDe(request.Context()),
	)
	if err != nil {
		controlador.responderError(response, err)
		return
	}

	controlador.renderizarListadoVinculaciones(response, vinculaciones)
}

func (controlador *VinculacionControlador) consultarVinculacion(response http.ResponseWriter, request *http.Request) {
	var payload tokenVinculacionJSON
	if !controlador.decodificarCuerpo(response, request, &payload) {
		return
	}

	vinculacion, err := controlador.servicio.Vinculacion.ConsultarVinculacion(
		request.Context(), UsuarioDe(request.Context()), payload.Token,
	)
	if err != nil {
		controlador.responderError(response, err)
		return
	}

	controlador.renderizarVinculacion(response, vinculacion)
}

func (controlador *VinculacionControlador) aceptarVinculacion(response http.ResponseWriter, request *http.Request) {
	var payload tokenVinculacionJSON
	if !controlador.decodificarCuerpo(response, request, &payload) {
		return
	}

	vinculacion, err := controlador.servicio.Vinculacion.AceptarVinculacion(
		request.Context(), UsuarioDe(request.Context()), payload.Token,
	)
	if err != nil {
		controlador.responderError(response, err)
		return
	}

	controlador.renderizarAceptacionVinculacion(response, vinculacion)
}

func (controlador *VinculacionControlador) miApoderado(response http.ResponseWriter, request *http.Request) {
	vinculacion, err := controlador.servicio.Vinculacion.VinculacionDeEscolar(
		request.Context(), UsuarioDe(request.Context()),
	)
	if err != nil {
		controlador.responderError(response, err)
		return
	}

	controlador.renderizarMiApoderado(response, vinculacion)
}

func (controlador *VinculacionControlador) renderizarSolicitudVinculacion(response http.ResponseWriter, vinculacion *modelo.Vinculacion) {
	vista.EscribirJSON(response, http.StatusAccepted, map[string]any{
		"mensaje":     "Solicitud enviada. El escolar recibirá un enlace para aceptarla.",
		"vinculacion": vista.PresentarVinculacion(vinculacion),
	})
}

func (controlador *VinculacionControlador) renderizarListadoVinculaciones(response http.ResponseWriter, vinculaciones []modelo.Vinculacion) {
	vista.EscribirJSON(response, http.StatusOK, map[string]any{
		"vinculaciones": vista.PresentarVinculaciones(vinculaciones),
	})
}

func (controlador *VinculacionControlador) renderizarVinculacion(response http.ResponseWriter, vinculacion *modelo.Vinculacion) {
	vista.EscribirJSON(response, http.StatusOK, map[string]any{
		"vinculacion": vista.PresentarVinculacion(vinculacion),
	})
}

func (controlador *VinculacionControlador) renderizarAceptacionVinculacion(response http.ResponseWriter, vinculacion *modelo.Vinculacion) {
	vista.EscribirJSON(response, http.StatusOK, map[string]any{
		"mensaje":     "Vinculación aceptada. Tu apoderado ya está registrado.",
		"vinculacion": vista.PresentarVinculacion(vinculacion),
	})
}

func (controlador *VinculacionControlador) renderizarMiApoderado(response http.ResponseWriter, vinculacion *modelo.Vinculacion) {
	var presentada any
	if vinculacion != nil {
		presentada = vista.PresentarVinculacion(vinculacion)
	}
	vista.EscribirJSON(response, http.StatusOK, map[string]any{"vinculacion": presentada})
}
