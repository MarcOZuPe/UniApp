package acceso

import (
	"net/http"

	vista "uniapp/internal/adapters/presentation/acceso"
	modelo "uniapp/internal/application/acceso"
)

type SesionControlador struct{ *controladorBase }

type ingresoJSON struct {
	Email string `json:"email"`
	Clave string `json:"clave"`
}

func (controlador *SesionControlador) Montar(multiplexor *http.ServeMux) {
	multiplexor.HandleFunc("POST /api/auth/ingreso", controlador.ingreso)
	multiplexor.Handle("POST /api/auth/salida", Encadenar(http.HandlerFunc(controlador.salida), controlador.Autenticar()))
	multiplexor.Handle("GET /api/auth/yo", Encadenar(http.HandlerFunc(controlador.yo), controlador.Autenticar()))
}

func (controlador *SesionControlador) ingreso(response http.ResponseWriter, request *http.Request) {
	var payload ingresoJSON
	if !controlador.decodificarCuerpo(response, request, &payload) {
		return
	}

	credenciales := controlador.compilarCredenciales(payload)
	token, usuario, err := controlador.servicio.Sesion.Ingresar(request.Context(), credenciales)
	if err != nil {
		controlador.responderError(response, err)
		return
	}

	controlador.renderizarIngreso(response, token, usuario)
}

func (controlador *SesionControlador) salida(response http.ResponseWriter, request *http.Request) {
	_ = controlador.servicio.Sesion.Salir(request.Context(), controlador.leerCookie(request))
	http.SetCookie(response, controlador.cookieExpirada())
	vista.EscribirJSON(response, http.StatusNoContent, nil)
}

func (controlador *SesionControlador) yo(response http.ResponseWriter, request *http.Request) {
	vista.EscribirJSON(response, http.StatusOK, vista.PresentarUsuario(UsuarioDe(request.Context())))
}

func (controlador *SesionControlador) compilarCredenciales(payload ingresoJSON) modelo.CredencialesIngreso {
	return modelo.CredencialesIngreso{Email: payload.Email, Clave: payload.Clave}
}

func (controlador *SesionControlador) renderizarIngreso(response http.ResponseWriter, token string, usuario *modelo.Usuario) {
	http.SetCookie(response, controlador.cookieDeSesion(token))
	vista.EscribirJSON(response, http.StatusOK, vista.PresentarUsuario(usuario))
}
