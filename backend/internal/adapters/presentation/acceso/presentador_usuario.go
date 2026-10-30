package acceso

import modelo "uniapp/internal/application/acceso"

type UsuarioJSON struct {
	ID               string `json:"id"`
	Email            string `json:"email"`
	NombreUso        string `json:"nombre_uso"`
	Rol              string `json:"rol"`
	CorreoConfirmado bool   `json:"correo_confirmado"`
}

func PresentarUsuario(usuario *modelo.Usuario) UsuarioJSON {
	return UsuarioJSON{
		ID:               usuario.ID,
		Email:            usuario.Email,
		NombreUso:        usuario.NombreUso,
		Rol:              string(usuario.Rol),
		CorreoConfirmado: usuario.EmailConfirmado,
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
