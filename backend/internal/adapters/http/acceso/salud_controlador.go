package acceso

import (
	"net/http"

	vista "uniapp/internal/adapters/presentation/acceso"
)

type SaludControlador struct{}

func (controlador *SaludControlador) Montar(multiplexor *http.ServeMux) {
	multiplexor.HandleFunc("GET /api/salud", controlador.salud)
}

func (controlador *SaludControlador) salud(response http.ResponseWriter, _ *http.Request) {
	vista.EscribirJSON(response, http.StatusOK, map[string]string{"estado": "ok"})
}
