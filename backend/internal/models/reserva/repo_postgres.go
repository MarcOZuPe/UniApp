package reserva

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

const columnasReserva = "id, escolar_id, tutor_id, tutor_nombre, slot_id, curso, grado, inicio, fin, tema, objetivo, estado, creada_en"

// RepositorioPostgres usa solo las tablas reservas y reserva_adjuntos (migración 004).
// La unicidad por slot la garantiza el índice reservas_slot_uq, no la aplicación.
type RepositorioPostgres struct{ DB *sql.DB }

func (r RepositorioPostgres) Crear(ctx context.Context, reserva *Reserva) (*Reserva, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var id string
	err = tx.QueryRowContext(ctx, `INSERT INTO reservas
		(escolar_id, tutor_id, tutor_nombre, slot_id, curso, grado, inicio, fin, tema, objetivo, estado, creada_en)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`,
		reserva.EscolarID, reserva.TutorID, reserva.TutorNombre, reserva.SlotID, reserva.Curso, reserva.Grado,
		reserva.Inicio, reserva.Fin, reserva.Tema, reserva.Objetivo, string(reserva.Estado), reserva.CreadaEn,
	).Scan(&id)
	if err != nil {
		return nil, traducirError(err)
	}
	for _, adjunto := range reserva.Adjuntos {
		if _, err := tx.ExecContext(ctx, `INSERT INTO reserva_adjuntos (reserva_id, nombre, tipo, tamano_bytes)
			VALUES ($1, $2, $3, $4)`, id, adjunto.Nombre, adjunto.Tipo, adjunto.TamanoBytes); err != nil {
			return nil, traducirError(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	guardada := *reserva
	guardada.ID = id
	guardada.Adjuntos = copiarAdjuntos(reserva.Adjuntos)
	return &guardada, nil
}

func (r RepositorioPostgres) PorID(ctx context.Context, id string) (*Reserva, error) {
	reserva, err := escanearReserva(r.DB.QueryRowContext(ctx, `SELECT `+columnasReserva+` FROM reservas WHERE id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrReservaNoEncontrada
	}
	if err != nil {
		return nil, err
	}
	adjuntos, err := r.adjuntosDe(ctx, reserva.ID)
	if err != nil {
		return nil, err
	}
	reserva.Adjuntos = adjuntos
	return reserva, nil
}

func (r RepositorioPostgres) PorEscolar(ctx context.Context, escolarID string) ([]Reserva, error) {
	filas, err := r.DB.QueryContext(ctx, `SELECT `+columnasReserva+` FROM reservas WHERE escolar_id = $1 ORDER BY inicio`, escolarID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = filas.Close() }()

	reservas := []Reserva{}
	for filas.Next() {
		reserva, err := escanearReserva(filas)
		if err != nil {
			return nil, err
		}
		adjuntos, err := r.adjuntosDe(ctx, reserva.ID)
		if err != nil {
			return nil, err
		}
		reserva.Adjuntos = adjuntos
		reservas = append(reservas, *reserva)
	}
	return reservas, filas.Err()
}

func (r RepositorioPostgres) SlotTomado(ctx context.Context, slotID string) (bool, error) {
	var tomado bool
	err := r.DB.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM reservas WHERE slot_id = $1)`, slotID).Scan(&tomado)
	return tomado, err
}

type filaEscaner interface{ Scan(destino ...any) error }

func escanearReserva(fila filaEscaner) (*Reserva, error) {
	var reserva Reserva
	var estado string
	if err := fila.Scan(
		&reserva.ID, &reserva.EscolarID, &reserva.TutorID, &reserva.TutorNombre, &reserva.SlotID,
		&reserva.Curso, &reserva.Grado, &reserva.Inicio, &reserva.Fin, &reserva.Tema, &reserva.Objetivo,
		&estado, &reserva.CreadaEn,
	); err != nil {
		return nil, err
	}
	reserva.Estado = Estado(estado)
	return &reserva, nil
}

func (r RepositorioPostgres) adjuntosDe(ctx context.Context, reservaID string) ([]Adjunto, error) {
	filas, err := r.DB.QueryContext(ctx, `SELECT nombre, tipo, tamano_bytes FROM reserva_adjuntos
		WHERE reserva_id = $1 ORDER BY nombre`, reservaID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = filas.Close() }()

	var adjuntos []Adjunto
	for filas.Next() {
		var adjunto Adjunto
		if err := filas.Scan(&adjunto.Nombre, &adjunto.Tipo, &adjunto.TamanoBytes); err != nil {
			return nil, err
		}
		adjuntos = append(adjuntos, adjunto)
	}
	return adjuntos, filas.Err()
}

// traducirError convierte las violaciones de integridad en errores del dominio, para que
// el controlador responda 409 o 403 sin filtrar detalles de Postgres.
func traducirError(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			return ErrSlotOcupado
		case "23503":
			return ErrNoAutorizado
		case "23514":
			return ErrReservaInvalida
		}
	}
	return err
}
