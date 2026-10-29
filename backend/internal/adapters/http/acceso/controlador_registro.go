package acceso

import (
	"net/http"

	vista "uniapp/internal/adapters/presentation/acceso"
	modelo "uniapp/internal/application/acceso"
)

func (c *Controlador) registro(w http.ResponseWriter, r *http.Request) {
	var entrada struct {
		Email     string `json:"email"`
		NombreUso string `json:"nombre_uso"`
		Clave     string `json:"clave"`
		Rol       string `json:"rol"`
	}
	if !c.decodificarCuerpo(w, r, &entrada) {
		return
	}
	datos := modelo.DatosRegistro{Email: entrada.Email, NombreUso: entrada.NombreUso, Clave: entrada.Clave, Rol: modelo.Rol(entrada.Rol)}
	usuario, err := c.servicio.Registrar(r.Context(), datos)
	if err != nil {
		c.responderError(w, err)
		return
	}
	vista.EscribirJSON(w, http.StatusCreated, vista.PresentarUsuario(usuario))
}
