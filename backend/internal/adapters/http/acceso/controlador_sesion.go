package acceso

import (
	"net/http"

	vista "uniapp/internal/adapters/presentation/acceso"
)

func (c *Controlador) ingreso(w http.ResponseWriter, r *http.Request) {
	var entrada struct{ Email, Clave string }
	if !c.decodificarCuerpo(w, r, &entrada) {
		return
	}
	token, usuario, err := c.servicio.Ingresar(r.Context(), entrada.Email, entrada.Clave)
	if err != nil {
		c.responderError(w, err)
		return
	}
	http.SetCookie(w, c.cookieDeSesion(token))
	vista.EscribirJSON(w, http.StatusOK, vista.PresentarUsuario(usuario))
}

func (c *Controlador) salida(w http.ResponseWriter, r *http.Request) {
	_ = c.servicio.Salir(r.Context(), c.leerCookie(r))
	http.SetCookie(w, c.cookieExpirada())
	vista.EscribirJSON(w, http.StatusNoContent, nil)
}

func (c *Controlador) yo(w http.ResponseWriter, r *http.Request) {
	vista.EscribirJSON(w, http.StatusOK, vista.PresentarUsuario(UsuarioDe(r.Context())))
}
