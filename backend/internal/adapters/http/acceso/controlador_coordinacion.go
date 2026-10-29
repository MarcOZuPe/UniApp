package acceso

import (
	"net/http"

	vista "uniapp/internal/adapters/presentation/acceso"
)

func (c *Controlador) panelCoordinacion(w http.ResponseWriter, r *http.Request) {
	vista.EscribirJSON(w, http.StatusOK, vista.PresentarPanelCoordinacion(UsuarioDe(r.Context())))
}
