package panel

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	modelo "uniapp/internal/domain/panel"
	"uniapp/migrations"
)

func TestRepositorioPostgresMetricasAuditoriaEInmutabilidad(t *testing.T) {
	db := baseAisladaPanel(t)
	ctx := context.Background()
	repositorio := RepositorioPostgres{DB: db}

	escolarID := crearUsuarioPanel(t, db, "escolar-i14@example.test", 1)
	apoderadoID := crearUsuarioPanel(t, db, "apoderado-i14@example.test", 2)
	tutorID := crearUsuarioPanel(t, db, "tutor-i14@uni.edu.pe", 3)
	generalID := crearUsuarioPanel(t, db, "general-i14@example.test", 5)

	ahora := time.Now().UTC().Truncate(time.Microsecond)
	inicio := ahora.Add(-2 * time.Hour)
	fin := inicio.Add(time.Hour)
	reservaID := insertarReservaPanel(t, db, escolarID, tutorID, inicio, fin)
	insertarSesionPanel(t, db, "sesion-i14-test", reservaID, tutorID, escolarID, inicio, fin)

	incidenteID := insertarIncidentePanel(t, db, escolarID, escolarID, tutorID, reservaID, inicio, ahora)
	if _, err := db.ExecContext(ctx, `
		UPDATE incidentes
		SET estado = 'revisado', resultado = 'procede',
		    comentario_revision = 'Procede según la evidencia.',
		    revisado_por = $2, revisado_en = now()
		WHERE id = $1`, incidenteID, generalID); err != nil {
		t.Fatal(err)
	}

	vinculacionID := insertarVinculacionPanel(t, db, apoderadoID, escolarID, ahora)
	if _, err := db.ExecContext(ctx, `
		INSERT INTO autorizaciones_apoderado (vinculacion_id, apoderado_id, escolar_id, estado)
		VALUES ($1, $2, $3, 'autorizado')`, vinculacionID, apoderadoID, escolarID); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO cupos_semanales (tutor_id, minutos) VALUES ($1, 180)`, tutorID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
		UPDATE cupos_semanales SET minutos = 240 WHERE tutor_id = $1`, tutorID); err != nil {
		t.Fatal(err)
	}

	metricas, err := repositorio.Metricas(ctx, ahora)
	if err != nil {
		t.Fatal(err)
	}
	if metricas.Reservas.Totales != 1 || metricas.Reservas.Finalizadas != 1 ||
		metricas.Reservas.Futuras != 0 {
		t.Fatalf("métricas de reservas: %+v", metricas.Reservas)
	}
	if metricas.Asistencia.SesionesFinalizadas != 1 ||
		metricas.Asistencia.ConAsistenciaEscolar != 1 ||
		metricas.Asistencia.ConAsistenciaTutor != 1 ||
		metricas.Asistencia.TasaEscolarPorcentaje != 100 {
		t.Fatalf("métricas de asistencia: %+v", metricas.Asistencia)
	}
	if metricas.Horas.Internas != 1 {
		t.Fatalf("horas internas: %+v", metricas.Horas)
	}

	eventos, err := repositorio.Auditoria(ctx, modelo.FiltroAuditoria{Limite: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(eventos) < 4 {
		t.Fatalf("faltan eventos de auditoría: %+v", eventos)
	}
	var revisado, autorizacion, cupo bool
	for _, evento := range eventos {
		switch evento.Accion {
		case "incidente.revisado":
			revisado = evento.ActorRol == "coordinador_general" && evento.ActorID == generalID
		case "autorizacion.cambiada":
			autorizacion = evento.ActorRol == "apoderado" && evento.ActorID == apoderadoID
		case "cupo.actualizado":
			cupo = evento.ActorRol == "tutor" && evento.ActorID == tutorID
		}
		for _, clave := range []string{"token", "email", "clave", "huella", "clave_almacenamiento"} {
			if _, existe := evento.Detalle[clave]; existe {
				t.Fatalf("la auditoría no debe guardar secretos (%s): %+v", clave, evento.Detalle)
			}
		}
	}
	if !revisado || !autorizacion || !cupo {
		t.Fatalf("auditoría incompleta: revisado=%v autorizacion=%v cupo=%v", revisado, autorizacion, cupo)
	}

	if _, err := db.ExecContext(ctx, `UPDATE auditoria_eventos SET accion = 'alterada'`); err == nil ||
		!strings.Contains(err.Error(), "inmutable") {
		t.Fatalf("la auditoría no debe permitir UPDATE: %v", err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM auditoria_eventos`); err == nil ||
		!strings.Contains(err.Error(), "inmutable") {
		t.Fatalf("la auditoría no debe permitir DELETE: %v", err)
	}
	if _, err := db.ExecContext(ctx, `TRUNCATE auditoria_eventos`); err == nil ||
		!strings.Contains(err.Error(), "inmutable") {
		t.Fatalf("la auditoría no debe permitir TRUNCATE: %v", err)
	}
}

func baseAisladaPanel(t *testing.T) *sql.DB {
	t.Helper()
	conexion := os.Getenv("TEST_DATABASE_URL")
	if conexion == "" {
		t.Skip("TEST_DATABASE_URL no definida")
	}
	administrador, err := sql.Open("postgres", conexion)
	if err != nil {
		t.Fatal(err)
	}
	nombre := fmt.Sprintf("uniapp_panel_%d", time.Now().UnixNano())
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

func crearUsuarioPanel(t *testing.T, db *sql.DB, email string, rol int) string {
	t.Helper()
	var id string
	if err := db.QueryRow(`
		INSERT INTO usuarios (email, nombre_uso, clave_hash, rol_id)
		VALUES ($1, $2, 'hash', $3) RETURNING id::text`,
		email, "Prueba I14", rol,
	).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func insertarReservaPanel(
	t *testing.T,
	db *sql.DB,
	escolarID, tutorID string,
	inicio, fin time.Time,
) string {
	t.Helper()
	var id string
	if err := db.QueryRow(`
		INSERT INTO reservas (
			escolar_id, tutor_id, tutor_nombre, slot_id, curso, grado,
			inicio, fin, tema, objetivo
		) VALUES ($1, $2, 'Tutor I14', 'slot-i14', 'Química', '5to de secundaria',
			$3, $4, 'Repaso de enlaces químicos', 'Comprender los enlaces químicos y sus ejemplos.')
		RETURNING id::text`,
		escolarID, tutorID, inicio, fin,
	).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func insertarSesionPanel(
	t *testing.T,
	db *sql.DB,
	sesionID, reservaID, tutorID, escolarID string,
	inicio, fin time.Time,
) {
	t.Helper()
	if _, err := db.Exec(`
		INSERT INTO sesion_sesiones (id, reserva_id, tutor_id, escolar_id, inicio, fin)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		sesionID, reservaID, tutorID, escolarID, inicio, fin,
	); err != nil {
		t.Fatal(err)
	}
	for _, asistencia := range []struct {
		usuarioID string
		rol       string
	}{{escolarID, "escolar"}, {tutorID, "tutor"}} {
		if _, err := db.Exec(`
			INSERT INTO sesion_asistencias (sesion_id, usuario_id, rol, declarada_en)
			VALUES ($1, $2, $3, now())`, sesionID, asistencia.usuarioID, asistencia.rol,
		); err != nil {
			t.Fatal(err)
		}
	}
}

func insertarIncidentePanel(
	t *testing.T,
	db *sql.DB,
	escolarID, reportadoPor, tutorID, reservaID string,
	inicio, creadoEn time.Time,
) string {
	t.Helper()
	var id string
	if err := db.QueryRow(`
		INSERT INTO incidentes (
			escolar_id, reportado_por, tutor_id, tutor_nombre, reserva_id, curso,
			inicio_sesion, tipo, descripcion, estado, creado_en
		) VALUES ($1, $2, $3, 'Tutor I14', $4, 'Química', $5,
			'trato_inadecuado', 'Descripción suficiente del incidente para la prueba.', 'pendiente', $6)
		RETURNING id::text`,
		escolarID, reportadoPor, tutorID, reservaID, inicio, creadoEn,
	).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func insertarVinculacionPanel(
	t *testing.T,
	db *sql.DB,
	apoderadoID, escolarID string,
	ahora time.Time,
) string {
	t.Helper()
	var id string
	if err := db.QueryRow(`
		INSERT INTO vinculaciones_apoderado (
			apoderado_id, escolar_id, estado, expira_en, creada_en, aceptada_en
		) VALUES ($1, $2, 'aceptada', $3, $4, $4)
		RETURNING id::text`,
		apoderadoID, escolarID, ahora.Add(24*time.Hour), ahora,
	).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}
