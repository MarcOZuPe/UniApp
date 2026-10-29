package catalogo

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

func TestRepositorioPostgresImportacionAprobacionYSolicitudes(t *testing.T) {
	db := baseAisladaCatalogo(t)
	ctx := context.Background()
	repositorio := RepositorioPostgres{DB: db}
	semilla, err := repositorio.ImportacionActual(ctx)
	if err != nil || semilla.Fuente != "sintetico: fixture inicial I07" || len(semilla.Cursos) == 0 || len(semilla.Colegios) == 0 {
		t.Fatalf("la migración debe registrar catálogo sintético y su origen: %+v, error: %v", semilla, err)
	}
	coordinadorID := crearUsuarioCatalogo(t, db, "coord-i07@example.test", 4)
	escolarID := crearUsuarioCatalogo(t, db, "escolar-i07@example.test", 1)
	ahora := time.Now().UTC().Add(time.Minute).Truncate(time.Microsecond)
	importacion, err := NuevaImportacion(datosValidosCatalogo(), ahora)
	if err != nil {
		t.Fatal(err)
	}
	if err := repositorio.Importar(ctx, importacion); err != nil {
		t.Fatal(err)
	}

	actual, err := repositorio.ImportacionActual(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if actual.ID != importacion.ID || actual.Fuente != "sintetico: prueba" || !actual.ImportadoEn.Equal(ahora) {
		t.Fatalf("metadatos de importación no persistidos: %+v", actual)
	}
	curso, err := repositorio.Curso(ctx, actual.Cursos[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := curso.Revisar(EstadoAprobado, coordinadorID, "", ahora); err != nil {
		t.Fatal(err)
	}
	if err := repositorio.RevisarCurso(ctx, curso, coordinadorID); err != nil {
		t.Fatal(err)
	}
	revisado, err := repositorio.Curso(ctx, curso.ID)
	if err != nil || revisado.Estado != EstadoAprobado || revisado.RevisadoPor != coordinadorID {
		t.Fatalf("la revisión debe persistir su responsable y estado: %+v, error: %v", revisado, err)
	}

	solicitud, err := NuevaSolicitudColegio(SolicitudColegioEntrada{
		Nombre: "Colegio Faltante", Distrito: "Distrito Nuevo",
	}, ahora)
	if err != nil {
		t.Fatal(err)
	}
	if err := repositorio.SolicitarColegio(ctx, solicitud, escolarID); err != nil {
		t.Fatal(err)
	}
	if err := repositorio.SolicitarColegio(ctx, solicitud, escolarID); !errors.Is(err, ErrSolicitudDuplicada) {
		t.Fatalf("la unicidad de solicitud debe persistir: %v", err)
	}

	datos := datosValidosCatalogo()
	datos.Colegios = append(datos.Colegios, ColegioEntrada{Nombre: "Colegio Faltante", Distrito: "Distrito Nuevo"})
	siguiente, err := NuevaImportacion(datos, ahora.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if err := repositorio.Importar(ctx, siguiente); err != nil {
		t.Fatal(err)
	}
	pendientes, err := repositorio.SolicitudesPendientes(ctx)
	if err != nil || len(pendientes) != 0 {
		t.Fatalf("la importación debe resolver la solicitud coincidente: %+v, error: %v", pendientes, err)
	}
	if _, err := repositorio.Curso(ctx, curso.ID); !errors.Is(err, ErrNoEncontrado) {
		t.Fatalf("solo se puede revisar un curso de la importación actual: %v", err)
	}
}

func baseAisladaCatalogo(t *testing.T) *sql.DB {
	t.Helper()
	conexion := os.Getenv("TEST_DATABASE_URL")
	if conexion == "" {
		t.Skip("TEST_DATABASE_URL no definida")
	}
	administrador, err := sql.Open("postgres", conexion)
	if err != nil {
		t.Fatal(err)
	}
	nombre := fmt.Sprintf("uniapp_catalogo_%d", time.Now().UnixNano())
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

func crearUsuarioCatalogo(t *testing.T, db *sql.DB, email string, rol int) string {
	t.Helper()
	var id string
	err := db.QueryRow(`INSERT INTO usuarios (email, nombre_uso, clave_hash, rol_id)
		VALUES ($1, 'Prueba I07', 'hash', $2) RETURNING id`, email, rol).Scan(&id)
	if err != nil {
		t.Fatalf("crear usuario de prueba: %v", err)
	}
	return id
}

func datosValidosCatalogo() DatosImportacion {
	return DatosImportacion{
		Fuente: "sintetico: prueba",
		Cursos: []CursoEntrada{
			{Nombre: "Álgebra", Grado: "1.° de secundaria"},
			{Nombre: "Comunicación", Grado: "2.° de secundaria"},
		},
		Colegios: []ColegioEntrada{
			{Nombre: "Colegio Sintético", Distrito: "Distrito Sintético"},
		},
	}
}
