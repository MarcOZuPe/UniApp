package oferta_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	repo "uniapp/internal/adapters/repositories/postgres/oferta"
	"uniapp/migrations"
)

// baseAislada crea una base propia para este paquete y aplica todas las migraciones.
func baseAislada(t *testing.T) *sql.DB {
	t.Helper()
	base := os.Getenv("TEST_DATABASE_URL")
	if base == "" {
		t.Skip("TEST_DATABASE_URL no definida")
	}
	administrador, err := sql.Open("postgres", base)
	if err != nil {
		t.Fatal(err)
	}
	nombre := fmt.Sprintf("uniapp_oferta_%d", time.Now().UnixNano())
	if _, err := administrador.Exec("CREATE DATABASE " + nombre); err != nil {
		t.Fatalf("crear base de prueba: %v", err)
	}
	t.Cleanup(func() {
		if _, err := administrador.Exec("DROP DATABASE IF EXISTS " + nombre); err != nil {
			t.Logf("no se pudo borrar la base de prueba: %v", err)
		}
		_ = administrador.Close()
	})

	destino, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	destino.Path = "/" + nombre
	baseDatos, err := migrations.Abrir(destino.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = baseDatos.Close() })
	if err := migrations.Migrar(context.Background(), baseDatos, migrations.FS); err != nil {
		t.Fatalf("migraciones: %v", err)
	}
	return baseDatos
}

func crearUsuario(t *testing.T, baseDatos *sql.DB, email string, rol int) string {
	t.Helper()
	var id string
	err := baseDatos.QueryRow(
		`INSERT INTO usuarios (email, nombre_uso, clave_hash, rol_id)
		 VALUES ($1, 'Prueba', 'hash', $2) RETURNING id`, email, rol).Scan(&id)
	if err != nil {
		t.Fatalf("crear usuario %s: %v", email, err)
	}
	return id
}

func reservaFutura(t *testing.T, baseDatos *sql.DB, escolarID, tutorID string) {
	t.Helper()
	inicio := time.Now().UTC().Add(72 * time.Hour)
	_, err := baseDatos.Exec(
		`INSERT INTO reservas (escolar_id, tutor_id, tutor_nombre, slot_id, curso, grado,
		                       inicio, fin, tema, objetivo, estado)
		 VALUES ($1, $2, 'Tutor de prueba', $3, 'Álgebra', '2° de secundaria',
		         $4, $5, 'Ecuaciones lineales', 'Resolver diez ejercicios y verificar', 'confirmada')`,
		escolarID, tutorID, fmt.Sprintf("slot-%d", time.Now().UnixNano()), inicio, inicio.Add(time.Hour))
	if err != nil {
		t.Fatalf("crear reserva: %v", err)
	}
}

// Reducir el cupo sin compromisos debe permitirse; con una reserva futura, no.
func TestCambiarCupoSemanal(t *testing.T) {
	baseDatos := baseAislada(t)
	repositorio := repo.RepositorioPostgres{DB: baseDatos}
	ctx := context.Background()

	tutor := crearUsuario(t, baseDatos, "tutor.oferta@example.test", 3)
	escolar := crearUsuario(t, baseDatos, "escolar.oferta@example.test", 1)

	if err := repositorio.CambiarCupo(ctx, tutor, 360); err != nil {
		t.Fatalf("ampliar el cupo: %v", err)
	}
	if err := repositorio.CambiarCupo(ctx, tutor, 300); err != nil {
		t.Fatalf("reducir sin compromisos debe permitirse: %v", err)
	}

	reservaFutura(t, baseDatos, escolar, tutor)
	if err := repositorio.CambiarCupo(ctx, tutor, 240); !errors.Is(err, repo.ErrReduccionConActivos) {
		t.Fatalf("reducir con compromisos debe rechazarse, se obtuvo: %v", err)
	}

	if err := repositorio.CambiarCupo(ctx, tutor, 360); err != nil {
		t.Fatalf("ampliar con compromisos debe permitirse: %v", err)
	}
}

// La disponibilidad publicada no puede superar el cupo nuevo.
func TestCambiarCupoNoExcedeDisponibilidad(t *testing.T) {
	baseDatos := baseAislada(t)
	repositorio := repo.RepositorioPostgres{DB: baseDatos}
	ctx := context.Background()
	tutor := crearUsuario(t, baseDatos, "tutor.slots@example.test", 3)

	if err := repositorio.CambiarCupo(ctx, tutor, 360); err != nil {
		t.Fatalf("cupo inicial: %v", err)
	}
	for hora := 8; hora < 15; hora++ {
		_, err := baseDatos.Exec(
			`INSERT INTO slots (tutor_id, dia_semana, inicio_minuto, fin_minuto)
			 VALUES ($1, 1, $2, $3)`, tutor, hora*60, (hora+1)*60)
		if err != nil {
			t.Fatalf("insertar slot: %v", err)
		}
	}
	if err := repositorio.CambiarCupo(ctx, tutor, 300); !errors.Is(err, repo.ErrCapacidadSemanal) {
		t.Fatalf("siete slots no caben en 300 minutos, se obtuvo: %v", err)
	}
}
