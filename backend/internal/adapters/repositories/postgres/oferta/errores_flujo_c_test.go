package oferta_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	repo "uniapp/internal/adapters/repositories/postgres/oferta"
	"uniapp/migrations"
)

// baseParaErrores crea una base propia de este archivo para no depender de otros helpers.
func baseParaErrores(t *testing.T) *sql.DB {
	t.Helper()
	base := os.Getenv("TEST_DATABASE_URL")
	if base == "" {
		t.Skip("TEST_DATABASE_URL no definida")
	}
	administrador, err := sql.Open("postgres", base)
	if err != nil {
		t.Fatal(err)
	}
	nombre := fmt.Sprintf("uniapp_oferta_c_%d", time.Now().UnixNano())
	if _, err := administrador.Exec("CREATE DATABASE " + nombre); err != nil {
		t.Fatalf("crear base de prueba: %v", err)
	}
	t.Cleanup(func() {
		_, _ = administrador.Exec("DROP DATABASE IF EXISTS " + nombre)
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

func usuarioParaErrores(t *testing.T, baseDatos *sql.DB, email string, rol int) string {
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

// F-31: eliminar un horario publicado con una reserva asociada deja la reserva
// apuntando a un slot inexistente y permite que el mismo horario se vuelva a publicar.
func TestF31NoDebeEliminarSlotConReserva(t *testing.T) {
	baseDatos := baseParaErrores(t)
	repositorio := repo.RepositorioPostgres{DB: baseDatos}
	ctx := context.Background()

	tutor := usuarioParaErrores(t, baseDatos, "tutor.f31@example.test", 3)
	escolar := usuarioParaErrores(t, baseDatos, "escolar.f31@example.test", 1)

	var slotID int64
	if err := baseDatos.QueryRow(
		`INSERT INTO slots (tutor_id, dia_semana, inicio_minuto, fin_minuto)
		 VALUES ($1, 1, 540, 600) RETURNING id`, tutor).Scan(&slotID); err != nil {
		t.Fatalf("crear slot: %v", err)
	}

	inicio := time.Now().UTC().Add(72 * time.Hour)
	if _, err := baseDatos.Exec(
		`INSERT INTO reservas (escolar_id, tutor_id, tutor_nombre, slot_id, curso, grado,
		                       inicio, fin, tema, objetivo, estado)
		 VALUES ($1, $2, 'Tutor de prueba', $3, 'Álgebra', '2° de secundaria',
		         $4, $5, 'Ecuaciones lineales', 'Resolver diez ejercicios y verificar', 'confirmada')`,
		escolar, tutor, fmt.Sprintf("%d", slotID), inicio, inicio.Add(time.Hour)); err != nil {
		t.Fatalf("crear reserva: %v", err)
	}

	if err := repositorio.EliminarSlot(ctx, tutor, slotID); err == nil {
		t.Fatalf("F-31: se eliminó el slot %d con una reserva asociada; la reserva queda huérfana", slotID)
	}
}
