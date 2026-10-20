package acceso

import (
	"encoding/json"
	"net/http"

	modelo "uniapp/internal/models/acceso"
)

type UsuarioJSON struct {
	ID               string `json:"id"`
	Email            string `json:"email"`
	NombreUso        string `json:"nombre_uso"`
	Rol              string `json:"rol"`
	CorreoConfirmado bool   `json:"correo_confirmado"`
}

func PresentarUsuario(usuario *modelo.Usuario) UsuarioJSON {
	return UsuarioJSON{
		ID: usuario.ID, Email: usuario.Email, NombreUso: usuario.NombreUso,
		Rol: string(usuario.Rol), CorreoConfirmado: usuario.EmailConfirmado,
	}
}

type PanelCoordinacionJSON struct {
	Mensaje string      `json:"mensaje"`
	Usuario UsuarioJSON `json:"usuario"`
}

func PresentarPanelCoordinacion(usuario *modelo.Usuario) PanelCoordinacionJSON {
	return PanelCoordinacionJSON{
		Mensaje: "Panel de coordinación",
		Usuario: PresentarUsuario(usuario),
	}
}

type ErrorJSON struct {
	Codigo  string            `json:"codigo"`
	Mensaje string            `json:"mensaje"`
	Campos  map[string]string `json:"campos,omitempty"`
}

type RespuestaError struct {
	Estado  int
	Codigo  string
	Mensaje string
	Campos  map[string]string
}

func EscribirJSON(w http.ResponseWriter, estado int, valor any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(estado)
	if valor != nil {
		_ = json.NewEncoder(w).Encode(valor)
	}
}

func EscribirError(w http.ResponseWriter, respuesta RespuestaError) {
	EscribirJSON(w, respuesta.Estado, ErrorJSON{Codigo: respuesta.Codigo, Mensaje: respuesta.Mensaje, Campos: respuesta.Campos})
}
