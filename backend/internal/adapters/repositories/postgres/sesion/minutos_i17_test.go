package sesion

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"uniapp/migrations"
)

// baseAisladaMinutos crea una base nueva con TODAS las migraciones, para
// comprobar tambien que 013_sesion_minutos convive con las de otros flujos.
func baseAisladaMinutos(t *testing.T) *sql.DB {
	t.Helper()
	conexion := os.Getenv("TEST_DATABASE_URL")
	if conexion == "" {
		t.Skip("TEST_DATABASE_URL no definida")
	}
	administrador, err := sql.Open("postgres", conexion)
	if err != nil {
		t.Fatal(err)
	}
	nombre := fmt.Sprintf("uniapp_minutos_%d", time.Now().UnixNano())
	if _, err := administrador.Exec("CREATE DATABASE " + nombre); err != nil {
		_ = administrador.Close()
		t.Fatalf("crear base de prueba: %v", err)
	}
	t.Cleanup(func() {
		if _, err := administrador.Exec("DROP DATABASE IF EXISTS " + nombre); err != nil {
			t.Logf("no se pudo borrar la base de prueba: %v", err)
		}
		_ = administrador.Close()
	})
	destino, err := url.Parse(conexion)
	if err != nil {
		t.Fatal(err)
	}
	destino.Path = "/" + nombre
	db, err := migrations.Abrir(destino.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := migrations.Migrar(context.Background(), db, migrations.FS); err != nil {
		t.Fatalf("aplicar migraciones: %v", err)
	}
	return db
}

func TestI17PostgresMinutosYAuditoria(t *testing.T) {
	db := baseAisladaMinutos(t)
	ctx := context.Background()
	inicio := time.Date(2026, 11, 10, 15, 0, 0, 0, time.UTC)
	sesiones := RepositorioPostgres{DB: db}
	s, err := NuevaSesion("sesion-i17-pg", "reserva-i17-pg", "tutor-pg", "escolar-pg", inicio, inicio.Add(DuracionSesion))
	if err != nil {
		t.Fatal(err)
	}
	if err := sesiones.CrearSesion(ctx, s); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"tutor-pg", "escolar-pg"} {
		if _, err := sesiones.DeclararAsistencia(ctx, s.ID(), id, inicio); err != nil {
			t.Fatal(err)
		}
	}
	conAsistencia, err := sesiones.SesionPorID(ctx, s.ID())
	if err != nil {
		t.Fatal(err)
	}
	ahora := inicio.Add(DuracionSesion + time.Minute)
	registro, err := ProponerMinutos(conAsistencia, "tutor-pg", 45, ahora)
	if err != nil {
		t.Fatal(err)
	}

	repo := RepositorioMinutosPostgres{DB: db}
	if err := repo.CrearMinutos(ctx, registro); err != nil {
		t.Fatal(err)
	}
	if err := repo.CrearMinutos(ctx, registro); !errors.Is(err, ErrMinutosDuplicados) {
		t.Fatalf("esperado ErrMinutosDuplicados: %v", err)
	}
	if _, err := repo.MinutosPorSesion(ctx, "no-existe"); !errors.Is(err, ErrMinutosNoEncontrados) {
		t.Fatalf("esperado ErrMinutosNoEncontrados: %v", err)
	}

	pendientes, err := repo.MinutosPendientes(ctx)
	if err != nil || len(pendientes) != 1 {
		t.Fatalf("esperado 1 pendiente: %d %v", len(pendientes), err)
	}

	// Un cambio rechazado no persiste nada (rollback).
	if _, err := repo.ActualizarMinutos(ctx, s.ID(), func(r *RegistroMinutos) error {
		return r.Aprobar("tutor-pg", ahora)
	}); !errors.Is(err, ErrAutoaprobacion) {
		t.Fatalf("esperado ErrAutoaprobacion: %v", err)
	}

	pasos := []func(*RegistroMinutos) error{
		func(r *RegistroMinutos) error { return r.Aprobar("coord-pg", ahora) },
		func(r *RegistroMinutos) error { return r.Ajustar("coord-pg", 40, "Retraso de 5 minutos", ahora) },
		func(r *RegistroMinutos) error { return r.Disputar("tutor-pg", "Dicte 45", ahora) },
		func(r *RegistroMinutos) error { return r.Resolver("coord-pg", 42, "Acuerdo", ahora) },
	}
	for i, paso := range pasos {
		if _, err := repo.ActualizarMinutos(ctx, s.ID(), paso); err != nil {
			t.Fatalf("paso %d: %v", i, err)
		}
	}

	recuperado, err := repo.MinutosPorSesion(ctx, s.ID())
	if err != nil {
		t.Fatal(err)
	}
	if recuperado.Estado() != MinutosVerificados || recuperado.MinutosComputables() != 42 {
		t.Fatalf("estado persistido incorrecto: %s %d", recuperado.Estado(), recuperado.MinutosComputables())
	}
	auditoria := recuperado.Auditoria()
	esperadas := []AccionMinutos{AccionProponer, AccionAprobar, AccionAjustar, AccionDisputar, AccionResolver}
	if len(auditoria) != len(esperadas) {
		t.Fatalf("esperadas %d entradas, hay %d", len(esperadas), len(auditoria))
	}
	for i, accion := range esperadas {
		if auditoria[i].Accion() != accion {
			t.Fatalf("entrada %d: esperado %s, hay %s", i, accion, auditoria[i].Accion())
		}
	}
	if anterior, ok := auditoria[2].ValorAnterior(); !ok || anterior != 45 {
		t.Fatal("el ajuste debe conservar el valor anterior")
	}
	if _, ok := auditoria[3].ValorNuevo(); ok {
		t.Fatal("la disputa no fija valor nuevo")
	}

	delTutor, err := repo.MinutosDeTutor(ctx, "tutor-pg")
	if err != nil || len(delTutor) != 1 {
		t.Fatalf("esperado 1 registro del tutor: %d %v", len(delTutor), err)
	}
	if pendientes, _ := repo.MinutosPendientes(ctx); len(pendientes) != 0 {
		t.Fatal("un verificado no debe quedar pendiente")
	}

	// La auditoria es inmutable en la base de datos.
	if _, err := db.ExecContext(ctx, `UPDATE sesion_minutos_auditoria SET motivo = 'alterado'`); err == nil {
		t.Fatal("la auditoria no debe poder modificarse")
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM sesion_minutos_auditoria`); err == nil {
		t.Fatal("la auditoria no debe poder borrarse")
	}
}
