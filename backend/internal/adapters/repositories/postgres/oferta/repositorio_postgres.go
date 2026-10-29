package oferta

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

func (r RepositorioPostgres) ObtenerDisponibilidad(ctx context.Context, tutorID string) (CupoSemanal, []SlotSemanal, error) {
	cupo := CupoSemanal{TutorID: tutorID, Minutos: CupoSemanalMinimo}
	err := r.DB.QueryRowContext(ctx,
		`SELECT minutos FROM cupos_semanales WHERE tutor_id = $1`, tutorID,
	).Scan(&cupo.Minutos)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return CupoSemanal{}, nil, err
	}

	filas, err := r.DB.QueryContext(ctx, `
		SELECT id, dia_semana, inicio_minuto, fin_minuto
		FROM slots
		WHERE tutor_id = $1
		ORDER BY dia_semana, inicio_minuto`, tutorID)
	if err != nil {
		return CupoSemanal{}, nil, err
	}
	defer filas.Close()

	slots := make([]SlotSemanal, 0)
	for filas.Next() {
		var slot SlotSemanal
		var dia, inicio, fin int
		if err := filas.Scan(&slot.ID, &dia, &inicio, &fin); err != nil {
			return CupoSemanal{}, nil, err
		}
		slot.TutorID = tutorID
		slot.DiaSemana = time.Weekday(dia)
		slot.Inicio = HorarioDesdeMinutos(inicio)
		slot.Fin = HorarioDesdeMinutos(fin)
		slots = append(slots, slot)
	}
	if err := filas.Err(); err != nil {
		return CupoSemanal{}, nil, err
	}
	return cupo, slots, nil
}

func (r RepositorioPostgres) AgregarSlots(ctx context.Context, tutorID string, slots []SlotSemanal) (CupoSemanal, []SlotSemanal, error) {
	transaccion, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return CupoSemanal{}, nil, err
	}
	defer transaccion.Rollback()

	if _, err := transaccion.ExecContext(ctx, `
		INSERT INTO cupos_semanales (tutor_id, minutos)
		VALUES ($1, $2)
		ON CONFLICT (tutor_id) DO NOTHING`, tutorID, CupoSemanalMinimo); err != nil {
		return CupoSemanal{}, nil, err
	}

	var cupo int
	if err := transaccion.QueryRowContext(ctx,
		`SELECT minutos FROM cupos_semanales WHERE tutor_id = $1 FOR UPDATE`, tutorID,
	).Scan(&cupo); err != nil {
		return CupoSemanal{}, nil, err
	}

	filas, err := transaccion.QueryContext(ctx, `
		SELECT dia_semana, inicio_minuto, fin_minuto
		FROM slots
		WHERE tutor_id = $1`, tutorID)
	if err != nil {
		return CupoSemanal{}, nil, err
	}
	existentes := make([]SlotSemanal, 0)
	for filas.Next() {
		var dia, inicio, fin int
		if err := filas.Scan(&dia, &inicio, &fin); err != nil {
			filas.Close()
			return CupoSemanal{}, nil, err
		}
		existentes = append(existentes, SlotSemanal{
			TutorID: tutorID, DiaSemana: time.Weekday(dia),
			Inicio: HorarioDesdeMinutos(inicio), Fin: HorarioDesdeMinutos(fin),
		})
	}
	if err := filas.Err(); err != nil {
		filas.Close()
		return CupoSemanal{}, nil, err
	}
	if err := filas.Close(); err != nil {
		return CupoSemanal{}, nil, err
	}

	for i := range slots {
		slots[i].TutorID = tutorID
	}
	todos := append(existentes, slots...)
	if err := ValidarSlots(todos); err != nil {
		return CupoSemanal{}, nil, err
	}
	if len(todos)*MinutosEnUnaHora > cupo {
		return CupoSemanal{}, nil, ErrCapacidadSemanal
	}

	creados := make([]SlotSemanal, 0, len(slots))
	for _, slot := range slots {
		var id int64
		err := transaccion.QueryRowContext(ctx, `
			INSERT INTO slots (tutor_id, dia_semana, inicio_minuto, fin_minuto)
			VALUES ($1, $2, $3, $4)
			RETURNING id`, tutorID, int(slot.DiaSemana),
			int(slot.Inicio/time.Minute), int(slot.Fin/time.Minute),
		).Scan(&id)
		if err != nil {
			var errorPostgres *pq.Error
			if errors.As(err, &errorPostgres) && errorPostgres.Code == "23P01" {
				return CupoSemanal{}, nil, ErrSolapamiento
			}
			return CupoSemanal{}, nil, err
		}
		slot.ID = id
		creados = append(creados, slot)
	}
	if err := transaccion.Commit(); err != nil {
		return CupoSemanal{}, nil, fmt.Errorf("guardar disponibilidad: %w", err)
	}
	return CupoSemanal{TutorID: tutorID, Minutos: cupo}, creados, nil
}

func (r RepositorioPostgres) EliminarSlot(ctx context.Context, tutorID string, slotID int64) error {
	resultado, err := r.DB.ExecContext(ctx,
		`DELETE FROM slots WHERE id = $1 AND tutor_id = $2`, slotID, tutorID)
	if err != nil {
		return err
	}
	afectados, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if afectados == 0 {
		return ErrSlotNoEncontrado
	}
	return nil
}

func (r RepositorioPostgres) CambiarCupo(ctx context.Context, tutorID string, minutos int) error {
	nuevo := CupoSemanal{TutorID: tutorID, Minutos: minutos}
	if err := nuevo.Validar(); err != nil {
		return err
	}

	transaccion, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer transaccion.Rollback()

	if _, err := transaccion.ExecContext(ctx, `
		INSERT INTO cupos_semanales (tutor_id, minutos)
		VALUES ($1, $2)
		ON CONFLICT (tutor_id) DO NOTHING`, tutorID, CupoSemanalMinimo); err != nil {
		return err
	}

	var actual int
	if err := transaccion.QueryRowContext(ctx,
		`SELECT minutos FROM cupos_semanales WHERE tutor_id = $1 FOR UPDATE`, tutorID,
	).Scan(&actual); err != nil {
		return err
	}
	if minutos < actual {
		return ErrReduccionNoVerificable
	}

	var cantidadSlots int
	if err := transaccion.QueryRowContext(ctx,
		`SELECT count(*) FROM slots WHERE tutor_id = $1`, tutorID,
	).Scan(&cantidadSlots); err != nil {
		return err
	}
	if cantidadSlots*MinutosEnUnaHora > minutos {
		return ErrCapacidadSemanal
	}
	_, err = transaccion.ExecContext(ctx, `
		UPDATE cupos_semanales
		SET minutos = $2, actualizado_en = now()
		WHERE tutor_id = $1`, tutorID, minutos)
	if err != nil {
		return err
	}
	return transaccion.Commit()
}
