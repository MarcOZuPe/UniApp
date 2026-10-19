package reserva

import (
	"net/http"
)

func Rutas() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/reserva/salud", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"flujo":"reserva","estado":"esqueleto"}`))
	})
	return mux
}
