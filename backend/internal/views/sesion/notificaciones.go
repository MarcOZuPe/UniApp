package sesion

import (
	"time"
	modelo "uniapp/internal/models/sesion"
)

type NotificacionJSON struct {
	ID        int64     `json:"id"`
	ReservaID string    `json:"reserva_id"`
	Tipo      string    `json:"tipo"`
	Asunto    string    `json:"asunto"`
	Cuerpo    string    `json:"cuerpo"`
	Fecha     time.Time `json:"fecha"`
}

func PresentarNotificaciones(entregas []modelo.Entrega) []NotificacionJSON {
	resultado := make([]NotificacionJSON, 0, len(entregas))
	for _, e := range entregas {
		resultado = append(resultado, NotificacionJSON{e.ID, e.ReservaID, e.Tipo, e.Asunto, e.Cuerpo, e.EntregadoEn})
	}
	return resultado
}
