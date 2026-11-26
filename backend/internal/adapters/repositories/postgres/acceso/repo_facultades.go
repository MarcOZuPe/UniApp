package acceso

import (
	"context"
	"database/sql"
	"errors"
)

func (repositorio RepositorioPostgres) Facultades(ctx context.Context) ([]Facultad, error) {
	filas, err := repositorio.DB.QueryContext(ctx, `SELECT id::text, nombre FROM facultades ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	facultades := make([]Facultad, 0)
	for filas.Next() {
		var facultad Facultad
		if err := filas.Scan(&facultad.ID, &facultad.Nombre); err != nil {
			return nil, err
		}
		facultades = append(facultades, facultad)
	}
	return facultades, filas.Err()
}

func (repositorio RepositorioPostgres) FacultadPorID(ctx context.Context, facultadID string) (*Facultad, error) {
	var facultad Facultad
	err := repositorio.DB.QueryRowContext(ctx,
		`SELECT id::text, nombre FROM facultades WHERE id = NULLIF($1,'')::smallint`, facultadID,
	).Scan(&facultad.ID, &facultad.Nombre)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	if err != nil {
		return nil, err
	}
	return &facultad, nil
}
