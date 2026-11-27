package acceso

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"
	"time"
)

// baseDeDatosSinFilas simula un UPDATE que no afecta ninguna fila, como cuando
// el usuario ya no existe en PostgreSQL.
type baseDeDatosSinFilas struct{}

func (baseDeDatosSinFilas) Open(string) (driver.Conn, error) { return conexionSinFilas{}, nil }

type conexionSinFilas struct{}

func (conexionSinFilas) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare no soportado")
}
func (conexionSinFilas) Close() error              { return nil }
func (conexionSinFilas) Begin() (driver.Tx, error) { return nil, errors.New("begin no soportado") }

func (conexionSinFilas) ExecContext(
	context.Context,
	string,
	[]driver.NamedValue,
) (driver.Result, error) {
	return driver.RowsAffected(0), nil
}

func abrirDBSinFilas(t *testing.T) *sql.DB {
	t.Helper()
	nombre := fmt.Sprintf("acceso-sin-filas-%d", time.Now().UnixNano())
	sql.Register(nombre, baseDeDatosSinFilas{})
	baseDatos, err := sql.Open(nombre, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = baseDatos.Close() })
	return baseDatos
}

// B-14: confirmar un correo que ya no existe actualiza cero filas y debe
// reportarse como usuario no encontrado.
func TestB14MarcarCorreoConfirmadoDebeDetectarCeroFilas(t *testing.T) {
	repositorio := RepositorioPostgres{DB: abrirDBSinFilas(t)}

	err := repositorio.MarcarCorreoConfirmado(
		context.Background(), "00000000-0000-0000-0000-000000000000",
	)

	if !errors.Is(err, ErrNoEncontrado) {
		t.Fatalf("B-14: con 0 filas afectadas se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}

// B-14: actualizar la clave de un usuario inexistente también debe detectar
// que el UPDATE no afectó filas.
func TestB14ActualizarClaveDebeDetectarCeroFilas(t *testing.T) {
	repositorio := RepositorioPostgres{DB: abrirDBSinFilas(t)}

	err := repositorio.ActualizarClave(
		context.Background(), "00000000-0000-0000-0000-000000000000", "nuevo-hash",
	)

	if !errors.Is(err, ErrNoEncontrado) {
		t.Fatalf("B-14: con 0 filas afectadas se esperaba ErrNoEncontrado, se obtuvo %v", err)
	}
}
