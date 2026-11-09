package sesion

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
	"uniapp/migrations"
)

// Tablas temporales en una única conexión: nunca escribe ni borra tablas públicas.
func prepararPostgresSesionAislado(t *testing.T) (*sql.DB, context.Context) {
	t.Helper()
	conexion := os.Getenv("I05_TEST_DATABASE_URL")
	if conexion == "" {
		t.Skip("I05_TEST_DATABASE_URL no definida; requiere PostgreSQL disponible")
	}
	// El aislamiento se aplica también a cualquier reconexión del driver.
	if strings.HasPrefix(conexion, "postgres://") || strings.HasPrefix(conexion, "postgresql://") {
		u, err := url.Parse(conexion)
		if err != nil {
			t.Fatal("URL de PostgreSQL inválida")
		}
		q := u.Query()
		q.Set("search_path", "pg_temp")
		u.RawQuery = q.Encode()
		conexion = u.String()
	} else {
		conexion += " search_path=pg_temp"
	}
	db, err := sql.Open("postgres", conexion)
	if err != nil {
		t.Fatal("no se pudo abrir PostgreSQL")
	}
	t.Cleanup(func() { db.Close() })
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)
	if err := db.PingContext(ctx); err != nil {
		t.Fatal("PostgreSQL no disponible")
	}
	// Si se pierde la conexión, el search_path impide caer en las tablas públicas.
	if _, err := db.ExecContext(ctx, "SET search_path TO pg_temp"); err != nil {
		t.Fatal("no se pudo aislar el esquema")
	}
	for _, nombre := range []string{"005_sesion.up.sql", "010_sesion_ausencias.up.sql", "011_sesion_resultados.up.sql"} {
		migracion, err := migrations.FS.ReadFile(nombre)
		if err != nil {
			t.Fatal(err)
		}
		sqlTemporal := strings.ReplaceAll(string(migracion), "CREATE TABLE", "CREATE TEMP TABLE")
		if _, err := db.ExecContext(ctx, sqlTemporal); err != nil {
			t.Fatal("no se pudieron crear tablas temporales")
		}
	}
	return db, ctx
}

func TestI05PostgresPersistenciaAislada(t *testing.T) {
	db, ctx := prepararPostgresSesionAislado(t)
	inicio := time.Date(2026, 10, 20, 15, 0, 0, 0, time.UTC)
	s, err := NuevaSesion("sesion-test-aislada", "reserva-test-aislada", "tutor-test", "escolar-test", inicio, inicio.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	repo := RepositorioPostgres{DB: db}
	if err := repo.CrearSesion(ctx, s); err != nil {
		t.Fatal("falló crear sesión temporal")
	}
	for _, usuario := range []string{"tutor-test", "escolar-test"} {
		if _, err := repo.DeclararAsistencia(ctx, s.ID(), usuario, inicio); err != nil {
			t.Fatal("falló persistir asistencia temporal")
		}
		if _, err := repo.DeclararAsistencia(ctx, s.ID(), usuario, inicio); !errors.Is(err, ErrAsistenciaDuplicada) {
			t.Fatal("no rechazó duplicado")
		}
	}
	if _, err := repo.DeclararAsistencia(ctx, s.ID(), "ajeno", inicio); !errors.Is(err, ErrNoAutorizado) {
		t.Fatal("no rechazó usuario ajeno")
	}
	// Nueva instancia del repositorio: lectura del almacenamiento, no del objeto original.
	otra := RepositorioPostgres{DB: db}
	guardada, err := otra.SesionPorID(ctx, s.ID())
	if err != nil {
		t.Fatal("falló recargar la sesión temporal")
	}
	if len(guardada.Asistencias()) != 2 {
		t.Fatal("asistencias no persistidas")
	}
	for i, rol := range []Rol{RolEscolar, RolTutor} {
		if guardada.Asistencias()[i].Rol() != rol {
			t.Fatal("rol persistido incorrecto")
		}
	}
}

func TestI11PostgresAusenciaAislada(t *testing.T) {
	db, ctx := prepararPostgresSesionAislado(t)
	inicio := time.Date(2026, 10, 20, 15, 0, 0, 0, time.UTC)
	s, err := NuevaSesion("sesion-i11", "reserva-i11", "tutor", "escolar", inicio, inicio.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	repo := RepositorioPostgres{DB: db}
	if err := repo.CrearSesion(ctx, s); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ReportarAusencia(ctx, s.ID(), "ajeno", inicio); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("autorizacion: %v", err)
	}
	a, err := repo.ReportarAusencia(ctx, s.ID(), s.TutorID(), inicio)
	if err != nil {
		t.Fatal(err)
	}
	if a.UsuarioID() != s.TutorID() || a.Rol() != RolTutor {
		t.Fatal("identidad incorrecta")
	}
	if _, err := repo.ReportarAusencia(ctx, s.ID(), s.TutorID(), inicio); !errors.Is(err, ErrAusenciaDuplicada) {
		t.Fatalf("duplicado: %v", err)
	}
	if _, err := repo.DeclararAsistencia(ctx, s.ID(), s.TutorID(), inicio); !errors.Is(err, ErrDeclaracionContradictoria) {
		t.Fatalf("contradiccion: %v", err)
	}
	if _, err := repo.DeclararAsistencia(ctx, s.ID(), s.EscolarID(), inicio); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ReportarAusencia(ctx, s.ID(), s.EscolarID(), inicio); !errors.Is(err, ErrDeclaracionContradictoria) {
		t.Fatalf("contradiccion inversa: %v", err)
	}
	otra := RepositorioPostgres{DB: db}
	guardada, err := otra.SesionPorID(ctx, s.ID())
	if err != nil {
		t.Fatal(err)
	}
	if len(guardada.Ausencias()) != 1 || len(guardada.Asistencias()) != 1 {
		t.Fatal("declaraciones no persistidas")
	}
	if guardada.Ausencias()[0].UsuarioID() != s.TutorID() || !guardada.Ausencias()[0].ReportadaEn().Equal(inicio) {
		t.Fatal("ausencia persistida incorrecta")
	}
}

func TestI11PostgresResultadoAislado(t *testing.T) {
	db, ctx := prepararPostgresSesionAislado(t)
	inicio := time.Date(2026, 10, 20, 15, 0, 0, 0, time.UTC)
	s, err := NuevaSesion("resultado-i11", "reserva-resultado-i11", "tutor", "escolar", inicio, inicio.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	repo := RepositorioPostgres{DB: db}
	if err := repo.CrearSesion(ctx, s); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.RegistrarResultado(ctx, s.ID(), s.EscolarID(), "Resumen", s.Fin()); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("autorizacion: %v", err)
	}
	if _, err := repo.RegistrarResultado(ctx, s.ID(), s.TutorID(), "  ", s.Fin()); !errors.Is(err, ErrResultadoInvalido) {
		t.Fatalf("validacion: %v", err)
	}
	if _, err := repo.RegistrarResultado(ctx, s.ID(), s.TutorID(), "  Ecuaciones resueltas.  ", s.Fin()); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.RegistrarResultado(ctx, s.ID(), s.TutorID(), "Sobrescribir", s.Fin()); !errors.Is(err, ErrResultadoDuplicado) {
		t.Fatalf("duplicado: %v", err)
	}
	otra := RepositorioPostgres{DB: db}
	guardada, err := otra.SesionPorReservaID(ctx, s.ReservaID())
	if err != nil {
		t.Fatal(err)
	}
	r := guardada.Resultado()
	if r == nil || r.Resumen() != "Ecuaciones resueltas." || r.UsuarioID() != s.TutorID() || !r.RegistradoEn().Equal(s.Fin()) {
		t.Fatal("resultado no persistido correctamente")
	}
	if len(guardada.Asistencias()) != 0 || len(guardada.Ausencias()) != 0 {
		t.Fatal("el resultado genero declaraciones")
	}
}

// Los datos historicos declarados antes de existir la politica temporal deben
// seguir cargando: la reconstitucion no aplica plazos, solo invariantes.
func TestI11PostgresRehidrataDatosHistoricosFueraDePlazo(t *testing.T) {
	db, ctx := prepararPostgresSesionAislado(t)
	inicio := time.Date(2026, 10, 20, 15, 0, 0, 0, time.UTC)
	s, err := NuevaSesion("historica-i11", "reserva-historica-i11", "tutor", "escolar", inicio, inicio.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	repo := RepositorioPostgres{DB: db}
	if err := repo.CrearSesion(ctx, s); err != nil {
		t.Fatal(err)
	}

	antes := inicio.Add(-30 * 24 * time.Hour)
	despues := s.Fin().Add(30 * 24 * time.Hour)
	inserciones := []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO sesion_asistencias (sesion_id, usuario_id, rol, declarada_en) VALUES ($1, $2, $3, $4)`,
			[]any{s.ID(), s.TutorID(), string(RolTutor), antes}},
		{`INSERT INTO sesion_ausencias (sesion_id, usuario_id, rol, reportada_en) VALUES ($1, $2, $3, $4)`,
			[]any{s.ID(), s.EscolarID(), string(RolEscolar), despues}},
		{`INSERT INTO sesion_resultados (sesion_id, usuario_id, resumen, registrado_en) VALUES ($1, $2, $3, $4)`,
			[]any{s.ID(), s.TutorID(), "Resumen historico", despues}},
	}
	for _, insercion := range inserciones {
		if _, err := db.ExecContext(ctx, insercion.sql, insercion.args...); err != nil {
			t.Fatal(err)
		}
	}

	guardada, err := repo.SesionPorID(ctx, s.ID())
	if err != nil {
		t.Fatalf("la rehidratacion no debe aplicar plazos: %v", err)
	}
	if len(guardada.Asistencias()) != 1 || len(guardada.Ausencias()) != 1 || guardada.Resultado() == nil {
		t.Fatal("datos historicos incompletos")
	}
}
