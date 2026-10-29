package sesion

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type RepositorioPostgres struct {
	DB *sql.DB
}

func (r RepositorioPostgres) CrearSesion(
	ctx context.Context,
	s *Sesion,
) error {
	if err := s.Validar(); err != nil {
		return err
	}

	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sesion_sesiones (
			id, reserva_id, tutor_id, escolar_id, inicio, fin
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		s.ID(),
		s.ReservaID(),
		s.TutorID(),
		s.EscolarID(),
		s.Inicio(),
		s.Fin(),
	)

	return traducirErrorDuplicado(err)
}

func (r RepositorioPostgres) SesionPorID(
	ctx context.Context,
	id string,
) (*Sesion, error) {
	tx, err := r.DB.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	s, err := cargarSesion(ctx, tx, id, false)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s, nil
}

func (r RepositorioPostgres) SesionPorReservaID(
	ctx context.Context,
	reservaID string,
) (*Sesion, error) {
	var sesionID string

	err := r.DB.QueryRowContext(ctx, `
        SELECT id
        FROM sesion_sesiones
        WHERE reserva_id = $1
    `, reservaID).Scan(&sesionID)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSesionNoEncontrada
	}

	if err != nil {
		return nil, err
	}

	return r.SesionPorID(ctx, sesionID)
}

func (r RepositorioPostgres) DeclararAsistencia(
	ctx context.Context,
	sesionID string,
	usuarioID string,
	ahora time.Time,
) (Asistencia, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return Asistencia{}, err
	}
	defer tx.Rollback()

	// Bloquea la sesión mientras se valida y registra.
	s, err := cargarSesion(ctx, tx, sesionID, true)
	if err != nil {
		return Asistencia{}, err
	}

	if err := s.DeclararAsistencia(usuarioID, ahora); err != nil {
		return Asistencia{}, err
	}

	asistencias := s.Asistencias()
	nueva := asistencias[len(asistencias)-1]

	_, err = tx.ExecContext(ctx, `
		INSERT INTO sesion_asistencias (
			sesion_id, usuario_id, rol, declarada_en
		)
		VALUES ($1, $2, $3, $4)
	`,
		sesionID,
		nueva.UsuarioID(),
		string(nueva.Rol()),
		nueva.DeclaradaEn(),
	)
	if err != nil {
		return Asistencia{}, traducirErrorAsistenciaDuplicada(err)
	}

	if err := tx.Commit(); err != nil {
		return Asistencia{}, err
	}

	return nueva, nil
}

func cargarSesion(
	ctx context.Context,
	tx *sql.Tx,
	id string,
	bloquear bool,
) (*Sesion, error) {
	consulta := `
		SELECT
			id, reserva_id, tutor_id, escolar_id, inicio, fin
		FROM sesion_sesiones
		WHERE id = $1
	`

	if bloquear {
		consulta += " FOR UPDATE"
	}

	var sesionID string
	var reservaID string
	var tutorID string
	var escolarID string
	var inicio time.Time
	var fin time.Time

	err := tx.QueryRowContext(ctx, consulta, id).Scan(
		&sesionID,
		&reservaID,
		&tutorID,
		&escolarID,
		&inicio,
		&fin,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSesionNoEncontrada
	}
	if err != nil {
		return nil, err
	}

	s, err := NuevaSesion(
		sesionID,
		reservaID,
		tutorID,
		escolarID,
		inicio,
		fin,
	)
	if err != nil {
		return nil, fmt.Errorf("sesion almacenada invalida: %w", err)
	}

	if err := cargarAsistencias(ctx, tx, s); err != nil {
		return nil, err
	}

	return s, nil
}

func cargarAsistencias(
	ctx context.Context,
	tx *sql.Tx,
	s *Sesion,
) error {
	filas, err := tx.QueryContext(ctx, `
		SELECT usuario_id, rol, declarada_en
		FROM sesion_asistencias
		WHERE sesion_id = $1
		ORDER BY declarada_en, usuario_id
	`, s.ID())
	if err != nil {
		return err
	}
	defer filas.Close()

	for filas.Next() {
		var usuarioID string
		var rol string
		var declaradaEn time.Time

		if err := filas.Scan(
			&usuarioID,
			&rol,
			&declaradaEn,
		); err != nil {
			return err
		}

		// Reconstruye la asistencia aplicando las
		// reglas de autorización del modelo.
		if err := s.DeclararAsistencia(usuarioID, declaradaEn); err != nil {
			return fmt.Errorf("asistencia almacenada invalida: %w", err)
		}

		asistencias := s.Asistencias()
		ultima := asistencias[len(asistencias)-1]

		if string(ultima.Rol()) != rol {
			return fmt.Errorf("rol de asistencia inconsistente")
		}
	}

	return filas.Err()
}

func traducirErrorDuplicado(err error) error {
	var errorPostgres *pq.Error

	if errors.As(err, &errorPostgres) &&
		errorPostgres.Code == "23505" {
		return ErrSesionDuplicada
	}

	return err
}

func traducirErrorAsistenciaDuplicada(err error) error {
	var errorPostgres *pq.Error

	if errors.As(err, &errorPostgres) &&
		errorPostgres.Code == "23505" {
		return ErrAsistenciaDuplicada
	}

	return err
}
