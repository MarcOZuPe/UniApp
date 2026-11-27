package sesion

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

// RepositorioReportesPostgres implementa el puerto de I18.
type RepositorioReportesPostgres struct {
	DB *sql.DB
}

// SesionesEntre lee las sesiones del periodo en una sola transaccion de
// lectura, para que el reporte sea una foto coherente.
func (r RepositorioReportesPostgres) SesionesEntre(ctx context.Context, filtro FiltroReporte) ([]*Sesion, error) {
	transaccion, err := r.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer transaccion.Rollback()

	filas, err := transaccion.QueryContext(ctx, `
		SELECT id
		FROM sesion_sesiones
		WHERE inicio >= $1 AND inicio < $2 AND ($3 = '' OR tutor_id = $3)
		ORDER BY inicio, id
	`, filtro.Desde(), filtro.Hasta(), filtro.TutorID())
	if err != nil {
		return nil, err
	}
	var ids []string
	for filas.Next() {
		var id string
		if err := filas.Scan(&id); err != nil {
			filas.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := filas.Close(); err != nil {
		return nil, err
	}
	if err := filas.Err(); err != nil {
		return nil, err
	}

	sesiones := make([]*Sesion, 0, len(ids))
	for _, id := range ids {
		sesion, err := cargarSesion(ctx, transaccion, consultaSesion{sesionID: id})
		if err != nil {
			return nil, err
		}
		sesiones = append(sesiones, sesion)
	}
	if err := transaccion.Commit(); err != nil {
		return nil, err
	}
	return sesiones, nil
}

func (r RepositorioReportesPostgres) RegistrarReporte(ctx context.Context, registro RegistroReporte) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sesion_reportes (id, generado_por, generado_en, desde, hasta, tutor_id, formato, filas, sha256)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, registro.ID, registro.GeneradoPor, registro.GeneradoEn, registro.Desde, registro.Hasta,
		registro.TutorID, string(registro.Formato), registro.Filas, registro.SHA256)
	var errPostgres *pq.Error
	if errors.As(err, &errPostgres) && (errPostgres.Code == "23505" || errPostgres.Code.Class() == "23") {
		return ErrReporteInvalido
	}
	return err
}

func (r RepositorioReportesPostgres) ReportesRecientes(ctx context.Context, limite int) ([]RegistroReporte, error) {
	if limite <= 0 {
		limite = 50
	}
	filas, err := r.DB.QueryContext(ctx, `
		SELECT id, generado_por, generado_en, desde, hasta, tutor_id, formato, filas, sha256
		FROM sesion_reportes
		ORDER BY generado_en DESC, id
		LIMIT $1
	`, limite)
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	registros := []RegistroReporte{}
	for filas.Next() {
		var registro RegistroReporte
		var formato string
		if err := filas.Scan(&registro.ID, &registro.GeneradoPor, &registro.GeneradoEn, &registro.Desde,
			&registro.Hasta, &registro.TutorID, &formato, &registro.Filas, &registro.SHA256); err != nil {
			return nil, err
		}
		registro.Formato = FormatoReporte(formato)
		registro.GeneradoEn, registro.Desde, registro.Hasta = registro.GeneradoEn.UTC(), registro.Desde.UTC(), registro.Hasta.UTC()
		registros = append(registros, registro)
	}
	return registros, filas.Err()
}
