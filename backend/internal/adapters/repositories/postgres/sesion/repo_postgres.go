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

type consultaSesion struct {
	sesionID string
	bloquear bool
}

func (repositorio RepositorioPostgres) CrearSesion(
	ctx context.Context,
	sesion *Sesion,
) error {
	if err := sesion.Validar(); err != nil {
		return err
	}

	_, err := repositorio.DB.ExecContext(ctx, `
		INSERT INTO sesion_sesiones (
			id, reserva_id, tutor_id, escolar_id, inicio, fin
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		sesion.ID(),
		sesion.ReservaID(),
		sesion.TutorID(),
		sesion.EscolarID(),
		sesion.Inicio(),
		sesion.Fin(),
	)
	return traducirErrorDuplicado(err)
}

func (repositorio RepositorioPostgres) SesionPorID(
	ctx context.Context,
	sesionID string,
) (*Sesion, error) {
	transaccion, err := repositorio.DB.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		return nil, err
	}
	defer transaccion.Rollback()

	sesion, err := cargarSesion(ctx, transaccion, consultaSesion{sesionID: sesionID})
	if err != nil {
		return nil, err
	}
	if err := transaccion.Commit(); err != nil {
		return nil, err
	}
	return sesion, nil
}

func (repositorio RepositorioPostgres) DeclararAsistencia(
	ctx context.Context,
	sesionID string,
	usuarioID string,
	momento time.Time,
) (Asistencia, error) {
	transaccion, err := repositorio.DB.BeginTx(ctx, nil)
	if err != nil {
		return Asistencia{}, err
	}
	defer transaccion.Rollback()

	sesion, err := cargarSesion(ctx, transaccion, consultaSesion{sesionID: sesionID, bloquear: true})
	if err != nil {
		return Asistencia{}, err
	}
	if err := sesion.DeclararAsistencia(usuarioID, momento); err != nil {
		return Asistencia{}, err
	}

	asistencias := sesion.Asistencias()
	asistencia := asistencias[len(asistencias)-1]

	_, err = transaccion.ExecContext(ctx, `
		INSERT INTO sesion_asistencias (
			sesion_id, usuario_id, rol, declarada_en
		)
		VALUES ($1, $2, $3, $4)
	`,
		sesionID,
		asistencia.UsuarioID(),
		string(asistencia.Rol()),
		asistencia.DeclaradaEn(),
	)
	if err != nil {
		return Asistencia{}, traducirErrorAsistenciaDuplicada(err)
	}
	if err := transaccion.Commit(); err != nil {
		return Asistencia{}, err
	}
	return asistencia, nil
}

func cargarSesion(
	ctx context.Context,
	transaccion *sql.Tx,
	consulta consultaSesion,
) (*Sesion, error) {
	consultaSQL := `
		SELECT
			id, reserva_id, tutor_id, escolar_id, inicio, fin
		FROM sesion_sesiones
		WHERE id = $1
	`
	if consulta.bloquear {
		consultaSQL += " FOR UPDATE"
	}

	var sesionID string
	var reservaID string
	var tutorID string
	var escolarID string
	var inicio time.Time
	var fin time.Time

	err := transaccion.QueryRowContext(ctx, consultaSQL, consulta.sesionID).Scan(
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

	sesion, err := NuevaSesion(
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
	if err := cargarAsistencias(ctx, transaccion, sesion); err != nil {
		return nil, err
	}
	return sesion, nil
}

func cargarAsistencias(
	ctx context.Context,
	transaccion *sql.Tx,
	sesion *Sesion,
) error {
	filas, err := transaccion.QueryContext(ctx, `
		SELECT usuario_id, rol, declarada_en
		FROM sesion_asistencias
		WHERE sesion_id = $1
		ORDER BY declarada_en, usuario_id
	`, sesion.ID())
	if err != nil {
		return err
	}
	defer filas.Close()

	for filas.Next() {
		var usuarioID string
		var rol string
		var declaradaEn time.Time

		if err := filas.Scan(&usuarioID, &rol, &declaradaEn); err != nil {
			return err
		}

		if err := sesion.DeclararAsistencia(usuarioID, declaradaEn); err != nil {
			return fmt.Errorf("asistencia almacenada invalida: %w", err)
		}

		asistencias := sesion.Asistencias()
		ultima := asistencias[len(asistencias)-1]
		if string(ultima.Rol()) != rol {
			return fmt.Errorf("rol de asistencia inconsistente")
		}
	}
	return filas.Err()
}

func traducirErrorDuplicado(err error) error {
	var errorPostgres *pq.Error
	if errors.As(err, &errorPostgres) && errorPostgres.Code == "23505" {
		return ErrSesionDuplicada
	}
	return err
}

func traducirErrorAsistenciaDuplicada(err error) error {
	var errorPostgres *pq.Error
	if errors.As(err, &errorPostgres) && errorPostgres.Code == "23505" {
		return ErrAsistenciaDuplicada
	}
	return err
}
