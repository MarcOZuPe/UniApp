package reserva

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

const columnasReserva = "id, escolar_id, tutor_id, tutor_nombre, slot_id, curso, grado, inicio, fin, tema, objetivo, estado, clave_idempotencia, creada_en"

const columnasAdjunto = "id, nombre, tipo, tamano_bytes, estado, motivo, huella, clave_almacenamiento, creado_en"

// RepositorioPostgres usa solo las tablas reservas y reserva_adjuntos (migraciones 004 y
// 007). La unicidad por slot, el solapamiento del escolar y la idempotencia los garantiza
// la base de datos; la cuota de D07 se cuenta dentro de la misma transacción.
type RepositorioPostgres struct {
	DB                       *sql.DB
	AnticipacionRecordatorio time.Duration
}

func (r RepositorioPostgres) ReservarTransaccional(ctx context.Context, reserva *Reserva, limites Limites) (*Reserva, bool, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = tx.Rollback() }()

	// Serializa las reservas de un mismo escolar: evita que dos peticiones simultáneas
	// pasen la cuota y el solapamiento a la vez.
	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtext($1)::bigint)", reserva.EscolarID); err != nil {
		return nil, false, err
	}

	if reserva.ClaveIdempotencia != "" {
		existente, err := reservaPorClave(ctx, tx, reserva.EscolarID, reserva.ClaveIdempotencia)
		switch {
		case err == nil:
			return existente, false, nil
		case !errors.Is(err, ErrReservaNoEncontrada):
			return nil, false, err
		}
	}

	// El slot se revisa antes que la cuota para dar el mensaje correcto; la carrera real
	// la resuelve el índice único al insertar.
	var tomado bool
	if err := tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM reservas WHERE slot_id = $1)", reserva.SlotID).Scan(&tomado); err != nil {
		return nil, false, err
	}
	if tomado {
		return nil, false, ErrSlotOcupado
	}

	var futuras int
	if err := tx.QueryRowContext(ctx, "SELECT count(*) FROM reservas WHERE escolar_id = $1 AND estado = 'confirmada' AND inicio > $2", reserva.EscolarID, reserva.CreadaEn).Scan(&futuras); err != nil {
		return nil, false, err
	}
	if futuras >= limites.MaxFuturas {
		return nil, false, ErrCuotaEscolar
	}

	inicioSemana, finSemana := SemanaLima(reserva.Inicio)
	var delCurso int
	if err := tx.QueryRowContext(ctx, "SELECT count(*) FROM reservas WHERE escolar_id = $1 AND curso = $2 AND estado = 'confirmada' AND inicio >= $3 AND inicio < $4", reserva.EscolarID, reserva.Curso, inicioSemana, finSemana).Scan(&delCurso); err != nil {
		return nil, false, err
	}
	if delCurso >= limites.MaxPorCursoSemana {
		return nil, false, ErrCuotaCursoSemana
	}

	var id string
	err = tx.QueryRowContext(ctx, "INSERT INTO reservas (escolar_id, tutor_id, tutor_nombre, slot_id, curso, grado, inicio, fin, tema, objetivo, estado, clave_idempotencia, creada_en) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id",
		reserva.EscolarID, reserva.TutorID, reserva.TutorNombre, reserva.SlotID, reserva.Curso, reserva.Grado,
		reserva.Inicio, reserva.Fin, reserva.Tema, reserva.Objetivo, string(reserva.Estado), nullable(reserva.ClaveIdempotencia), reserva.CreadaEn,
	).Scan(&id)
	if err != nil {
		if esCodigo(err, "23505") && nombreRestriccion(err) == "reservas_idempotencia_uq" && reserva.ClaveIdempotencia != "" {
			existente, errExistente := reservaPorClave(ctx, tx, reserva.EscolarID, reserva.ClaveIdempotencia)
			if errExistente != nil {
				return nil, false, errExistente
			}
			return existente, false, nil
		}
		return nil, false, traducirError(err)
	}
	// I10: persistir entregas en la misma transacción; I09 conserva sus validaciones.
	if err := encolarNotificaciones(ctx, tx, id, reserva, r.AnticipacionRecordatorio); err != nil {
		return nil, false, err
	}
	if err := tx.Commit(); err != nil {
		return nil, false, err
	}
	guardada := *reserva
	guardada.ID = id
	return &guardada, true, nil
}

func (r RepositorioPostgres) PorClaveIdempotencia(ctx context.Context, escolarID, clave string) (*Reserva, error) {
	return reservaPorClave(ctx, r.DB, escolarID, clave)
}

type consultaReserva interface {
	QueryRowContext(ctx context.Context, consulta string, argumentos ...any) *sql.Row
}

func reservaPorClave(ctx context.Context, base consultaReserva, escolarID, clave string) (*Reserva, error) {
	reserva, err := escanearReserva(base.QueryRowContext(ctx, "SELECT "+columnasReserva+" FROM reservas WHERE escolar_id = $1 AND clave_idempotencia = $2", escolarID, clave))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrReservaNoEncontrada
	}
	if err != nil {
		return nil, err
	}
	return reserva, nil
}

func (r RepositorioPostgres) PorID(ctx context.Context, id string) (*Reserva, error) {
	reserva, err := escanearReserva(r.DB.QueryRowContext(ctx, "SELECT "+columnasReserva+" FROM reservas WHERE id = $1", id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrReservaNoEncontrada
	}
	if err != nil {
		return nil, err
	}
	adjuntos, err := r.AdjuntosDe(ctx, reserva.ID)
	if err != nil {
		return nil, err
	}
	reserva.Adjuntos = adjuntos
	return reserva, nil
}

func (r RepositorioPostgres) PorEscolar(ctx context.Context, escolarID string) ([]Reserva, error) {
	filas, err := r.DB.QueryContext(ctx, "SELECT "+columnasReserva+" FROM reservas WHERE escolar_id = $1 ORDER BY inicio", escolarID)
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
		reserva.Adjuntos, err = r.AdjuntosDe(ctx, reserva.ID)
		if err != nil {
			return nil, err
		}
		reservas = append(reservas, *reserva)
	}
	return reservas, filas.Err()
}

func (r RepositorioPostgres) PorTutor(ctx context.Context, tutorID string) ([]Reserva, error) {
	filas, err := r.DB.QueryContext(
		ctx,
		"SELECT "+columnasReserva+" FROM reservas WHERE tutor_id = $1 ORDER BY inicio, id",
		tutorID,
	)
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	reservas := []Reserva{}

	for filas.Next() {
		reserva, err := escanearReserva(filas)
		if err != nil {
			return nil, err
		}

		reservas = append(reservas, *reserva)
	}

	return reservas, filas.Err()
}

func (r RepositorioPostgres) SlotTomado(ctx context.Context, slotID string) (bool, error) {
	var tomado bool
	err := r.DB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM reservas WHERE slot_id = $1)", slotID).Scan(&tomado)
	return tomado, err
}

func (r RepositorioPostgres) ContarAdjuntos(ctx context.Context, reservaID string) (int, error) {
	var total int
	err := r.DB.QueryRowContext(ctx, "SELECT count(*) FROM reserva_adjuntos WHERE reserva_id = $1", reservaID).Scan(&total)
	return total, err
}

func (r RepositorioPostgres) AgregarAdjunto(ctx context.Context, reservaID string, adjunto *Adjunto) (*Adjunto, error) {
	guardado := *adjunto
	err := r.DB.QueryRowContext(ctx, "INSERT INTO reserva_adjuntos (reserva_id, nombre, tipo, tamano_bytes, estado, motivo, huella, clave_almacenamiento, creado_en) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id",
		reservaID, adjunto.Nombre, adjunto.Tipo, adjunto.TamanoBytes, string(adjunto.Estado), adjunto.Motivo,
		adjunto.Huella, adjunto.ClaveAlmacenamiento, adjunto.CreadoEn,
	).Scan(&guardado.ID)
	if err != nil {
		return nil, traducirError(err)
	}
	return &guardado, nil
}

func (r RepositorioPostgres) AdjuntosDe(ctx context.Context, reservaID string) ([]Adjunto, error) {
	filas, err := r.DB.QueryContext(ctx, "SELECT "+columnasAdjunto+" FROM reserva_adjuntos WHERE reserva_id = $1 ORDER BY creado_en", reservaID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = filas.Close() }()

	adjuntos := []Adjunto{}
	for filas.Next() {
		adjunto, err := escanearAdjunto(filas)
		if err != nil {
			return nil, err
		}
		adjuntos = append(adjuntos, *adjunto)
	}
	return adjuntos, filas.Err()
}

func (r RepositorioPostgres) AdjuntoPorID(ctx context.Context, reservaID, adjuntoID string) (*Adjunto, error) {
	adjunto, err := escanearAdjunto(r.DB.QueryRowContext(ctx, "SELECT "+columnasAdjunto+" FROM reserva_adjuntos WHERE id = $1 AND reserva_id = $2", adjuntoID, reservaID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAdjuntoNoEncontrado
	}
	return adjunto, err
}

func (r RepositorioPostgres) ActualizarAdjunto(ctx context.Context, adjuntoID string, estado EstadoAdjunto, motivo string) error {
	resultado, err := r.DB.ExecContext(ctx, "UPDATE reserva_adjuntos SET estado = $2, motivo = $3, actualizado_en = now() WHERE id = $1", adjuntoID, string(estado), motivo)
	if err != nil {
		return err
	}
	if filas, _ := resultado.RowsAffected(); filas == 0 {
		return ErrAdjuntoNoEncontrado
	}
	return nil
}

type filaEscaner interface{ Scan(destino ...any) error }

func escanearReserva(fila filaEscaner) (*Reserva, error) {
	var reserva Reserva
	var estado string
	var clave sql.NullString
	if err := fila.Scan(
		&reserva.ID, &reserva.EscolarID, &reserva.TutorID, &reserva.TutorNombre, &reserva.SlotID,
		&reserva.Curso, &reserva.Grado, &reserva.Inicio, &reserva.Fin, &reserva.Tema, &reserva.Objetivo,
		&estado, &clave, &reserva.CreadaEn,
	); err != nil {
		return nil, err
	}
	reserva.Estado = Estado(estado)
	reserva.ClaveIdempotencia = clave.String
	return &reserva, nil
}

func escanearAdjunto(fila filaEscaner) (*Adjunto, error) {
	var adjunto Adjunto
	var estado string
	var motivo sql.NullString
	var huella sql.NullString
	var clave sql.NullString
	var creadoEn time.Time
	if err := fila.Scan(
		&adjunto.ID, &adjunto.Nombre, &adjunto.Tipo, &adjunto.TamanoBytes, &estado, &motivo,
		&huella, &clave, &creadoEn,
	); err != nil {
		return nil, err
	}
	adjunto.Estado = EstadoAdjunto(estado)
	adjunto.Motivo = motivo.String
	adjunto.Huella = huella.String
	adjunto.ClaveAlmacenamiento = clave.String
	adjunto.CreadoEn = creadoEn
	return &adjunto, nil
}

func nullable(texto string) any {
	if texto == "" {
		return nil
	}
	return texto
}

func esCodigo(err error, codigo string) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && string(pqErr.Code) == codigo
}

func nombreRestriccion(err error) string {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Constraint
	}
	return ""
}

// traducirError convierte las violaciones de integridad en errores del dominio, para que
// el controlador responda 409 o 403 sin filtrar detalles de Postgres.
func traducirError(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			if pqErr.Constraint == "reservas_idempotencia_uq" {
				return ErrIdempotenciaConflicto
			}
			return ErrSlotOcupado
		case "23P01":
			return ErrSolapamiento
		case "23503":
			return ErrNoAutorizado
		case "23514":
			return ErrReservaInvalida
		}
	}
	return err
}
