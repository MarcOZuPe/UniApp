package acceso

import (
	"time"

	modelo "uniapp/internal/application/acceso"
)

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

type LogSistemaJSON struct {
	ID        int64          `json:"id"`
	CreadoEn  time.Time      `json:"creado_en"`
	Categoria string         `json:"categoria"`
	Evento    string         `json:"evento"`
	ActorID   string         `json:"actor_id,omitempty"`
	ActorRol  string         `json:"actor_rol"`
	IP        string         `json:"ip"`
	Detalle   map[string]any `json:"detalle"`
}

func PresentarLogsSistema(logs []modelo.LogSistema) []LogSistemaJSON {
	presentados := make([]LogSistemaJSON, 0, len(logs))
	for _, log := range logs {
		detalle := log.Detalle
		if detalle == nil {
			detalle = map[string]any{}
		}
		presentados = append(presentados, LogSistemaJSON{
			ID: log.ID, CreadoEn: log.CreadoEn, Categoria: log.Categoria,
			Evento: log.Evento, ActorID: log.ActorID, ActorRol: log.ActorRol,
			IP: log.IP, Detalle: detalle,
		})
	}
	return presentados
}
