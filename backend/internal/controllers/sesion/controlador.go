package sesion

import (
	"net/http"
)

func Rutas() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/sesion/salud", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"flujo":"sesion","estado":"esqueleto"}`))
	})
	return mux
}
