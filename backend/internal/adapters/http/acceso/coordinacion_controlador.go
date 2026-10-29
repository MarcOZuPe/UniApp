package acceso

import (
	"net/http"

	vista "uniapp/internal/adapters/presentation/acceso"
	modelo "uniapp/internal/application/acceso"
)

type CoordinacionControlador struct{ *controladorBase }

func (controlador *CoordinacionControlador) Montar(multiplexor *http.ServeMux) {
	multiplexor.Handle(
		"GET /api/coordinacion/panel",
		Encadenar(
			http.HandlerFunc(controlador.mostrarPanel),
			controlador.Autenticar(),
			RequierePermiso(modelo.PermisoCoordinacionAcceder),
		),
	)
}

func (controlador *CoordinacionControlador) mostrarPanel(response http.ResponseWriter, request *http.Request) {
	usuario := UsuarioDe(request.Context())
	panel := vista.PresentarPanelCoordinacion(usuario)

	vista.EscribirJSON(response, http.StatusOK, panel)
}

