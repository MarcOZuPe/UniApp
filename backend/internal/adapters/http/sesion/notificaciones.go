package sesion

import (
	"net/http"
	acceso "uniapp/internal/adapters/http/acceso"
	comun "uniapp/internal/adapters/presentation/acceso"
	vista "uniapp/internal/adapters/presentation/sesion"
	modelo "uniapp/internal/application/sesion"
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
