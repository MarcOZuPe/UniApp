package reserva_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"
	"uniapp/migrations"
)

// Esquema exclusivo por prueba: sin CREATE DATABASE ni escrituras en tablas públicas.
func baseOutboxAislada(t *testing.T) *sql.DB {
	t.Helper()
	conexion := os.Getenv("I10_TEST_DATABASE_URL")
	if conexion == "" {
		t.Skip("I10_TEST_DATABASE_URL no definida")
	}
	u, err := url.Parse(conexion)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		t.Fatal("I10_TEST_DATABASE_URL debe ser URL PostgreSQL")
	}
	admin, err := sql.Open("postgres", conexion)
	if err != nil {
		t.Fatal("conexión de prueba inválida")
	}
	t.Cleanup(func() { _ = admin.Close() })
	// No instalar extensiones ni cambiar la base compartida durante estas pruebas.
	var extensiones int
	if err := admin.QueryRow("SELECT count(*) FROM pg_extension WHERE extname IN ('pgcrypto','btree_gist')").Scan(&extensiones); err != nil {
		t.Fatal(err)
	}
	if extensiones != 2 {
		t.Skip("se requieren pgcrypto y btree_gist previamente instaladas en la base de pruebas")
	}
	nombre := fmt.Sprintf("i10_test_%d", time.Now().UnixNano())
	if _, err := admin.Exec("CREATE SCHEMA " + nombre); err != nil {
		t.Fatal("no se pudo crear esquema aislado: ", err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec("DROP SCHEMA " + nombre + " CASCADE"); err != nil {
			t.Error("no se pudo limpiar el esquema de prueba: ", err)
		}
	})
	q := u.Query()
	q.Set("search_path", nombre+",public")
	u.RawQuery = q.Encode()
	db, err := migrations.Abrir(u.String())
	if err != nil {
		t.Fatal("no se pudo conectar al esquema aislado")
	}
	t.Cleanup(func() { _ = db.Close() })
	// Evita reutilizar schema_migrations de public a través del search_path.
	if _, err := db.Exec("CREATE TABLE " + nombre + ".schema_migrations(nombre TEXT PRIMARY KEY, aplicada_en TIMESTAMPTZ NOT NULL DEFAULT now())"); err != nil {
		t.Fatal(err)
	}
	if err := migrations.Migrar(context.Background(), db, migrations.FS); err != nil {
		t.Fatal(err)
	}
	return db
}
