package acceso

import (
	"net/http"
	"strconv"

	vista "uniapp/internal/adapters/presentation/acceso"
	modelo "uniapp/internal/application/acceso"
)

type AdministracionControlador struct{ *controladorBase }

func (controlador *AdministracionControlador) Montar(multiplexor *http.ServeMux) {
	proteger := func(manejador http.HandlerFunc) http.Handler {
		return Encadenar(
			http.HandlerFunc(manejador),
			controlador.Autenticar(),
			RequierePermiso(modelo.PermisoAdministracionAcceder),
		)
	}
	multiplexor.Handle("GET /api/administracion/coordinadores-generales", proteger(controlador.listarGenerales))
	multiplexor.Handle("POST /api/administracion/coordinadores-generales", proteger(controlador.crearGeneral))
	multiplexor.Handle("DELETE /api/administracion/coordinadores-generales/{id}", proteger(controlador.eliminarGeneral))
	multiplexor.Handle("GET /api/administracion/logs", proteger(controlador.logs))
}

func (controlador *AdministracionControlador) listarGenerales(response http.ResponseWriter, request *http.Request) {
	usuarios, err := controlador.servicio.Administracion.ListarGenerales(
		request.Context(), UsuarioDe(request.Context()),
	)
	if err != nil {
		controlador.responderError(response, err)
		return
	}
	resultado := make([]vista.UsuarioJSON, 0, len(usuarios))
	for _, usuario := range usuarios {
		resultado = append(resultado, vista.PresentarUsuario(usuario))
	}
	vista.EscribirJSON(response, http.StatusOK, map[string]any{
		"coordinadores_generales": resultado,
	})
}

type entradaNuevoGeneral struct {
	Email      string `json:"email"`
	NombreUso  string `json:"nombre_uso"`
	Clave      string `json:"clave"`
	FacultadID string `json:"facultad_id"`
}

func (controlador *AdministracionControlador) crearGeneral(response http.ResponseWriter, request *http.Request) {
	var entrada entradaNuevoGeneral
	if !controlador.decodificarCuerpo(response, request, &entrada) {
		return
	}
	usuario, err := controlador.servicio.Administracion.CrearGeneral(
		request.Context(),
		modelo.OrigenSolicitud{Actor: UsuarioDe(request.Context()), IP: ipCliente(request)},
		modelo.DatosNuevoGeneral{
			Email: entrada.Email, NombreUso: entrada.NombreUso,
			Clave: entrada.Clave, FacultadID: entrada.FacultadID,
		},
	)
	if err != nil {
		controlador.responderError(response, err)
		return
	}
	vista.EscribirJSON(response, http.StatusCreated, vista.PresentarUsuario(usuario))
}

func (controlador *AdministracionControlador) eliminarGeneral(response http.ResponseWriter, request *http.Request) {
	if err := controlador.servicio.Administracion.EliminarGeneral(
		request.Context(),
		modelo.OrigenSolicitud{Actor: UsuarioDe(request.Context()), IP: ipCliente(request)},
		request.PathValue("id"),
	); err != nil {
		controlador.responderError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (controlador *AdministracionControlador) logs(response http.ResponseWriter, request *http.Request) {
	limite := 0
	if valor := request.URL.Query().Get("limite"); valor != "" {
		convertido, err := strconv.Atoi(valor)
		if err != nil || convertido <= 0 {
			vista.EscribirError(response, vista.RespuestaError{
				Estado: http.StatusBadRequest, Codigo: "limite_invalido",
				Mensaje: "El límite debe ser un número mayor que cero.",
			})
			return
		}
		limite = convertido
	}
	logs, err := controlador.servicio.Administracion.Logs(
		request.Context(), UsuarioDe(request.Context()), limite,
	)
	if err != nil {
		controlador.responderError(response, err)
		return
	}
	vista.EscribirJSON(response, http.StatusOK, map[string]any{
		"logs": vista.PresentarLogsSistema(logs),
	})
}
