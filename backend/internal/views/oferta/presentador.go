package oferta

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	modelo "uniapp/internal/models/oferta"
)

type SlotJSON struct {
	ID        int64  `json:"id"`
	DiaSemana int    `json:"dia_semana"`
	Inicio    string `json:"inicio"`
	Fin       string `json:"fin"`
	Zona      string `json:"zona_horaria"`
}

type DisponibilidadJSON struct {
	CupoSemanalMinutos int        `json:"cupo_semanal_minutos"`
	Slots              []SlotJSON `json:"slots"`
}

type ErrorJSON struct {
	Codigo  string            `json:"codigo"`
	Mensaje string            `json:"mensaje"`
	Campos  map[string]string `json:"campos,omitempty"`
}

func PresentarDisponibilidad(cupo modelo.CupoSemanal, slots []modelo.SlotSemanal) DisponibilidadJSON {
	respuesta := DisponibilidadJSON{
		CupoSemanalMinutos: cupo.Minutos,
		Slots:              make([]SlotJSON, 0, len(slots)),
	}
	for _, slot := range slots {
		respuesta.Slots = append(respuesta.Slots, SlotJSON{
			ID: slot.ID, DiaSemana: int(slot.DiaSemana),
			Inicio: formatoHora(slot.Inicio), Fin: formatoHora(slot.Fin),
			Zona: modelo.ZonaHoraria,
		})
	}
	return respuesta
}

func EscribirJSON(w http.ResponseWriter, estado int, valor any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(estado)
	if valor != nil {
		_ = json.NewEncoder(w).Encode(valor)
	}
}

func EscribirError(w http.ResponseWriter, estado int, codigo, mensaje string, campos map[string]string) {
	EscribirJSON(w, estado, ErrorJSON{Codigo: codigo, Mensaje: mensaje, Campos: campos})
}

func formatoHora(duracion time.Duration) string {
	totalMinutos := int(duracion / time.Minute)
	return fmt.Sprintf("%02d:%02d", totalMinutos/60, totalMinutos%60)
}
