package acceso

import (
	"net/http"

	vista "uniapp/internal/adapters/presentation/acceso"
	modelo "uniapp/internal/application/acceso"
)

type RegistroControlador struct{ *controladorBase }

type registroCuentaJSON struct {
	Email      string `json:"email"`
	NombreUso  string `json:"nombre_uso"`
	Clave      string `json:"clave"`
	Rol        string `json:"rol"`
	FacultadID string `json:"facultad_id"`
}

func (controlador *RegistroControlador) Montar(multiplexor *http.ServeMux) {
	multiplexor.HandleFunc("POST /api/auth/registro", controlador.registro)
	multiplexor.HandleFunc("GET /api/facultades", controlador.facultades)
}

func (controlador *RegistroControlador) facultades(response http.ResponseWriter, request *http.Request) {
	facultades, err := controlador.servicio.Registro.Facultades(request.Context())
	if err != nil {
		controlador.responderError(response, err)
		return
	}
	vista.EscribirJSONCacheable(response, request, http.StatusOK, map[string]any{
		"facultades": vista.PresentarFacultades(facultades),
	}, 3600)
}

func (controlador *RegistroControlador) registro(response http.ResponseWriter, request *http.Request) {
	var payload registroCuentaJSON
	if !controlador.decodificarCuerpo(response, request, &payload) {
		return
	}

	solicitud := controlador.compilarSolicitudRegistro(payload)
	usuario, err := controlador.servicio.Registro.Registrar(request.Context(), solicitud)
	if err != nil {
		controlador.responderError(response, err)
		return
	}

	controlador.renderizarRespuestaRegistro(response, usuario)
}

func (controlador *RegistroControlador) compilarSolicitudRegistro(payload registroCuentaJSON) modelo.SolicitudRegistro {
	return modelo.SolicitudRegistro{
		Email:      payload.Email,
		NombreUso:  payload.NombreUso,
		Clave:      payload.Clave,
		Rol:        modelo.Rol(payload.Rol),
		FacultadID: payload.FacultadID,
	}
}

func (controlador *RegistroControlador) renderizarRespuestaRegistro(response http.ResponseWriter, usuario *modelo.Usuario) {
	vista.EscribirJSON(response, http.StatusCreated, vista.PresentarUsuario(usuario))
}
