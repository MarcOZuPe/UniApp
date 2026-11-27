package sesion

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRepositorioMemoriaSesion(t *testing.T) {
	repo := NuevoRepositorioMemoria()
	ctx := context.Background()
	inicio := time.Now().UTC().Add(-2 * time.Hour)
	sesion, err := NuevaSesion("s-1", "r-1", "tutor-1", "escolar-1", inicio, inicio.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.CrearSesion(ctx, sesion); err != nil {
		t.Fatalf("crear sesión falló: %v", err)
	}
	if err := repo.CrearSesion(ctx, sesion); err != ErrSesionDuplicada {
		t.Fatalf("se esperaba sesión duplicada: %v", err)
	}
	otra, _ := NuevaSesion("s-2", "r-1", "tutor-2", "escolar-2", inicio, inicio.Add(time.Hour))
	if err := repo.CrearSesion(ctx, otra); err != ErrSesionDuplicada {
		t.Fatalf("se esperaba reserva duplicada: %v", err)
	}
	if _, err := repo.SesionPorID(ctx, "s-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.SesionPorID(ctx, "inexistente"); err != ErrSesionNoEncontrada {
		t.Fatalf("se esperaba no encontrada: %v", err)
	}
	if _, err := repo.SesionPorReservaID(ctx, "r-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.SesionPorReservaID(ctx, "inexistente"); err != ErrSesionNoEncontrada {
		t.Fatalf("se esperaba no encontrada: %v", err)
	}
	ahora := inicio.Add(30 * time.Minute)
	if _, err := repo.DeclararAsistencia(ctx, "s-1", "escolar-1", ahora); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.DeclararAsistencia(ctx, "inexistente", "escolar-1", ahora); err != ErrSesionNoEncontrada {
		t.Fatalf("se esperaba no encontrada: %v", err)
	}
	if _, err := repo.ReportarAusencia(ctx, "s-1", "tutor-1", ahora); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ReportarAusencia(ctx, "inexistente", "tutor-1", ahora); err != ErrSesionNoEncontrada {
		t.Fatalf("se esperaba no encontrada: %v", err)
	}
	resultado, err := repo.RegistrarResultado(ctx, "s-1", "tutor-1", "Resultado", inicio.Add(2*time.Hour))
	if err != nil || resultado.Resumen() != "Resultado" {
		t.Fatalf("resultado inesperado: %+v err=%v", resultado, err)
	}
	if _, err := repo.RegistrarResultado(ctx, "inexistente", "tutor-1", "Resultado", ahora); err != ErrSesionNoEncontrada {
		t.Fatalf("se esperaba no encontrada: %v", err)
	}
}

func TestRepositorioMemoriaMinutos(t *testing.T) {
	repo := NuevoRepositorioMinutosMemoria()
	ctx := context.Background()
	inicio := time.Now().UTC().Add(-2 * time.Hour)
	fin := inicio.Add(time.Hour)
	sesion, _ := NuevaSesion("s-1", "r-1", "tutor-1", "escolar-1", inicio, fin)
	_ = sesion.DeclararAsistencia("tutor-1", inicio.Add(30*time.Minute))
	_ = sesion.DeclararAsistencia("escolar-1", inicio.Add(30*time.Minute))
	registro, err := ProponerMinutos(sesion, "tutor-1", 60, fin.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.CrearMinutos(ctx, registro); err != nil {
		t.Fatalf("crear minutos falló: %v", err)
	}
	if err := repo.CrearMinutos(ctx, registro); err != ErrMinutosDuplicados {
		t.Fatalf("se esperaba duplicado: %v", err)
	}
	if _, err := repo.MinutosPorSesion(ctx, "s-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.MinutosPorSesion(ctx, "inexistente"); err != ErrMinutosNoEncontrados {
		t.Fatalf("se esperaba no encontrado: %v", err)
	}
	if pendientes, err := repo.MinutosPendientes(ctx); err != nil || len(pendientes) != 1 {
		t.Fatalf("pendientes inesperados: %+v err=%v", pendientes, err)
	}
	if delTutor, err := repo.MinutosDeTutor(ctx, "tutor-1"); err != nil || len(delTutor) != 1 {
		t.Fatalf("minutos del tutor inesperados: %+v err=%v", delTutor, err)
	}
	actualizado, err := repo.ActualizarMinutos(ctx, "s-1", func(registro *RegistroMinutos) error {
		return registro.Aprobar("coord-1", fin.Add(2*time.Hour))
	})
	if err != nil || actualizado.Estado() != MinutosVerificados {
		t.Fatalf("actualización inesperada: %+v err=%v", actualizado, err)
	}
	if _, err := repo.ActualizarMinutos(ctx, "s-1", func(*RegistroMinutos) error { return ErrTransicionMinutos }); err != ErrTransicionMinutos {
		t.Fatalf("el error del cambio debe propagarse: %v", err)
	}
	if _, err := repo.ActualizarMinutos(ctx, "inexistente", func(*RegistroMinutos) error { return nil }); err != ErrMinutosNoEncontrados {
		t.Fatalf("se esperaba no encontrado: %v", err)
	}
}

func TestCancelacionesYFixture(t *testing.T) {
	cancelaciones := NuevasCancelacionesFijas("r-1", "r-2")
	if cancelada, err := cancelaciones.Cancelada(context.Background(), "r-1"); err != nil || !cancelada {
		t.Fatalf("r-1 debería estar cancelada: %v %v", cancelada, err)
	}
	if cancelada, _ := cancelaciones.Cancelada(context.Background(), "r-3"); cancelada {
		t.Fatal("r-3 no debería estar cancelada")
	}
	if _, err := CargarCancelacionesFixture("/ruta/inexistente"); err == nil {
		t.Fatal("un fixture inexistente debe fallar")
	}
	ruta := filepath.Join(t.TempDir(), "cancelaciones.json")
	if err := os.WriteFile(ruta, []byte(`{"reservas_canceladas":["r-4"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	fixture, err := CargarCancelacionesFixture(ruta)
	if err != nil {
		t.Fatal(err)
	}
	if cancelada, _ := fixture.Cancelada(context.Background(), "r-4"); !cancelada {
		t.Fatal("r-4 debería estar cancelada")
	}
}
