package sesion

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type DependenciasSiembraSesion struct {
	BaseDatos    *sql.DB
	Repositorio  Repositorio
	DominioTutor string
}

func SembrarSintetico(ctx context.Context, dependencias DependenciasSiembraSesion) error {
	tutorID, err := buscarUsuarioID(ctx, dependencias.BaseDatos, "tutor.prueba@"+dependencias.DominioTutor)
	if err != nil {
		return err
	}
	escolarID, err := buscarUsuarioID(ctx, dependencias.BaseDatos, "escolar.prueba@example.test")
	if err != nil {
		return err
	}

	inicio := inicioSesionDemo()
	sesion, err := NuevaSesion(
		"sesion-i05-demo",
		"reserva-i05-demo",
		tutorID,
		escolarID,
		inicio,
		inicio.Add(DuracionSesion),
	)
	if err != nil {
		return fmt.Errorf("crear sesion sintetica: %w", err)
	}

	if _, err := dependencias.Repositorio.SesionPorID(ctx, "sesion-i05-demo"); err == nil {
		return nil
	} else if !errors.Is(err, ErrSesionNoEncontrada) {
		return err
	}

	err = dependencias.Repositorio.CrearSesion(ctx, sesion)
	if errors.Is(err, ErrSesionDuplicada) {
		return nil
	}
	return err
}

func buscarUsuarioID(ctx context.Context, baseDatos *sql.DB, email string) (string, error) {
	var usuarioID string
	err := baseDatos.QueryRowContext(
		ctx,
		`SELECT id FROM usuarios WHERE email = $1`,
		email,
	).Scan(&usuarioID)
	if err != nil {
		return "", fmt.Errorf("buscar usuario sintetico %s: %w", email, err)
	}
	return usuarioID, nil
}

func inicioSesionDemo() time.Time {
	return time.Now().UTC().
		Truncate(24 * time.Hour).
		Add(24*time.Hour + 15*time.Hour)
}
