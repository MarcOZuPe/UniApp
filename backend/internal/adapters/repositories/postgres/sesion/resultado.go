package sesion

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"time"
)

func (r RepositorioPostgres) RegistrarResultado(ctx context.Context, sesionID, usuarioID, resumen string, ahora time.Time) (Resultado, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return Resultado{}, err
	}
	defer tx.Rollback()
	s, err := cargarSesion(ctx, tx, sesionID, true)
	if err != nil {
		return Resultado{}, err
	}
	if err := s.RegistrarResultado(usuarioID, resumen, ahora); err != nil {
		return Resultado{}, err
	}
	nuevo := *s.Resultado()
	_, err = tx.ExecContext(ctx, `INSERT INTO sesion_resultados
        (sesion_id, usuario_id, resumen, registrado_en) VALUES ($1, $2, $3, $4)`,
		sesionID, nuevo.UsuarioID(), nuevo.Resumen(), nuevo.RegistradoEn())
	if err != nil {
		var pg *pq.Error
		if errors.As(err, &pg) && pg.Code == "23505" {
			return Resultado{}, ErrResultadoDuplicado
		}
		return Resultado{}, err
	}
	if err := tx.Commit(); err != nil {
		return Resultado{}, err
	}
	return nuevo, nil
}

func cargarResultado(ctx context.Context, tx *sql.Tx, s *Sesion) error {
	var usuarioID, resumen string
	var registradoEn time.Time
	err := tx.QueryRowContext(ctx, `SELECT usuario_id, resumen, registrado_en
        FROM sesion_resultados WHERE sesion_id = $1`, s.ID()).Scan(&usuarioID, &resumen, &registradoEn)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := s.RegistrarResultado(usuarioID, resumen, registradoEn); err != nil {
		return fmt.Errorf("resultado almacenado invalido: %w", err)
	}
	return nil
}
