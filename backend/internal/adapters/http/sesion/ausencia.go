package sesion

import (
	"net/http"
	"time"
	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	vista "uniapp/internal/adapters/presentation/acceso"
	modelo "uniapp/internal/application/sesion"
)

func (c *Controlador) reportarAusencia(w http.ResponseWriter, r *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	if usuario == nil {
		responderError(w, http.StatusUnauthorized, "no_autenticado", "Debes iniciar sesion.")
		return
	}
	// No se admite identidad ni participante objetivo proporcionado por el cliente.
	if r.Body != nil && r.Body != http.NoBody {
		responderError(w, http.StatusBadRequest, "solicitud_invalida", "No se admite un cuerpo.")
		return
	}
	a, err := c.repositorio.ReportarAusencia(r.Context(), r.PathValue("id"), usuario.ID, time.Now().UTC())
	if err != nil {
		responderErrorRepositorio(w, err)
		return
	}
	vista.EscribirJSON(w, http.StatusCreated, presentarAusencia(a))
}

func presentarAusencia(a modelo.Ausencia) map[string]any {
	return map[string]any{
		"usuario_id":   a.UsuarioID(),
		"rol":          string(a.Rol()),
		"reportada_en": a.ReportadaEn(),
	}
}
