package reserva

import (
	"time"

	modelo "uniapp/internal/models/reserva"
)

// Los instantes viajan en UTC (RFC 3339, que es lo que serializa time.Time) y además se
// presentan en America/Lima para que la vista no tenga que convertir nada (RN-04).
const formatoLima = "2006-01-02 15:04"

type AdjuntoJSON struct {
	Nombre      string `json:"nombre"`
	Tipo        string `json:"tipo"`
	TamanoBytes int64  `json:"tamano_bytes"`
}

type ReservaJSON struct {
	ID          string        `json:"id"`
	SlotID      string        `json:"slot_id"`
	TutorID     string        `json:"tutor_id"`
	TutorNombre string        `json:"tutor_nombre"`
	Curso       string        `json:"curso"`
	Grado       string        `json:"grado"`
	Tema        string        `json:"tema"`
	Objetivo    string        `json:"objetivo"`
	Estado      string        `json:"estado"`
	Inicio      time.Time     `json:"inicio"`
	Fin         time.Time     `json:"fin"`
	InicioLima  string        `json:"inicio_lima"`
	FinLima     string        `json:"fin_lima"`
	CreadaEn    time.Time     `json:"creada_en"`
	Adjuntos    []AdjuntoJSON `json:"adjuntos"`
}

func PresentarReserva(reserva *modelo.Reserva) ReservaJSON {
	return ReservaJSON{
		ID:          reserva.ID,
		SlotID:      reserva.SlotID,
		TutorID:     reserva.TutorID,
		TutorNombre: reserva.TutorNombre,
		Curso:       reserva.Curso,
		Grado:       reserva.Grado,
		Tema:        reserva.Tema,
		Objetivo:    reserva.Objetivo,
		Estado:      string(reserva.Estado),
		Inicio:      reserva.Inicio.UTC(),
		Fin:         reserva.Fin.UTC(),
		InicioLima:  presentarLima(reserva.Inicio),
		FinLima:     presentarLima(reserva.Fin),
		CreadaEn:    reserva.CreadaEn.UTC(),
		Adjuntos:    presentarAdjuntos(reserva.Adjuntos),
	}
}

func PresentarReservas(reservas []modelo.Reserva) []ReservaJSON {
	presentadas := make([]ReservaJSON, 0, len(reservas))
	for i := range reservas {
		presentadas = append(presentadas, PresentarReserva(&reservas[i]))
	}
	return presentadas
}

func presentarAdjuntos(adjuntos []modelo.Adjunto) []AdjuntoJSON {
	presentados := make([]AdjuntoJSON, 0, len(adjuntos))
	for _, adjunto := range adjuntos {
		presentados = append(presentados, AdjuntoJSON{
			Nombre: adjunto.Nombre, Tipo: adjunto.Tipo, TamanoBytes: adjunto.TamanoBytes,
		})
	}
	return presentados
}

func presentarLima(instante time.Time) string {
	return instante.In(modelo.ZonaLima).Format(formatoLima)
}

type SlotJSON struct {
	ID          string    `json:"id"`
	TutorID     string    `json:"tutor_id"`
	TutorNombre string    `json:"tutor_nombre"`
	Curso       string    `json:"curso"`
	Inicio      time.Time `json:"inicio"`
	Fin         time.Time `json:"fin"`
	InicioLima  string    `json:"inicio_lima"`
	FinLima     string    `json:"fin_lima"`
	Disponible  bool      `json:"disponible"`
	Motivo      string    `json:"motivo,omitempty"`
}

func PresentarSlots(disposiciones []modelo.Disposicion) []SlotJSON {
	presentados := make([]SlotJSON, 0, len(disposiciones))
	for _, disposicion := range disposiciones {
		presentados = append(presentados, SlotJSON{
			ID:          disposicion.Slot.ID,
			TutorID:     disposicion.Slot.TutorID,
			TutorNombre: disposicion.Slot.TutorNombre,
			Curso:       disposicion.Slot.Curso,
			Inicio:      disposicion.Slot.Inicio.UTC(),
			Fin:         disposicion.Slot.Fin.UTC(),
			InicioLima:  presentarLima(disposicion.Slot.Inicio),
			FinLima:     presentarLima(disposicion.Slot.Fin),
			Disponible:  disposicion.Disponible(),
			Motivo:      disposicion.Motivo,
		})
	}
	return presentados
}
