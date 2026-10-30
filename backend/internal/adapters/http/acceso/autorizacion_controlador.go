package acceso

import (
	"net/http"

	vista "uniapp/internal/adapters/presentation/acceso"
	modelo "uniapp/internal/application/acceso"
)

type AutorizacionControlador struct{ *controladorBase }

type cambioAutorizacionJSON struct {
	Autorizado *bool `json:"autorizado"`
}

func (controlador *AutorizacionControlador) Montar(mux *http.ServeMux) {
	mux.Handle("POST /api/apoderado/escolares/{escolarID}/autorizacion", Encadenar(
		http.HandlerFunc(controlador.cambiarAutorizacion),
		controlador.Autenticar(),
		RequierePermiso(modelo.PermisoAutorizacionOtorgar),
	))
	mux.Handle("GET /api/apoderado/escolares/{escolarID}/autorizaciones", Encadenar(
		http.HandlerFunc(controlador.historialAutorizaciones),
		controlador.Autenticar(),
		RequierePermiso(modelo.PermisoAutorizacionOtorgar),
	))
}

func (controlador *AutorizacionControlador) cambiarAutorizacion(writer http.ResponseWriter, request *http.Request) {
	var payload cambioAutorizacionJSON
	if !controlador.decodificarCuerpo(writer, request, &payload) {
		return
	}

	if err := controlador.validarPayloadCambio(payload.Autorizado); err != nil {
		controlador.responderError(writer, err)
		return
	}

	solicitud := controlador.compilarSolicitudCambio(request, *payload.Autorizado)
	autorizacion, err := controlador.servicio.Autorizacion.CambiarAutorizacion(request.Context(), solicitud)
	if err != nil {
		controlador.responderError(writer, err)
		return
	}

	controlador.renderizarRespuestaCambio(writer, autorizacion)
}

func (controlador *AutorizacionControlador) historialAutorizaciones(writer http.ResponseWriter, request *http.Request) {
	solicitud := controlador.compilarSolicitudBase(request)
	autorizaciones, err := controlador.servicio.Autorizacion.HistorialAutorizaciones(request.Context(), solicitud)
	if err != nil {
		controlador.responderError(writer, err)
		return
	}

	vista.EscribirJSON(writer, http.StatusOK, map[string]any{
		"autorizaciones": vista.PresentarAutorizaciones(autorizaciones),
	})
}

func (controlador *AutorizacionControlador) validarPayloadCambio(autorizado *bool) error {
	if autorizado == nil {
		return modelo.ErrValidacion{
			Campos: map[string]string{"autorizado": "indica si el escolar puede reservar"},
		}
	}
	return nil
}

func (controlador *AutorizacionControlador) compilarSolicitudCambio(request *http.Request, autorizado bool) modelo.SolicitudCambioAutorizacion {
	return modelo.SolicitudCambioAutorizacion{
		SolicitudAutorizacion: controlador.compilarSolicitudBase(request),
		Autorizado:            autorizado,
	}
}

func (controlador *AutorizacionControlador) compilarSolicitudBase(request *http.Request) modelo.SolicitudAutorizacion {
	return modelo.SolicitudAutorizacion{
		Apoderado: UsuarioDe(request.Context()),
		EscolarID: request.PathValue("escolarID"),
	}
}

func (controlador *AutorizacionControlador) renderizarRespuestaCambio(writer http.ResponseWriter, autorizacion *modelo.Autorizacion) {
	mensaje := "Autorización revocada. El escolar no podrá crear nuevas reservas."
	if autorizacion.Vigente() {
		mensaje = "Autorización otorgada. El escolar puede crear nuevas reservas."
	}

	vista.EscribirJSON(writer, http.StatusOK, map[string]any{
		"mensaje":      mensaje,
		"autorizacion": vista.PresentarAutorizacion(autorizacion),
	})
}
