package acceso

import (
	"net/http"

	vista "uniapp/internal/adapters/presentation/acceso"
	modelo "uniapp/internal/application/acceso"
)

type RegistroControlador struct{ *controladorBase }

type registroCuentaJSON struct {
	Email     string `json:"email"`
	NombreUso string `json:"nombre_uso"`
	Clave     string `json:"clave"`
	Rol       string `json:"rol"`
}

func (controlador *RegistroControlador) Montar(multiplexor *http.ServeMux) {
	multiplexor.HandleFunc("POST /api/auth/registro", controlador.registro)
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
		Email:     payload.Email,
		NombreUso: payload.NombreUso,
		Clave:     payload.Clave,
		Rol:       modelo.Rol(payload.Rol),
	}
}

func (controlador *RegistroControlador) renderizarRespuestaRegistro(response http.ResponseWriter, usuario *modelo.Usuario) {
	vista.EscribirJSON(response, http.StatusCreated, vista.PresentarUsuario(usuario))
}
