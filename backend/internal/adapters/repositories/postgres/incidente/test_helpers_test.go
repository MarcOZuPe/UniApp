package incidente

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"testing"
	"time"
)

// baseDeDatosConError simula un fallo de PostgreSQL sin abrir una conexión real.
type baseDeDatosConError struct{ err error }

func (d baseDeDatosConError) Open(string) (driver.Conn, error) { return conexionConError(d), nil }

type conexionConError struct{ err error }

func (c conexionConError) Prepare(string) (driver.Stmt, error) { return nil, c.err }
func (c conexionConError) Close() error                        { return nil }
func (c conexionConError) Begin() (driver.Tx, error)           { return nil, c.err }

func (c conexionConError) QueryContext(
	context.Context,
	string,
	[]driver.NamedValue,
) (driver.Rows, error) {
	return nil, c.err
}

func abrirDBConError(t *testing.T, err error) *sql.DB {
	t.Helper()
	nombre := fmt.Sprintf("incidente-error-%d", time.Now().UnixNano())
	sql.Register(nombre, baseDeDatosConError{err: err})
	baseDatos, errAbrir := sql.Open(nombre, "")
	if errAbrir != nil {
		t.Fatal(errAbrir)
	}
	t.Cleanup(func() { _ = baseDatos.Close() })
	return baseDatos
}
