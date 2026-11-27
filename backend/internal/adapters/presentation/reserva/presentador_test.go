package reserva

import (
	"testing"
	"time"

	modelo "uniapp/internal/application/reserva"
)

func TestPresentarReservasYAdjuntos(t *testing.T) {
	inicio := time.Date(2026, 11, 26, 18, 0, 0, 0, time.UTC)
	reserva := modelo.Reserva{
		ID: "r-1", SlotID: "s-1", TutorID: "t-1", TutorNombre: "Tutor", Curso: "Álgebra",
		Grado: "5to", Tema: "Tema", Objetivo: "Objetivo", Estado: modelo.EstadoConfirmada,
		Inicio: inicio, Fin: inicio.Add(time.Hour), CreadaEn: inicio,
		Adjuntos: []modelo.Adjunto{{ID: "a-1", Nombre: "a.pdf", Tipo: "application/pdf", TamanoBytes: 10, Estado: modelo.AdjuntoValidado, CreadoEn: inicio}},
	}
	presentada := PresentarReserva(&reserva)
	if presentada.ID != "r-1" || presentada.InicioLima != "2026-11-26 13:00" || presentada.FinLima != "2026-11-26 14:00" {
		t.Fatalf("reserva presentada inesperada: %+v", presentada)
	}
	if len(presentada.Adjuntos) != 1 || !presentada.Adjuntos[0].Descargable {
		t.Fatalf("adjuntos inesperados: %+v", presentada.Adjuntos)
	}
	if len(PresentarReservas([]modelo.Reserva{reserva})) != 1 {
		t.Fatal("debería presentar una reserva")
	}
	adjunto := PresentarAdjunto(&reserva.Adjuntos[0])
	if adjunto.Nombre != "a.pdf" || !adjunto.Descargable {
		t.Fatalf("adjunto presentado inesperado: %+v", adjunto)
	}
	if len(PresentarAdjuntos(reserva.Adjuntos)) != 1 {
		t.Fatal("debería presentar un adjunto")
	}

	slots := PresentarSlots([]modelo.Disposicion{{
		Slot:   modelo.Slot{ID: "s-1", TutorID: "t-1", TutorNombre: "Tutor", Curso: "Álgebra", Inicio: inicio, Fin: inicio.Add(time.Hour)},
		Motivo: "no disponible",
	}})
	if len(slots) != 1 || slots[0].Disponible || slots[0].Motivo == "" || slots[0].InicioLima != "2026-11-26 13:00" {
		t.Fatalf("slot presentado inesperado: %+v", slots)
	}
}
