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
)

func TestI05PostgresPersistenciaAislada(testingT *testing.T) {
	cadenaConexion := os.Getenv("I05_TEST_DATABASE_URL")
	if cadenaConexion == "" {
		testingT.Skip("I05_TEST_DATABASE_URL no definida; requiere PostgreSQL disponible")
	}

	if strings.HasPrefix(cadenaConexion, "postgres://") || strings.HasPrefix(cadenaConexion, "postgresql://") {
		urlPostgres, err := url.Parse(cadenaConexion)
		if err != nil {
			testingT.Fatal("URL de PostgreSQL inválida")
		}
		parametros := urlPostgres.Query()
		parametros.Set("search_path", "pg_temp")
		urlPostgres.RawQuery = parametros.Encode()
		cadenaConexion = urlPostgres.String()
	} else {
		cadenaConexion += " search_path=pg_temp"
	}

	baseDatos, err := sql.Open("postgres", cadenaConexion)
	if err != nil {
		testingT.Fatal("no se pudo abrir PostgreSQL")
	}
	defer baseDatos.Close()
	baseDatos.SetMaxOpenConns(1)
	baseDatos.SetMaxIdleConns(1)

	contexto, cancelar := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelar()
	if err := baseDatos.PingContext(contexto); err != nil {
		testingT.Fatal("PostgreSQL no disponible")
	}
	if _, err := baseDatos.ExecContext(contexto, "SET search_path TO pg_temp"); err != nil {
		testingT.Fatal("no se pudo aislar el esquema")
	}

	migracionSQL, err := os.ReadFile("../../../migrations/005_sesion.up.sql")
	if err != nil {
		testingT.Fatal(err)
	}
	sqlTemporal := strings.ReplaceAll(string(migracionSQL), "CREATE TABLE", "CREATE TEMP TABLE")
	if _, err := baseDatos.ExecContext(contexto, sqlTemporal); err != nil {
		testingT.Fatal("no se pudieron crear tablas temporales")
	}

	inicio := time.Date(2026, 10, 20, 15, 0, 0, 0, time.UTC)
	sesion, err := NuevaSesion(
		"sesion-test-aislada", "reserva-test-aislada",
		"tutor-test", "escolar-test", inicio, inicio.Add(time.Hour),
	)
	if err != nil {
		testingT.Fatal(err)
	}

	repositorio := RepositorioPostgres{DB: baseDatos}
	if err := repositorio.CrearSesion(contexto, sesion); err != nil {
		testingT.Fatal("falló crear sesión temporal")
	}
	for _, usuario := range []string{"tutor-test", "escolar-test"} {
		if _, err := repositorio.DeclararAsistencia(contexto, sesion.ID(), usuario, inicio); err != nil {
			testingT.Fatal("falló persistir asistencia temporal")
		}
		if _, err := repositorio.DeclararAsistencia(contexto, sesion.ID(), usuario, inicio); !errors.Is(err, ErrAsistenciaDuplicada) {
			testingT.Fatal("no rechazó duplicado")
		}
	}
	if _, err := repositorio.DeclararAsistencia(contexto, sesion.ID(), "ajeno", inicio); !errors.Is(err, ErrNoAutorizado) {
		testingT.Fatal("no rechazó usuario ajeno")
	}

	otroRepositorio := RepositorioPostgres{DB: baseDatos}
	guardada, err := otroRepositorio.SesionPorID(contexto, sesion.ID())
	if err != nil {
		testingT.Fatal("falló recargar la sesión temporal")
	}
	if len(guardada.Asistencias()) != 2 {
		testingT.Fatal("asistencias no persistidas")
	}
	for indice, rolEsperado := range []Rol{RolEscolar, RolTutor} {
		if guardada.Asistencias()[indice].Rol() != rolEsperado {
			testingT.Fatal("rol persistido incorrecto")
		}
	}
}
