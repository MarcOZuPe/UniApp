package acceso

import (
	"net/http"

	modelo "uniapp/internal/application/acceso"
)

type PerfilControlador struct{ *controladorBase }

func (controlador *PerfilControlador) Montar(multiplexor *http.ServeMux) {
	multiplexor.Handle(
		"PUT /api/perfil/clave",
		Encadenar(http.HandlerFunc(controlador.cambiarClave), controlador.Autenticar()),
	)
}

type entradaCambioClave struct {
	ClaveActual string `json:"clave_actual"`
	ClaveNueva  string `json:"clave_nueva"`
}

func (controlador *PerfilControlador) cambiarClave(response http.ResponseWriter, request *http.Request) {
	usuario := UsuarioDe(request.Context())
	if usuario == nil {
		controlador.responderError(response, modelo.ErrNoAutorizado)
		return
	}
	var entrada entradaCambioClave
	if !controlador.decodificarCuerpo(response, request, &entrada) {
		return
	}
	if err := controlador.servicio.Perfil.CambiarClave(request.Context(), usuario.ID, modelo.DatosCambioClave{
		ClaveActual: entrada.ClaveActual,
		ClaveNueva:  entrada.ClaveNueva,
		IP:          ipCliente(request),
	}); err != nil {
		controlador.responderError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}
