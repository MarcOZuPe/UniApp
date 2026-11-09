package sesion

import (
	"net/http"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	vista "uniapp/internal/adapters/presentation/acceso"
	modelo "uniapp/internal/application/sesion"
)

// reportarAusencia registra la ausencia propia del usuario autenticado.
// No se admite identidad ni participante objetivo enviado por el cliente.
func (controlador *Controlador) reportarAusencia(response http.ResponseWriter, request *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(request.Context())
	if usuario == nil {
		responderError(response, detalleErrorHTTP{
			estado: http.StatusUnauthorized, codigo: "no_autenticado", mensaje: "Debes iniciar sesion.",
		})
		return
	}
	if request.Body != nil && request.Body != http.NoBody {
		responderError(response, detalleErrorHTTP{
			estado: http.StatusBadRequest, codigo: "solicitud_invalida", mensaje: "No se admite un cuerpo.",
		})
		return
	}

	ausencia, err := controlador.repositorio.ReportarAusencia(
		request.Context(),
		request.PathValue("id"),
		usuario.ID,
		controlador.ahora(),
	)
	if err != nil {
		responderErrorRepositorio(response, err)
		return
	}

	vista.EscribirJSON(response, http.StatusCreated, presentarAusencia(ausencia))
}

func presentarAusencia(ausencia modelo.Ausencia) map[string]any {
	return map[string]any{
		"usuario_id":   ausencia.UsuarioID(),
		"rol":          string(ausencia.Rol()),
		"reportada_en": ausencia.ReportadaEn(),
	}
}
