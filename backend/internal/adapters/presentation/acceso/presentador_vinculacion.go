package acceso

import (
	"time"

	modelo "uniapp/internal/application/acceso"
)

type VinculacionUsuarioJSON struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	NombreUso string `json:"nombre_uso"`
	Rol       string `json:"rol"`
}

type VinculacionJSON struct {
	ID           string                 `json:"id"`
	Estado       string                 `json:"estado"`
	Apoderado    VinculacionUsuarioJSON `json:"apoderado"`
	Escolar      VinculacionUsuarioJSON `json:"escolar"`
	Autorizacion *AutorizacionJSON      `json:"autorizacion,omitempty"`
	CreadaEn     time.Time              `json:"creada_en"`
	AceptadaEn   *time.Time             `json:"aceptada_en,omitempty"`
	ExpiraEn     time.Time              `json:"expira_en"`
}

func PresentarVinculacion(vinculacion *modelo.Vinculacion) VinculacionJSON {
	return VinculacionJSON{
		ID:     vinculacion.ID,
		Estado: string(vinculacion.Estado),
		Apoderado: presentarUsuarioVinculado(usuarioVinculacion{
			id:        vinculacion.ApoderadoID,
			email:     vinculacion.ApoderadoEmail,
			nombreUso: vinculacion.ApoderadoNombre,
			rol:       string(modelo.RolApoderado),
		}),
		Escolar: presentarUsuarioVinculado(usuarioVinculacion{
			id:        vinculacion.EscolarID,
			email:     vinculacion.EscolarEmail,
			nombreUso: vinculacion.EscolarNombre,
			rol:       string(modelo.RolEscolar),
		}),
		CreadaEn:     vinculacion.CreadaEn,
		AceptadaEn:   vinculacion.AceptadaEn,
		ExpiraEn:     vinculacion.ExpiraEn,
		Autorizacion: PresentarAutorizacion(vinculacion.Autorizacion),
	}
}

func PresentarVinculaciones(vinculaciones []modelo.Vinculacion) []VinculacionJSON {
	presentadas := make([]VinculacionJSON, 0, len(vinculaciones))
	for indice := range vinculaciones {
		presentadas = append(presentadas, PresentarVinculacion(&vinculaciones[indice]))
	}
	return presentadas
}

type usuarioVinculacion struct {
	id        string
	email     string
	nombreUso string
	rol       string
}

func presentarUsuarioVinculado(datos usuarioVinculacion) VinculacionUsuarioJSON {
	return VinculacionUsuarioJSON{
		ID:        datos.id,
		Email:     datos.email,
		NombreUso: datos.nombreUso,
		Rol:       datos.rol,
	}
}
