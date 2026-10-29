package acceso

import (
	"net/http"

	vista "uniapp/internal/adapters/presentation/acceso"
)

func (c *Controlador) confirmar(w http.ResponseWriter, r *http.Request) {
	var entrada struct{ Token string }
	if !c.decodificarCuerpo(w, r, &entrada) {
		return
	}
	if err := c.servicio.ConfirmarCorreo(r.Context(), entrada.Token); err != nil {
		c.responderError(w, err)
		return
	}
	vista.EscribirJSON(w, http.StatusOK, map[string]string{"mensaje": "Correo confirmado. Ya puedes ingresar."})
}

func (c *Controlador) reenviarConfirmacion(w http.ResponseWriter, r *http.Request) {
	var entrada struct{ Email string }
	if !c.decodificarCuerpo(w, r, &entrada) {
		return
	}
	if err := c.servicio.ReenviarConfirmacion(r.Context(), entrada.Email); err != nil {
		c.responderError(w, err)
		return
	}
	vista.EscribirJSON(w, http.StatusAccepted, map[string]string{"mensaje": "Si el correo está registrado y sin confirmar, recibirás un nuevo enlace."})
}
