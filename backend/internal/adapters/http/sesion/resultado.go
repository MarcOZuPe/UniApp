package sesion

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	vista "uniapp/internal/adapters/presentation/acceso"
	modelo "uniapp/internal/application/sesion"
)

func (c *Controlador) registrarResultado(w http.ResponseWriter, r *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	if usuario == nil {
		responderError(w, http.StatusUnauthorized, "no_autenticado", "Debes iniciar sesion.")
		return
	}
	var entrada struct {
		Resumen string `json:"resumen"`
	}
	// Cuerpo acotado; no se admiten autor, fecha ni campos ajenos al resultado.
	decodificador := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16*1024))
	decodificador.DisallowUnknownFields()
	if err := decodificador.Decode(&entrada); err != nil {
		responderError(w, http.StatusBadRequest, "solicitud_invalida", "Envia un objeto JSON con el resumen del resultado.")
		return
	}
	if err := decodificador.Decode(new(any)); err != io.EOF {
		responderError(w, http.StatusBadRequest, "solicitud_invalida", "Envia un unico objeto JSON.")
		return
	}
	resultado, err := c.repositorio.RegistrarResultado(r.Context(), r.PathValue("id"), usuario.ID, entrada.Resumen, time.Now().UTC())
	if err != nil {
		responderErrorRepositorio(w, err)
		return
	}
	vista.EscribirJSON(w, http.StatusCreated, presentarResultado(&resultado))
}

func presentarResultado(r *modelo.Resultado) any {
	if r == nil {
		return nil
	}
	return map[string]any{
		"usuario_id":    r.UsuarioID(),
		"resumen":       r.Resumen(),
		"registrado_en": r.RegistradoEn(),
	}
}
