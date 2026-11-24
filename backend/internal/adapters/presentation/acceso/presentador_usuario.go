package acceso

import modelo "uniapp/internal/application/acceso"

type UsuarioJSON struct {
	ID               string `json:"id"`
	Email            string `json:"email"`
	NombreUso        string `json:"nombre_uso"`
	Rol              string `json:"rol"`
	FacultadID       string `json:"facultad_id,omitempty"`
	FacultadNombre   string `json:"facultad_nombre,omitempty"`
	CorreoConfirmado bool   `json:"correo_confirmado"`
}

func PresentarUsuario(usuario *modelo.Usuario) UsuarioJSON {
	return UsuarioJSON{
		ID:               usuario.ID,
		Email:            usuario.Email,
		NombreUso:        usuario.NombreUso,
		Rol:              string(usuario.Rol),
		FacultadID:       usuario.FacultadID,
		FacultadNombre:   usuario.FacultadNombre,
		CorreoConfirmado: usuario.EmailConfirmado,
	}
}

type FacultadJSON struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
}

func PresentarFacultades(facultades []modelo.Facultad) []FacultadJSON {
	presentadas := make([]FacultadJSON, 0, len(facultades))
	for _, facultad := range facultades {
		presentadas = append(presentadas, FacultadJSON{ID: facultad.ID, Nombre: facultad.Nombre})
	}
	return presentadas
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
