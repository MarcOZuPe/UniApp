package sesion

import (
	"net/http"
	acceso "uniapp/internal/controllers/acceso"
	modelo "uniapp/internal/models/sesion"
	comun "uniapp/internal/views/acceso"
	vista "uniapp/internal/views/sesion"
)

func RutasNotificaciones(bandeja modelo.Bandeja, autenticar func(http.Handler) http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /api/notificaciones", autenticar(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := acceso.UsuarioDe(r.Context())
		if u == nil {
			responderError(w, 401, "no_autenticado", "Debes iniciar sesión.")
			return
		}
		if u.Rol != "tutor" && u.Rol != "escolar" {
			responderError(w, 403, "rol_no_permitido", "Tu rol no tiene acceso a notificaciones.")
			return
		}
		datos, err := bandeja.NotificacionesDe(r.Context(), u.ID)
		if err != nil {
			responderError(w, 500, "error_notificaciones", "No se pudieron consultar las notificaciones.")
			return
		}
		comun.EscribirJSON(w, 200, map[string]any{"notificaciones": vista.PresentarNotificaciones(datos)})
	})))
	return mux
}
