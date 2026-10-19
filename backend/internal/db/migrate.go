package db

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"

	_ "github.com/lib/pq"
)

const maxConexionesAbiertas = 10

func Abrir(urlConexion string) (*sql.DB, error) {
	baseDatos, err := sql.Open("postgres", urlConexion)
	if err != nil {
		return nil, err
	}
	baseDatos.SetMaxOpenConns(maxConexionesAbiertas)
	return baseDatos, baseDatos.Ping()
}

func Migrar(ctx context.Context, baseDatos *sql.DB, archivos fs.FS) error {
	if err := crearTablaMigraciones(ctx, baseDatos); err != nil {
		return err
	}
	nombres, err := fs.Glob(archivos, "*.up.sql")
	if err != nil {
		return err
	}
	sort.Strings(nombres)
	for _, nombre := range nombres {
		if err := aplicarSiFalta(ctx, baseDatos, archivos, nombre); err != nil {
			return err
		}
	}
	return nil
}

func crearTablaMigraciones(ctx context.Context, baseDatos *sql.DB) error {
	_, err := baseDatos.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		nombre TEXT PRIMARY KEY, aplicada_en TIMESTAMPTZ NOT NULL DEFAULT now())`)
	return err
}

func aplicarSiFalta(ctx context.Context, baseDatos *sql.DB, archivos fs.FS, nombre string) error {
	aplicada, err := migracionAplicada(ctx, baseDatos, nombre)
	if err != nil || aplicada {
		return err
	}
	contenido, err := fs.ReadFile(archivos, nombre)
	if err != nil {
		return err
	}
	return ejecutarEnTransaccion(ctx, baseDatos, nombre, string(contenido))
}

func migracionAplicada(ctx context.Context, baseDatos *sql.DB, nombre string) (bool, error) {
	var existe bool
	err := baseDatos.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE nombre=$1)`, nombre).Scan(&existe)
	return existe, err
}

func ejecutarEnTransaccion(ctx context.Context, baseDatos *sql.DB, nombre, contenido string) error {
	transaccion, err := baseDatos.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := transaccion.ExecContext(ctx, contenido); err != nil {
		_ = transaccion.Rollback()
		return fmt.Errorf("migración %s: %w", nombre, err)
	}
	if _, err := transaccion.ExecContext(ctx, `INSERT INTO schema_migrations(nombre) VALUES($1)`, nombre); err != nil {
		_ = transaccion.Rollback()
		return err
	}
	return transaccion.Commit()
}
