package incidente

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"uniapp/migrations"
)

func TestRepositorioPostgresIncidentesTutoriasYRevision(t *testing.T) {
	db := baseAisladaIncidente(t)
	ctx := context.Background()
	repositorio := RepositorioPostgres{DB: db}

	escolarID := crearUsuarioIncidente(t, db, "escolar-i13@example.test", 1)
	apoderadoID := crearUsuarioIncidente(t, db, "apoderado-i13@example.test", 2)
	tutorID := crearUsuarioIncidente(t, db, "tutor-i13@uni.edu.pe", 3)
	otroEscolarID := crearUsuarioIncidente(t, db, "otro-i13@example.test", 1)
	coordinadorID := crearUsuarioIncidente(t, db, "coord-i13@example.test", 4)

	reservaID := crearReservaI13(t, db, escolarID, tutorID)
	if _, err := db.ExecContext(ctx, `
		INSERT INTO vinculaciones_apoderado (apoderado_id, escolar_id, estado, expira_en, aceptada_en)
		VALUES ($1, $2, 'aceptada', now() + interval '1 day', now())`, apoderadoID, escolarID); err != nil {
		t.Fatal(err)
	}

	ahora := time.Now().UTC().Add(time.Minute).Truncate(time.Microsecond)
	tutorias, err := repositorio.TutoriasDeEscolar(ctx, escolarID, ahora)
	if err != nil || len(tutorias) != 1 {
		t.Fatalf("tutorías: %+v, error: %v", tutorias, err)
	}
	if !tutorias[0].Asistio || tutorias[0].TutorID != tutorID || tutorias[0].Curso != "Álgebra" {
		t.Fatalf("tutoría con asistencia: %+v", tutorias[0])
	}
	tutoria, err := repositorio.TutoriaDeEscolar(ctx, escolarID, reservaID, ahora)
	if err != nil || tutoria.ReservaID != reservaID {
		t.Fatalf("tutoría puntual: %+v, error: %v", tutoria, err)
	}
	if _, err := repositorio.TutoriaDeEscolar(ctx, otroEscolarID, reservaID, ahora); !errors.Is(err, ErrTutoriaNoEncontrada) {
		t.Fatalf("tutoría de otro escolar: %v", err)
	}
	if _, err := repositorio.TutoriaDeEscolar(ctx, escolarID, reservaID, time.Now().Add(-48*time.Hour)); !errors.Is(err, ErrTutoriaNoEncontrada) {
		t.Fatalf("tutoría futura: %v", err)
	}

	escolares, err := repositorio.EscolaresDeApoderado(ctx, apoderadoID)
	if err != nil || len(escolares) != 1 || escolares[0].ID != escolarID {
		t.Fatalf("escolares del apoderado: %+v, error: %v", escolares, err)
	}
	if vinculado, err := repositorio.EsApoderadoDe(ctx, apoderadoID, escolarID); err != nil || !vinculado {
		t.Fatalf("vinculación: %v, %v", vinculado, err)
	}
	if vinculado, err := repositorio.EsApoderadoDe(ctx, apoderadoID, otroEscolarID); err != nil || vinculado {
		t.Fatalf("vinculación ajena: %v, %v", vinculado, err)
	}

	incidente := &Incidente{
		EscolarID: escolarID, EscolarNombre: "Escolar I13", ReportadoPor: escolarID,
		TutorID: tutorID, TutorNombre: "Tutor I13", ReservaID: reservaID, Curso: "Álgebra",
		Inicio: tutoria.Inicio, Tipo: TipoTratoInadecuado,
		Descripcion: strings.Repeat("Descripción del incidente. ", 2),
		Estado:      EstadoPendiente, CreadoEn: ahora,
		Evidencias: []Evidencia{
			{Nombre: "chat.pdf", Tipo: "application/pdf", TamanoBytes: 20, ClaveAlmacenamiento: "0123456789abcdef0123456789abcdef", CreadoEn: ahora},
			{Nombre: "foto.png", Tipo: "image/png", TamanoBytes: 20, ClaveAlmacenamiento: "fedcba9876543210fedcba9876543210", CreadoEn: ahora},
		},
	}
	creado, err := repositorio.Crear(ctx, incidente)
	if err != nil {
		t.Fatal(err)
	}
	if creado.ID == "" || len(creado.Evidencias) != 2 || creado.Evidencias[0].ID == "" {
		t.Fatalf("incidente persistido: %+v", creado)
	}
	detalle, err := repositorio.PorID(ctx, creado.ID)
	if err != nil || detalle.EscolarNombre != "Prueba I13" || len(detalle.Evidencias) != 2 {
		t.Fatalf("detalle: %+v, error: %v", detalle, err)
	}
	var evidenciaCreada Evidencia
	for _, candidata := range creado.Evidencias {
		if candidata.Nombre == "chat.pdf" {
			evidenciaCreada = candidata
		}
	}
	evidencia, err := repositorio.EvidenciaDe(ctx, creado.ID, evidenciaCreada.ID)
	if err != nil || evidencia.ClaveAlmacenamiento != "0123456789abcdef0123456789abcdef" {
		t.Fatalf("evidencia: %+v, error: %v", evidencia, err)
	}

	if lista, err := repositorio.Listar(ctx, Filtro{EscolarID: escolarID}); err != nil || len(lista) != 1 {
		t.Fatalf("listado del escolar: %+v, error: %v", lista, err)
	}
	if lista, err := repositorio.Listar(ctx, Filtro{EscolaresID: []string{escolarID, otroEscolarID}}); err != nil || len(lista) != 1 {
		t.Fatalf("listado del apoderado: %+v, error: %v", lista, err)
	}
	if lista, err := repositorio.Listar(ctx, Filtro{Todos: true}); err != nil || len(lista) != 1 {
		t.Fatalf("listado de coordinación: %+v, error: %v", lista, err)
	}
	if lista, err := repositorio.Listar(ctx, Filtro{}); err != nil || len(lista) != 0 {
		t.Fatalf("listado sin filtro no debe exponer todo: %+v, error: %v", lista, err)
	}

	revisado, err := repositorio.Revisar(ctx, creado.ID, coordinadorID, ResultadoProcede, "Procede según la evidencia.", ahora.Add(time.Minute))
	if err != nil || revisado.Estado != EstadoRevisado || revisado.RevisadoPor != coordinadorID || revisado.RevisadoEn == nil {
		t.Fatalf("revisión: %+v, error: %v", revisado, err)
	}
	if _, err := repositorio.Revisar(ctx, creado.ID, coordinadorID, ResultadoNoProcede, "Otra revisión.", ahora); !errors.Is(err, ErrYaRevisado) {
		t.Fatalf("segunda revisión: %v", err)
	}
	if _, err := repositorio.Revisar(ctx, "00000000-0000-0000-0000-000000000000", coordinadorID, ResultadoNoProcede, "Otra revisión.", ahora); !errors.Is(err, ErrIncidenteNoEncontrado) {
		t.Fatalf("incidente inexistente: %v", err)
	}
}

func baseAisladaIncidente(t *testing.T) *sql.DB {
	t.Helper()
	conexion := os.Getenv("TEST_DATABASE_URL")
	if conexion == "" {
		t.Skip("TEST_DATABASE_URL no definida")
	}
	administrador, err := sql.Open("postgres", conexion)
	if err != nil {
		t.Fatal(err)
	}
	nombre := fmt.Sprintf("uniapp_incidente_%d", time.Now().UnixNano())
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

func crearUsuarioIncidente(t *testing.T, db *sql.DB, email string, rol int) string {
	t.Helper()
	var id string
	err := db.QueryRow(`INSERT INTO usuarios (email, nombre_uso, clave_hash, rol_id)
		VALUES ($1, 'Prueba I13', 'hash', $2) RETURNING id`, email, rol).Scan(&id)
	if err != nil {
		t.Fatalf("crear usuario de prueba: %v", err)
	}
	return id
}

func crearReservaI13(t *testing.T, db *sql.DB, escolarID, tutorID string) string {
	t.Helper()
	var id string
	err := db.QueryRow(`
		INSERT INTO reservas (
			escolar_id, tutor_id, tutor_nombre, slot_id, curso, grado, inicio, fin, tema, objetivo
		) VALUES (
			$1, $2, 'Tutor I13', 'slot-i13', 'Álgebra', '5.° de secundaria',
			now() - interval '2 hours', now() - interval '1 hour',
			'Repaso de factorización', 'Resolver ejercicios de factorización con acompañamiento.'
		) RETURNING id`, escolarID, tutorID).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		INSERT INTO sesion_sesiones (id, reserva_id, tutor_id, escolar_id, inicio, fin)
		VALUES ('sesion-i13', $1::text, $2::text, $3::text,
		        now() - interval '2 hours', now() - interval '1 hour')`, id, tutorID, escolarID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		INSERT INTO sesion_asistencias (sesion_id, usuario_id, rol, declarada_en)
		VALUES ('sesion-i13', $1::text, 'escolar', now())`, escolarID); err != nil {
		t.Fatal(err)
	}
	return id
}
