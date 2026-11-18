package incidente

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const (
	slotTutoriaSintetica   = "slot-i13-demo"
	sesionTutoriaSintetica = "sesion-i13-demo"
	emailEscolarSintetico  = "escolar.prueba@example.test"
)

// SembrarTutoriaSintetica deja una reserva ya finalizada entre el escolar y el tutor
// sintéticos, con sesión y asistencia declarada, para poder probar I13 desde la
// interfaz sin crear antes una reserva manualmente. Es idempotente por slot y sesión.
func SembrarTutoriaSintetica(ctx context.Context, baseDatos *sql.DB, dominioTutor string) error {
	if dominioTutor == "" {
		return fmt.Errorf("dominio del tutor sintético vacío")
	}

	var escolarID string
	if err := baseDatos.QueryRowContext(ctx, `
		SELECT id::text FROM usuarios WHERE email = $1`, emailEscolarSintetico,
	).Scan(&escolarID); err != nil {
		return fmt.Errorf("buscar escolar sintético: %w", err)
	}

	var tutorID, tutorNombre string
	if err := baseDatos.QueryRowContext(ctx, `
		SELECT id::text, nombre_uso FROM usuarios WHERE email = $1`,
		"tutor.prueba@"+dominioTutor,
	).Scan(&tutorID, &tutorNombre); err != nil {
		return fmt.Errorf("buscar tutor sintético: %w", err)
	}

	inicio := inicioTutoriaSintetica(time.Now().UTC())
	fin := inicio.Add(60 * time.Minute)

	var reservaID string
	err := baseDatos.QueryRowContext(ctx, `
		SELECT id::text FROM reservas WHERE slot_id = $1`,
		slotTutoriaSintetica,
	).Scan(&reservaID)
	if errors.Is(err, sql.ErrNoRows) {
		err = baseDatos.QueryRowContext(ctx, `
			INSERT INTO reservas (
				escolar_id, tutor_id, tutor_nombre, slot_id, curso, grado,
				inicio, fin, tema, objetivo
			) VALUES (
				$1, $2, $3, $4, 'Álgebra', '5to de secundaria', $5, $6,
				'Repaso de factorización',
				'Resolver ejercicios de factorización con acompañamiento.'
			) RETURNING id::text`,
			escolarID, tutorID, tutorNombre, slotTutoriaSintetica, inicio, fin,
		).Scan(&reservaID)
	}
	if err != nil {
		return fmt.Errorf("crear reserva sintética I13: %w", err)
	}

	if _, err := baseDatos.ExecContext(ctx, `
		INSERT INTO sesion_sesiones (id, reserva_id, tutor_id, escolar_id, inicio, fin)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT DO NOTHING`,
		sesionTutoriaSintetica, reservaID, tutorID, escolarID, inicio, fin,
	); err != nil {
		return fmt.Errorf("crear sesión sintética I13: %w", err)
	}

	if _, err := baseDatos.ExecContext(ctx, `
		INSERT INTO sesion_asistencias (sesion_id, usuario_id, rol, declarada_en)
		VALUES ($1, $2, 'escolar', now())
		ON CONFLICT DO NOTHING`,
		sesionTutoriaSintetica, escolarID,
	); err != nil {
		return fmt.Errorf("declarar asistencia sintética I13: %w", err)
	}
	return nil
}

func inicioTutoriaSintetica(ahora time.Time) time.Time {
	lima := time.FixedZone("America/Lima", -5*60*60)
	local := ahora.In(lima).AddDate(0, 0, -3)
	return time.Date(local.Year(), local.Month(), local.Day(), 15, 0, 0, 0, lima).UTC()
}
