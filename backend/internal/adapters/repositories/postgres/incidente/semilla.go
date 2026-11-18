package incidente

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const (
	slotTutoriaSintetica      = "slot-i13-demo"
	sesionTutoriaSintetica    = "sesion-i13-demo"
	emailEscolarSintetico     = "escolar.prueba@example.test"
	slotTutoriaSinteticaDos   = "slot-i13-demo-2"
	sesionTutoriaSinteticaDos = "sesion-i13-demo-2"
	emailEscolarSinteticoDos  = "escolar2.prueba@example.test"
)

type tutoriaSintetica struct {
	slot         string
	sesion       string
	emailEscolar string
	emailTutor   string
	curso        string
	grado        string
	tema         string
	objetivo     string
	diasAtras    int
	horaLima     int
}

var tutoriasSinteticas = []tutoriaSintetica{
	{
		slot: slotTutoriaSintetica, sesion: sesionTutoriaSintetica,
		emailEscolar: emailEscolarSintetico, emailTutor: "tutor.prueba",
		curso: "Álgebra", grado: "5to de secundaria",
		tema:      "Repaso de factorización",
		objetivo:  "Resolver ejercicios de factorización con acompañamiento.",
		diasAtras: 3, horaLima: 15,
	},
	{
		slot: slotTutoriaSinteticaDos, sesion: sesionTutoriaSinteticaDos,
		emailEscolar: emailEscolarSinteticoDos, emailTutor: "tutor2.prueba",
		curso: "Biología", grado: "4to de secundaria",
		tema:      "La célula y sus funciones",
		objetivo:  "Reforzar el reconocimiento de estructuras celulares.",
		diasAtras: 2, horaLima: 10,
	},
}

func SembrarTutoriasSinteticas(ctx context.Context, baseDatos *sql.DB, dominioTutor string) error {
	if dominioTutor == "" {
		return fmt.Errorf("dominio del tutor sintético vacío")
	}
	for _, tutoria := range tutoriasSinteticas {
		if err := sembrarTutoriaSintetica(ctx, baseDatos, dominioTutor, tutoria); err != nil {
			return fmt.Errorf("sembrar tutoría %s: %w", tutoria.slot, err)
		}
	}
	return nil
}

func sembrarTutoriaSintetica(
	ctx context.Context,
	baseDatos *sql.DB,
	dominioTutor string,
	tutoria tutoriaSintetica,
) error {
	var escolarID string
	if err := baseDatos.QueryRowContext(ctx, `
		SELECT id::text FROM usuarios WHERE email = $1`, tutoria.emailEscolar,
	).Scan(&escolarID); err != nil {
		return fmt.Errorf("buscar escolar sintético: %w", err)
	}

	var tutorID, tutorNombre string
	if err := baseDatos.QueryRowContext(ctx, `
		SELECT id::text, nombre_uso FROM usuarios WHERE email = $1`,
		tutoria.emailTutor+"@"+dominioTutor,
	).Scan(&tutorID, &tutorNombre); err != nil {
		return fmt.Errorf("buscar tutor sintético: %w", err)
	}

	inicio := inicioTutoriaSintetica(time.Now().UTC(), tutoria.diasAtras, tutoria.horaLima)
	fin := inicio.Add(60 * time.Minute)

	var reservaID string
	err := baseDatos.QueryRowContext(ctx, `
		SELECT id::text FROM reservas WHERE slot_id = $1`,
		tutoria.slot,
	).Scan(&reservaID)
	if errors.Is(err, sql.ErrNoRows) {
		err = baseDatos.QueryRowContext(ctx, `
			INSERT INTO reservas (
				escolar_id, tutor_id, tutor_nombre, slot_id, curso, grado,
				inicio, fin, tema, objetivo
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
			) RETURNING id::text`,
			escolarID, tutorID, tutorNombre, tutoria.slot, tutoria.curso,
			tutoria.grado, inicio, fin, tutoria.tema, tutoria.objetivo,
		).Scan(&reservaID)
	}
	if err != nil {
		return fmt.Errorf("crear reserva sintética I13: %w", err)
	}

	if _, err := baseDatos.ExecContext(ctx, `
		INSERT INTO sesion_sesiones (id, reserva_id, tutor_id, escolar_id, inicio, fin)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT DO NOTHING`,
		tutoria.sesion, reservaID, tutorID, escolarID, inicio, fin,
	); err != nil {
		return fmt.Errorf("crear sesión sintética I13: %w", err)
	}

	if _, err := baseDatos.ExecContext(ctx, `
		INSERT INTO sesion_asistencias (sesion_id, usuario_id, rol, declarada_en)
		VALUES ($1, $2, 'escolar', now())
		ON CONFLICT DO NOTHING`,
		tutoria.sesion, escolarID,
	); err != nil {
		return fmt.Errorf("declarar asistencia sintética I13: %w", err)
	}
	return nil
}

func inicioTutoriaSintetica(ahora time.Time, diasAtras, horaLima int) time.Time {
	lima := time.FixedZone("America/Lima", -5*60*60)
	local := ahora.In(lima).AddDate(0, 0, -diasAtras)
	return time.Date(local.Year(), local.Month(), local.Day(), horaLima, 0, 0, 0, lima).UTC()
}
