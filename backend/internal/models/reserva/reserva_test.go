package reserva

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// Reloj fijo: las pruebas de 48 horas no dependen de la hora real (RN-03).
var ahora = time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)

func datosValidos() DatosReserva {
	return DatosReserva{
		SlotID:   "slot-i04-01",
		Tema:     "Repaso de factorización",
		Objetivo: "Resolver ejercicios de factorización de trinomios sin ayuda del tutor.",
		Grado:    "5to de secundaria",
	}
}

func servicioDePrueba(base time.Time) *Servicio {
	s := NuevoServicio(Dependencias{
		Repositorio:  NuevoRepositorioMemoria(),
		Oferta:       NuevaOfertaSintetica(base),
		Autorizacion: autorizacionFija(true),
	})
	s.FijarReloj(func() time.Time { return ahora })
	return s
}

func slotEn(inicio time.Time) Slot {
	return Slot{
		ID:          "slot-fijo",
		TutorID:     TutorSinteticoID,
		TutorNombre: TutorSinteticoNombre,
		Curso:       "Álgebra",
		Inicio:      inicio,
		Fin:         inicio.Add(DuracionSesion),
	}
}

type autorizacionFija bool

func (a autorizacionFija) AutorizacionVigente(context.Context, string) (bool, error) {
	return bool(a), nil
}

func TestReservarSoloConAutorizacionVigente(t *testing.T) {
	servicio := NuevoServicio(Dependencias{
		Repositorio:  NuevoRepositorioMemoria(),
		Oferta:       NuevaOfertaSintetica(ahora),
		Autorizacion: autorizacionFija(false),
	})
	servicio.FijarReloj(func() time.Time { return ahora })
	if _, err := servicio.Reservar(context.Background(), "escolar-1", datosValidos()); !errors.Is(err, ErrSinAutorizacion) {
		t.Fatalf("sin autorización no se reserva: %v", err)
	}
}

func TestAnticipacionExactaDe48Horas(t *testing.T) {
	casos := []struct {
		nombre string
		inicio time.Time
		espera error
	}{
		{"exactamente 48 h", ahora.Add(48 * time.Hour), nil},
		{"un nanosegundo menos", ahora.Add(48*time.Hour - time.Nanosecond), ErrAnticipacion},
		{"49 horas", ahora.Add(49 * time.Hour), nil},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			_, err := NuevaReserva("escolar-1", datosValidos(), slotEn(caso.inicio), ahora)
			if caso.espera == nil && err != nil {
				t.Fatalf("debe aceptar: %v", err)
			}
			if caso.espera != nil && !errors.Is(err, caso.espera) {
				t.Fatalf("debe rechazar con %v: %v", caso.espera, err)
			}
		})
	}
}

func TestHorizonteDe28Dias(t *testing.T) {
	if _, err := NuevaReserva("e", datosValidos(), slotEn(ahora.Add(28*24*time.Hour)), ahora); err != nil {
		t.Fatalf("el borde del horizonte es válido: %v", err)
	}
	if _, err := NuevaReserva("e", datosValidos(), slotEn(ahora.Add(29*24*time.Hour)), ahora); !errors.Is(err, ErrHorizonte) {
		t.Fatalf("fuera del horizonte: %v", err)
	}
}

func TestSlotVencido(t *testing.T) {
	if _, err := NuevaReserva("e", datosValidos(), slotEn(ahora.Add(-72*time.Hour)), ahora); !errors.Is(err, ErrSlotVencido) {
		t.Fatalf("un horario pasado no se reserva: %v", err)
	}
}

func TestValidacionDeDatosPreparacion(t *testing.T) {
	casos := []struct {
		nombre string
		cambio func(*DatosReserva)
		campo  string
	}{
		{"tema corto", func(d *DatosReserva) { d.Tema = "corto" }, "tema"},
		{"tema largo", func(d *DatosReserva) { d.Tema = strings.Repeat("a", 201) }, "tema"},
		{"objetivo corto", func(d *DatosReserva) { d.Objetivo = "muy corto" }, "objetivo"},
		{"objetivo largo", func(d *DatosReserva) { d.Objetivo = strings.Repeat("a", 1001) }, "objetivo"},
		{"grado vacío", func(d *DatosReserva) { d.Grado = "  " }, "grado"},
		{"sin slot", func(d *DatosReserva) { d.SlotID = "" }, "slot_id"},
		{"seis adjuntos", func(d *DatosReserva) {
			for i := 0; i < 6; i++ {
				d.Adjuntos = append(d.Adjuntos, Adjunto{Nombre: "a.pdf", Tipo: "application/pdf", TamanoBytes: 10})
			}
		}, "adjuntos"},
		{"adjunto vacío", func(d *DatosReserva) {
			d.Adjuntos = []Adjunto{{Nombre: "a.pdf", TamanoBytes: 0}}
		}, "adjuntos"},
		{"adjunto pesado", func(d *DatosReserva) {
			d.Adjuntos = []Adjunto{{Nombre: "a.pdf", TamanoBytes: AdjuntoMaximoBytes + 1}}
		}, "adjuntos"},
		{"adjunto con ruta", func(d *DatosReserva) {
			d.Adjuntos = []Adjunto{{Nombre: "../../etc/passwd", TamanoBytes: 10}}
		}, "adjuntos"},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			datos := datosValidos()
			caso.cambio(&datos)
			var errValidacion ErrValidacion
			_, err := NuevaReserva("escolar-1", datos, slotEn(ahora.Add(72*time.Hour)), ahora)
			if !errors.As(err, &errValidacion) {
				t.Fatalf("se esperaba ErrValidacion: %v", err)
			}
			if errValidacion.Campos[caso.campo] == "" {
				t.Fatalf("falta el campo %q en %v", caso.campo, errValidacion.Campos)
			}
		})
	}
}

func TestAdjuntoOpcionalNormalizado(t *testing.T) {
	datos := datosValidos()
	datos.Adjuntos = []Adjunto{{Nombre: " practica.pdf ", TamanoBytes: 1024}}
	reserva, err := NuevaReserva("escolar-1", datos, slotEn(ahora.Add(72*time.Hour)), ahora)
	if err != nil {
		t.Fatal(err)
	}
	if reserva.Adjuntos[0].Nombre != "practica.pdf" || reserva.Adjuntos[0].Tipo != "application/octet-stream" {
		t.Fatalf("adjunto normalizado: %+v", reserva.Adjuntos[0])
	}
	if reserva.Estado != EstadoConfirmada || !reserva.CreadaEn.Equal(ahora) {
		t.Fatalf("la reserva válida se confirma de inmediato: %+v", reserva)
	}
}

func TestReservarYNoDuplicarSlot(t *testing.T) {
	servicio := servicioDePrueba(ahora)
	ctx := context.Background()
	reserva, err := servicio.Reservar(ctx, "escolar-1", datosValidos())
	if err != nil {
		t.Fatal(err)
	}
	if reserva.Curso != "Álgebra" || reserva.TutorID != TutorSinteticoID || reserva.ID == "" {
		t.Fatalf("la reserva guarda el horario elegido: %+v", reserva)
	}
	if _, err := servicio.Reservar(ctx, "escolar-2", datosValidos()); !errors.Is(err, ErrSlotOcupado) {
		t.Fatalf("un slot no se reserva dos veces: %v", err)
	}
}

func TestSlotInexistente(t *testing.T) {
	servicio := servicioDePrueba(ahora)
	datos := datosValidos()
	datos.SlotID = "slot-no-existe"
	if _, err := servicio.Reservar(context.Background(), "escolar-1", datos); !errors.Is(err, ErrSlotInexistente) {
		t.Fatalf("slot inexistente: %v", err)
	}
}

func TestSlotsExplicanElRechazo(t *testing.T) {
	servicio := servicioDePrueba(ahora)
	disposiciones, err := servicio.Slots(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(disposiciones) != 6 {
		t.Fatalf("se esperaban 6 horarios sintéticos: %d", len(disposiciones))
	}
	motivos := map[string]string{}
	for _, disposicion := range disposiciones {
		motivos[disposicion.Slot.ID] = disposicion.Motivo
		if disposicion.Slot.Fin.Sub(disposicion.Slot.Inicio) != DuracionSesion {
			t.Fatalf("toda sesión dura 60 minutos: %v", disposicion.Slot)
		}
	}
	if motivos["slot-i04-01"] != "" {
		t.Fatalf("un horario a 3 días debe estar disponible: %q", motivos["slot-i04-01"])
	}
	if !strings.Contains(motivos["slot-i04-cercano"], "48 horas") {
		t.Fatalf("el horario cercano debe explicar las 48 horas: %q", motivos["slot-i04-cercano"])
	}
	if !strings.Contains(motivos["slot-i04-lejano"], "28 días") {
		t.Fatalf("el horario lejano debe explicar el horizonte: %q", motivos["slot-i04-lejano"])
	}
	if _, err := servicio.Reservar(context.Background(), "escolar-1", datosValidos()); err != nil {
		t.Fatal(err)
	}
	disposiciones, _ = servicio.Slots(context.Background())
	for _, disposicion := range disposiciones {
		if disposicion.Slot.ID == "slot-i04-01" && disposicion.Motivo == "" {
			t.Fatal("un horario reservado debe dejar de estar disponible")
		}
	}
}

func TestReservaSoloParaSuEscolar(t *testing.T) {
	servicio := servicioDePrueba(ahora)
	ctx := context.Background()
	reserva, err := servicio.Reservar(ctx, "escolar-1", datosValidos())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := servicio.ReservaDe(ctx, reserva.ID, "escolar-2"); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("otra persona no ve la reserva: %v", err)
	}
	propia, err := servicio.ReservaDe(ctx, reserva.ID, "escolar-1")
	if err != nil || propia.ID != reserva.ID {
		t.Fatalf("el dueño sí la ve: %v", err)
	}
	if _, err := servicio.ReservaDe(ctx, "no-existe", "escolar-1"); !errors.Is(err, ErrReservaNoEncontrada) {
		t.Fatalf("reserva inexistente: %v", err)
	}
	if _, err := servicio.Reservar(ctx, "  ", datosValidos()); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("sin escolar autenticado no hay reserva: %v", err)
	}
}

func TestHorariosEnAmericaLima(t *testing.T) {
	// 12:00 UTC son las 07:00 en Lima; el slot de las 15:00 de Lima son las 20:00 UTC.
	inicio := horaEnLima(ahora, 3, 15)
	if inicio.Hour() != 15 || inicio.Location() != ZonaLima {
		t.Fatalf("hora local esperada 15:00: %v", inicio)
	}
	if inicio.UTC().Hour() != 20 {
		t.Fatalf("15:00 de Lima son 20:00 UTC: %v", inicio.UTC())
	}
}
