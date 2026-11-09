package acceso

import (
	"time"

	modelo "uniapp/internal/application/acceso"
)

type AutorizacionJSON struct {
	ID         string    `json:"id"`
	Estado     string    `json:"estado"`
	Autorizado bool      `json:"autorizado"`
	CreadaEn   time.Time `json:"creada_en"`
}

func PresentarAutorizacion(autorizacion *modelo.Autorizacion) *AutorizacionJSON {
	if autorizacion == nil {
		return nil
	}
	return &AutorizacionJSON{
		ID:         autorizacion.ID,
		Estado:     string(autorizacion.Estado),
		Autorizado: autorizacion.Vigente(),
		CreadaEn:   autorizacion.CreadaEn,
	}
}

func PresentarAutorizaciones(autorizaciones []modelo.Autorizacion) []AutorizacionJSON {
	presentadas := make([]AutorizacionJSON, 0, len(autorizaciones))
	for indice := range autorizaciones {
		presentadas = append(presentadas, *PresentarAutorizacion(&autorizaciones[indice]))
	}
	return presentadas
}
