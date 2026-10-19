package acceso

import (
	"net/http"

	vista "uniapp/internal/views/acceso"
)

func (c *Controlador) recuperar(w http.ResponseWriter, r *http.Request) {
	var entrada struct{ Email string }
	if !c.decodificarCuerpo(w, r, &entrada) {
		return
	}
	if err := c.servicio.SolicitarRecuperacion(r.Context(), entrada.Email); err != nil {
		c.responderError(w, err)
		return
	}

	vista.EscribirJSON(w, http.StatusAccepted, map[string]string{"mensaje": "Si el correo está registrado, recibirás un enlace para restablecer tu contraseña."})
}

func (c *Controlador) restablecer(w http.ResponseWriter, r *http.Request) {
	var entrada struct{ Token, Clave string }
	if !c.decodificarCuerpo(w, r, &entrada) {
		return
	}
	if err := c.servicio.RestablecerClave(r.Context(), entrada.Token, entrada.Clave); err != nil {
		c.responderError(w, err)
		return
	}
	vista.EscribirJSON(w, http.StatusOK, map[string]string{"mensaje": "Contraseña actualizada. Inicia sesión con la nueva."})
}
