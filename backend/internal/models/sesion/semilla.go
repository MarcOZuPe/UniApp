package sesion

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// SembrarSintetico crea una sesión demostrativa de I05.
// Solo debe invocarse cuando SEED_SYNTHETIC esté activado.
func SembrarSintetico(
	ctx context.Context,
	db *sql.DB,
	repo Repositorio,
	dominioTutor string,
) error {
	obtenerID := func(email string) (string, error) {
		var id string

		err := db.QueryRowContext(
			ctx,
			`SELECT id FROM usuarios WHERE email = $1`,
			email,
		).Scan(&id)

		if err != nil {
			return "", fmt.Errorf(
				"buscar usuario sintetico %s: %w",
				email,
				err,
			)
		}

		return id, nil
	}

	tutorID, err := obtenerID("tutor.prueba@" + dominioTutor)
	if err != nil {
		return err
	}

	escolarID, err := obtenerID("escolar.prueba@example.test")
	if err != nil {
		return err
	}

	// La sesión demostrativa dura exactamente 60 minutos.
	inicio := time.Now().UTC().
		Truncate(24 * time.Hour).
		Add(24*time.Hour + 15*time.Hour)

	fin := inicio.Add(DuracionSesion)

	sesion, err := NuevaSesion(
		"sesion-i05-demo",
		"reserva-i05-demo",
		tutorID,
		escolarID,
		inicio,
		fin,
	)
	if err != nil {
		return fmt.Errorf("crear sesion sintetica: %w", err)
	}

	err = repo.CrearSesion(ctx, sesion)

	// Permite reiniciar el servidor sin insertar duplicados.
	if errors.Is(err, ErrSesionDuplicada) {
		return nil
	}

	return err
}
