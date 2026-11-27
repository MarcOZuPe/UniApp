package reserva

import (
	"context"
	"testing"
	"time"
)

func reservaPrueba(slot string, inicio time.Time, curso, clave string) *Reserva {
	return &Reserva{
		EscolarID: "e-1", TutorID: "t-1", TutorNombre: "Tutor", SlotID: slot,
		Curso: curso, Inicio: inicio, Fin: inicio.Add(time.Hour), Estado: EstadoConfirmada,
		CreadaEn: inicio.Add(-48 * time.Hour), ClaveIdempotencia: clave,
	}
}

func TestRepositorioMemoriaReserva(t *testing.T) {
	repo := NuevoRepositorioMemoria()
	ctx := context.Background()
	base := time.Now().UTC().Add(72 * time.Hour)

	primera, creada, err := repo.ReservarTransaccional(ctx, reservaPrueba("s-1", base, "Álgebra", "clave-1"), LimitesPorDefecto)
	if err != nil || !creada || primera.ID == "" {
		t.Fatalf("primera reserva inesperada: %+v creada=%v err=%v", primera, creada, err)
	}
	repetida, creada, err := repo.ReservarTransaccional(ctx, reservaPrueba("s-1", base, "Álgebra", "clave-1"), LimitesPorDefecto)
	if err != nil || creada || repetida.ID != primera.ID {
		t.Fatalf("la idempotencia debía devolver la reserva original: %+v creada=%v err=%v", repetida, creada, err)
	}
	if _, _, err := repo.ReservarTransaccional(ctx, reservaPrueba("s-1", base, "Álgebra", "otra"), LimitesPorDefecto); err != ErrSlotOcupado {
		t.Fatalf("se esperaba slot ocupado: %v", err)
	}
	if _, _, err := repo.ReservarTransaccional(ctx, reservaPrueba("s-2", base.Add(30*time.Minute), "Física", "clave-2"), LimitesPorDefecto); err != ErrSolapamiento {
		t.Fatalf("se esperaba solapamiento: %v", err)
	}
	if _, _, err := repo.ReservarTransaccional(ctx, reservaPrueba("s-2", base.Add(2*time.Hour), "Física", "clave-3"), LimitesPorDefecto); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.ReservarTransaccional(ctx, reservaPrueba("s-3", base.Add(4*time.Hour), "Química", "clave-4"), LimitesPorDefecto); err != ErrCuotaEscolar {
		t.Fatalf("se esperaba cuota escolar: %v", err)
	}
	if _, err := repo.PorClaveIdempotencia(ctx, "e-1", "clave-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.PorClaveIdempotencia(ctx, "e-1", "inexistente"); err != ErrReservaNoEncontrada {
		t.Fatalf("se esperaba no encontrada: %v", err)
	}
	if _, err := repo.PorID(ctx, "inexistente"); err != ErrReservaNoEncontrada {
		t.Fatalf("se esperaba no encontrada: %v", err)
	}
	porEscolar, err := repo.PorEscolar(ctx, "e-1")
	if err != nil || len(porEscolar) != 2 {
		t.Fatalf("reservas del escolar inesperadas: %+v err=%v", porEscolar, err)
	}
	porTutor, err := repo.PorTutor(ctx, "t-1")
	if err != nil || len(porTutor) != 2 {
		t.Fatalf("reservas del tutor inesperadas: %+v err=%v", porTutor, err)
	}
	if tomado, err := repo.SlotTomado(ctx, "s-1"); err != nil || !tomado {
		t.Fatalf("el slot debería estar tomado: %v %v", tomado, err)
	}
	if tomado, _ := repo.SlotTomado(ctx, "otro"); tomado {
		t.Fatal("un slot libre no debe estar tomado")
	}

	adjunto, err := repo.AgregarAdjunto(ctx, primera.ID, &Adjunto{Nombre: "a.pdf", Tipo: "application/pdf", Estado: AdjuntoValidado})
	if err != nil || adjunto.ID == "" {
		t.Fatalf("adjunto inesperado: %+v err=%v", adjunto, err)
	}
	if _, err := repo.AgregarAdjunto(ctx, "inexistente", &Adjunto{}); err != ErrReservaNoEncontrada {
		t.Fatalf("se esperaba reserva no encontrada: %v", err)
	}
	if cantidad, err := repo.ContarAdjuntos(ctx, primera.ID); err != nil || cantidad != 1 {
		t.Fatalf("conteo inesperado: %d %v", cantidad, err)
	}
	adjuntos, err := repo.AdjuntosDe(ctx, primera.ID)
	if err != nil || len(adjuntos) != 1 {
		t.Fatalf("adjuntos inesperados: %+v err=%v", adjuntos, err)
	}
	if porID, err := repo.AdjuntoPorID(ctx, primera.ID, adjunto.ID); err != nil || porID.Nombre != "a.pdf" {
		t.Fatalf("adjunto por ID inesperado: %+v err=%v", porID, err)
	}
	if _, err := repo.AdjuntoPorID(ctx, primera.ID, "inexistente"); err != ErrAdjuntoNoEncontrado {
		t.Fatalf("se esperaba adjunto no encontrado: %v", err)
	}
	if err := repo.ActualizarAdjunto(ctx, adjunto.ID, AdjuntoRechazado, "motivo"); err != nil {
		t.Fatal(err)
	}
	if err := repo.ActualizarAdjunto(ctx, "inexistente", AdjuntoRechazado, "motivo"); err == nil {
		t.Fatal("un adjunto inexistente debe fallar")
	}
}
