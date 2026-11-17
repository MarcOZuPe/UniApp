package tutor

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

func TestRepositorioPostgresExpedienteRevisionYPerfiles(t *testing.T) {
	db := baseAisladaTutor(t)
	ctx := context.Background()
	repositorio := RepositorioPostgres{DB: db}
	tutorID := crearUsuarioTutor(t, db, "tutor-i02@example.test", 3)
	revisorID := crearUsuarioTutor(t, db, "coord-i02@example.test", 4)
	ahora := time.Now().UTC().Truncate(time.Microsecond)

	expediente := expedienteValido(t, tutorID, ahora, "11111111-1111-4111-8111-111111111111", "constancia.pdf")
	if err := repositorio.Guardar(ctx, expediente); err != nil {
		t.Fatal(err)
	}
	guardado, err := repositorio.PorTutor(ctx, tutorID)
	if err != nil {
		t.Fatal(err)
	}
	if guardado.Estado != EstadoPendiente || len(guardado.Documentos) != 1 {
		t.Fatalf("expediente no persistido: %+v", guardado)
	}
	if guardado.Documentos[0].ID != expediente.Documentos[0].ID {
		t.Fatalf("documento no persistido: %+v", guardado.Documentos)
	}
	pendientes, err := repositorio.Pendientes(ctx)
	if err != nil || len(pendientes) != 1 || pendientes[0].TutorID != tutorID {
		t.Fatalf("cola de revisión: %+v, error: %v", pendientes, err)
	}
	documento, propietario, err := repositorio.Documento(ctx, expediente.Documentos[0].ID)
	if err != nil || documento.Nombre != "constancia.pdf" || propietario != tutorID {
		t.Fatalf("metadatos del documento: %+v, propietario=%q, error=%v", documento, propietario, err)
	}

	venceEn := ahora.Add(24 * time.Hour)
	if err := guardado.Revisar(EstadoAprobado, revisorID, "", venceEn, ahora); err != nil {
		t.Fatal(err)
	}
	if err := repositorio.Revisar(ctx, guardado); err != nil {
		t.Fatal(err)
	}
	aprobado, err := repositorio.PorTutor(ctx, tutorID)
	if err != nil {
		t.Fatal(err)
	}
	if aprobado.Estado != EstadoAprobado || aprobado.VenceEn == nil || !aprobado.VenceEn.Equal(venceEn) || aprobado.RevisadoPor != revisorID {
		t.Fatalf("revisión no persistida: %+v", aprobado)
	}
	if err := repositorio.Revisar(ctx, guardado); !errors.Is(err, ErrEstadoInvalido) {
		t.Fatalf("una segunda revisión debe ser rechazada: %v", err)
	}

	perfiles, err := repositorio.PerfilesPublicos(ctx, venceEn.Add(-time.Second))
	if err != nil || len(perfiles) != 1 || perfiles[0].TutorID != tutorID {
		t.Fatalf("perfil vigente no publicado: %+v, error: %v", perfiles, err)
	}
	perfiles, err = repositorio.PerfilesPublicos(ctx, venceEn)
	if err != nil || len(perfiles) != 0 {
		t.Fatalf("perfil vencido no debe publicarse: %+v, error: %v", perfiles, err)
	}
	pendientes, err = repositorio.Pendientes(ctx)
	if err != nil || len(pendientes) != 0 {
		t.Fatalf("la aprobación debe vaciar la cola: %+v, error: %v", pendientes, err)
	}
}

func TestRepositorioPostgresReemplazaDocumentosAlVolverAPresentar(t *testing.T) {
	db := baseAisladaTutor(t)
	ctx := context.Background()
	repositorio := RepositorioPostgres{DB: db}
	tutorID := crearUsuarioTutor(t, db, "tutor-reintento@example.test", 3)
	revisorID := crearUsuarioTutor(t, db, "coord-reintento@example.test", 4)
	ahora := time.Now().UTC().Truncate(time.Microsecond)

	primero := expedienteValido(t, tutorID, ahora, "22222222-2222-4222-8222-222222222222", "anterior.pdf")
	if err := repositorio.Guardar(ctx, primero); err != nil {
		t.Fatal(err)
	}
	primero, err := repositorio.PorTutor(ctx, tutorID)
	if err != nil {
		t.Fatal(err)
	}
	if err := primero.Revisar(EstadoRechazado, revisorID, "Falta sustento", time.Time{}, ahora); err != nil {
		t.Fatal(err)
	}
	if err := repositorio.Revisar(ctx, primero); err != nil {
		t.Fatal(err)
	}

	segundo := expedienteValido(t, tutorID, ahora.Add(time.Minute), "33333333-3333-4333-8333-333333333333", "nueva.pdf")
	segundo.Presentacion = "Nueva presentación"
	if err := repositorio.Guardar(ctx, segundo); err != nil {
		t.Fatal(err)
	}
	actual, err := repositorio.PorTutor(ctx, tutorID)
	if err != nil {
		t.Fatal(err)
	}
	if actual.Estado != EstadoPendiente || len(actual.Documentos) != 1 || actual.Documentos[0].ID != segundo.Documentos[0].ID {
		t.Fatalf("el reenvío debe conservar solo los documentos nuevos: %+v", actual)
	}
	if actual.Motivo != "" || actual.VenceEn != nil || actual.RevisadoPor != "" {
		t.Fatalf("el reenvío debe limpiar la revisión anterior: %+v", actual)
	}
	if _, _, err := repositorio.Documento(ctx, primero.Documentos[0].ID); !errors.Is(err, ErrNoEncontrado) {
		t.Fatalf("el documento anterior debe dejar de existir: %v", err)
	}
	if _, propietario, err := repositorio.Documento(ctx, segundo.Documentos[0].ID); err != nil || propietario != tutorID {
		t.Fatalf("el documento nuevo debe pertenecer al tutor: %q, %v", propietario, err)
	}
}

func baseAisladaTutor(t *testing.T) *sql.DB {
	t.Helper()
	conexion := os.Getenv("TEST_DATABASE_URL")
	if conexion == "" {
		t.Skip("TEST_DATABASE_URL no definida")
	}
	administrador, err := sql.Open("postgres", conexion)
	if err != nil {
		t.Fatal(err)
	}
	nombre := fmt.Sprintf("uniapp_tutor_%d", time.Now().UnixNano())
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

func crearUsuarioTutor(t *testing.T, db *sql.DB, email string, rol int) string {
	t.Helper()
	var id string
	err := db.QueryRow(`INSERT INTO usuarios (email, nombre_uso, clave_hash, rol_id)
		VALUES ($1, 'Prueba I02', 'hash', $2) RETURNING id`, email, rol).Scan(&id)
	if err != nil {
		t.Fatalf("crear usuario de prueba: %v", err)
	}
	return id
}

func expedienteValido(t *testing.T, tutorID string, ahora time.Time, documentoID, nombre string) *Expediente {
	t.Helper()
	expediente, err := NuevoExpediente(tutorID, DatosExpediente{
		Presentacion: "Presentación sintética de prueba",
		Cursos:       "Álgebra",
		Experiencia:  "Experiencia sintética de prueba",
		Documentos: []Archivo{{
			Nombre:    nombre,
			Contenido: []byte("%PDF-1.4\ncontenido sintético"),
		}},
	}, ahora)
	if err != nil {
		t.Fatal(err)
	}
	expediente.Documentos[0].ID = documentoID
	return expediente
}
