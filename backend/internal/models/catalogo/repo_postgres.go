package catalogo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type RepositorioPostgres struct {
	DB *sql.DB
}

func (r RepositorioPostgres) Importar(ctx context.Context, importacion *Importacion) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := tx.QueryRowContext(ctx, `INSERT INTO catalogo_importaciones (fuente, importado_en)
		VALUES ($1, $2) RETURNING id`, importacion.Fuente, importacion.ImportadoEn).Scan(&importacion.ID); err != nil {
		return err
	}
	for indice := range importacion.Cursos {
		curso := &importacion.Cursos[indice]
		if err := tx.QueryRowContext(ctx, `INSERT INTO catalogo_cursos
			(importacion_id, nombre, grado, estado)
			VALUES ($1, $2, $3, $4) RETURNING id`,
			importacion.ID, curso.Nombre, curso.Grado, string(EstadoPendiente)).Scan(&curso.ID); err != nil {
			return err
		}
		curso.Estado = EstadoPendiente
	}
	for _, colegio := range importacion.Colegios {
		if _, err := tx.ExecContext(ctx, `INSERT INTO catalogo_colegios
			(importacion_id, nombre, distrito) VALUES ($1, $2, $3)`,
			importacion.ID, colegio.Nombre, colegio.Distrito); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE catalogo_solicitudes_colegio solicitud
		SET estado = 'incorporada'
		WHERE solicitud.estado = 'pendiente' AND EXISTS (
			SELECT 1 FROM catalogo_colegios colegio
			WHERE colegio.importacion_id = $1
				AND lower(trim(colegio.nombre)) = lower(trim(solicitud.nombre))
				AND lower(trim(colegio.distrito)) = lower(trim(solicitud.distrito))
		)`, importacion.ID); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("guardar importación de catálogo: %w", err)
	}
	return nil
}

func (r RepositorioPostgres) ImportacionActual(ctx context.Context) (*Importacion, error) {
	var importacion Importacion
	err := r.DB.QueryRowContext(ctx, `SELECT id, fuente, importado_en
		FROM catalogo_importaciones ORDER BY importado_en DESC, id DESC LIMIT 1`).
		Scan(&importacion.ID, &importacion.Fuente, &importacion.ImportadoEn)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	if err != nil {
		return nil, err
	}
	importacion.Cursos, err = r.cursosDe(ctx, importacion.ID)
	if err != nil {
		return nil, err
	}
	importacion.Colegios, err = r.colegiosDe(ctx, importacion.ID)
	if err != nil {
		return nil, err
	}
	return &importacion, nil
}

func (r RepositorioPostgres) Curso(ctx context.Context, id int64) (Curso, error) {
	curso, err := escanearCurso(r.DB.QueryRowContext(ctx, `SELECT c.id, c.nombre, c.grado, c.estado,
		c.motivo_rechazo, c.revisado_en, COALESCE(c.revisado_por::text, '')
		FROM catalogo_cursos c
		WHERE c.id = $1 AND c.importacion_id = (
			SELECT id FROM catalogo_importaciones ORDER BY importado_en DESC, id DESC LIMIT 1
		)`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Curso{}, ErrNoEncontrado
	}
	return curso, err
}

func (r RepositorioPostgres) RevisarCurso(ctx context.Context, curso Curso, revisorID string) error {
	resultado, err := r.DB.ExecContext(ctx, `UPDATE catalogo_cursos
		SET estado = $2, motivo_rechazo = $3, revisado_por = $4, revisado_en = $5
		WHERE id = $1 AND estado = $6 AND importacion_id = (
			SELECT id FROM catalogo_importaciones ORDER BY importado_en DESC, id DESC LIMIT 1
		)`, curso.ID, string(curso.Estado), curso.Motivo, revisorID, curso.RevisadoEn, string(EstadoPendiente))
	if err != nil {
		return err
	}
	afectados, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if afectados == 0 {
		return ErrEstadoInvalido
	}
	return nil
}

func (r RepositorioPostgres) SolicitarColegio(ctx context.Context, solicitud SolicitudColegio, requesterID string) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO catalogo_solicitudes_colegio
		(nombre, distrito, solicitante_id, solicitado_en)
		VALUES ($1, $2, $3, $4)`, solicitud.Nombre, solicitud.Distrito, requesterID, solicitud.SolicitadoEn)
	if esViolacionUnica(err) {
		return ErrSolicitudDuplicada
	}
	return err
}

func (r RepositorioPostgres) SolicitudesPendientes(ctx context.Context) ([]SolicitudColegio, error) {
	filas, err := r.DB.QueryContext(ctx, `SELECT id, nombre, distrito, solicitado_en
		FROM catalogo_solicitudes_colegio WHERE estado = 'pendiente'
		ORDER BY solicitado_en, id`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = filas.Close() }()
	resultado := make([]SolicitudColegio, 0)
	for filas.Next() {
		var solicitud SolicitudColegio
		if err := filas.Scan(&solicitud.ID, &solicitud.Nombre, &solicitud.Distrito, &solicitud.SolicitadoEn); err != nil {
			return nil, err
		}
		solicitud.Estado = EstadoSolicitudPendiente
		resultado = append(resultado, solicitud)
	}
	return resultado, filas.Err()
}

func (r RepositorioPostgres) cursosDe(ctx context.Context, importacionID int64) ([]Curso, error) {
	filas, err := r.DB.QueryContext(ctx, `SELECT id, nombre, grado, estado, motivo_rechazo,
		revisado_en, COALESCE(revisado_por::text, '')
		FROM catalogo_cursos WHERE importacion_id = $1 ORDER BY grado, nombre`, importacionID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = filas.Close() }()
	resultado := make([]Curso, 0)
	for filas.Next() {
		curso, err := escanearCurso(filas)
		if err != nil {
			return nil, err
		}
		resultado = append(resultado, curso)
	}
	return resultado, filas.Err()
}

func (r RepositorioPostgres) colegiosDe(ctx context.Context, importacionID int64) ([]Colegio, error) {
	filas, err := r.DB.QueryContext(ctx, `SELECT nombre, distrito
		FROM catalogo_colegios WHERE importacion_id = $1 ORDER BY distrito, nombre`, importacionID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = filas.Close() }()
	resultado := make([]Colegio, 0)
	for filas.Next() {
		var colegio Colegio
		if err := filas.Scan(&colegio.Nombre, &colegio.Distrito); err != nil {
			return nil, err
		}
		resultado = append(resultado, colegio)
	}
	return resultado, filas.Err()
}

type escanerCurso interface {
	Scan(...any) error
}

func escanearCurso(fila escanerCurso) (Curso, error) {
	var curso Curso
	var estado string
	err := fila.Scan(&curso.ID, &curso.Nombre, &curso.Grado, &estado, &curso.Motivo, &curso.RevisadoEn, &curso.RevisadoPor)
	curso.Estado = EstadoCurso(estado)
	return curso, err
}

func esViolacionUnica(err error) bool {
	var numeroError interface{ SQLState() string }
	return errors.As(err, &numeroError) && numeroError.SQLState() == "23505"
}
