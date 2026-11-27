package reserva

import (
	"time"

	modelo "uniapp/internal/application/reserva"
)

// Los instantes viajan en UTC (RFC 3339, que es lo que serializa time.Time) y además se
// presentan en America/Lima para que la vista no tenga que convertir nada (RN-04).
const formatoLima = "2006-01-02 15:04"

type AdjuntoJSON struct {
	ID          string `json:"id"`
	Nombre      string `json:"nombre"`
	Tipo        string `json:"tipo"`
	TamanoBytes int64  `json:"tamano_bytes"`
	Estado      string `json:"estado"`
	Motivo      string `json:"motivo,omitempty"`
	Descargable bool   `json:"descargable"`
	CreadoEn    string `json:"creado_en"`
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
	CanceladaEn *time.Time    `json:"cancelada_en,omitempty"`
	Motivo      string        `json:"motivo_cancelacion,omitempty"`
	Adjuntos    []AdjuntoJSON `json:"adjuntos"`
}

type HistorialJSON struct {
	ID             string `json:"id"`
	EstadoAnterior string `json:"estado_anterior"`
	EstadoNuevo    string `json:"estado_nuevo"`
	Motivo         string `json:"motivo,omitempty"`
	ActorID        string `json:"actor_id"`
	CreadoEn       string `json:"creado_en"`
}

func PresentarHistorial(eventos []modelo.EventoHistorial) []HistorialJSON {
	salida := make([]HistorialJSON, 0, len(eventos))
	for _, evento := range eventos {
		salida = append(salida, HistorialJSON{
			ID:             evento.ID,
			EstadoAnterior: string(evento.EstadoAnterior),
			EstadoNuevo:    string(evento.EstadoNuevo),
			Motivo:         evento.Motivo,
			ActorID:        evento.ActorID,
			CreadoEn:       presentarLima(evento.CreadoEn),
		})
	}
	return salida
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
		CanceladaEn: reserva.CanceladaEn,
		Motivo:      reserva.MotivoCancelacion,
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

func PresentarAdjuntos(adjuntos []modelo.Adjunto) []AdjuntoJSON {
	return presentarAdjuntos(adjuntos)
}

func PresentarAdjunto(adjunto *modelo.Adjunto) AdjuntoJSON {
	return AdjuntoJSON{
		ID:          adjunto.ID,
		Nombre:      adjunto.Nombre,
		Tipo:        adjunto.Tipo,
		TamanoBytes: adjunto.TamanoBytes,
		Estado:      string(adjunto.Estado),
		Motivo:      adjunto.Motivo,
		Descargable: adjunto.Descargable(),
		CreadoEn:    presentarLima(adjunto.CreadoEn),
	}
}

func presentarAdjuntos(adjuntos []modelo.Adjunto) []AdjuntoJSON {
	presentados := make([]AdjuntoJSON, 0, len(adjuntos))
	for i := range adjuntos {
		presentados = append(presentados, PresentarAdjunto(&adjuntos[i]))
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
