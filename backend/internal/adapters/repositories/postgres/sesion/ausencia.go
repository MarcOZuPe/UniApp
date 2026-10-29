package sesion

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"time"
)

func (r RepositorioPostgres) ReportarAusencia(ctx context.Context, sesionID string, usuarioID string, ahora time.Time) (Ausencia, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return Ausencia{}, err
	}
	defer tx.Rollback()
	// El mismo bloqueo que usa asistencia serializa ambas declaraciones.
	s, err := cargarSesion(ctx, tx, sesionID, true)
	if err != nil {
		return Ausencia{}, err
	}
	if err := s.ReportarAusencia(usuarioID, ahora); err != nil {
		return Ausencia{}, err
	}
	ausencias := s.Ausencias()
	nueva := ausencias[len(ausencias)-1]
	_, err = tx.ExecContext(ctx, `INSERT INTO sesion_ausencias
        (sesion_id, usuario_id, rol, reportada_en) VALUES ($1, $2, $3, $4)`,
		sesionID, nueva.UsuarioID(), string(nueva.Rol()), nueva.ReportadaEn())
	if err != nil {
		var pg *pq.Error
		if errors.As(err, &pg) && pg.Code == "23505" {
			return Ausencia{}, ErrAusenciaDuplicada
		}
		return Ausencia{}, err
	}
	if err := tx.Commit(); err != nil {
		return Ausencia{}, err
	}
	return nueva, nil
}

func cargarAusencias(ctx context.Context, tx *sql.Tx, s *Sesion) error {
	filas, err := tx.QueryContext(ctx, `SELECT usuario_id, rol, reportada_en
        FROM sesion_ausencias WHERE sesion_id = $1 ORDER BY reportada_en, usuario_id`, s.ID())
	if err != nil {
		return err
	}
	defer filas.Close()
	for filas.Next() {
		var usuarioID, rol string
		var reportadaEn time.Time
		if err := filas.Scan(&usuarioID, &rol, &reportadaEn); err != nil {
			return err
		}
		if err := s.ReportarAusencia(usuarioID, reportadaEn); err != nil {
			return fmt.Errorf("ausencia almacenada invalida: %w", err)
		}
		ausencias := s.Ausencias()
		if string(ausencias[len(ausencias)-1].Rol()) != rol {
			return fmt.Errorf("rol de ausencia inconsistente")
		}
	}
	return filas.Err()
}
