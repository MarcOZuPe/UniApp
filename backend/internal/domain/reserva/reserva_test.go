package reserva

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func slotFuturo(ahora time.Time) Slot {
	inicio := ahora.Add(72 * time.Hour)
	return Slot{ID: "slot-1", TutorID: "tutor-1", TutorNombre: "Tutor", Curso: "Álgebra", Inicio: inicio, Fin: inicio.Add(time.Hour)}
}

func datosReservaValidos() DatosReserva {
	return DatosReserva{
		SlotID: "slot-1", Tema: "Ecuaciones lineales",
		Objetivo: "Resolver ejercicios de ecuaciones lineales",
		Grado:    "5to de secundaria", ClaveIdempotencia: "clave-123456",
	}
}

func TestMotivoPorTiempo(t *testing.T) {
	ahora := time.Now().UTC()
	slot := slotFuturo(ahora)
	if motivo := MotivoPorTiempo(slot, ahora); motivo != "" {
		t.Fatalf("el slot futuro debería estar disponible: %q", motivo)
	}
	if motivo := MotivoPorTiempo(slot, slot.Fin.Add(time.Hour)); motivo == "" {
		t.Fatal("un slot vencido debería tener motivo")
	}
	cercano := slot
	cercano.Inicio = ahora.Add(time.Hour)
	cercano.Fin = cercano.Inicio.Add(time.Hour)
	if motivo := MotivoPorTiempo(cercano, ahora); !strings.Contains(motivo, "48") {
		t.Fatalf("motivo de anticipación inesperado: %q", motivo)
	}
	lejano := slot
	lejano.Inicio = ahora.Add(40 * 24 * time.Hour)
	lejano.Fin = lejano.Inicio.Add(time.Hour)
	if motivo := MotivoPorTiempo(lejano, ahora); motivo == "" {
		t.Fatal("un slot fuera del horizonte debería tener motivo")
	}
}

func TestSemanaLimaEmpiezaLunes(t *testing.T) {
	momento := time.Date(2026, 11, 26, 18, 0, 0, 0, time.UTC) // jueves
	inicio, fin := SemanaLima(momento)
	if inicio.Weekday() != time.Monday || fin.Sub(inicio) != 7*24*time.Hour {
		t.Fatalf("semana inesperada: %v - %v", inicio, fin)
	}
	if !inicio.Before(momento) || !fin.After(momento) {
		t.Fatal("la semana debe contener el momento")
	}
}

func TestNuevaReservaYValidaciones(t *testing.T) {
	ahora := time.Now().UTC()
	reserva, err := NuevaReserva("escolar-1", datosReservaValidos(), slotFuturo(ahora), ahora)
	if err != nil {
		t.Fatalf("reserva válida rechazada: %v", err)
	}
	if reserva.Estado != EstadoConfirmada || reserva.EscolarID != "escolar-1" || reserva.Inicio.IsZero() {
		t.Fatalf("reserva inesperada: %+v", reserva)
	}
	if _, err := NuevaReserva("", datosReservaValidos(), slotFuturo(ahora), ahora); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("se esperaba ErrNoAutorizado, se obtuvo %v", err)
	}
	datos := datosReservaValidos()
	datos.Tema = "corto"
	datos.Objetivo = "corto"
	datos.Grado = ""
	datos.ClaveIdempotencia = "mala clave!"
	if _, err := NuevaReserva("escolar-1", datos, slotFuturo(ahora), ahora); err == nil {
		t.Fatal("se esperaban errores de validación")
	}
	slot := slotFuturo(ahora)
	slot.ID = ""
	if _, err := NuevaReserva("escolar-1", datosReservaValidos(), slot, ahora); !errors.Is(err, ErrSlotInexistente) {
		t.Fatalf("se esperaba ErrSlotInexistente, se obtuvo %v", err)
	}
}

func TestMismosDatos(t *testing.T) {
	ahora := time.Now().UTC()
	pedida, _ := NuevaReserva("escolar-1", datosReservaValidos(), slotFuturo(ahora), ahora)
	guardada := *pedida
	if !MismosDatos(&guardada, pedida) {
		t.Fatal("las reservas con mismos datos deberían coincidir")
	}
	guardada.Tema = "otro tema"
	if MismosDatos(&guardada, pedida) {
		t.Fatal("un tema distinto no debería coincidir")
	}
}

func TestAnalizadorYAdjuntos(t *testing.T) {
	if estado, motivo, err := (AnalizadorNoDisponible{}).Analizar("a.pdf", "application/pdf", []byte("x")); estado != AdjuntoEnCuarentena || motivo == "" || !errors.Is(err, ErrAnalisisNoDisponible) {
		t.Fatalf("analizador no disponible inesperado: %v %q %v", estado, motivo, err)
	}
	basico := AnalizadorBasico{}
	if estado, _, err := basico.Analizar("a.pdf", "application/pdf", nil); estado != AdjuntoRechazado || err != nil {
		t.Fatalf("archivo vacío inesperado: %v %v", estado, err)
	}
	if estado, _, err := basico.Analizar("a.txt", "text/plain", []byte("EICAR-STANDARD-ANTIVIRUS-TEST-FILE")); estado != AdjuntoRechazado || err != nil {
		t.Fatalf("firma EICAR inesperada: %v %v", estado, err)
	}
	if estado, _, err := basico.Analizar("a.exe", "application/octet-stream", []byte("x")); estado != AdjuntoRechazado || err != nil {
		t.Fatalf("tipo no permitido inesperado: %v %v", estado, err)
	}
	if estado, _, err := basico.Analizar("a.pdf", "application/pdf", []byte("x")); estado != AdjuntoValidado || err != nil {
		t.Fatalf("PDF válido inesperado: %v %v", estado, err)
	}
	if err := ValidarEntradaAdjunto("", 1); err == nil {
		t.Fatal("nombre vacío debe rechazarse")
	}
	if err := ValidarEntradaAdjunto("../a.pdf", 1); err == nil {
		t.Fatal("nombre con ruta debe rechazarse")
	}
	if err := ValidarEntradaAdjunto("a.pdf", 0); err == nil {
		t.Fatal("tamaño cero debe rechazarse")
	}
	if err := ValidarEntradaAdjunto("a.pdf", AdjuntoMaximoBytes+1); err == nil {
		t.Fatal("tamaño excesivo debe rechazarse")
	}
	if err := ValidarEntradaAdjunto("a.pdf", 100); err != nil {
		t.Fatalf("adjunto válido rechazado: %v", err)
	}
	if TipoDetectado(nil) != "application/octet-stream" {
		t.Fatal("contenido vacío debería tener tipo octet-stream")
	}
	if extension, ok := ExtensionPermitida("application/pdf"); !ok || extension != ".pdf" {
		t.Fatalf("extensión inesperada: %q %v", extension, ok)
	}
	if _, ok := ExtensionPermitida("application/zip"); ok {
		t.Fatal("zip no debería estar permitido")
	}
	if HuellaAdjunto([]byte("abc")) == "" {
		t.Fatal("la huella no debe estar vacía")
	}
	adjunto, err := NuevoAdjunto("a.pdf", "application/pdf", 3, "huella", "clave", "", "", time.Now())
	if err != nil || adjunto.Estado != AdjuntoEnCuarentena || adjunto.Descargable() {
		t.Fatalf("adjunto inesperado: %+v err=%v", adjunto, err)
	}
	adjunto.Estado = AdjuntoValidado
	if !adjunto.Descargable() {
		t.Fatal("un adjunto validado debería ser descargable")
	}
	if _, err := NuevoAdjunto("", "application/pdf", 3, "huella", "clave", "", "", time.Now()); err == nil {
		t.Fatal("nombre vacío debe rechazarse")
	}
	if _, err := NuevoAdjunto("a.pdf", "application/zip", 3, "huella", "clave", "", "", time.Now()); err == nil {
		t.Fatal("tipo no permitido debe rechazarse")
	}
}
