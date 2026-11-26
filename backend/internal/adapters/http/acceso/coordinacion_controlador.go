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
		controlador.proteger(controlador.mostrarPanel, modelo.PermisoCoordinacionAcceder),
	)
	multiplexor.Handle(
		"GET /api/coordinacion/coordinadores",
		controlador.proteger(controlador.listarCoordinadores, modelo.PermisoCoordinadoresGestionar),
	)
	multiplexor.Handle(
		"POST /api/coordinacion/coordinadores",
		controlador.proteger(controlador.crearCoordinador, modelo.PermisoCoordinadoresGestionar),
	)
	multiplexor.Handle(
		"DELETE /api/coordinacion/coordinadores/{id}",
		controlador.proteger(controlador.eliminarCoordinador, modelo.PermisoCoordinadoresGestionar),
	)
}

func (controlador *CoordinacionControlador) proteger(
	manejador http.HandlerFunc,
	permiso modelo.Permiso,
) http.Handler {
	return Encadenar(
		http.HandlerFunc(manejador),
		controlador.Autenticar(),
		RequierePermiso(permiso),
	)
}

func (controlador *CoordinacionControlador) mostrarPanel(response http.ResponseWriter, request *http.Request) {
	usuario := UsuarioDe(request.Context())
	panel := vista.PresentarPanelCoordinacion(usuario)

	vista.EscribirJSON(response, http.StatusOK, panel)
}

type entradaCoordinador struct {
	Email     string `json:"email"`
	NombreUso string `json:"nombre_uso"`
	Clave     string `json:"clave"`
}

func (controlador *CoordinacionControlador) listarCoordinadores(response http.ResponseWriter, request *http.Request) {
	usuarios, err := controlador.servicio.Coordinadores.Listar(request.Context(), UsuarioDe(request.Context()))
	if err != nil {
		controlador.responderError(response, err)
		return
	}
	coordinadores := make([]vista.UsuarioJSON, 0, len(usuarios))
	for _, usuario := range usuarios {
		coordinadores = append(coordinadores, vista.PresentarUsuario(usuario))
	}
	vista.EscribirJSON(response, http.StatusOK, map[string]any{"coordinadores": coordinadores})
}

func (controlador *CoordinacionControlador) crearCoordinador(response http.ResponseWriter, request *http.Request) {
	var entrada entradaCoordinador
	if !controlador.decodificarCuerpo(response, request, &entrada) {
		return
	}
	usuario, err := controlador.servicio.Coordinadores.Crear(
		request.Context(),
		UsuarioDe(request.Context()),
		modelo.SolicitudRegistro{
			Email: entrada.Email, NombreUso: entrada.NombreUso, Clave: entrada.Clave,
		},
	)
	if err != nil {
		controlador.responderError(response, err)
		return
	}
	vista.EscribirJSON(response, http.StatusCreated, vista.PresentarUsuario(usuario))
}

func (controlador *CoordinacionControlador) eliminarCoordinador(response http.ResponseWriter, request *http.Request) {
	if err := controlador.servicio.Coordinadores.Eliminar(
		request.Context(),
		UsuarioDe(request.Context()),
		request.PathValue("id"),
	); err != nil {
		controlador.responderError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}
